package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"terrestrial-dtn/pkg/hdtn"

	"gopkg.in/yaml.v3"
)

// Config represents the node configuration file
type Config struct {
	NodeID          string `yaml:"node_id" json:"node_id"`
	NodeNumber      int    `yaml:"node_number" json:"node_number"`
	Callsign        string `yaml:"callsign" json:"callsign"`
	HDTNBinary      string `yaml:"hdtn_binary" json:"hdtn_binary"`
	HDTNConfig      string `yaml:"hdtn_config" json:"hdtn_config"`
	ContactPlanFile string `yaml:"contact_plan_file" json:"contact_plan_file"`
	TelemetryPort   int    `yaml:"telemetry_port" json:"telemetry_port"`
	TelemetryFile   string `yaml:"telemetry_file" json:"telemetry_file"`
	HealthInterval  int    `yaml:"health_interval" json:"health_interval"` // seconds
}

var (
	configFile    = flag.String("config", "", "Path to configuration file (YAML or JSON)")
	nodeID        = flag.String("node-id", "", "Node identifier (overrides config file)")
	nodeNumber    = flag.Int("node-number", 0, "HDTN node number (overrides config file)")
	hdtnBinary    = flag.String("hdtn-binary", "", "Path to HDTN binary (overrides config file)")
	hdtnConfig    = flag.String("hdtn-config", "", "Path to HDTN JSON configuration file (overrides config file)")
	telemetryPort = flag.Int("telemetry-port", 8080, "HTTP port for telemetry endpoint")
	showVersion   = flag.Bool("version", false, "Show version information")
)

const version = "2.0.0"

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("dtn-node version %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting DTN node: %s (node %d)", config.NodeID, config.NodeNumber)
	log.Printf("  Callsign: %s", config.Callsign)
	log.Printf("  HDTN binary: %s", config.HDTNBinary)
	log.Printf("  HDTN config: %s", config.HDTNConfig)
	log.Printf("  Telemetry port: %d", config.TelemetryPort)

	// Create HDTN lifecycle manager
	lifecycle, err := hdtn.NewLifecycleManager(hdtn.LifecycleConfig{
		BinaryPath: config.HDTNBinary,
		ConfigPath: config.HDTNConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create lifecycle manager: %v", err)
	}

	// Start HDTN
	log.Printf("Starting HDTN...")
	if err := lifecycle.Start(); err != nil {
		log.Fatalf("Failed to start HDTN: %v", err)
	}
	log.Printf("HDTN is ready")

	// Create telemetry collector
	telemetryBaseURL := fmt.Sprintf("http://localhost:%d", 10305)
	telemetryCollector := hdtn.NewTelemetryCollector(telemetryBaseURL, config.NodeID, config.NodeNumber)
	telemetryCollector.SetRunningCheck(lifecycle.IsRunning)

	// Create contact plan manager
	contactPlanManager := hdtn.NewContactPlanManager(telemetryBaseURL)

	// Load contact plan if specified
	if config.ContactPlanFile != "" {
		log.Printf("Loading contact plan from %s...", config.ContactPlanFile)
		if err := contactPlanManager.LoadFromFile(config.ContactPlanFile); err != nil {
			log.Printf("Warning: Failed to load contact plan: %v", err)
		} else {
			log.Printf("Contact plan loaded successfully")
			if err := contactPlanManager.Apply(); err != nil {
				log.Printf("Warning: Failed to apply contact plan: %v", err)
			} else {
				log.Printf("Contact plan applied to HDTN")
			}
		}
	}

	// Start telemetry HTTP server
	if config.TelemetryPort > 0 {
		go startTelemetryServer(config.TelemetryPort, telemetryCollector, contactPlanManager)
		log.Printf("Telemetry server started on http://localhost:%d", config.TelemetryPort)
	}

	// Start health monitoring
	healthInterval := time.Duration(config.HealthInterval) * time.Second
	if healthInterval == 0 {
		healthInterval = 10 * time.Second
	}
	go monitorHealth(lifecycle, telemetryCollector, config.TelemetryFile, healthInterval)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Node is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Graceful shutdown
	log.Printf("Stopping HDTN...")
	if err := lifecycle.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Printf("Node stopped")
}

