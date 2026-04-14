package ion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Telemetry represents ION-DTN node telemetry data
type Telemetry struct {
	Timestamp         time.Time         `json:"timestamp"`
	NodeID            string            `json:"node_id"`
	NodeNumber        int               `json:"node_number"`
	BundleProtocol    BPTelemetry       `json:"bundle_protocol"`
	LTP               LTPTelemetry      `json:"ltp"`
	ContactPlan       ContactTelemetry  `json:"contact_plan"`
	Health            HealthStatus      `json:"health"`
}

// BPTelemetry represents Bundle Protocol statistics
type BPTelemetry struct {
	BundlesStored     int64   `json:"bundles_stored"`
	BundlesReceived   int64   `json:"bundles_received"`
	BundlesSent       int64   `json:"bundles_sent"`
	BundlesForwarded  int64   `json:"bundles_forwarded"`
	BundlesExpired    int64   `json:"bundles_expired"`
	BytesReceived     int64   `json:"bytes_received"`
	BytesSent         int64   `json:"bytes_sent"`
	StorageUsedBytes  int64   `json:"storage_used_bytes"`
	StorageQuotaBytes int64   `json:"storage_quota_bytes"`
}

// LTPTelemetry represents LTP statistics
type LTPTelemetry struct {
	SessionsActive    int     `json:"sessions_active"`
	SessionsCompleted int64   `json:"sessions_completed"`
	SessionsFailed    int64   `json:"sessions_failed"`
	SegmentsSent      int64   `json:"segments_sent"`
	SegmentsReceived  int64   `json:"segments_received"`
	Retransmissions   int64   `json:"retransmissions"`
}

// ContactTelemetry represents contact plan statistics
type ContactTelemetry struct {
	ContactsActive    int     `json:"contacts_active"`
	ContactsCompleted int64   `json:"contacts_completed"`
	ContactsMissed    int64   `json:"contacts_missed"`
	NextContactTime   *int64  `json:"next_contact_time,omitempty"`
}

// HealthStatus represents node health
type HealthStatus struct {
	Running           bool    `json:"running"`
	UptimeSeconds     int64   `json:"uptime_seconds"`
	StoragePercent    float64 `json:"storage_percent"`
	ErrorCount        int     `json:"error_count"`
}

// TelemetryCollector collects telemetry from ION-DTN
type TelemetryCollector struct {
	lifecycle  *NodeLifecycle
	startTime  time.Time
}

// NewTelemetryCollector creates a new telemetry collector
func NewTelemetryCollector(lifecycle *NodeLifecycle) *TelemetryCollector {
	return &TelemetryCollector{
		lifecycle: lifecycle,
		startTime: time.Now(),
	}
}

// Collect gathers current telemetry from ION-DTN
func (tc *TelemetryCollector) Collect() (*Telemetry, error) {
	if !tc.lifecycle.IsRunning() {
		return nil, fmt.Errorf("ION-DTN is not running")
	}

	telemetry := &Telemetry{
		Timestamp:  time.Now(),
		NodeID:     tc.lifecycle.GetNodeID(),
		NodeNumber: tc.lifecycle.GetNodeNumber(),
	}

	// Collect BP statistics
	bpStats, err := tc.collectBPStats()
	if err != nil {
		return nil, fmt.Errorf("failed to collect BP stats: %w", err)
	}
	telemetry.BundleProtocol = bpStats

	// Collect LTP statistics
	ltpStats, err := tc.collectLTPStats()
	if err != nil {
		return nil, fmt.Errorf("failed to collect LTP stats: %w", err)
	}
	telemetry.LTP = ltpStats

	// Collect contact plan info
	contactStats, err := tc.collectContactStats()
	if err != nil {
		// Non-fatal - contact stats may not be available
		contactStats = ContactTelemetry{}
	}
	telemetry.ContactPlan = contactStats

	// Calculate health status
	telemetry.Health = tc.calculateHealth(bpStats)

	return telemetry, nil
}

