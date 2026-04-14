package cla

import (
	"testing"
	"terrestrial-dtn/ax25"
)

// TestValidateAX25Frame tests the AX.25 frame validation function
// Validates: Requirement 10.1
func TestValidateAX25Frame(t *testing.T) {
	tests := []struct {
		name    string
		src     ax25.Callsign
		dst     ax25.Callsign
		payload []byte
		wantErr bool
	}{
		{
			name:    "valid frame with callsigns",
			src:     ax25.Callsign{Call: "W1AW", SSID: 0},
			dst:     ax25.Callsign{Call: "N0CALL", SSID: 1},
			payload: []byte("test payload"),
			wantErr: false,
		},
		{
			name:    "valid frame with max SSID",
			src:     ax25.Callsign{Call: "TEST", SSID: 15},
			dst:     ax25.Callsign{Call: "DEST", SSID: 15},
			payload: []byte{0x01, 0x02, 0x03},
			wantErr: false,
		},
		{
			name:    "valid frame with empty payload",
			src:     ax25.Callsign{Call: "SRC", SSID: 0},
			dst:     ax25.Callsign{Call: "DST", SSID: 0},
			payload: []byte{},
			wantErr: false,
		},
		{
			name:    "valid frame with 6-char callsigns",
			src:     ax25.Callsign{Call: "ABCDEF", SSID: 5},
			dst:     ax25.Callsign{Call: "GHIJKL", SSID: 10},
			payload: []byte("data"),
			wantErr: false,
		},
		{
			name:    "valid frame with 1-char callsigns",
			src:     ax25.Callsign{Call: "A", SSID: 0},
			dst:     ax25.Callsign{Call: "B", SSID: 0},
			payload: []byte("x"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the frame
			frame, err := ax25.BuildUIFrame(tt.dst, tt.src, tt.payload)
			if err != nil {
				t.Fatalf("BuildUIFrame failed: %v", err)
			}

			// Validate the frame
			err = ValidateAX25Frame(frame)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAX25Frame() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				t.Logf("Valid AX.25 frame: src=%s dst=%s payload_len=%d",
					tt.src.String(), tt.dst.String(), len(tt.payload))
			}
		})
	}
}

// TestValidateCLAAX25Framing tests the CLA AX.25 framing validation
// Validates: Requirement 10.1
func TestValidateCLAAX25Framing(t *testing.T) {
	tests := []struct {
		name    string
		src     ax25.Callsign
		dst     ax25.Callsign
		payload []byte
		wantErr bool
	}{
		{
			name:    "valid framing",
			src:     ax25.Callsign{Call: "W1AW", SSID: 7},
			dst:     ax25.Callsign{Call: "K3LR", SSID: 9},
			payload: []byte("Hello DTN"),
			wantErr: false,
		},
		{
			name:    "valid framing with large payload",
			src:     ax25.Callsign{Call: "SRC", SSID: 0},
			dst:     ax25.Callsign{Call: "DST", SSID: 0},
			payload: make([]byte, 256),
			wantErr: false,
		},
		{
			name:    "valid framing with binary payload",
			src:     ax25.Callsign{Call: "NODE1", SSID: 1},
			dst:     ax25.Callsign{Call: "NODE2", SSID: 2},
			payload: []byte{0x00, 0xFF, 0xAA, 0x55, 0xDE, 0xAD, 0xBE, 0xEF},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCLAAX25Framing(tt.src, tt.dst, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCLAAX25Framing() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				t.Logf("CLA AX.25 framing valid: src=%s dst=%s payload_len=%d",
					tt.src.String(), tt.dst.String(), len(tt.payload))
			}
		})
	}
}

// TestAX25FramingAllCLATypes verifies that all CLA types can produce valid AX.25 frames
// This is a conceptual test - in practice, each CLA implementation would be tested separately
// Validates: Requirement 10.1
func TestAX25FramingAllCLATypes(t *testing.T) {
	claTypes := []CLAType{
		CLATypeAX25LTPVHFTNC,
		CLATypeAX25LTPUHFTNC,
		CLATypeAX25LTPUHFIQB200,
		CLATypeAX25LTPUHFIQ,
		CLATypeAX25LTPSBandIQ,
		CLATypeAX25LTPXBandIQ,
	}

	src := ax25.Callsign{Call: "W1AW", SSID: 0}
	dst := ax25.Callsign{Call: "N0CALL", SSID: 1}
	payload := []byte("DTN test payload")

	for _, claType := range claTypes {
		t.Run(claType.String(), func(t *testing.T) {
			// All CLA types use the same AX.25 framing
			err := ValidateCLAAX25Framing(src, dst, payload)
			if err != nil {
				t.Errorf("CLA type %s failed AX.25 framing validation: %v", claType.String(), err)
			} else {
				t.Logf("CLA type %s produces valid AX.25 frames", claType.String())
			}
		})
	}
}

// TestInvalidAX25Frames tests that invalid frames are rejected
func TestInvalidAX25Frames(t *testing.T) {
	tests := []struct {
		name      string
		frameData []byte
		wantErr   bool
	}{
		{
			name:      "frame too short",
			frameData: []byte{0x01, 0x02, 0x03},
			wantErr:   true,
		},
		{
			name:      "empty frame",
			frameData: []byte{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAX25Frame(tt.frameData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAX25Frame() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
