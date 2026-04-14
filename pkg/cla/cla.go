package cla

import (
	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/contact"
)

// CLAType represents the type of convergence layer adapter
type CLAType int

const (
	CLATypeAX25LTPVHFTNC CLAType = iota
	CLATypeAX25LTPUHFTNC
	CLATypeAX25LTPUHFIQB200
	CLATypeAX25LTPUHFIQ
	CLATypeAX25LTPSBandIQ
	CLATypeAX25LTPXBandIQ
)

func (ct CLAType) String() string {
	switch ct {
	case CLATypeAX25LTPVHFTNC:
		return "ax25ltp_vhf_tnc"
	case CLATypeAX25LTPUHFTNC:
		return "ax25ltp_uhf_tnc"
	case CLATypeAX25LTPUHFIQB200:
		return "ax25ltp_uhf_iq_b200"
	case CLATypeAX25LTPUHFIQ:
		return "ax25ltp_uhf_iq"
	case CLATypeAX25LTPSBandIQ:
		return "ax25ltp_sband_iq"
	case CLATypeAX25LTPXBandIQ:
		return "ax25ltp_xband_iq"
	default:
		return "unknown"
	}
}

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
