// Package kiss implements the KISS TNC protocol for communicating with
// a Mobilinkd TNC4 (or any KISS-compatible TNC) over a serial port.
//
// KISS frame format:
//   [FEND] [Command] [Data (with FEND/FESC escaped)] [FEND]
//
// Special bytes:
//   FEND  = 0xC0 (frame delimiter)
//   FESC  = 0xDB (escape character)
//   TFEND = 0xDC (transposed FEND: FESC TFEND = 0xC0 in data)
//   TFESC = 0xDD (transposed FESC: FESC TFESC = 0xDB in data)
//
// Command byte for data frames on port 0: 0x00
//
// Reference: KISS TNC protocol specification.
package kiss

import (
	"errors"
	"fmt"
)

const (
	FEND  = 0xC0 // Frame delimiter
	FESC  = 0xDB // Escape character
	TFEND = 0xDC // Transposed FEND
	TFESC = 0xDD // Transposed FESC

	CmdDataFrame = 0x00 // Data frame on port 0
)

var (
	ErrFrameEmpty    = errors.New("kiss: frame data is empty")
	ErrPortClosed    = errors.New("kiss: port is closed")
	ErrReadTimeout   = errors.New("kiss: read timeout")
	ErrInvalidEscape = errors.New("kiss: invalid escape sequence")
)

// Encode wraps raw AX.25 frame data into a KISS frame.
// The result is: FEND + CmdDataFrame + escaped(data) + FEND
func Encode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrFrameEmpty
	}

	// Worst case: every byte needs escaping (2x) + 3 overhead bytes
	buf := make([]byte, 0, len(data)*2+3)
	buf = append(buf, FEND)
	buf = append(buf, CmdDataFrame)

	for _, b := range data {
		switch b {
		case FEND:
			buf = append(buf, FESC, TFEND)
		case FESC:
			buf = append(buf, FESC, TFESC)
		default:
			buf = append(buf, b)
		}
	}

	buf = append(buf, FEND)
	return buf, nil
}

// Decode extracts the AX.25 frame data from a KISS frame.
// Input should be the bytes between (and including) the FEND delimiters.
// Returns the unescaped data payload (without command byte).
func Decode(frame []byte) ([]byte, error) {
	if len(frame) < 3 {
		return nil, fmt.Errorf("kiss: frame too short: %d bytes", len(frame))
	}

	// Strip leading/trailing FENDs
	start := 0
	for start < len(frame) && frame[start] == FEND {
		start++
	}
	end := len(frame) - 1
	for end > start && frame[end] == FEND {
		end--
	}

	if start > end {
		return nil, ErrFrameEmpty
	}

	// First byte after FEND is the command byte
	// cmd := frame[start]
	// We only handle data frames (cmd == 0x00) but accept any for flexibility
	dataStart := start + 1

	// Unescape the data
	buf := make([]byte, 0, end-dataStart+1)
	escaped := false
	for i := dataStart; i <= end; i++ {
		b := frame[i]
		if escaped {
			switch b {
			case TFEND:
				buf = append(buf, FEND)
			case TFESC:
				buf = append(buf, FESC)
			default:
				return nil, fmt.Errorf("%w: FESC followed by 0x%02X", ErrInvalidEscape, b)
			}
			escaped = false
		} else if b == FESC {
			escaped = true
		} else {
			buf = append(buf, b)
		}
	}

	if escaped {
		return nil, fmt.Errorf("%w: trailing FESC", ErrInvalidEscape)
	}

	return buf, nil
}
