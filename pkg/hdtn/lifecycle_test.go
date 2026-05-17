package hdtn

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// TestHelperProcess is a test helper that acts as a fake HDTN process.
// It is invoked as a subprocess by the tests. The behavior is controlled
// by the TEST_HELPER_MODE environment variable.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	mode := os.Getenv("TEST_HELPER_MODE")
	port := os.Getenv("TEST_HELPER_PORT")

	switch mode {
	case "healthy":
		// Start an HTTP server that responds to health checks immediately
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"running"}`)
		})
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			os.Exit(1)
		}
		server := &http.Server{Handler: mux}
		go server.Serve(listener)
		// Wait for signal
		select {}

	case "slow_start":
		// Wait before starting the health endpoint (simulates slow startup)
		delay := os.Getenv("TEST_HELPER_DELAY")
		d, _ := time.ParseDuration(delay)
		time.Sleep(d)
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"running"}`)
		})
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			os.Exit(1)
		}
		server := &http.Server{Handler: mux}
		go server.Serve(listener)
		select {}

	case "no_health":
		// Never respond to health checks (for timeout testing)
		select {}

	case "hang_on_signal":
		// Ignore SIGTERM (for SIGKILL fallback testing)
		// Trap SIGTERM so the process doesn't exit
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM)
		go func() {
			for range sigCh {
				// Ignore SIGTERM
			}
		}()

		// Start health endpoint so we can get to Running state
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"running"}`)
		})
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			os.Exit(1)
		}
		server := &http.Server{Handler: mux}
		go server.Serve(listener)
		// Block forever
		select {}

	case "exit_quickly":
		// Start health endpoint, then exit after a short delay
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"running"}`)
		})
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			os.Exit(1)
		}
		server := &http.Server{Handler: mux}
		go server.Serve(listener)
		delay := os.Getenv("TEST_HELPER_DELAY")
		d, _ := time.ParseDuration(delay)
		time.Sleep(d)
		os.Exit(42)

	default:
		os.Exit(1)
	}
}

// helperBinary returns the path to the test binary itself, which can be
// re-invoked as a subprocess with GO_TEST_HELPER_PROCESS=1.
func helperBinary(t *testing.T) string {
	t.Helper()
	// The test binary is os.Args[0]
	return os.Args[0]
}