// loadConfig loads configuration from file or command-line flags
func loadConfig() (*Config, error) {
	config := &Config{
		TelemetryPort:  8080,
		HealthInterval: 10,
	}

	// Load from config file if specified
	if *configFile != "" {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Determine format by extension
		ext := filepath.Ext(*configFile)
		if ext == ".yaml" || ext == ".yml" {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse YAML config: %w", err)
			}
		} else if ext == ".json" {
			if err := json.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse JSON config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("unsupported config file format: %s", ext)
		}
	}

	// Override with command-line flags
	if *nodeID != "" {
		config.NodeID = *nodeID
	}
	if *nodeNumber > 0 {
		config.NodeNumber = *nodeNumber
	}
	if *hdtnBinary != "" {
		config.HDTNBinary = *hdtnBinary
	}
	if *hdtnConfig != "" {
		config.HDTNConfig = *hdtnConfig
	}
	if *telemetryPort > 0 {
		config.TelemetryPort = *telemetryPort
	}

	// Validate required fields
	if config.NodeID == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	if config.NodeNumber <= 0 {
		return nil, fmt.Errorf("node_number is required and must be positive")
	}
	if config.HDTNBinary == "" {
		return nil, fmt.Errorf("hdtn_binary is required")
	}
	if config.HDTNConfig == "" {
		return nil, fmt.Errorf("hdtn_config is required")
	}

	// Expand paths
	if !filepath.IsAbs(config.HDTNBinary) {
		absPath, err := filepath.Abs(config.HDTNBinary)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve hdtn_binary: %w", err)
		}
		config.HDTNBinary = absPath
	}

	if !filepath.IsAbs(config.HDTNConfig) {
		absPath, err := filepath.Abs(config.HDTNConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve hdtn_config: %w", err)
		}
		config.HDTNConfig = absPath
	}

	return config, nil
}

// monitorHealth periodically collects and logs health information
func monitorHealth(lifecycle *hdtn.LifecycleManager, collector *hdtn.TelemetryCollector, telemetryFile string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if !lifecycle.IsRunning() {
			log.Printf("Warning: HDTN is not running")
			continue
		}

		telemetry, err := collector.Collect()
		if err != nil {
			log.Printf("Failed to collect telemetry: %v", err)
			continue
		}

		// Log health summary
		log.Printf("Health: uptime=%ds, storage=%.1f%%, bundles_stored=%d, bundles_sent=%d, bundles_received=%d",
			telemetry.Health.UptimeSeconds,
			telemetry.Health.StoragePercent,
			telemetry.BundleProtocol.BundlesStored,
			telemetry.BundleProtocol.BundlesSent,
			telemetry.BundleProtocol.BundlesReceived)

		// Save to file if specified
		if telemetryFile != "" {
			data, err := json.MarshalIndent(telemetry, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal telemetry: %v", err)
				continue
			}
			if err := os.WriteFile(telemetryFile, data, 0644); err != nil {
				log.Printf("Failed to save telemetry to file: %v", err)
			}
		}
	}
}

// startTelemetryServer starts an HTTP server for telemetry
func startTelemetryServer(port int, collector *hdtn.TelemetryCollector, cpm *hdtn.ContactPlanManager) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		telemetry, err := collector.Collect()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(telemetry)
	})

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		contacts, err := cpm.ListContacts()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"contacts": contacts,
		})
	})

	mux.HandleFunc("/contacts/active", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		currentTime := time.Now().Unix()
		active := cpm.GetActiveContacts(currentTime)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"active_contacts": active,
			"current_time":    currentTime,
		})
	})

	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}
