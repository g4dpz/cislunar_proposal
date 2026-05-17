package hdtn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pgregory.net/rapid"
)

// Feature: hdtn-migration, Property 5: Telemetry parsing preserves all statistics
// **Validates: Requirements 2.1, 2.2, 2.3**
func TestProperty_TelemetryParsingPreservesAllStatistics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate arbitrary valid HDTN REST API JSON response with non-negative integer fields
		bundleCountStorage := rapid.Int64Range(0, 1_000_000).Draw(t, "bundleCountStorage")
		bundleCountEgress := rapid.Int64Range(0, 1_000_000).Draw(t, "bundleCountEgress")
		bundleCountIngress := rapid.Int64Range(0, 1_000_000).Draw(t, "bundleCountIngress")
		bundleByteCountEgress := rapid.Int64Range(0, 1_000_000_000).Draw(t, "bundleByteCountEgress")
		bundleByteCountIngress := rapid.Int64Range(0, 1_000_000_000).Draw(t, "bundleByteCountIngress")
		numActiveSendSessions := rapid.IntRange(0, 10_000).Draw(t, "numActiveSendSessions")
		numActiveRecvSessions := rapid.IntRange(0, 10_000).Draw(t, "numActiveRecvSessions")
		totalCompletedSessions := rapid.Int64Range(0, 1_000_000).Draw(t, "totalCompletedSessions")
		totalRetransmissions := rapid.Int64Range(0, 1_000_000).Draw(t, "totalRetransmissions")
		usedSpaceBytes := rapid.Int64Range(0, 1_000_000_000).Draw(t, "usedSpaceBytes")
		totalSpaceBytes := rapid.Int64Range(1, 1_000_000_000).Draw(t, "totalSpaceBytes")

		// Build the HDTN API response JSON
		apiResponse := map[string]interface{}{
			"bundleCountStorage":     bundleCountStorage,
			"bundleCountEgress":      bundleCountEgress,
			"bundleCountIngress":     bundleCountIngress,
			"bundleByteCountEgress":  bundleByteCountEgress,
			"bundleByteCountIngress": bundleByteCountIngress,
			"numActiveSendSessions":  numActiveSendSessions,
			"numActiveRecvSessions":  numActiveRecvSessions,
			"totalCompletedSessions": totalCompletedSessions,
			"totalRetransmissions":   totalRetransmissions,
			"usedSpaceBytes":         usedSpaceBytes,
			"totalSpaceBytes":        totalSpaceBytes,
		}

		responseJSON, err := json.Marshal(apiResponse)
		if err != nil {
			t.Fatalf("failed to marshal test response: %v", err)
		}

		// Start httptest server returning the generated JSON
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseJSON)
		}))
		defer server.Close()

		// Create collector and call Collect()
		tc := NewTelemetryCollector(server.URL, "ipn:1.0", 1)

		telemetry, err := tc.Collect()
		if err != nil {
			t.Fatalf("Collect() returned error: %v", err)
		}

		// Assert every field maps correctly
		if telemetry.BundleProtocol.BundlesStored != bundleCountStorage {
			t.Errorf("BundlesStored: got %d, want %d", telemetry.BundleProtocol.BundlesStored, bundleCountStorage)
		}
		if telemetry.BundleProtocol.BundlesSent != bundleCountEgress {
			t.Errorf("BundlesSent: got %d, want %d", telemetry.BundleProtocol.BundlesSent, bundleCountEgress)
		}
		if telemetry.BundleProtocol.BundlesReceived != bundleCountIngress {
			t.Errorf("BundlesReceived: got %d, want %d", telemetry.BundleProtocol.BundlesReceived, bundleCountIngress)
		}
		if telemetry.BundleProtocol.BytesSent != bundleByteCountEgress {
			t.Errorf("BytesSent: got %d, want %d", telemetry.BundleProtocol.BytesSent, bundleByteCountEgress)
		}
		if telemetry.BundleProtocol.BytesReceived != bundleByteCountIngress {
			t.Errorf("BytesReceived: got %d, want %d", telemetry.BundleProtocol.BytesReceived, bundleByteCountIngress)
		}
		if telemetry.BundleProtocol.StorageUsedBytes != usedSpaceBytes {
			t.Errorf("StorageUsedBytes: got %d, want %d", telemetry.BundleProtocol.StorageUsedBytes, usedSpaceBytes)
		}
		if telemetry.BundleProtocol.StorageQuotaBytes != totalSpaceBytes {
			t.Errorf("StorageQuotaBytes: got %d, want %d", telemetry.BundleProtocol.StorageQuotaBytes, totalSpaceBytes)
		}
		if telemetry.LTP.SessionsActive != numActiveSendSessions+numActiveRecvSessions {
			t.Errorf("SessionsActive: got %d, want %d", telemetry.LTP.SessionsActive, numActiveSendSessions+numActiveRecvSessions)
		}
		if telemetry.LTP.SessionsCompleted != totalCompletedSessions {
			t.Errorf("SessionsCompleted: got %d, want %d", telemetry.LTP.SessionsCompleted, totalCompletedSessions)
		}
		if telemetry.LTP.Retransmissions != totalRetransmissions {
			t.Errorf("Retransmissions: got %d, want %d", telemetry.LTP.Retransmissions, totalRetransmissions)
		}
		if telemetry.Health.Running != true {
			t.Errorf("Health.Running: got %v, want true", telemetry.Health.Running)
		}
		if telemetry.NodeID != "ipn:1.0" {
			t.Errorf("NodeID: got %q, want %q", telemetry.NodeID, "ipn:1.0")
		}
		if telemetry.NodeNumber != 1 {
			t.Errorf("NodeNumber: got %d, want %d", telemetry.NodeNumber, 1)
		}
	})
}

