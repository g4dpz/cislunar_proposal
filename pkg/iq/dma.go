package iq

import (
	"fmt"
	"sync"
	"time"
)

// DMAController manages DMA-driven IQ sample streaming for STM32U585
// This is a simulation/abstraction layer for the actual STM32U585 DMA hardware
type DMAController struct {
	mu            sync.Mutex
	txBuffer      *IQBuffer
	rxBuffer      *IQBuffer
	txCallback    func(*IQBuffer) error
	rxCallback    func(*IQBuffer) error
	streaming     bool
	sampleRate    float64
	bufferSize    int
	txChan        chan *IQBuffer
	rxChan        chan *IQBuffer
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// DMAConfig holds DMA controller configuration
type DMAConfig struct {
	SampleRate float64 // Samples per second
	BufferSize int     // Number of samples per DMA transfer
	TXCallback func(*IQBuffer) error
	RXCallback func(*IQBuffer) error
}

// NewDMAController creates a new DMA controller
func NewDMAController(config DMAConfig) *DMAController {
	return &DMAController{
		sampleRate: config.SampleRate,
		bufferSize: config.BufferSize,
		txCallback: config.TXCallback,
		rxCallback: config.RXCallback,
		txChan:     make(chan *IQBuffer, 4), // Buffer 4 DMA transfers
		rxChan:     make(chan *IQBuffer, 4),
		stopChan:   make(chan struct{}),
	}
}

// Start begins DMA streaming
func (d *DMAController) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.streaming {
		return fmt.Errorf("DMA already streaming")
	}

	d.streaming = true
	d.stopChan = make(chan struct{})

	// Start TX DMA worker
	d.wg.Add(1)
	go d.txWorker()

	// Start RX DMA worker
	d.wg.Add(1)
	go d.rxWorker()

	return nil
}

// Stop halts DMA streaming
func (d *DMAController) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.streaming {
		return fmt.Errorf("DMA not streaming")
	}

	close(d.stopChan)
	d.wg.Wait()
	d.streaming = false

	return nil
}

// QueueTX queues an IQ buffer for transmission via DMA
func (d *DMAController) QueueTX(buffer *IQBuffer) error {
	select {
	case d.txChan <- buffer:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("TX queue full")
	}
}

// QueueRX queues an IQ buffer for reception via DMA
func (d *DMAController) QueueRX(buffer *IQBuffer) error {
	select {
	case d.rxChan <- buffer:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("RX queue full")
	}
}

// txWorker handles TX DMA transfers
func (d *DMAController) txWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(time.Duration(float64(d.bufferSize) / d.sampleRate * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			// Simulate DMA transfer timing
			select {
			case buffer := <-d.txChan:
				if d.txCallback != nil {
					if err := d.txCallback(buffer); err != nil {
						// Log error but continue
						fmt.Printf("TX DMA callback error: %v\n", err)
					}
				}
			default:
				// No buffer available, send silence
			}
		}
	}
}

// rxWorker handles RX DMA transfers
func (d *DMAController) rxWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(time.Duration(float64(d.bufferSize) / d.sampleRate * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			// Simulate DMA transfer timing
			buffer := NewIQBuffer(d.bufferSize, d.sampleRate)
			buffer.Timestamp = time.Now().UnixNano()

			if d.rxCallback != nil {
				if err := d.rxCallback(buffer); err != nil {
					// Log error but continue
					fmt.Printf("RX DMA callback error: %v\n", err)
				}
			}

			// Queue received buffer
			select {
			case d.rxChan <- buffer:
			default:
				// RX queue full, drop buffer
				fmt.Println("RX queue full, dropping buffer")
			}
		}
	}
}

// IsStreaming returns whether DMA is currently streaming
func (d *DMAController) IsStreaming() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.streaming
}

// GetSampleRate returns the configured sample rate
func (d *DMAController) GetSampleRate() float64 {
	return d.sampleRate
}

// GetBufferSize returns the configured buffer size
func (d *DMAController) GetBufferSize() int {
	return d.bufferSize
}
