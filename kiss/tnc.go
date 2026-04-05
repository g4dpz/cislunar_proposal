package kiss

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"go.bug.st/serial"
)

// TNCConfig holds configuration for the TNC serial connection.
type TNCConfig struct {
	Device   string        // USB serial device path, e.g. "/dev/ttyACM0"
	BaudRate int           // Serial baud rate (9600 for TNC4)
	Timeout  time.Duration // Read timeout for Receive channel (0 = block forever)
}

// TNC represents a connection to a KISS TNC over USB serial.
// Receive runs in a dedicated goroutine, delivering frames via a channel.
// Send writes directly to the port with a write-only mutex.
type TNC struct {
	config  TNCConfig
	port    serial.Port
	rxCh    chan []byte   // received AX.25 frames
	errCh   chan error    // receive errors
	writeMu sync.Mutex   // protects writes only
	closed  atomic.Bool
	done    chan struct{} // signals receive goroutine exit
}

// Open opens the USB serial connection to the TNC and starts the
// background receive goroutine.
func Open(config TNCConfig) (*TNC, error) {
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(config.Device, mode)
	if err != nil {
		return nil, fmt.Errorf("kiss: failed to open %s: %w", config.Device, err)
	}

	t := &TNC{
		config: config,
		port:   port,
		rxCh:   make(chan []byte, 16),
		errCh:  make(chan error, 16),
		done:   make(chan struct{}),
	}

	go t.receiveLoop()

	return t, nil
}

// Send transmits an AX.25 frame through the TNC using KISS framing.
func (t *TNC) Send(ax25Frame []byte) error {
	if t.closed.Load() {
		return ErrPortClosed
	}

	kissFrame, err := Encode(ax25Frame)
	if err != nil {
		return err
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if t.closed.Load() {
		return ErrPortClosed
	}

	_, err = t.port.Write(kissFrame)
	if err != nil {
		return fmt.Errorf("kiss: write failed: %w", err)
	}

	return nil
}

// Receive returns the next received AX.25 frame.
// Blocks until a frame is available, the TNC is closed, or the timeout expires.
func (t *TNC) Receive() ([]byte, error) {
	if t.config.Timeout > 0 {
		timer := time.NewTimer(t.config.Timeout)
		defer timer.Stop()
		select {
		case frame := <-t.rxCh:
			return frame, nil
		case err := <-t.errCh:
			return nil, err
		case <-timer.C:
			return nil, ErrReadTimeout
		case <-t.done:
			return nil, ErrPortClosed
		}
	}

	select {
	case frame := <-t.rxCh:
		return frame, nil
	case err := <-t.errCh:
		return nil, err
	case <-t.done:
		return nil, ErrPortClosed
	}
}

// Close closes the serial connection and stops the receive goroutine.
// Safe to call multiple times and from any goroutine.
func (t *TNC) Close() error {
	if t.closed.Swap(true) {
		return nil // already closed
	}
	// Close the port — this will unblock the receive goroutine's read
	err := t.port.Close()
	// Wait for receive goroutine to exit
	<-t.done
	return err
}

// IsOpen returns true if the TNC connection is open.
func (t *TNC) IsOpen() bool {
	return !t.closed.Load()
}

// receiveLoop runs in a background goroutine, reading KISS frames from the
// serial port and delivering decoded AX.25 frames to the rxCh channel.
func (t *TNC) receiveLoop() {
	defer close(t.done)

	reader := bufio.NewReader(t.port)

	for {
		if t.closed.Load() {
			return
		}

		frame, err := readKISSFrameFromReader(reader)
		if err != nil {
			if t.closed.Load() {
				return // port closed, expected
			}
			// Non-blocking send to error channel
			select {
			case t.errCh <- err:
			default:
			}
			continue
		}

		data, err := Decode(frame)
		if err != nil {
			select {
			case t.errCh <- err:
			default:
			}
			continue
		}

		// Non-blocking send to frame channel
		select {
		case t.rxCh <- data:
		default:
			// Channel full — drop oldest frame to make room
			select {
			case <-t.rxCh:
			default:
			}
			t.rxCh <- data
		}
	}
}

// readKISSFrameFromReader reads bytes from a buffered reader until a complete
// KISS frame (FEND ... FEND) is received. Returns the raw KISS frame including FENDs.
// Skips consecutive FENDs (inter-frame fill) per KISS spec.
func readKISSFrameFromReader(reader *bufio.Reader) ([]byte, error) {
	// Skip bytes until we find a FEND (frame start marker)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("kiss: read error waiting for frame start: %w", err)
		}
		if b == FEND {
			break
		}
	}

	// Skip any additional consecutive FENDs (inter-frame fill)
	var firstDataByte byte
	foundData := false
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("kiss: read error skipping inter-frame FENDs: %w", err)
		}
		if b != FEND {
			firstDataByte = b
			foundData = true
			break
		}
	}

	if !foundData {
		return nil, ErrFrameEmpty
	}

	// Read until we hit the closing FEND
	buf := []byte{FEND, firstDataByte}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("kiss: read error during frame: %w", err)
		}
		buf = append(buf, b)
		if b == FEND {
			return buf, nil
		}
	}
}

// readKISSFrame reads from an io.Reader (used by tests).
func readKISSFrame(r io.Reader) ([]byte, error) {
	return readKISSFrameFromReader(bufio.NewReader(r))
}
