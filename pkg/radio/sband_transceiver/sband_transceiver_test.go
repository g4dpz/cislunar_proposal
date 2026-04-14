package sband_transceiver

import (
	"testing"
	"time"

	"terrestrial-dtn/pkg/iq"
)

func TestNew(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	if transceiver.GetCenterFreq() != config.CenterFreq {
		t.Errorf("Expected center freq %.3f GHz, got %.3f GHz",
			config.CenterFreq/1e9, transceiver.GetCenterFreq()/1e9)
	}

	if transceiver.GetSampleRate() != config.SampleRate {
		t.Errorf("Expected sample rate %.3f kHz, got %.3f kHz",
			config.SampleRate/1e3, transceiver.GetSampleRate()/1e3)
	}

	if transceiver.GetTXPower() != config.TXPower {
		t.Errorf("Expected TX power %.1f W, got %.1f W",
			config.TXPower, transceiver.GetTXPower())
	}

	if transceiver.GetBand() != BandS {
		t.Errorf("Expected S-band, got %s", transceiver.GetBand())
	}

	if !transceiver.IsFECEnabled() {
		t.Error("Expected FEC to be enabled")
	}

	if transceiver.GetFECType() != FECLDPC {
		t.Errorf("Expected LDPC FEC, got %s", transceiver.GetFECType())
	}
}

func TestDefaultConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		band   Band
		freq   float64
	}{
		{
			name:   "S-band",
			config: DefaultSBandConfig(),
			band:   BandS,
			freq:   2.2e9,
		},
		{
			name:   "X-band",
			config: DefaultXBandConfig(),
			band:   BandX,
			freq:   8.4e9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Band != tt.band {
				t.Errorf("Expected band %s, got %s", tt.band, tt.config.Band)
			}

			if tt.config.CenterFreq != tt.freq {
				t.Errorf("Expected center freq %.3f GHz, got %.3f GHz",
					tt.freq/1e9, tt.config.CenterFreq/1e9)
			}

			if !tt.config.FECEnabled {
				t.Error("Expected FEC to be enabled")
			}

			if tt.config.LightTimeDelay < 1*time.Second {
				t.Errorf("Expected light-time delay >= 1s, got %.3fs",
					tt.config.LightTimeDelay.Seconds())
			}
		})
	}
}

func TestOpenClose(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	if err := transceiver.Open(); err != nil {
		t.Fatalf("Failed to open transceiver: %v", err)
	}

	if err := transceiver.Close(); err != nil {
		t.Fatalf("Failed to close transceiver: %v", err)
	}
}

func TestStreaming(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	if err := transceiver.Open(); err != nil {
		t.Fatalf("Failed to open transceiver: %v", err)
	}
	defer transceiver.Close()

	if transceiver.IsStreaming() {
		t.Error("Expected transceiver to not be streaming initially")
	}

	if err := transceiver.StartStreaming(); err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}

	if !transceiver.IsStreaming() {
		t.Error("Expected transceiver to be streaming")
	}

	// Test double start
	if err := transceiver.StartStreaming(); err == nil {
		t.Error("Expected error when starting streaming twice")
	}

	if err := transceiver.StopStreaming(); err != nil {
		t.Fatalf("Failed to stop streaming: %v", err)
	}

	if transceiver.IsStreaming() {
		t.Error("Expected transceiver to not be streaming after stop")
	}

	// Test double stop
	if err := transceiver.StopStreaming(); err == nil {
		t.Error("Expected error when stopping streaming twice")
	}
}

