package hdtn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTelemetry_ProcessNotRunning(t *testing.T) {
	tc := NewTelemetryCollector("http://localhost:10305", "ipn:1.0", 1)
	tc.SetRunningCheck(func() bool { return false })

	_, err := tc.Collect()
	if err == nil {
		t.Fatal("expected error when process is not running")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error should mention 'not running', got: %v", err)
	}
}

func TestTelemetry_APITimeout(t *testing.T) {
	// Create a server that delays longer than the client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tc := NewTelemetryCollector(server.URL, "ipn:1.0", 1)
	// Override the client timeout to something shorter for testing
	tc.httpClient = &http.Client{Timeout: 50 * time.Millisecond}

	_, err := tc.Collect()
	if err == nil {
		t.Fatal("expected error on API timeout")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("error should mention 'unavailable', got: %v", err)
	}
}

func TestTelemetry_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json!!!`))
	}))
	defer server.Close()

	tc := NewTelemetryCollector(server.URL, "ipn:1.0", 1)

	_, err := tc.Collect()
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention 'parse', got: %v", err)
	}
}

func TestTelemetry_SuccessfulCollection(t *testing.T) {
	responseJSON := `{
		"bundleCountStorage": 42,
		"bundleCountEgress": 100,
		"bundleCountIngress": 200,
		"bundleByteCountEgress": 5000,
		"bundleByteCountIngress": 8000,
		"numActiveSendSessions": 3,
		"numActiveRecvSessions": 2,
		"totalCompletedSessions": 50,
		"totalRetransmissions": 5,
		"usedSpaceBytes": 1048576,
		"totalSpaceBytes": 10485760
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the correct endpoint is called
		if r.URL.Path != "/api/v1/telemetry" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	tc := NewTelemetryCollector(server.URL, "ipn:5.0", 5)
	tc.SetRunningCheck(func() bool { return true })

	telemetry, err := tc.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify node info
	if telemetry.NodeID != "ipn:5.0" {
		t.Errorf("NodeID: got %q, want %q", telemetry.NodeID, "ipn:5.0")
	}
	if telemetry.NodeNumber != 5 {
		t.Errorf("NodeNumber: got %d, want %d", telemetry.NodeNumber, 5)
	}

	// Verify timestamp is recent (within last 5 seconds)
	if time.Since(telemetry.Timestamp) > 5*time.Second {
		t.Errorf("Timestamp too old: %v", telemetry.Timestamp)
	}
	// Verify RFC 3339 format
	formatted := telemetry.Timestamp.Format(time.RFC3339)
	if formatted == "" {
		t.Error("Timestamp should format as RFC 3339")
	}

	// Verify bundle protocol fields
	if telemetry.BundleProtocol.BundlesStored != 42 {
		t.Errorf("BundlesStored: got %d, want 42", telemetry.BundleProtocol.BundlesStored)
	}
	if telemetry.BundleProtocol.BundlesSent != 100 {
		t.Errorf("BundlesSent: got %d, want 100", telemetry.BundleProtocol.BundlesSent)
	}
	if telemetry.BundleProtocol.BundlesReceived != 200 {
		t.Errorf("BundlesReceived: got %d, want 200", telemetry.BundleProtocol.BundlesReceived)
	}
	if telemetry.BundleProtocol.BytesSent != 5000 {
		t.Errorf("BytesSent: got %d, want 5000", telemetry.BundleProtocol.BytesSent)
	}
	if telemetry.BundleProtocol.BytesReceived != 8000 {
		t.Errorf("BytesReceived: got %d, want 8000", telemetry.BundleProtocol.BytesReceived)
	}
	if telemetry.BundleProtocol.StorageUsedBytes != 1048576 {
		t.Errorf("StorageUsedBytes: got %d, want 1048576", telemetry.BundleProtocol.StorageUsedBytes)
	}
	if telemetry.BundleProtocol.StorageQuotaBytes != 10485760 {
		t.Errorf("StorageQuotaBytes: got %d, want 10485760", telemetry.BundleProtocol.StorageQuotaBytes)
	}

	// Verify LTP fields
	if telemetry.LTP.SessionsActive != 5 { // 3 send + 2 recv
		t.Errorf("SessionsActive: got %d, want 5", telemetry.LTP.SessionsActive)
	}
	if telemetry.LTP.SessionsCompleted != 50 {
		t.Errorf("SessionsCompleted: got %d, want 50", telemetry.LTP.SessionsCompleted)
	}
	if telemetry.LTP.Retransmissions != 5 {
		t.Errorf("Retransmissions: got %d, want 5", telemetry.LTP.Retransmissions)
	}

	// Verify health
	if !telemetry.Health.Running {
		t.Error("Health.Running should be true")
	}
	// Storage percent: 1048576 / 10485760 * 100 = 10.0
	expectedPercent := 10.0
	if telemetry.Health.StoragePercent != expectedPercent {
		t.Errorf("StoragePercent: got %f, want %f", telemetry.Health.StoragePercent, expectedPercent)
	}
}
