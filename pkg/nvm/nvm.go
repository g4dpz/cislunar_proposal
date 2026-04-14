// Package nvm provides an interface to external non-volatile memory (NVM)
// for persistent bundle storage on STM32U585-based nodes.
//
// Supports SPI/QSPI flash (64-256 MB) with atomic store/delete and CRC validation.
package nvm

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"sync"
)

// NVM represents an external NVM device
type NVM struct {
	mu       sync.Mutex
	config   Config
	storage  map[string][]byte // Simulated flash storage
	capacity uint64
	used     uint64
	open     bool
}

// Config holds NVM configuration
type Config struct {
	// Device path (e.g., "/dev/spidev0.0" for SPI, or memory-mapped address)
	Device string

	// Capacity in bytes (64 MB to 256 MB typical)
	Capacity uint64

	// SPI/QSPI mode
	UseQSPI bool

	// Clock speed in Hz
	ClockSpeed uint32

	// Page size for write operations (typically 256 bytes)
	PageSize uint32

	// Sector size for erase operations (typically 4 KB)
	SectorSize uint32
}

// DefaultConfig returns default NVM configuration for STM32U585
func DefaultConfig() Config {
	return Config{
		Device:     "/dev/spidev0.0",
		Capacity:   128 * 1024 * 1024, // 128 MB
		UseQSPI:    true,
		ClockSpeed: 50000000, // 50 MHz
		PageSize:   256,
		SectorSize: 4096,
	}
}

// New creates a new NVM interface
func New(config Config) (*NVM, error) {
	return &NVM{
		config:  config,
		storage: make(map[string][]byte),
		capacity: config.Capacity,
	}, nil
}

// Open initializes the NVM device
func (n *NVM) Open() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.open {
		return fmt.Errorf("NVM already open")
	}

	// In real implementation:
	// - Open SPI/QSPI device
	// - Initialize flash chip (read ID, check status)
	// - Configure QSPI mode if enabled
	// For simulation, just mark as open

	mode := "SPI"
	if n.config.UseQSPI {
		mode = "QSPI"
	}
	fmt.Printf("NVM: Opened %s device %s (%.1f MB, %s mode)\n",
		mode, n.config.Device, float64(n.config.Capacity)/(1024*1024), mode)

	n.open = true
	return nil
}

// Close shuts down the NVM device
func (n *NVM) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return fmt.Errorf("NVM not open")
	}

	n.open = false
	fmt.Println("NVM: Closed")
	return nil
}

// Write writes data to NVM with atomic operation and CRC
func (n *NVM) Write(key string, data []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return fmt.Errorf("NVM not open")
	}

	// Check capacity
	newSize := uint64(len(data))
	if existing, ok := n.storage[key]; ok {
		newSize -= uint64(len(existing))
	}
	if n.used+newSize > n.capacity {
		return fmt.Errorf("NVM capacity exceeded")
	}

	// Calculate CRC
	crc := crc32.ChecksumIEEE(data)

	// Create entry with CRC header
	entry := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(entry[0:4], crc)
	copy(entry[4:], data)

	// Atomic write (in real implementation, this would use flash write operations)
	n.storage[key] = entry
	n.used += newSize

	fmt.Printf("NVM: Wrote %s (%d bytes, CRC=0x%08x)\n", key, len(data), crc)
	return nil
}

// Read reads data from NVM with CRC validation
func (n *NVM) Read(key string) ([]byte, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return nil, fmt.Errorf("NVM not open")
	}

	entry, ok := n.storage[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if len(entry) < 4 {
		return nil, fmt.Errorf("corrupted entry: too short")
	}

	// Extract CRC and data
	storedCRC := binary.LittleEndian.Uint32(entry[0:4])
	data := entry[4:]

	// Validate CRC
	computedCRC := crc32.ChecksumIEEE(data)
	if storedCRC != computedCRC {
		return nil, fmt.Errorf("CRC mismatch: stored=0x%08x computed=0x%08x", storedCRC, computedCRC)
	}

	fmt.Printf("NVM: Read %s (%d bytes, CRC OK)\n", key, len(data))
	return data, nil
}

// Delete removes data from NVM
func (n *NVM) Delete(key string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return fmt.Errorf("NVM not open")
	}

	entry, ok := n.storage[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	n.used -= uint64(len(entry))
	delete(n.storage, key)

	fmt.Printf("NVM: Deleted %s\n", key)
	return nil
}

// List returns all keys in NVM
func (n *NVM) List() ([]string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return nil, fmt.Errorf("NVM not open")
	}

	keys := make([]string, 0, len(n.storage))
	for key := range n.storage {
		keys = append(keys, key)
	}

	return keys, nil
}

// Capacity returns total and used capacity
func (n *NVM) Capacity() (total, used uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.capacity, n.used
}

// Sync ensures all writes are flushed to NVM
func (n *NVM) Sync() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return fmt.Errorf("NVM not open")
	}

	// In real implementation, this would ensure all pending writes are complete
	fmt.Println("NVM: Synced")
	return nil
}

// Validate checks integrity of all stored data
func (n *NVM) Validate() ([]string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.open {
		return nil, fmt.Errorf("NVM not open")
	}

	corrupted := make([]string, 0)

	for key, entry := range n.storage {
		if len(entry) < 4 {
			corrupted = append(corrupted, key)
			continue
		}

		storedCRC := binary.LittleEndian.Uint32(entry[0:4])
		data := entry[4:]
		computedCRC := crc32.ChecksumIEEE(data)

		if storedCRC != computedCRC {
			corrupted = append(corrupted, key)
		}
	}

	if len(corrupted) > 0 {
		fmt.Printf("NVM: Validation found %d corrupted entries\n", len(corrupted))
	} else {
		fmt.Println("NVM: Validation passed")
	}

	return corrupted, nil
}

// IsOpen returns whether the NVM is open
func (n *NVM) IsOpen() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.open
}
