package cla

import (
	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/contact"
)

// CLAType represents the type of convergence layer adapter
type CLAType int

const (
	CLATypeKISSLTPVHFTNC CLAType = iota
	CLATypeKISSLTPUHFTNC
	CLATypeKISSLTPUHFIQB200
	CLATypeKISSLTPUHFIQ
	CLATypeKISSLTPSBandIQ
	CLATypeKISSLTPXBandIQ
)

func (ct CLAType) String() string {
	switch ct {
	case CLATypeKISSLTPVHFTNC:
		return "kissltp_vhf_tnc"
	case CLATypeKISSLTPUHFTNC:
		return "kissltp_uhf_tnc"
	case CLATypeKISSLTPUHFIQB200:
		return "kissltp_uhf_iq_b200"
	case CLATypeKISSLTPUHFIQ:
		return "kissltp_uhf_iq"
	case CLATypeKISSLTPSBandIQ:
		return "kissltp_sband_iq"
	case CLATypeKISSLTPXBandIQ:
		return "kissltp_xband_iq"
	default:
		return "unknown"
	}
}

// Legacy aliases for backward compatibility
const (
	CLATypeAX25LTPVHFTNC    = CLATypeKISSLTPVHFTNC
	CLATypeAX25LTPUHFTNC    = CLATypeKISSLTPUHFTNC
	CLATypeAX25LTPUHFIQB200 = CLATypeKISSLTPUHFIQB200
	CLATypeAX25LTPUHFIQ     = CLATypeKISSLTPUHFIQ
	CLATypeAX25LTPSBandIQ   = CLATypeKISSLTPSBandIQ
	CLATypeAX25LTPXBandIQ   = CLATypeKISSLTPXBandIQ
)

// CLAStatus represents the current status of the CLA
type CLAStatus int

const (
	CLAStatusIdle CLAStatus = iota
	CLAStatusTransmitting
	CLAStatusReceiving
	CLAStatusError
)

func (cs CLAStatus) String() string {
	switch cs {
	case CLAStatusIdle:
		return "idle"
	case CLAStatusTransmitting:
		return "transmitting"
	case CLAStatusReceiving:
		return "receiving"
	case CLAStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// LinkMetrics represents link quality metrics
type LinkMetrics struct {
	RSSI             int     // dBm
	SNR              float64 // dB
	BitErrorRate     float64
	BytesTransferred int64
}

// ConvergenceLayerAdapter is the interface for all CLA implementations
type ConvergenceLayerAdapter interface {
	// Type returns the CLA type
	Type() CLAType

	// Open establishes the link for a contact window
	Open(window contact.ContactWindow) error

	// Close closes the link
	Close() error

	// SendBundle transmits a bundle over the link
	SendBundle(bundle *bpa.Bundle) (*LinkMetrics, error)

	// RecvBundle receives a bundle from the link
	RecvBundle() (*bpa.Bundle, *LinkMetrics, error)

	// Status returns the current CLA status
	Status() CLAStatus

	// LinkMetrics returns the current link metrics
	LinkMetrics() *LinkMetrics
}
