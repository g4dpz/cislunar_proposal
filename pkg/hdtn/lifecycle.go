package hdtn

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// LifecycleConfig holds configuration for the HDTN lifecycle manager.
type LifecycleConfig struct {
	BinaryPath      string        // Path to hdtn-one-process binary
	ConfigPath      string        // Path to hdtn-config.json
	RESTPort        int           // HDTN REST API port (default 10305)
	StartTimeout    time.Duration // Max wait for health check (default 30s)
	StopTimeout     time.Duration // Max wait for graceful stop (default 10s)
	MonitorInterval time.Duration // Process monitor poll interval (default 1s)
}

// ProcessState represents the current state of the HDTN process.
type ProcessState int

const (
	StateStopped  ProcessState = iota
	StateStarting
	StateRunning
	StateStopping
	StateFailed
)

// String returns a human-readable representation of the process state.
func (s ProcessState) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ExitEvent is sent when the process exits unexpectedly.
type ExitEvent struct {
	ExitCode int
	Error    error
	Time     time.Time
}

// LifecycleManager manages the HDTN process lifecycle.
type LifecycleManager struct {
	config LifecycleConfig

	mu       sync.Mutex
	state    ProcessState
	cmd      *exec.Cmd
	exitCh   chan struct{} // closed when process exits
	onExitFn func(ExitEvent)

	// For testing: allow overriding the health check function
	healthCheckFn func() error
}

// NewLifecycleManager creates a new lifecycle manager with the given config.
// Returns an error if the config is invalid.
func NewLifecycleManager(config LifecycleConfig) (*LifecycleManager, error) {
	if config.BinaryPath == "" {
		return nil, fmt.Errorf("binary path cannot be empty")
	}
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	// Apply defaults
	if config.RESTPort == 0 {
		config.RESTPort = 10305
	}
	if config.StartTimeout == 0 {
		config.StartTimeout = 30 * time.Second
	}
	if config.StopTimeout == 0 {
		config.StopTimeout = 10 * time.Second
	}
	if config.MonitorInterval == 0 {
		config.MonitorInterval = 1 * time.Second
	}

	lm := &LifecycleManager{
		config: config,
		state:  StateStopped,
	}
	lm.healthCheckFn = lm.defaultHealthCheck

	return lm, nil
}

// Start spawns the HDTN process and waits for readiness via REST API health check.
func (lm *LifecycleManager) Start() error {
	lm.mu.Lock()
	if lm.state == StateRunning || lm.state == StateStarting {
		lm.mu.Unlock()
		return fmt.Errorf("hdtn process is already running")
	}
	lm.state = StateStarting
	lm.mu.Unlock()

	// Verify binary exists
	if _, err := os.Stat(lm.config.BinaryPath); err != nil {
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return fmt.Errorf("failed to spawn hdtn process: %w", err)
	}

	// Spawn the HDTN process
	cmd := exec.Command(lm.config.BinaryPath, "--contact-plan-file="+lm.config.ConfigPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return fmt.Errorf("failed to spawn hdtn process: %w", err)
	}

	lm.mu.Lock()
	lm.cmd = cmd
	lm.exitCh = make(chan struct{})
	lm.mu.Unlock()

	// Start process monitor goroutine
	go lm.monitorProcess()

	// Poll REST API health check at 500ms intervals
	deadline := time.Now().Add(lm.config.StartTimeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := lm.healthCheckFn(); err == nil {
				lm.mu.Lock()
				lm.state = StateRunning
				lm.mu.Unlock()
				return nil
			}
			if time.Now().After(deadline) {
				// Timeout: kill the process and return error
				_ = lm.forceStop()
				lm.mu.Lock()
				lm.state = StateStopped
				lm.mu.Unlock()
				return fmt.Errorf("startup timeout: HDTN REST API did not respond within %s", lm.config.StartTimeout)
			}
		case <-lm.exitCh:
			// Process exited before becoming ready
			lm.mu.Lock()
			lm.state = StateStopped
			lm.mu.Unlock()
			return fmt.Errorf("hdtn process exited during startup")
		}
	}
}

// Stop sends SIGTERM to the HDTN process and waits for graceful termination.
// If the process doesn't terminate within StopTimeout, SIGKILL is sent.
func (lm *LifecycleManager) Stop() error {
	lm.mu.Lock()
	if lm.state != StateRunning && lm.state != StateStarting && lm.state != StateFailed {
		lm.mu.Unlock()
		return nil // Already stopped
	}
	lm.state = StateStopping
	cmd := lm.cmd
	exitCh := lm.exitCh
	lm.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return nil
	}

	// Send SIGTERM
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process may have already exited
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return nil
	}

	// Wait for process to exit or timeout
	select {
	case <-exitCh:
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return nil
	case <-time.After(lm.config.StopTimeout):
		// Force kill
		_ = cmd.Process.Signal(syscall.SIGKILL)
		<-exitCh
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return nil
	}
}

// Restart stops the running HDTN process and starts a new instance.
func (lm *LifecycleManager) Restart() error {
	if err := lm.Stop(); err != nil {
		return fmt.Errorf("failed to stop during restart: %w", err)
	}
	return lm.Start()
}

// IsRunning returns true if the process is in the Running state.
func (lm *LifecycleManager) IsRunning() bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.state == StateRunning
}

// State returns the current process state.
func (lm *LifecycleManager) State() ProcessState {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.state
}

// OnExit registers a callback for unexpected process exits.
// The callback is invoked when the process exits while in the Running state.
func (lm *LifecycleManager) OnExit(fn func(ExitEvent)) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.onExitFn = fn
}

// WaitForReady blocks until the process reaches the Running state or the timeout expires.
func (lm *LifecycleManager) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if lm.IsRunning() {
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for HDTN to be ready")
			}
		}
	}
}

// monitorProcess watches the HDTN process and detects unexpected exits.
func (lm *LifecycleManager) monitorProcess() {
	cmd := lm.cmd
	if cmd == nil {
		return
	}

	// Wait for the process to exit
	err := cmd.Wait()

	// Close the exit channel to signal that the process has exited
	lm.mu.Lock()
	exitCh := lm.exitCh
	lm.mu.Unlock()

	if exitCh != nil {
		close(exitCh)
	}

	// Determine if this was an unexpected exit
	lm.mu.Lock()
	wasRunning := lm.state == StateRunning
	if wasRunning {
		lm.state = StateFailed
	}
	onExitFn := lm.onExitFn
	lm.mu.Unlock()

	// If the process was running (not being intentionally stopped), fire the callback
	if wasRunning && onExitFn != nil {
		exitCode := 0
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		onExitFn(ExitEvent{
			ExitCode: exitCode,
			Error:    err,
			Time:     time.Now(),
		})
	}
}

// defaultHealthCheck performs an HTTP GET to the HDTN REST API health endpoint.
func (lm *LifecycleManager) defaultHealthCheck() error {
	url := fmt.Sprintf("http://localhost:%d/api/v1/status", lm.config.RESTPort)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	return nil
}

// forceStop kills the process immediately without waiting.
func (lm *LifecycleManager) forceStop() error {
	lm.mu.Lock()
	cmd := lm.cmd
	exitCh := lm.exitCh
	lm.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	_ = cmd.Process.Signal(syscall.SIGKILL)
	if exitCh != nil {
		<-exitCh
	}
	return nil
}
