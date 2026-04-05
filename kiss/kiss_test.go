package kiss

import (
	"bytes"
	"testing"
)

// --- KISS Encode ---

func TestEncodeBasic(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	frame, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Expected: FEND + CMD(0x00) + data + FEND
	expected := []byte{FEND, CmdDataFrame, 0x01, 0x02, 0x03, FEND}
	if !bytes.Equal(frame, expected) {
		t.Errorf("got %v, want %v", frame, expected)
	}
}

func TestEncodeEscapesFEND(t *testing.T) {
	data := []byte{0xAA, FEND, 0xBB}
	frame, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	expected := []byte{FEND, CmdDataFrame, 0xAA, FESC, TFEND, 0xBB, FEND}
	if !bytes.Equal(frame, expected) {
		t.Errorf("got %v, want %v", frame, expected)
	}
}

func TestEncodeEscapesFESC(t *testing.T) {
	data := []byte{0xAA, FESC, 0xBB}
	frame, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	expected := []byte{FEND, CmdDataFrame, 0xAA, FESC, TFESC, 0xBB, FEND}
	if !bytes.Equal(frame, expected) {
		t.Errorf("got %v, want %v", frame, expected)
	}
}

func TestEncodeEmpty(t *testing.T) {
	_, err := Encode(nil)
	if err == nil {
		t.Error("expected error for empty data")
	}
	_, err = Encode([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestEncodeBothSpecialBytes(t *testing.T) {
	data := []byte{FEND, FESC}
	frame, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	expected := []byte{FEND, CmdDataFrame, FESC, TFEND, FESC, TFESC, FEND}
	if !bytes.Equal(frame, expected) {
		t.Errorf("got %v, want %v", frame, expected)
	}
}

// --- KISS Decode ---

func TestDecodeBasic(t *testing.T) {
	frame := []byte{FEND, CmdDataFrame, 0x01, 0x02, 0x03, FEND}
	data, err := Decode(frame)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	expected := []byte{0x01, 0x02, 0x03}
	if !bytes.Equal(data, expected) {
		t.Errorf("got %v, want %v", data, expected)
	}
}

func TestDecodeUnescapesFEND(t *testing.T) {
	frame := []byte{FEND, CmdDataFrame, 0xAA, FESC, TFEND, 0xBB, FEND}
	data, err := Decode(frame)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	expected := []byte{0xAA, FEND, 0xBB}
	if !bytes.Equal(data, expected) {
		t.Errorf("got %v, want %v", data, expected)
	}
}

func TestDecodeUnescapesFESC(t *testing.T) {
	frame := []byte{FEND, CmdDataFrame, 0xAA, FESC, TFESC, 0xBB, FEND}
	data, err := Decode(frame)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	expected := []byte{0xAA, FESC, 0xBB}
	if !bytes.Equal(data, expected) {
		t.Errorf("got %v, want %v", data, expected)
	}
}

func TestDecodeInvalidEscape(t *testing.T) {
	frame := []byte{FEND, CmdDataFrame, FESC, 0x42, FEND}
	_, err := Decode(frame)
	if err == nil {
		t.Error("expected error for invalid escape sequence")
	}
}

func TestDecodeTrailingFESC(t *testing.T) {
	frame := []byte{FEND, CmdDataFrame, 0x01, FESC, FEND}
	_, err := Decode(frame)
	if err == nil {
		t.Error("expected error for trailing FESC")
	}
}

func TestDecodeTooShort(t *testing.T) {
	_, err := Decode([]byte{FEND, FEND})
	if err == nil {
		t.Error("expected error for frame too short")
	}
}

// --- Round-trip: Encode then Decode ---

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testCases := [][]byte{
		{0x01, 0x02, 0x03},
		{FEND},
		{FESC},
		{FEND, FESC, FEND, FESC},
		{0x00, 0xFF, FEND, 0x42, FESC, 0x99},
		bytes.Repeat([]byte{0xAB}, 256),
	}

	for i, original := range testCases {
		frame, err := Encode(original)
		if err != nil {
			t.Fatalf("case %d: Encode: %v", i, err)
		}
		decoded, err := Decode(frame)
		if err != nil {
			t.Fatalf("case %d: Decode: %v", i, err)
		}
		if !bytes.Equal(decoded, original) {
			t.Errorf("case %d: round-trip mismatch: got %v, want %v", i, decoded, original)
		}
	}
}

// --- readKISSFrame ---

func TestReadKISSFrame(t *testing.T) {
	// Simulate a serial stream: garbage + FEND + data + FEND + trailing
	stream := []byte{0xFF, 0xFF, FEND, CmdDataFrame, 0x01, 0x02, FEND, 0xAA}
	reader := bytes.NewReader(stream)

	frame, err := readKISSFrame(reader)
	if err != nil {
		t.Fatalf("readKISSFrame: %v", err)
	}

	expected := []byte{FEND, CmdDataFrame, 0x01, 0x02, FEND}
	if !bytes.Equal(frame, expected) {
		t.Errorf("got %v, want %v", frame, expected)
	}
}

func TestReadKISSFrameMultipleFENDs(t *testing.T) {
	// Multiple leading FENDs (common in KISS streams) followed by a valid frame
	stream := []byte{FEND, FEND, FEND, CmdDataFrame, 0xAA, FEND}
	reader := bytes.NewReader(stream)

	frame, err := readKISSFrame(reader)
	if err != nil {
		t.Fatalf("readKISSFrame: %v", err)
	}

	data, err := Decode(frame)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	expected := []byte{0xAA}
	if !bytes.Equal(data, expected) {
		t.Errorf("got %v, want %v", data, expected)
	}
}

func TestReadKISSFrameEOF(t *testing.T) {
	// Stream ends before closing FEND
	stream := []byte{FEND, CmdDataFrame, 0x01}
	reader := bytes.NewReader(stream)

	_, err := readKISSFrame(reader)
	if err == nil {
		t.Error("expected error for EOF before closing FEND")
	}
}
