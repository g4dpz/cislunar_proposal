// Package ax25 implements AX.25 UI frame construction and parsing
// for amateur radio DTN use. Callsigns are encoded per the AX.25 spec:
// 6 characters space-padded, each byte left-shifted by 1 bit.
// SSID byte format: 0b0SSSS0RR where bits 5-1 = SSID value.
//
// Frame structure:
//   [Destination Address (7 bytes)] [Source Address (7 bytes)]
//   [Control (1 byte = 0x03)] [PID (1 byte = 0xF0)]
//   [Information field (variable)]
//
// No FCS computation — the TNC handles that.
package ax25

import (
	"errors"
	"fmt"
)

const (
	// ControlUI is the control field value for UI (Unnumbered Information) frames.
	ControlUI = 0x03
	// PIDNoLayer3 indicates no layer 3 protocol.
	PIDNoLayer3 = 0xF0
	// CallsignLen is the fixed length of a callsign in the address field.
	CallsignLen = 6
	// AddressFieldLen is the length of a single address field (callsign + SSID byte).
	AddressFieldLen = 7
	// HeaderLen is the total header length: dest(7) + src(7) + control(1) + PID(1).
	HeaderLen = 2*AddressFieldLen + 2
	// MaxSSID is the maximum SSID value.
	MaxSSID = 15
)

// Callsign represents an amateur radio callsign with SSID.
type Callsign struct {
	Call string // e.g. "W1AW", max 6 characters
	SSID uint8  // 0-15
}

// Frame represents a parsed AX.25 UI frame.
type Frame struct {
	Dst  Callsign
	Src  Callsign
	Info []byte // information field payload
}

// String returns the callsign in standard amateur radio format: CALL-SSID.
func (cs Callsign) String() string {
	return fmt.Sprintf("%s-%d", cs.Call, cs.SSID)
}

// ParseCallsign parses a "CALL-SSID" string into a Callsign.
// If no dash is present, SSID defaults to 0.
// Examples: "W1AW-7", "N0CALL-0", "AB1CD"
func ParseCallsign(s string) (Callsign, error) {
	if s == "" {
		return Callsign{}, ErrEmptyCallsign
	}

	dashIdx := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '-' {
			dashIdx = i
			break
		}
	}

	if dashIdx < 0 {
		// No dash — SSID defaults to 0
		cs := Callsign{Call: s, SSID: 0}
		return cs, validateCallsign(cs)
	}

	call := s[:dashIdx]
	ssidStr := s[dashIdx+1:]

	var ssid uint8
	for _, c := range ssidStr {
		if c < '0' || c > '9' {
			return Callsign{}, fmt.Errorf("ax25: invalid SSID in %q: not a number", s)
		}
		ssid = ssid*10 + uint8(c-'0')
	}

	cs := Callsign{Call: call, SSID: ssid}
	return cs, validateCallsign(cs)
}

var (
	ErrEmptyCallsign   = errors.New("ax25: callsign must not be empty")
	ErrCallsignTooLong = errors.New("ax25: callsign must be at most 6 characters")
	ErrSSIDOutOfRange  = errors.New("ax25: SSID must be 0-15")
	ErrFrameTooShort   = errors.New("ax25: frame too short to contain header")
	ErrBadControl      = errors.New("ax25: unexpected control field (expected UI 0x03)")
	ErrBadPID          = errors.New("ax25: unexpected PID field (expected 0xF0)")
)

// validateCallsign checks that a callsign is valid for encoding.
func validateCallsign(cs Callsign) error {
	if len(cs.Call) == 0 {
		return ErrEmptyCallsign
	}
	if len(cs.Call) > CallsignLen {
		return ErrCallsignTooLong
	}
	if cs.SSID > MaxSSID {
		return ErrSSIDOutOfRange
	}
	return nil
}

// encodeAddress encodes a callsign into a 7-byte AX.25 address field.
// The callsign is space-padded to 6 characters, each byte left-shifted by 1.
// The SSID byte is: 0b0SSSS0E where bits 5-1 = SSID, bit 0 = extension bit.
func encodeAddress(cs Callsign, lastAddress bool) ([AddressFieldLen]byte, error) {
	var addr [AddressFieldLen]byte
	if err := validateCallsign(cs); err != nil {
		return addr, err
	}

	// Pad callsign to 6 characters with spaces, then left-shift each byte by 1.
	padded := fmt.Sprintf("%-6s", cs.Call)
	for i := 0; i < CallsignLen; i++ {
		addr[i] = padded[i] << 1
	}

	// SSID byte: bits 6,5 reserved (set to 1 per convention), bits 4-1 = SSID, bit 0 = extension.
	// Format: 0b011SSSS0 with extension bit set if this is the last address.
	ssidByte := byte(0x60) | (cs.SSID << 1)
	if lastAddress {
		ssidByte |= 0x01 // set extension bit to mark end of address fields
	}
	addr[CallsignLen] = ssidByte

	return addr, nil
}

// decodeAddress decodes a 7-byte AX.25 address field into a Callsign.
func decodeAddress(data []byte) (Callsign, error) {
	if len(data) < AddressFieldLen {
		return Callsign{}, fmt.Errorf("ax25: address field too short: need %d bytes, got %d", AddressFieldLen, len(data))
	}

	// Right-shift each byte by 1 to recover the ASCII character, then trim trailing spaces.
	var raw [CallsignLen]byte
	for i := 0; i < CallsignLen; i++ {
		raw[i] = data[i] >> 1
	}

	// Trim trailing spaces.
	callLen := CallsignLen
	for callLen > 0 && raw[callLen-1] == ' ' {
		callLen--
	}

	ssid := (data[CallsignLen] >> 1) & 0x0F

	return Callsign{
		Call: string(raw[:callLen]),
		SSID: ssid,
	}, nil
}

// BuildUIFrame constructs an AX.25 UI frame from source/destination callsigns and payload.
func BuildUIFrame(dst, src Callsign, info []byte) ([]byte, error) {
	dstAddr, err := encodeAddress(dst, false)
	if err != nil {
		return nil, fmt.Errorf("ax25: invalid destination: %w", err)
	}
	srcAddr, err := encodeAddress(src, true)
	if err != nil {
		return nil, fmt.Errorf("ax25: invalid source: %w", err)
	}

	frame := make([]byte, 0, HeaderLen+len(info))
	frame = append(frame, dstAddr[:]...)
	frame = append(frame, srcAddr[:]...)
	frame = append(frame, ControlUI)
	frame = append(frame, PIDNoLayer3)
	frame = append(frame, info...)

	return frame, nil
}

// ParseFrame parses an AX.25 UI frame, extracting destination/source callsigns
// and the information field.
func ParseFrame(data []byte) (*Frame, error) {
	if len(data) < HeaderLen {
		return nil, ErrFrameTooShort
	}

	dst, err := decodeAddress(data[0:AddressFieldLen])
	if err != nil {
		return nil, fmt.Errorf("ax25: parsing destination: %w", err)
	}

	src, err := decodeAddress(data[AddressFieldLen : 2*AddressFieldLen])
	if err != nil {
		return nil, fmt.Errorf("ax25: parsing source: %w", err)
	}

	control := data[2*AddressFieldLen]
	if control != ControlUI {
		return nil, ErrBadControl
	}

	pid := data[2*AddressFieldLen+1]
	if pid != PIDNoLayer3 {
		return nil, ErrBadPID
	}

	info := make([]byte, len(data)-HeaderLen)
	copy(info, data[HeaderLen:])

	return &Frame{
		Dst:  dst,
		Src:  src,
		Info: info,
	}, nil
}