// collectBPStats queries bpadmin for Bundle Protocol statistics
func (tc *TelemetryCollector) collectBPStats() (BPTelemetry, error) {
	stats := BPTelemetry{}

	// Run bpadmin with 'i' (info) command
	output, err := tc.runAdminCommand("bpadmin", "i")
	if err != nil {
		return stats, err
	}

	// Parse output for statistics
	// Example output format:
	// bundles stored: 5
	// bundles received: 10
	// bundles sent: 8
	// bytes received: 1024
	// bytes sent: 2048

	stats.BundlesStored = tc.extractInt64(output, `bundles?\s+stored[:\s]+(\d+)`)
	stats.BundlesReceived = tc.extractInt64(output, `bundles?\s+received[:\s]+(\d+)`)
	stats.BundlesSent = tc.extractInt64(output, `bundles?\s+sent[:\s]+(\d+)`)
	stats.BundlesForwarded = tc.extractInt64(output, `bundles?\s+forwarded[:\s]+(\d+)`)
	stats.BundlesExpired = tc.extractInt64(output, `bundles?\s+expired[:\s]+(\d+)`)
	stats.BytesReceived = tc.extractInt64(output, `bytes?\s+received[:\s]+(\d+)`)
	stats.BytesSent = tc.extractInt64(output, `bytes?\s+sent[:\s]+(\d+)`)
	stats.StorageUsedBytes = tc.extractInt64(output, `storage\s+used[:\s]+(\d+)`)
	stats.StorageQuotaBytes = tc.extractInt64(output, `storage\s+quota[:\s]+(\d+)`)

	return stats, nil
}

// collectLTPStats queries ltpadmin for LTP statistics
func (tc *TelemetryCollector) collectLTPStats() (LTPTelemetry, error) {
	stats := LTPTelemetry{}

	// Run ltpadmin with 'i' (info) command
	output, err := tc.runAdminCommand("ltpadmin", "i")
	if err != nil {
		return stats, err
	}

	// Parse LTP statistics
	stats.SessionsActive = int(tc.extractInt64(output, `sessions?\s+active[:\s]+(\d+)`))
	stats.SessionsCompleted = tc.extractInt64(output, `sessions?\s+completed[:\s]+(\d+)`)
	stats.SessionsFailed = tc.extractInt64(output, `sessions?\s+failed[:\s]+(\d+)`)
	stats.SegmentsSent = tc.extractInt64(output, `segments?\s+sent[:\s]+(\d+)`)
	stats.SegmentsReceived = tc.extractInt64(output, `segments?\s+received[:\s]+(\d+)`)
	stats.Retransmissions = tc.extractInt64(output, `retransmissions?[:\s]+(\d+)`)

	return stats, nil
}

// collectContactStats queries ionadmin for contact plan information
func (tc *TelemetryCollector) collectContactStats() (ContactTelemetry, error) {
	stats := ContactTelemetry{}

	// Run ionadmin with 'l contact' command
	output, err := tc.runAdminCommand("ionadmin", "l contact")
	if err != nil {
		return stats, err
	}

	// Count active contacts (simplified - would need current time comparison)
	contactLines := strings.Count(output, "contact")
	stats.ContactsActive = contactLines

	return stats, nil
}

// calculateHealth determines node health status
func (tc *TelemetryCollector) calculateHealth(bpStats BPTelemetry) HealthStatus {
	health := HealthStatus{
		Running:       tc.lifecycle.IsRunning(),
		UptimeSeconds: int64(time.Since(tc.startTime).Seconds()),
	}

	// Calculate storage percentage
	if bpStats.StorageQuotaBytes > 0 {
		health.StoragePercent = float64(bpStats.StorageUsedBytes) / float64(bpStats.StorageQuotaBytes) * 100.0
	}

	return health
}

// runAdminCommand executes an ION admin command and returns output
func (tc *TelemetryCollector) runAdminCommand(adminCmd, command string) (string, error) {
	cmdPath := tc.lifecycle.ionBinPath + "/" + adminCmd
	
	cmd := exec.Command(cmdPath, ".")
	cmd.Stdin = strings.NewReader(command + "\nq\n")
	cmd.Env = tc.lifecycle.buildEnv()
	cmd.Dir = tc.lifecycle.config.WorkingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s failed: %w, stderr: %s", adminCmd, err, stderr.String())
	}

	return stdout.String(), nil
}

// extractInt64 extracts an integer value from text using a regex pattern
func (tc *TelemetryCollector) extractInt64(text, pattern string) int64 {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		val, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return val
		}
	}
	return 0
}

// ToJSON converts telemetry to JSON format
func (t *Telemetry) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// ToJSONString converts telemetry to JSON string
func (t *Telemetry) ToJSONString() (string, error) {
	data, err := t.ToJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveToFile saves telemetry to a JSON file
func (t *Telemetry) SaveToFile(path string) error {
	data, err := t.ToJSON()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
