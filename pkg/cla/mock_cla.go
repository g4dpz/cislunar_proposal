package cla

import (
	"fmt"
	"sync"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/contact"
)

// MockCLA is a mock implementation of ConvergenceLayerAdapter for testing
type MockCLA struct {
	claType       CLAType
	status        CLAStatus
	metrics       LinkMetrics
	isOpen        bool
	currentWindow *contact.ContactWindow
	openError     error
	mu            sync.RWMutex
}

// NewMockCLA creates a new mock CLA with default UHF IQ type
func NewMockCLA() *MockCLA {
	return &MockCLA{
		claType: CLATypeAX25LTPUHFIQ,
		status:  CLAStatusIdle,
		metrics: LinkMetrics{
			RSSI:             -80,
			SNR:              15.0,
			BitErrorRate:     0.001,
			BytesTransferred: 0,
		},
	}
}

// NewMockCLAWithType creates a new mock CLA with a specific type
func NewMockCLAWithType(claType CLAType) *MockCLA {
	return &MockCLA{
		claType: claType,
		status:  CLAStatusIdle,
		metrics: LinkMetrics{
			RSSI:             -80,
			SNR:              15.0,
			BitErrorRate:     0.001,
			BytesTransferred: 0,
		},
	}
}

// SetOpenError sets an error to be returned by Open (for testing)
func (m *MockCLA) SetOpenError(errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if errMsg == "" {
		m.openError = nil
	} else {
		m.openError = fmt.Errorf("%s", errMsg)
	}
}

// Type returns the CLA type
func (m *MockCLA) Type() CLAType {
	return m.claType
}

// Open establishes the link for a contact window
func (m *MockCLA) Open(window contact.ContactWindow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we should simulate an error
	if m.openError != nil {
		return m.openError
	}

	if m.isOpen {
		return fmt.Errorf("link already open")
	}

	m.isOpen = true
	m.currentWindow = &window
	m.status = CLAStatusIdle
	return nil
}

// Close closes the link
func (m *MockCLA) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isOpen {
		return fmt.Errorf("link not open")
	}

	m.isOpen = false
	m.currentWindow = nil
	m.status = CLAStatusIdle
	return nil
}

// SendBundle transmits a bundle over the link
func (m *MockCLA) SendBundle(bundle *bpa.Bundle) (*LinkMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isOpen {
		return nil, fmt.Errorf("link not open")
	}

	m.status = CLAStatusTransmitting

	// Simulate transmission
	bundleSize := int64(bundle.Size())
	m.metrics.BytesTransferred += bundleSize

	m.status = CLAStatusIdle

	// Return a copy of the metrics
	metricsCopy := m.metrics
	return &metricsCopy, nil
}

// RecvBundle receives a bundle from the link
func (m *MockCLA) RecvBundle() (*bpa.Bundle, *LinkMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isOpen {
		return nil, nil, fmt.Errorf("link not open")
	}

	m.status = CLAStatusReceiving

	// Mock implementation - would normally receive from hardware
	// For now, return error indicating no bundle available
	m.status = CLAStatusIdle
	return nil, nil, fmt.Errorf("no bundle available")
}

// Status returns the current CLA status
func (m *MockCLA) Status() CLAStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// LinkMetrics returns the current link metrics
func (m *MockCLA) LinkMetrics() *LinkMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	metricsCopy := m.metrics
	return &metricsCopy
}

// SetLinkMetrics allows setting link metrics for testing
func (m *MockCLA) SetLinkMetrics(metrics LinkMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = metrics
}
