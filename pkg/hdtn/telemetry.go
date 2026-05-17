package hdtn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Telemetry represents HDTN node telemetry data.
// Field names and nesting match the existing telemetry HTTP server format for backward compatibility.
type Telemetry struct {
	Timestamp      time.Time        `json:"timestamp"`
	NodeID         string           `json:"node_id"`
	NodeNumber     int              `json:"node_number"`
	BundleProtocol BPTelemetry      `json:"bundle_protocol"`
	LTP            LTPTelemetry     `json:"ltp"`
	ContactPlan    ContactTelemetry `json:"contact_plan"`
	Health         HealthStatus     `json:"health"`
}

// BPTelemetry holds bundle protocol statistics.
type BPTelemetry struct {
	BundlesStored     int64 `json:"bundles_stored"`
	BundlesReceived   int64 `json:"bundles_received"`
	BundlesSent       int64 `json:"bundles_sent"`
	BundlesForwarded  int64 `json:"bundles_forwarded"`
	BundlesExpired    int64 `json:"bundles_expired"`
	BytesReceived     int64 `json:"bytes_received"`
	BytesSent         int64 `json:"bytes_sent"`
	StorageUsedBytes  int64 `json:"storage_used_bytes"`
	StorageQuotaBytes int64 `json:"storage_quota_bytes"`
}

// LTPTelemetry holds LTP session statistics.
type LTPTelemetry struct {
	SessionsActive    int   `json:"sessions_active"`
	SessionsCompleted int64 `json:"sessions_completed"`
	SessionsFailed    int64 `json:"sessions_failed"`
	SegmentsSent      int64 `json:"segments_sent"`
	SegmentsReceived  int64 `json:"segments_received"`
	Retransmissions   int64 `json:"retransmissions"`
}

// ContactTelemetry holds contact plan statistics.
type ContactTelemetry struct {
	ContactsActive    int    `json:"contacts_active"`
	ContactsCompleted int64  `json:"contacts_completed"`
	ContactsMissed    int64  `json:"contacts_missed"`
	NextContactTime   *int64 `json:"next_contact_time,omitempty"`
}

// HealthStatus holds process health information.
type HealthStatus struct {
	Running        bool    `json:"running"`
	UptimeSeconds  int64   `json:"uptime_seconds"`
	StoragePercent float64 `json:"storage_percent"`
	ErrorCount     int     `json:"error_count"`
}

// hdtnAllTelemetry is the raw response from GET /api/v1/telemetry on the HDTN REST API.
type hdtnAllTelemetry struct {
	BundleCountStorage     int64 `json:"bundleCountStorage"`
	BundleCountEgress      int64 `json:"bundleCountEgress"`
	BundleCountIngress     int64 `json:"bundleCountIngress"`
	BundleByteCountEgress  int64 `json:"bundleByteCountEgress"`
	BundleByteCountIngress int64 `json:"bundleByteCountIngress"`
	// LTP fields
	NumActiveSendSessions  int   `json:"numActiveSendSessions"`
	NumActiveRecvSessions  int   `json:"numActiveRecvSessions"`
	TotalCompletedSessions int64 `json:"totalCompletedSessions"`
	TotalRetransmissions   int64 `json:"totalRetransmissions"`
	// Storage fields
	UsedSpaceBytes  int64 `json:"usedSpaceBytes"`
	TotalSpaceBytes int64 `json:"totalSpaceBytes"`
}

// TelemetryCollector queries the HDTN REST API for telemetry data.
type TelemetryCollector struct {
	baseURL    string
	nodeID     string
	nodeNumber int
	httpClient *http.Client
	// runningCheck is called before making API calls to verify the process is running.
	// If nil, the check is skipped.
	runningCheck func() bool
}

// NewTelemetryCollector creates a telemetry collector that queries the HDTN REST API.
func NewTelemetryCollector(baseURL string, nodeID string, nodeNumber int) *TelemetryCollector {
	return &TelemetryCollector{
		baseURL:    baseURL,
		nodeID:     nodeID,
		nodeNumber: nodeNumber,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// SetRunningCheck sets a function that checks if the HDTN process is running.
// The telemetry collector will call this before making API requests.
func (tc *TelemetryCollector) SetRunningCheck(fn func() bool) {
	tc.runningCheck = fn
}

// Collect queries the HDTN REST API and returns telemetry in the backward-compatible format.
// Returns an error if the process is not running or the API is unavailable.
func (tc *TelemetryCollector) Collect() (*Telemetry, error) {
	// Check if process is running before making API call
	if tc.runningCheck != nil && !tc.runningCheck() {
		return nil, fmt.Errorf("hdtn process is not running")
	}

	// Query HDTN REST API
	url := tc.baseURL + "/api/v1/telemetry"
	resp, err := tc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("telemetry API unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read telemetry response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telemetry API returned status %d", resp.StatusCode)
	}

	// Parse the HDTN response
	var raw hdtnAllTelemetry
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse telemetry response: %w", err)
	}

	// Map HDTN fields to backward-compatible Telemetry struct
	// Zero-filling happens naturally since Go zero-initializes struct fields,
	// and json.Unmarshal only sets fields present in the JSON.
	telemetry := &Telemetry{
		Timestamp:  time.Now().UTC(),
		NodeID:     tc.nodeID,
		NodeNumber: tc.nodeNumber,
		BundleProtocol: BPTelemetry{
			BundlesStored:     raw.BundleCountStorage,
			BundlesReceived:   raw.BundleCountIngress,
			BundlesSent:       raw.BundleCountEgress,
			BytesReceived:     raw.BundleByteCountIngress,
			BytesSent:         raw.BundleByteCountEgress,
			StorageUsedBytes:  raw.UsedSpaceBytes,
			StorageQuotaBytes: raw.TotalSpaceBytes,
		},
		LTP: LTPTelemetry{
			SessionsActive:    raw.NumActiveSendSessions + raw.NumActiveRecvSessions,
			SessionsCompleted: raw.TotalCompletedSessions,
			Retransmissions:   raw.TotalRetransmissions,
		},
		Health: HealthStatus{
			Running: true,
		},
	}

	// Calculate storage percent if quota is available
	if raw.TotalSpaceBytes > 0 {
		telemetry.Health.StoragePercent = float64(raw.UsedSpaceBytes) / float64(raw.TotalSpaceBytes) * 100.0
	}

	return telemetry, nil
}
