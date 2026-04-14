package b200mini

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"terrestrial-dtn/pkg/iq"
)

// SPIBridge provides SPI/UART communication between companion host (running UHD)
// and STM32U585 OBC for IQ sample streaming
type SPIBridge struct {
	mu     sync.Mutex
	config SPIBridgeConfig
	device interface{} // SPI device handle (platform-specific)
	open   bool
}

// SPIBridgeConfig holds SPI bridge configuration
type SPIBridgeConfig struct {
	// SPI device path (e.g., "/dev/spidev0.0")
	Device string

	// SPI clock speed in Hz
	Speed uint32

	// Buffer size for IQ samples
	BufferSize int

	// Use UART instead of SPI (alternative bridge mode)
	UseUART bool

	// UART device path (e.g., "/dev/ttyUSB0")
	UARTDevice string

	// UART baud rate
	UARTBaud int
}

// NewSPIBridge creates a new SPI bridge
func NewSPIBridge(config SPIBridgeConfig) (*SPIBridge, error) {
	return &SPIBridge{
		config: config,
	}, nil
}

// Open initializes the SPI/UART bridge
func (s *SPIBridge) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.open {
		return fmt.Errorf("bridge already open")
	}

	// In real implementation, this would:
	// - Open SPI device (/dev/spidev0.0)
	// - Configure SPI mode, speed, bits per word
	// - Or open UART device if UseUART is true
	// For simulation, we just mark as open

	if s.config.UseUART {
		fmt.Printf("SPI Bridge: Opened UART %s at %d baud\n", s.config.UARTDevice, s.config.UARTBaud)
	} else {
		fmt.Printf("SPI Bridge: Opened SPI %s at %d Hz\n", s.config.Device, s.config.Speed)
	}

	s.open = true
	return nil
}

// Close shuts down the SPI/UART bridge
func (s *SPIBridge) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.open {
		return fmt.Errorf("bridge not open")
	}

	// In real implementation, close the device handle
	s.open = false
	fmt.Println("SPI Bridge: Closed")
	return nil
}

// SendIQ sends IQ samples to STM32U585 via SPI/UART
func (s *SPIBridge) SendIQ(buffer *iq.IQBuffer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.open {
		return fmt.Errorf("bridge not open")
	}

	// Serialize IQ samples to binary format
	data := s.serializeIQ(buffer)

	// In real implementation:
	// - Write data to SPI device using ioctl() or spidev library
	// - Or write to UART device
	// - STM32U585 DMA receives the data
	// For simulation, we just log the transfer

	fmt.Printf("SPI Bridge: Sent %d IQ samples (%d bytes)\n", len(buffer.Samples), len(data))
	return nil
}

// ReceiveIQ receives IQ samples from STM32U585 via SPI/UART
func (s *SPIBridge) ReceiveIQ() (*iq.IQBuffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.open {
		return nil, fmt.Errorf("bridge not open")
	}

	// In real implementation:
	// - Read data from SPI device
	// - Or read from UART device
	// - Deserialize binary data to IQ samples
	// For simulation, create empty buffer

	buffer := iq.NewIQBuffer(s.config.BufferSize, 1e6)
	fmt.Printf("SPI Bridge: Received %d IQ samples\n", s.config.BufferSize)
	return buffer, nil
}

// serializeIQ converts IQ samples to binary format for SPI/UART transfer
func (s *SPIBridge) serializeIQ(buffer *iq.IQBuffer) []byte {
	// Format: [num_samples:4][sample_rate:8][timestamp:8][I:4 Q:4]...
	headerSize := 4 + 8 + 8 // num_samples + sample_rate + timestamp
	sampleSize := 8         // 4 bytes I + 4 bytes Q (float32)
	totalSize := headerSize + len(buffer.Samples)*sampleSize

	data := make([]byte, totalSize)

	// Write header
	binary.LittleEndian.PutUint32(data[0:4], uint32(len(buffer.Samples)))
	binary.LittleEndian.PutUint64(data[4:12], uint64(buffer.SampleRate))
	binary.LittleEndian.PutUint64(data[12:20], uint64(buffer.Timestamp))

	// Write samples
	offset := headerSize
	for _, sample := range buffer.Samples {
		binary.LittleEndian.PutUint32(data[offset:offset+4], floatToUint32(float32(sample.I)))
		binary.LittleEndian.PutUint32(data[offset+4:offset+8], floatToUint32(float32(sample.Q)))
		offset += 8
	}

	return data
}

// deserializeIQ converts binary data to IQ samples
func (s *SPIBridge) deserializeIQ(data []byte) (*iq.IQBuffer, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("data too short for header")
	}

	// Read header
	numSamples := binary.LittleEndian.Uint32(data[0:4])
	sampleRate := float64(binary.LittleEndian.Uint64(data[4:12]))
	timestamp := int64(binary.LittleEndian.Uint64(data[12:20]))

	// Validate size
	expectedSize := 20 + int(numSamples)*8
	if len(data) < expectedSize {
		return nil, fmt.Errorf("data too short for samples")
	}

	// Create buffer
	buffer := iq.NewIQBuffer(int(numSamples), sampleRate)
	buffer.Timestamp = timestamp

	// Read samples
	offset := 20
	for i := 0; i < int(numSamples); i++ {
		iVal := uint32ToFloat(binary.LittleEndian.Uint32(data[offset : offset+4]))
		qVal := uint32ToFloat(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		buffer.Append(iq.IQSample{I: float64(iVal), Q: float64(qVal)})
		offset += 8
	}

	return buffer, nil
}

// floatToUint32 converts float32 to uint32 for binary serialization
func floatToUint32(f float32) uint32 {
	return math.Float32bits(f)
}

// uint32ToFloat converts uint32 to float32 for binary deserialization
func uint32ToFloat(u uint32) float32 {
	return math.Float32frombits(u)
}

// IsOpen returns whether the bridge is open
func (s *SPIBridge) IsOpen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.open
}
