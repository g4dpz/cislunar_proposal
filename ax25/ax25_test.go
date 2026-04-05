package ax25

import (
	"bytes"
	"testing"
)

// --- Build + Parse round-trip ---

func TestBuildAndParseRoundTrip(t *testing.T) {
	dst := Callsign{Call: "G4DPZ", SSID: 2}
	src := Callsign{Call: "G4DPZ", SSID: 1}
	payload := []byte("Hello DTN")

	frame, err := BuildUIFrame(dst, src, payload)
	if err != nil {
		t.Fatalf("BuildUIFrame: %v", err)
	}

	parsed, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}

	if parsed.Dst.Call != dst.Call || parsed.Dst.SSID != dst.SSID {
		t.Errorf("dst mismatch: got %+v, want %+v", parsed.Dst, dst)
	}
	if parsed.Src.Call != src.Call || parsed.Src.SSID != src.SSID {
		t.Errorf("src mismatch: got %+v, want %+v", parsed.Src, src)
	}
	if !bytes.Equal(parsed.Info, payload) {
		t.Errorf("info mismatch: got %q, want %q", parsed.Info, payload)
	}
}

func TestRoundTripEmptyPayload(t *testing.T) {
	dst := Callsign{Call: "AB1CD", SSID: 3}
	src := Callsign{Call: "XY9Z", SSID: 15}

	frame, err := BuildUIFrame(dst, src, nil)
	if err != nil {
		t.Fatalf("BuildUIFrame: %v", err)
	}

	parsed, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}

	if parsed.Dst.Call != dst.Call || parsed.Dst.SSID != dst.SSID {
		t.Errorf("dst mismatch: got %+v, want %+v", parsed.Dst, dst)
	}
	if parsed.Src.Call != src.Call || parsed.Src.SSID != src.SSID {
		t.Errorf("src mismatch: got %+v, want %+v", parsed.Src, src)
	}
	if len(parsed.Info) != 0 {
		t.Errorf("expected empty info, got %q", parsed.Info)
	}
}

// --- Callsign encoding: space-padded, left-shifted ---

func TestCallsignEncoding(t *testing.T) {
	cs := Callsign{Call: "W1AW", SSID: 0}
	addr, err := encodeAddress(cs, false)
	if err != nil {
		t.Fatalf("encodeAddress: %v", err)
	}

	// "W1AW  " padded to 6 chars, each byte << 1
	expected := "W1AW  "
	for i := 0; i < CallsignLen; i++ {
		want := expected[i] << 1
		if addr[i] != want {
			t.Errorf("byte %d: got 0x%02X, want 0x%02X (char %c << 1)", i, addr[i], want, expected[i])
		}
	}
}

func TestCallsignDecodingRoundTrip(t *testing.T) {
	tests := []Callsign{
		{Call: "W1AW", SSID: 0},
		{Call: "N0CALL", SSID: 15},
		{Call: "A", SSID: 5},
		{Call: "ABCDEF", SSID: 8},
	}

	for _, cs := range tests {
		addr, err := encodeAddress(cs, true)
		if err != nil {
			t.Fatalf("encodeAddress(%+v): %v", cs, err)
		}
		decoded, err := decodeAddress(addr[:])
		if err != nil {
			t.Fatalf("decodeAddress: %v", err)
		}
		if decoded.Call != cs.Call {
			t.Errorf("call mismatch: got %q, want %q", decoded.Call, cs.Call)
		}
		if decoded.SSID != cs.SSID {
			t.Errorf("SSID mismatch: got %d, want %d", decoded.SSID, cs.SSID)
		}
	}
}

// --- SSID encoding/decoding ---

func TestSSIDEncoding(t *testing.T) {
	for ssid := uint8(0); ssid <= MaxSSID; ssid++ {
		cs := Callsign{Call: "TEST", SSID: ssid}
		addr, err := encodeAddress(cs, false)
		if err != nil {
			t.Fatalf("SSID %d: encodeAddress: %v", ssid, err)
		}
		decoded, err := decodeAddress(addr[:])
		if err != nil {
			t.Fatalf("SSID %d: decodeAddress: %v", ssid, err)
		}
		if decoded.SSID != ssid {
			t.Errorf("SSID round-trip failed: got %d, want %d", decoded.SSID, ssid)
		}
	}
}

