package sband_iq

import (
	"testing"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/radio/sband_transceiver"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "S-band default config",
			config:  DefaultSBandConfig("W1ABC"),
			wantErr: false,
		},
		{
			name:    "X-band default config",
			config:  DefaultXBandConfig("W1ABC"),
			wantErr: false,
		},
		{
			name: "custom S-band config",
			config: Config{
				Callsign:       "K2XYZ",
				Band:           sband_transceiver.BandS,
				CenterFreq:     2.2e9,
				SampleRate:     8000.0,
				DataRate:       500,
				TXPower:        7.0,
				TXGain:         12.0,
				RXGain:         38.0,
				FECEnabled:     true,
				FECType:        sband_transceiver.FECTurbo,
				LightTimeDelay: 1500 * time.Millisecond,
				LTPTimeout:     15 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cla, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cla == nil {
					t.Error("New() returned nil CLA")
					return
				}
				if cla.config.Callsign != tt.config.Callsign {
					t.Errorf("Callsign = %v, want %v", cla.config.Callsign, tt.config.Callsign)
				}
				if cla.config.DataRate != tt.config.DataRate {
					t.Errorf("DataRate = %v, want %v", cla.config.DataRate, tt.config.DataRate)
				}
				if cla.config.Band != tt.config.Band {
					t.Errorf("Band = %v, want %v", cla.config.Band, tt.config.Band)
				}
			}
		})
	}
}

func TestType(t *testing.T) {
	tests := []struct {
		name string
		band sband_transceiver.Band
		want string
	}{
		{
			name: "S-band type",
			band: sband_transceiver.BandS,
			want: "ax25ltp_sband_iq",
		},
		{
			name: "X-band type",
			band: sband_transceiver.BandX,
			want: "ax25ltp_xband_iq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSBandConfig("W1ABC")
			config.Band = tt.band
			cla, err := New(config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			claType := cla.Type()
			if claType.String() != tt.want {
				t.Errorf("Type() = %v, want %v", claType.String(), tt.want)
			}
		})
	}
}

func TestOpenClose(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "ground-station-1",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Add(10 * time.Minute).Unix(),
		DataRate:   500,
		LinkType:   contact.LinkTypeSBandIQ,
	}

	// Test Open
	err = cla.Open(window)
	if err != nil {
		t.Errorf("Open() error = %v", err)
	}

	if !cla.isOpen {
		t.Error("CLA should be open after Open()")
	}

	// Test double Open (should fail)
	err = cla.Open(window)
	if err == nil {
		t.Error("Open() should fail when already open")
	}

	// Test Close
	err = cla.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if cla.isOpen {
		t.Error("CLA should be closed after Close()")
	}

	// Test double Close (should succeed)
	err = cla.Close()
	if err != nil {
		t.Errorf("Close() should not error when already closed, got: %v", err)
	}
}

func TestSendBundle(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "ground-station-1",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Add(10 * time.Minute).Unix(),
		DataRate:   500,
		LinkType:   contact.LinkTypeSBandIQ,
	}

	err = cla.Open(window)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer cla.Close()

	// Create test bundle
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
		Payload:     []byte("Test payload for cislunar transmission"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600,
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	// Test SendBundle
	metrics, err := cla.SendBundle(bundle)
	if err != nil {
		t.Errorf("SendBundle() error = %v", err)
	}

	if metrics == nil {
		t.Error("SendBundle() returned nil metrics")
	}

	if metrics.BytesTransferred == 0 {
		t.Error("SendBundle() should have transferred bytes")
	}

	// Verify LTP session was created
	activeSessions := cla.GetActiveSessions()
	if activeSessions == 0 {
		t.Error("SendBundle() should have created an LTP session")
	}
}

func TestSendBundleWhenClosed(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
		Payload:     []byte("Test"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600,
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	// Try to send when closed
	_, err = cla.SendBundle(bundle)
	if err == nil {
		t.Error("SendBundle() should fail when CLA is closed")
	}
}

func TestRecvBundleWhenClosed(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Try to receive when closed
	_, _, err = cla.RecvBundle()
	if err == nil {
		t.Error("RecvBundle() should fail when CLA is closed")
	}
}

func TestStatus(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Initial status should be idle
	status := cla.Status()
	if status.String() != "idle" {
		t.Errorf("Initial status = %v, want idle", status)
	}
}

func TestLinkMetrics(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	metrics := cla.LinkMetrics()
	if metrics == nil {
		t.Error("LinkMetrics() returned nil")
	}

	// Initial metrics should be zero
	if metrics.BytesTransferred != 0 {
		t.Errorf("Initial BytesTransferred = %v, want 0", metrics.BytesTransferred)
	}
}

func TestLTPSessionManagement(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	config.LTPTimeout = 100 * time.Millisecond // Short timeout for testing
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	window := contact.ContactWindow{
		ContactID:  1,
		RemoteNode: "ground-station-1",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Add(10 * time.Minute).Unix(),
		DataRate:   500,
		LinkType:   contact.LinkTypeSBandIQ,
	}

	err = cla.Open(window)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer cla.Close()

	// Create and send bundle to create LTP session
	bundle := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    1,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
		Payload:     []byte("Test"),
		Priority:    bpa.PriorityNormal,
		Lifetime:    3600,
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	_, err = cla.SendBundle(bundle)
	if err != nil {
		t.Fatalf("SendBundle() error = %v", err)
	}

	// Verify session exists
	activeSessions := cla.GetActiveSessions()
	if activeSessions != 1 {
		t.Errorf("GetActiveSessions() = %v, want 1", activeSessions)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Cleanup expired sessions
	cla.CleanupSessions()

	// Verify session was cleaned up
	activeSessions = cla.GetActiveSessions()
	if activeSessions != 0 {
		t.Errorf("GetActiveSessions() after cleanup = %v, want 0", activeSessions)
	}
}

func TestLightTimeDelayConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		lightTimeDelay time.Duration
	}{
		{
			name:           "1 second delay",
			lightTimeDelay: 1 * time.Second,
		},
		{
			name:           "1.2 second delay (nominal)",
			lightTimeDelay: 1200 * time.Millisecond,
		},
		{
			name:           "2 second delay (max)",
			lightTimeDelay: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSBandConfig("W1ABC")
			config.LightTimeDelay = tt.lightTimeDelay
			cla, err := New(config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if cla.config.LightTimeDelay != tt.lightTimeDelay {
				t.Errorf("LightTimeDelay = %v, want %v", cla.config.LightTimeDelay, tt.lightTimeDelay)
			}
		})
	}
}

func TestFECConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		fecEnabled bool
		fecType    sband_transceiver.FECType
	}{
		{
			name:       "LDPC FEC enabled",
			fecEnabled: true,
			fecType:    sband_transceiver.FECLDPC,
		},
		{
			name:       "Turbo FEC enabled",
			fecEnabled: true,
			fecType:    sband_transceiver.FECTurbo,
		},
		{
			name:       "FEC disabled",
			fecEnabled: false,
			fecType:    sband_transceiver.FECLDPC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSBandConfig("W1ABC")
			config.FECEnabled = tt.fecEnabled
			config.FECType = tt.fecType
			cla, err := New(config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if cla.config.FECEnabled != tt.fecEnabled {
				t.Errorf("FECEnabled = %v, want %v", cla.config.FECEnabled, tt.fecEnabled)
			}
			if cla.config.FECType != tt.fecType {
				t.Errorf("FECType = %v, want %v", cla.config.FECType, tt.fecType)
			}
		})
	}
}

func TestBundleSerializationRoundTrip(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	original := &bpa.Bundle{
		ID: bpa.BundleID{
			SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "1.0"},
			CreationTimestamp: time.Now().Unix(),
			SequenceNumber:    42,
		},
		Destination: bpa.EndpointID{Scheme: "ipn", SSP: "2.0"},
		Payload:     []byte("Test payload for round-trip"),
		Priority:    bpa.PriorityExpedited,
		Lifetime:    7200,
		CreatedAt:   time.Now().Unix(),
		BundleType:  bpa.BundleTypeData,
	}

	// Serialize
	serialized := cla.serializeBundle(original)
	if len(serialized) == 0 {
		t.Error("serializeBundle() returned empty data")
	}

	// Deserialize
	deserialized, err := cla.deserializeBundle(serialized)
	if err != nil {
		t.Fatalf("deserializeBundle() error = %v", err)
	}

	// Verify key fields
	if deserialized.BundleType != original.BundleType {
		t.Errorf("BundleType = %v, want %v", deserialized.BundleType, original.BundleType)
	}
	if deserialized.Priority != original.Priority {
		t.Errorf("Priority = %v, want %v", deserialized.Priority, original.Priority)
	}
	if deserialized.Lifetime != original.Lifetime {
		t.Errorf("Lifetime = %v, want %v", deserialized.Lifetime, original.Lifetime)
	}
	if string(deserialized.Payload) != string(original.Payload) {
		t.Errorf("Payload = %v, want %v", string(deserialized.Payload), string(original.Payload))
	}
}

func TestAX25FramingRoundTrip(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	payload := []byte("Test AX.25 frame payload")

	// Create frame
	frame := cla.createAX25Frame(payload)
	if len(frame) <= len(payload) {
		t.Error("createAX25Frame() should add header")
	}

	// Extract payload
	extracted, err := cla.extractAX25Frame(frame)
	if err != nil {
		t.Fatalf("extractAX25Frame() error = %v", err)
	}

	if string(extracted) != string(payload) {
		t.Errorf("Extracted payload = %v, want %v", string(extracted), string(payload))
	}
}

func TestDataRateConfiguration(t *testing.T) {
	config := DefaultSBandConfig("W1ABC")
	if config.DataRate != 500 {
		t.Errorf("Default DataRate = %v, want 500", config.DataRate)
	}

	cla, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if cla.config.DataRate != 500 {
		t.Errorf("CLA DataRate = %v, want 500", cla.config.DataRate)
	}
}

func TestTransmitPowerConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		txPower float64
	}{
		{
			name:    "5W transmit power",
			txPower: 5.0,
		},
		{
			name:    "7W transmit power",
			txPower: 7.0,
		},
		{
			name:    "10W transmit power",
			txPower: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSBandConfig("W1ABC")
			config.TXPower = tt.txPower
			cla, err := New(config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if cla.config.TXPower != tt.txPower {
				t.Errorf("TXPower = %v, want %v", cla.config.TXPower, tt.txPower)
			}
		})
	}
}