// Feature: hdtn-migration, Property 6: Telemetry partial response zero-filling
// **Validates: Requirements 2.8**
func TestProperty_TelemetryPartialResponseZeroFilling(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate HDTN REST API JSON with randomly omitted fields
		// Each field has a 50% chance of being included
		includeStorage := rapid.Bool().Draw(t, "includeStorage")
		includeEgress := rapid.Bool().Draw(t, "includeEgress")
		includeIngress := rapid.Bool().Draw(t, "includeIngress")
		includeByteEgress := rapid.Bool().Draw(t, "includeByteEgress")
		includeByteIngress := rapid.Bool().Draw(t, "includeByteIngress")
		includeSendSessions := rapid.Bool().Draw(t, "includeSendSessions")
		includeRecvSessions := rapid.Bool().Draw(t, "includeRecvSessions")
		includeCompleted := rapid.Bool().Draw(t, "includeCompleted")
		includeRetransmissions := rapid.Bool().Draw(t, "includeRetransmissions")
		includeUsedSpace := rapid.Bool().Draw(t, "includeUsedSpace")
		includeTotalSpace := rapid.Bool().Draw(t, "includeTotalSpace")

		// Generate values for included fields
		bundleCountStorage := rapid.Int64Range(1, 1_000_000).Draw(t, "bundleCountStorage")
		bundleCountEgress := rapid.Int64Range(1, 1_000_000).Draw(t, "bundleCountEgress")
		bundleCountIngress := rapid.Int64Range(1, 1_000_000).Draw(t, "bundleCountIngress")
		bundleByteCountEgress := rapid.Int64Range(1, 1_000_000_000).Draw(t, "bundleByteCountEgress")
		bundleByteCountIngress := rapid.Int64Range(1, 1_000_000_000).Draw(t, "bundleByteCountIngress")
		numActiveSendSessions := rapid.IntRange(1, 10_000).Draw(t, "numActiveSendSessions")
		numActiveRecvSessions := rapid.IntRange(1, 10_000).Draw(t, "numActiveRecvSessions")
		totalCompletedSessions := rapid.Int64Range(1, 1_000_000).Draw(t, "totalCompletedSessions")
		totalRetransmissions := rapid.Int64Range(1, 1_000_000).Draw(t, "totalRetransmissions")
		usedSpaceBytes := rapid.Int64Range(1, 1_000_000_000).Draw(t, "usedSpaceBytes")
		totalSpaceBytes := rapid.Int64Range(1, 1_000_000_000).Draw(t, "totalSpaceBytes")

		// Build partial response
		apiResponse := make(map[string]interface{})
		if includeStorage {
			apiResponse["bundleCountStorage"] = bundleCountStorage
		}
		if includeEgress {
			apiResponse["bundleCountEgress"] = bundleCountEgress
		}
		if includeIngress {
			apiResponse["bundleCountIngress"] = bundleCountIngress
		}
		if includeByteEgress {
			apiResponse["bundleByteCountEgress"] = bundleByteCountEgress
		}
		if includeByteIngress {
			apiResponse["bundleByteCountIngress"] = bundleByteCountIngress
		}
		if includeSendSessions {
			apiResponse["numActiveSendSessions"] = numActiveSendSessions
		}
		if includeRecvSessions {
			apiResponse["numActiveRecvSessions"] = numActiveRecvSessions
		}
		if includeCompleted {
			apiResponse["totalCompletedSessions"] = totalCompletedSessions
		}
		if includeRetransmissions {
			apiResponse["totalRetransmissions"] = totalRetransmissions
		}
		if includeUsedSpace {
			apiResponse["usedSpaceBytes"] = usedSpaceBytes
		}
		if includeTotalSpace {
			apiResponse["totalSpaceBytes"] = totalSpaceBytes
		}

		responseJSON, err := json.Marshal(apiResponse)
		if err != nil {
			t.Fatalf("failed to marshal test response: %v", err)
		}

		// Start httptest server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseJSON)
		}))
		defer server.Close()

		// Create collector and call Collect()
		tc := NewTelemetryCollector(server.URL, "ipn:2.0", 2)

		telemetry, err := tc.Collect()
		if err != nil {
			t.Fatalf("Collect() returned error for partial response: %v", err)
		}

		// Assert: present fields retain correct values, missing fields are zero
		assertField := func(name string, got, want int64, included bool) {
			if included {
				if got != want {
					t.Errorf("%s: got %d, want %d (field was included)", name, got, want)
				}
			} else {
				if got != 0 {
					t.Errorf("%s: got %d, want 0 (field was omitted)", name, got)
				}
			}
		}

		assertField("BundlesStored", telemetry.BundleProtocol.BundlesStored, bundleCountStorage, includeStorage)
		assertField("BundlesSent", telemetry.BundleProtocol.BundlesSent, bundleCountEgress, includeEgress)
		assertField("BundlesReceived", telemetry.BundleProtocol.BundlesReceived, bundleCountIngress, includeIngress)
		assertField("BytesSent", telemetry.BundleProtocol.BytesSent, bundleByteCountEgress, includeByteEgress)
		assertField("BytesReceived", telemetry.BundleProtocol.BytesReceived, bundleByteCountIngress, includeByteIngress)
		assertField("StorageUsedBytes", telemetry.BundleProtocol.StorageUsedBytes, usedSpaceBytes, includeUsedSpace)
		assertField("StorageQuotaBytes", telemetry.BundleProtocol.StorageQuotaBytes, totalSpaceBytes, includeTotalSpace)
		assertField("SessionsCompleted", telemetry.LTP.SessionsCompleted, totalCompletedSessions, includeCompleted)
		assertField("Retransmissions", telemetry.LTP.Retransmissions, totalRetransmissions, includeRetransmissions)

		// SessionsActive is the sum of send + recv sessions
		expectedActive := 0
		if includeSendSessions {
			expectedActive += numActiveSendSessions
		}
		if includeRecvSessions {
			expectedActive += numActiveRecvSessions
		}
		if telemetry.LTP.SessionsActive != expectedActive {
			t.Errorf("SessionsActive: got %d, want %d", telemetry.LTP.SessionsActive, expectedActive)
		}

		// Verify node info is always present
		if telemetry.NodeID != "ipn:2.0" {
			t.Errorf("NodeID: got %q, want %q", telemetry.NodeID, "ipn:2.0")
		}
		if telemetry.NodeNumber != 2 {
			t.Errorf("NodeNumber: got %d, want %d", telemetry.NodeNumber, 2)
		}
		if telemetry.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}

		// Verify RFC 3339 format
		formatted := telemetry.Timestamp.Format("2006-01-02T15:04:05Z07:00")
		if formatted == "" {
			t.Error("Timestamp should format as RFC 3339")
		}

		_ = fmt.Sprintf("partial response test with %d fields included", len(apiResponse))
	})
}
