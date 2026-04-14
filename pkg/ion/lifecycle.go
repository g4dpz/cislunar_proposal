package ion

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NodeLifecycle manages ION-DTN node lifecycle operations
type NodeLifecycle struct {
	config     NodeConfig
	ionBinPath string
	ionLibPath string
	running    bool
}

// NodeConfig holds ION-DTN node configuration
type NodeConfig struct {
	NodeID      string
	NodeNumber  int
	ConfigDir   string
	WorkingDir  string
	IONInstall  string // Path to ion-install directory
}

// NewNodeLifecycle creates a new ION-DTN lifecycle manager
func NewNodeLifecycle(config NodeConfig) (*NodeLifecycle, error) {
	if config.NodeID == "" {
		return nil, fmt.Errorf("node ID cannot be empty")
	}
	if config.NodeNumber <= 0 {
		return nil, fmt.Errorf("node number must be positive")
	}
	if config.ConfigDir == "" {
		return nil, fmt.Errorf("config directory cannot be empty")
	}
	if config.IONInstall == "" {
		return nil, fmt.Errorf("ION install path cannot be empty")
	}

	ionBinPath := filepath.Join(config.IONInstall, "bin")
	ionLibPath := filepath.Join(config.IONInstall, "lib")

	// Verify ION binaries exist
	if _, err := os.Stat(filepath.Join(ionBinPath, "ionadmin")); err != nil {
		return nil, fmt.Errorf("ionadmin not found in %s: %w", ionBinPath, err)
	}

	// Set working directory to current directory if not specified
	if config.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		config.WorkingDir = wd
	}

	return &NodeLifecycle{
		config:     config,
		ionBinPath: ionBinPath,
		ionLibPath: ionLibPath,
		running:    false,
	}, nil
}

// Start initializes and starts the ION-DTN node
// Executes ionadmin, ltpadmin, bpadmin, ipnadmin with config files
func (nl *NodeLifecycle) Start() error {
	if nl.running {
		return fmt.Errorf("node is already running")
	}

	// Copy kiss.ionconfig to working directory (ION looks for it in cwd)
	kissConfig := filepath.Join(nl.config.ConfigDir, "kiss.ionconfig")
	if _, err := os.Stat(kissConfig); err == nil {
		destPath := filepath.Join(nl.config.WorkingDir, "kiss.ionconfig")
		if err := copyFile(kissConfig, destPath); err != nil {
			return fmt.Errorf("failed to copy kiss.ionconfig: %w", err)
		}
	}

	// Initialize ION
	if err := nl.runAdmin("ionadmin", "node.ionrc"); err != nil {
		return fmt.Errorf("ionadmin failed: %w", err)
	}

	// Initialize LTP (with KISS CLA)
	if err := nl.runAdmin("ltpadmin", "node.ltprc"); err != nil {
		nl.Stop() // Clean up
		return fmt.Errorf("ltpadmin failed: %w", err)
	}

	// Initialize BP
	if err := nl.runAdmin("bpadmin", "node.bprc"); err != nil {
		nl.Stop() // Clean up
		return fmt.Errorf("bpadmin failed: %w", err)
	}

	// Initialize IPN routing
	if err := nl.runAdmin("ipnadmin", "node.ipnrc"); err != nil {
		nl.Stop() // Clean up
		return fmt.Errorf("ipnadmin failed: %w", err)
	}

	// Initialize BPSec if config exists
	bpsecConfig := filepath.Join(nl.config.ConfigDir, "node.bpsecrc")
	if _, err := os.Stat(bpsecConfig); err == nil {
		if err := nl.runAdmin("bpsecadmin", "node.bpsecrc"); err != nil {
			nl.Stop() // Clean up
			return fmt.Errorf("bpsecadmin failed: %w", err)
		}
	}

	nl.running = true
	return nil
}

// Stop gracefully shuts down the ION-DTN node
func (nl *NodeLifecycle) Stop() error {
	if !nl.running {
		return nil // Already stopped
	}

	cmd := exec.Command(filepath.Join(nl.ionBinPath, "ionstop"))
	cmd.Dir = nl.config.WorkingDir
	cmd.Env = nl.buildEnv()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ionstop failed: %w", err)
	}

	nl.running = false

	// Give ION time to clean up
	time.Sleep(1 * time.Second)

	return nil
}

// IsRunning checks if ION-DTN processes are alive
func (nl *NodeLifecycle) IsRunning() bool {
	// Check if key ION processes are running
	processes := []string{"ionadmin", "ltpclock", "bpclock"}
	
	for _, proc := range processes {
		cmd := exec.Command("pgrep", "-f", proc)
		if err := cmd.Run(); err != nil {
			return false
		}
	}

	return true
}

// Restart stops and starts the node
func (nl *NodeLifecycle) Restart() error {
	if err := nl.Stop(); err != nil {
		return fmt.Errorf("failed to stop node: %w", err)
	}

	// Wait for processes to fully terminate
	time.Sleep(2 * time.Second)

	if err := nl.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	return nil
}

// runAdmin executes an ION admin command with a config file
func (nl *NodeLifecycle) runAdmin(adminCmd, configFile string) error {
	configPath := filepath.Join(nl.config.ConfigDir, configFile)
	
	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	cmd := exec.Command(filepath.Join(nl.ionBinPath, adminCmd), configPath)
	cmd.Dir = nl.config.WorkingDir
	cmd.Env = nl.buildEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", adminCmd, err)
	}

	return nil
}

// buildEnv constructs the environment variables for ION commands
func (nl *NodeLifecycle) buildEnv() []string {
	env := os.Environ()
	
	// Add ION bin to PATH
	pathEnv := fmt.Sprintf("PATH=%s:%s", nl.ionBinPath, os.Getenv("PATH"))
	env = append(env, pathEnv)

	// Add ION lib to library path (macOS and Linux)
	dyldPath := fmt.Sprintf("DYLD_LIBRARY_PATH=%s:%s", nl.ionLibPath, os.Getenv("DYLD_LIBRARY_PATH"))
	ldPath := fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", nl.ionLibPath, os.Getenv("LD_LIBRARY_PATH"))
	env = append(env, dyldPath, ldPath)

	return env
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetNodeNumber returns the node number
func (nl *NodeLifecycle) GetNodeNumber() int {
	return nl.config.NodeNumber
}

// GetNodeID returns the node ID
func (nl *NodeLifecycle) GetNodeID() string {
	return nl.config.NodeID
}

// GetConfigDir returns the configuration directory
func (nl *NodeLifecycle) GetConfigDir() string {
	return nl.config.ConfigDir
}

// WaitForReady waits for ION-DTN to be fully initialized
func (nl *NodeLifecycle) WaitForReady(timeout time.Duration) error {
	start := time.Now()
	
	for time.Since(start) < timeout {
		if nl.IsRunning() {
			// Additional check: try to query ION status
			cmd := exec.Command(filepath.Join(nl.ionBinPath, "bpadmin"), ".")
			cmd.Stdin = strings.NewReader("i\nq\n")
			cmd.Env = nl.buildEnv()
			
			if err := cmd.Run(); err == nil {
				return nil // ION is ready
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for ION-DTN to be ready")
}