// --- Edge cases ---

func TestMaxLengthCallsign(t *testing.T) {
	cs := Callsign{Call: "ABCDEF", SSID: 0}
	frame, err := BuildUIFrame(cs, cs, []byte("test"))
	if err != nil {
		t.Fatalf("BuildUIFrame with 6-char callsign: %v", err)
	}
	parsed, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if parsed.Dst.Call != "ABCDEF" || parsed.Src.Call != "ABCDEF" {
		t.Errorf("6-char callsign mismatch: dst=%q src=%q", parsed.Dst.Call, parsed.Src.Call)
	}
}

func TestSingleCharCallsign(t *testing.T) {
	cs := Callsign{Call: "A", SSID: 0}
	frame, err := BuildUIFrame(cs, cs, []byte("x"))
	if err != nil {
		t.Fatalf("BuildUIFrame with 1-char callsign: %v", err)
	}
	parsed, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if parsed.Dst.Call != "A" || parsed.Src.Call != "A" {
		t.Errorf("1-char callsign mismatch: dst=%q src=%q", parsed.Dst.Call, parsed.Src.Call)
	}
}

func TestSSIDBoundaries(t *testing.T) {
	// SSID 0
	cs0 := Callsign{Call: "TEST", SSID: 0}
	frame, err := BuildUIFrame(cs0, cs0, nil)
	if err != nil {
		t.Fatalf("SSID 0: %v", err)
	}
	parsed, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if parsed.Dst.SSID != 0 || parsed.Src.SSID != 0 {
		t.Errorf("SSID 0 mismatch: dst=%d src=%d", parsed.Dst.SSID, parsed.Src.SSID)
	}

	// SSID 15
	cs15 := Callsign{Call: "TEST", SSID: 15}
	frame, err = BuildUIFrame(cs15, cs15, nil)
	if err != nil {
		t.Fatalf("SSID 15: %v", err)
	}
	parsed, err = ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if parsed.Dst.SSID != 15 || parsed.Src.SSID != 15 {
		t.Errorf("SSID 15 mismatch: dst=%d src=%d", parsed.Dst.SSID, parsed.Src.SSID)
	}
}

// --- Invalid inputs ---

func TestEmptyCallsign(t *testing.T) {
	empty := Callsign{Call: "", SSID: 0}
	valid := Callsign{Call: "TEST", SSID: 0}

	_, err := BuildUIFrame(empty, valid, nil)
	if err == nil {
		t.Error("expected error for empty destination callsign")
	}

	_, err = BuildUIFrame(valid, empty, nil)
	if err == nil {
		t.Error("expected error for empty source callsign")
	}
}

func TestCallsignTooLong(t *testing.T) {
	long := Callsign{Call: "ABCDEFG", SSID: 0} // 7 chars
	valid := Callsign{Call: "TEST", SSID: 0}

	_, err := BuildUIFrame(long, valid, nil)
	if err == nil {
		t.Error("expected error for callsign > 6 chars (destination)")
	}

	_, err = BuildUIFrame(valid, long, nil)
	if err == nil {
		t.Error("expected error for callsign > 6 chars (source)")
	}
}

func TestSSIDOutOfRange(t *testing.T) {
	bad := Callsign{Call: "TEST", SSID: 16}
	valid := Callsign{Call: "TEST", SSID: 0}

	_, err := BuildUIFrame(bad, valid, nil)
	if err == nil {
		t.Error("expected error for SSID > 15 (destination)")
	}

	_, err = BuildUIFrame(valid, bad, nil)
	if err == nil {
		t.Error("expected error for SSID > 15 (source)")
	}
}