// getFreePort returns an available TCP port.
func getFreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// newTestLifecycleManager creates a LifecycleManager configured to use the test helper process.
// It sets up the health check function to use the given port.
func newTestLifecycleManager(t *testing.T, port int) *LifecycleManager {
	t.Helper()

	binary := helperBinary(t)
	config := LifecycleConfig{
		BinaryPath:      binary,
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     3 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	return lm
}

func TestNewLifecycleManager_Defaults(t *testing.T) {
	config := LifecycleConfig{
		BinaryPath: "/usr/local/bin/hdtn-one-process",
		ConfigPath: "/etc/hdtn/config.json",
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lm.config.RESTPort != 10305 {
		t.Errorf("expected default RESTPort 10305, got %d", lm.config.RESTPort)
	}
	if lm.config.StartTimeout != 30*time.Second {
		t.Errorf("expected default StartTimeout 30s, got %s", lm.config.StartTimeout)
	}
	if lm.config.StopTimeout != 10*time.Second {
		t.Errorf("expected default StopTimeout 10s, got %s", lm.config.StopTimeout)
	}
	if lm.config.MonitorInterval != 1*time.Second {
		t.Errorf("expected default MonitorInterval 1s, got %s", lm.config.MonitorInterval)
	}
	if lm.State() != StateStopped {
		t.Errorf("expected initial state Stopped, got %s", lm.State())
	}
}

func TestNewLifecycleManager_ValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		config LifecycleConfig
		errMsg string
	}{
		{
			name:   "empty binary path",
			config: LifecycleConfig{ConfigPath: "/etc/hdtn/config.json"},
			errMsg: "binary path cannot be empty",
		},
		{
			name:   "empty config path",
			config: LifecycleConfig{BinaryPath: "/usr/local/bin/hdtn-one-process"},
			errMsg: "config path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLifecycleManager(tt.config)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.errMsg {
				t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestLifecycleManager_InvalidBinaryPath(t *testing.T) {
	config := LifecycleConfig{
		BinaryPath:   "/nonexistent/path/hdtn-one-process",
		ConfigPath:   "/tmp/test-config.json",
		StartTimeout: 2 * time.Second,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("unexpected error creating manager: %v", err)
	}

	err = lm.Start()
	if err == nil {
		t.Fatal("expected error for invalid binary path, got nil")
	}
	if lm.State() != StateStopped {
		t.Errorf("expected state Stopped after spawn failure, got %s", lm.State())
	}
}

func TestLifecycleManager_StateTransitions(t *testing.T) {
	port := getFreePort(t)

	// Create a lifecycle manager that uses the test helper binary
	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     3 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	// Override the command creation to use our test helper
	origBinaryPath := lm.config.BinaryPath
	lm.config.BinaryPath = origBinaryPath

	// We need to intercept the Start() to set up the subprocess correctly.
	// Instead, let's create a wrapper that sets up the test helper process.
	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	// Override Start to use our test helper subprocess
	// We'll directly test using startWithCmd
	if lm.State() != StateStopped {
		t.Fatalf("expected initial state Stopped, got %s", lm.State())
	}

	// Start with test helper
	err = lm.startWithHelper(t, port, "healthy", "")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if lm.State() != StateRunning {
		t.Errorf("expected state Running after start, got %s", lm.State())
	}
	if !lm.IsRunning() {
		t.Error("expected IsRunning() to return true")
	}

	// Stop
	err = lm.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if lm.State() != StateStopped {
		t.Errorf("expected state Stopped after stop, got %s", lm.State())
	}
	if lm.IsRunning() {
		t.Error("expected IsRunning() to return false after stop")
	}
}

func TestLifecycleManager_StartTimeout(t *testing.T) {
	port := getFreePort(t)

	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    2 * time.Second,
		StopTimeout:     2 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	// Start with a helper that never responds to health checks
	err = lm.startWithHelper(t, port, "no_health", "")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if lm.State() != StateStopped {
		t.Errorf("expected state Stopped after timeout, got %s", lm.State())
	}
}

func TestLifecycleManager_StopWithSIGKILL(t *testing.T) {
	port := getFreePort(t)

	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     1 * time.Second, // Short timeout to trigger SIGKILL quickly
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	// Start with a helper that ignores SIGTERM
	err = lm.startWithHelper(t, port, "hang_on_signal", "")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if lm.State() != StateRunning {
		t.Fatalf("expected state Running, got %s", lm.State())
	}

	// Stop should eventually SIGKILL
	start := time.Now()
	err = lm.Stop()
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if lm.State() != StateStopped {
		t.Errorf("expected state Stopped after SIGKILL, got %s", lm.State())
	}

	// Should have taken at least StopTimeout (1s) since SIGTERM was ignored
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected stop to take at least ~1s (SIGKILL fallback), took %s", elapsed)
	}
}

func TestLifecycleManager_DoubleStart(t *testing.T) {
	port := getFreePort(t)

	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     3 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	// First start
	err = lm.startWithHelper(t, port, "healthy", "")
	if err != nil {
		t.Fatalf("first Start failed: %v", err)
	}
	defer lm.Stop()

	// Second start should fail
	err = lm.Start()
	if err == nil {
		t.Fatal("expected error on double start, got nil")
	}
	if err.Error() != "hdtn process is already running" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLifecycleManager_UnexpectedExit(t *testing.T) {
	port := getFreePort(t)

	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     3 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	lm.healthCheckFn = func() error {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/status", port)
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	}

	// Track exit callback
	var exitEvent ExitEvent
	var exitCalled atomic.Bool
	var exitMu sync.Mutex

	lm.OnExit(func(ev ExitEvent) {
		exitMu.Lock()
		exitEvent = ev
		exitMu.Unlock()
		exitCalled.Store(true)
	})

	// Start with a helper that exits after 1 second
	err = lm.startWithHelper(t, port, "exit_quickly", "1s")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if lm.State() != StateRunning {
		t.Fatalf("expected state Running, got %s", lm.State())
	}

	// Wait for the process to exit unexpectedly (should happen within 2s)
	deadline := time.After(4 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for unexpected exit detection")
		case <-ticker.C:
			if exitCalled.Load() {
				// Verify the exit event
				exitMu.Lock()
				ev := exitEvent
				exitMu.Unlock()

				if ev.ExitCode != 42 {
					t.Errorf("expected exit code 42, got %d", ev.ExitCode)
				}
				if lm.State() != StateFailed {
					t.Errorf("expected state Failed after unexpected exit, got %s", lm.State())
				}
				return
			}
		}
	}
}

func TestLifecycleManager_ProcessStateString(t *testing.T) {
	tests := []struct {
		state    ProcessState
		expected string
	}{
		{StateStopped, "stopped"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateFailed, "failed"},
		{ProcessState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLifecycleManager_WaitForReady(t *testing.T) {
	port := getFreePort(t)

	config := LifecycleConfig{
		BinaryPath:      helperBinary(t),
		ConfigPath:      "/tmp/test-hdtn-config.json",
		RESTPort:        port,
		StartTimeout:    5 * time.Second,
		StopTimeout:     3 * time.Second,
		MonitorInterval: 100 * time.Millisecond,
	}

	lm, err := NewLifecycleManager(config)
	if err != nil {
		t.Fatalf("failed to create lifecycle manager: %v", err)
	}

	// Test timeout when not running
	err = lm.WaitForReady(500 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// startWithHelper is a test helper that starts the lifecycle manager using
// the test helper subprocess pattern.
func (lm *LifecycleManager) startWithHelper(t *testing.T, port int, mode string, delay string) error {
	t.Helper()

	lm.mu.Lock()
	if lm.state == StateRunning || lm.state == StateStarting {
		lm.mu.Unlock()
		return fmt.Errorf("hdtn process is already running")
	}
	lm.state = StateStarting
	lm.mu.Unlock()

	// Create the subprocess command using the test binary
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_TEST_HELPER_PROCESS=1",
		"TEST_HELPER_MODE="+mode,
		"TEST_HELPER_PORT="+strconv.Itoa(port),
		"TEST_HELPER_DELAY="+delay,
	)

	if err := cmd.Start(); err != nil {
		lm.mu.Lock()
		lm.state = StateStopped
		lm.mu.Unlock()
		return fmt.Errorf("failed to spawn test helper: %w", err)
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
				_ = lm.forceStop()
				lm.mu.Lock()
				lm.state = StateStopped
				lm.mu.Unlock()
				return fmt.Errorf("startup timeout: HDTN REST API did not respond within %s", lm.config.StartTimeout)
			}
		case <-lm.exitCh:
			lm.mu.Lock()
			lm.state = StateStopped
			lm.mu.Unlock()
			return fmt.Errorf("hdtn process exited during startup")
		}
	}
}