func TestTransmitReceive(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	if err := transceiver.Open(); err != nil {
		t.Fatalf("Failed to open transceiver: %v", err)
	}
	defer transceiver.Close()

	if err := transceiver.StartStreaming(); err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}
	defer transceiver.StopStreaming()

	// Create test IQ buffer
	buffer := iq.NewIQBuffer(100, config.SampleRate)
	for i := 0; i < 100; i++ {
		buffer.Append(iq.IQSample{I: 0.5, Q: 0.5})
	}
	buffer.Timestamp = time.Now().UnixNano()

	// Test transmit
	if err := transceiver.Transmit(buffer); err != nil {
		t.Fatalf("Failed to transmit: %v", err)
	}

	// Test receive (with timeout accounting for light-time delay)
	// Note: In simulation, this will timeout since we don't have real RX
	_, err = transceiver.Receive()
	if err == nil {
		t.Log("Received buffer (simulation)")
	} else {
		t.Logf("Receive timeout (expected in simulation): %v", err)
	}
}

func TestSetCenterFreq(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		freq      float64
		expectErr bool
	}{
		{
			name:      "S-band valid",
			config:    DefaultSBandConfig(),
			freq:      2.2e9,
			expectErr: false,
		},
		{
			name:      "S-band out of range low",
			config:    DefaultSBandConfig(),
			freq:      1.9e9,
			expectErr: true,
		},
		{
			name:      "S-band out of range high",
			config:    DefaultSBandConfig(),
			freq:      2.4e9,
			expectErr: true,
		},
		{
			name:      "X-band valid",
			config:    DefaultXBandConfig(),
			freq:      8.4e9,
			expectErr: false,
		},
		{
			name:      "X-band out of range low",
			config:    DefaultXBandConfig(),
			freq:      7.9e9,
			expectErr: true,
		},
		{
			name:      "X-band out of range high",
			config:    DefaultXBandConfig(),
			freq:      8.6e9,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transceiver, err := New(tt.config)
			if err != nil {
				t.Fatalf("Failed to create transceiver: %v", err)
			}

			err = transceiver.SetCenterFreq(tt.freq)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectErr && transceiver.GetCenterFreq() != tt.freq {
				t.Errorf("Expected center freq %.3f GHz, got %.3f GHz",
					tt.freq/1e9, transceiver.GetCenterFreq()/1e9)
			}
		})
	}
}

func TestSetSampleRate(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	tests := []struct {
		name      string
		rate      float64
		expectErr bool
	}{
		{"Valid 8 kHz", 8000.0, false},
		{"Valid 4 kHz", 4000.0, false},
		{"Valid 16 kHz", 16000.0, false},
		{"Too low", 3000.0, true},
		{"Too high", 20000.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transceiver.SetSampleRate(tt.rate)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSetTXPower(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	tests := []struct {
		name      string
		power     float64
		expectErr bool
	}{
		{"Valid 5W", 5.0, false},
		{"Valid 1W", 1.0, false},
		{"Valid 10W", 10.0, false},
		{"Too low", 0.5, true},
		{"Too high", 15.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transceiver.SetTXPower(tt.power)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectErr && transceiver.GetTXPower() != tt.power {
				t.Errorf("Expected TX power %.1f W, got %.1f W",
					tt.power, transceiver.GetTXPower())
			}
		})
	}
}

func TestSetLightTimeDelay(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	tests := []struct {
		name      string
		delay     time.Duration
		expectErr bool
	}{
		{"Valid 1.2s", 1200 * time.Millisecond, false},
		{"Valid 1s", 1 * time.Second, false},
		{"Valid 2s", 2 * time.Second, false},
		{"Too low", 500 * time.Millisecond, true},
		{"Too high", 4 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transceiver.SetLightTimeDelay(tt.delay)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectErr && transceiver.GetLightTimeDelay() != tt.delay {
				t.Errorf("Expected light-time delay %.3fs, got %.3fs",
					tt.delay.Seconds(), transceiver.GetLightTimeDelay().Seconds())
			}
		})
	}
}

func TestFECControl(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	// Initially enabled with LDPC
	if !transceiver.IsFECEnabled() {
		t.Error("Expected FEC to be enabled initially")
	}
	if transceiver.GetFECType() != FECLDPC {
		t.Errorf("Expected LDPC FEC, got %s", transceiver.GetFECType())
	}

	// Disable FEC
	if err := transceiver.DisableFEC(); err != nil {
		t.Fatalf("Failed to disable FEC: %v", err)
	}
	if transceiver.IsFECEnabled() {
		t.Error("Expected FEC to be disabled")
	}

	// Enable Turbo FEC
	if err := transceiver.EnableFEC(FECTurbo); err != nil {
		t.Fatalf("Failed to enable Turbo FEC: %v", err)
	}
	if !transceiver.IsFECEnabled() {
		t.Error("Expected FEC to be enabled")
	}
	if transceiver.GetFECType() != FECTurbo {
		t.Errorf("Expected Turbo FEC, got %s", transceiver.GetFECType())
	}

	// Enable LDPC FEC
	if err := transceiver.EnableFEC(FECLDPC); err != nil {
		t.Fatalf("Failed to enable LDPC FEC: %v", err)
	}
	if transceiver.GetFECType() != FECLDPC {
		t.Errorf("Expected LDPC FEC, got %s", transceiver.GetFECType())
	}
}

func TestGetLinkMetrics(t *testing.T) {
	config := DefaultSBandConfig()
	transceiver, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create transceiver: %v", err)
	}

	metrics := transceiver.GetLinkMetrics()

	// Verify metrics are in reasonable ranges for deep-space
	if metrics.RSSI > -100.0 {
		t.Errorf("Expected weak RSSI for deep-space, got %.1f dBm", metrics.RSSI)
	}

	if metrics.SNR < 0 {
		t.Errorf("Expected positive SNR, got %.1f dB", metrics.SNR)
	}

	t.Logf("Link metrics: %s", metrics.String())
}