func TestParseFrameTooShort(t *testing.T) {
	_, err := ParseFrame([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Error("expected error for frame too short")
	}
}

func TestParseFrameBadControl(t *testing.T) {
	// Build a valid frame, then corrupt the control byte.
	dst := Callsign{Call: "W1AW", SSID: 0}
	src := Callsign{Call: "N0CALL", SSID: 1}
	frame, err := BuildUIFrame(dst, src, []byte("test"))
	if err != nil {
		t.Fatalf("BuildUIFrame: %v", err)
	}
	frame[2*AddressFieldLen] = 0xFF // corrupt control
	_, err = ParseFrame(frame)
	if err == nil {
		t.Error("expected error for bad control field")
	}
}

func TestParseFrameBadPID(t *testing.T) {
	dst := Callsign{Call: "W1AW", SSID: 0}
	src := Callsign{Call: "N0CALL", SSID: 1}
	frame, err := BuildUIFrame(dst, src, []byte("test"))
	if err != nil {
		t.Fatalf("BuildUIFrame: %v", err)
	}
	frame[2*AddressFieldLen+1] = 0x00 // corrupt PID
	_, err = ParseFrame(frame)
	if err == nil {
		t.Error("expected error for bad PID field")
	}
}

// --- Frame structure verification ---

func TestFrameStructure(t *testing.T) {
	dst := Callsign{Call: "W1AW", SSID: 2}
	src := Callsign{Call: "K3LR", SSID: 9}
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	frame, err := BuildUIFrame(dst, src, payload)
	if err != nil {
		t.Fatalf("BuildUIFrame: %v", err)
	}

	// Total length: 7 (dst) + 7 (src) + 1 (control) + 1 (PID) + 4 (payload) = 20
	if len(frame) != HeaderLen+len(payload) {
		t.Errorf("frame length: got %d, want %d", len(frame), HeaderLen+len(payload))
	}

	// Control byte at offset 14
	if frame[14] != ControlUI {
		t.Errorf("control byte: got 0x%02X, want 0x%02X", frame[14], ControlUI)
	}

	// PID byte at offset 15
	if frame[15] != PIDNoLayer3 {
		t.Errorf("PID byte: got 0x%02X, want 0x%02X", frame[15], PIDNoLayer3)
	}

	// Payload at offset 16
	if !bytes.Equal(frame[16:], payload) {
		t.Errorf("payload mismatch in raw frame")
	}
}

// --- ParseCallsign ---

func TestParseCallsignWithSSID(t *testing.T) {
	cs, err := ParseCallsign("W1AW-7")
	if err != nil {
		t.Fatalf("ParseCallsign: %v", err)
	}
	if cs.Call != "W1AW" || cs.SSID != 7 {
		t.Errorf("got %+v, want Call=W1AW SSID=7", cs)
	}
}

func TestParseCallsignNoSSID(t *testing.T) {
	cs, err := ParseCallsign("W1AW")
	if err != nil {
		t.Fatalf("ParseCallsign: %v", err)
	}
	if cs.Call != "W1AW" || cs.SSID != 0 {
		t.Errorf("got %+v, want Call=W1AW SSID=0", cs)
	}
}

func TestParseCallsignSSID15(t *testing.T) {
	cs, err := ParseCallsign("TEST-15")
	if err != nil {
		t.Fatalf("ParseCallsign: %v", err)
	}
	if cs.Call != "TEST" || cs.SSID != 15 {
		t.Errorf("got %+v, want Call=TEST SSID=15", cs)
	}
}

func TestParseCallsignSSID0(t *testing.T) {
	cs, err := ParseCallsign("AB1CD-0")
	if err != nil {
		t.Fatalf("ParseCallsign: %v", err)
	}
	if cs.Call != "AB1CD" || cs.SSID != 0 {
		t.Errorf("got %+v, want Call=AB1CD SSID=0", cs)
	}
}

func TestParseCallsignEmpty(t *testing.T) {
	_, err := ParseCallsign("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseCallsignTooLong(t *testing.T) {
	_, err := ParseCallsign("ABCDEFG-0")
	if err == nil {
		t.Error("expected error for callsign > 6 chars")
	}
}

func TestParseCallsignSSIDTooHigh(t *testing.T) {
	_, err := ParseCallsign("TEST-16")
	if err == nil {
		t.Error("expected error for SSID > 15")
	}
}

func TestParseCallsignBadSSID(t *testing.T) {
	_, err := ParseCallsign("TEST-abc")
	if err == nil {
		t.Error("expected error for non-numeric SSID")
	}
}

func TestCallsignString(t *testing.T) {
	cs := Callsign{Call: "W1AW", SSID: 7}
	if cs.String() != "W1AW-7" {
		t.Errorf("got %q, want %q", cs.String(), "W1AW-7")
	}
}