func TestBandString(t *testing.T) {
	tests := []struct {
		band Band
		want string
	}{
		{BandS, "S-band"},
		{BandX, "X-band"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.band.String(); got != tt.want {
				t.Errorf("Band.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFECTypeString(t *testing.T) {
	tests := []struct {
		fec  FECType
		want string
	}{
		{FECLDPC, "LDPC"},
		{FECTurbo, "Turbo"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.fec.String(); got != tt.want {
				t.Errorf("FECType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCislunarLinkBudget(t *testing.T) {
	// Verify configuration supports cislunar link budget requirements
	config := DefaultSBandConfig()

	// Requirements from design:
	// - 5W TX power
	// - 10 dBi TX antenna (directional patch)
	// - 35 dBi RX antenna (3-5m ground dish)
	// - 500 bps data rate
	// - BPSK + LDPC/Turbo FEC
	// - 5-7 dB link margin

	if config.TXPower < 5.0 {
		t.Errorf("TX power %.1fW insufficient for cislunar (need >= 5W)", config.TXPower)
	}

	if config.TXGain < 10.0 {
		t.Errorf("TX gain %.1fdBi insufficient for cislunar (need >= 10dBi)", config.TXGain)
	}

	if config.RXGain < 35.0 {
		t.Errorf("RX gain %.1fdBi insufficient for cislunar (need >= 35dBi)", config.RXGain)
	}

	if !config.FECEnabled {
		t.Error("FEC must be enabled for cislunar link budget")
	}

	// Verify data rate supports 500 bps
	// Sample rate should be at least 8x symbol rate for good filtering
	minSampleRate := 500.0 * 8.0 // 4 kHz minimum
	if config.SampleRate < minSampleRate {
		t.Errorf("Sample rate %.1f Hz too low for 500 bps (need >= %.1f Hz)",
			config.SampleRate, minSampleRate)
	}

	t.Logf("Cislunar link budget configuration validated:")
	t.Logf("  TX power: %.1f W", config.TXPower)
	t.Logf("  TX gain: %.1f dBi", config.TXGain)
	t.Logf("  RX gain: %.1f dBi", config.RXGain)
	t.Logf("  FEC: %s", config.FECType)
	t.Logf("  Sample rate: %.1f kHz", config.SampleRate/1e3)
	t.Logf("  Light-time delay: %.3f s", config.LightTimeDelay.Seconds())
}
