package bpa

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Priority represents bundle priority levels
type Priority int

const (
	PriorityBulk Priority = iota
	PriorityNormal
	PriorityExpedited
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityBulk:
		return "bulk"
	case PriorityNormal:
		return "normal"
	case PriorityExpedited:
		return "expedited"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// BundleType represents the type of bundle
type BundleType int

const (
	BundleTypeData BundleType = iota
	BundleTypePingRequest
	BundleTypePingResponse
)

func (bt BundleType) String() string {
	switch bt {
	case BundleTypeData:
		return "data"
	case BundleTypePingRequest:
		return "ping_request"
	case BundleTypePingResponse:
		return "ping_response"
	default:
		return "unknown"
	}
}

// EndpointID represents a DTN endpoint identifier
type EndpointID struct {
	Scheme string // "dtn" or "ipn"
	SSP    string // scheme-specific part
}

func (e EndpointID) String() string {
	return fmt.Sprintf("%s://%s", e.Scheme, e.SSP)
}

// BundleID uniquely identifies a bundle
type BundleID struct {
	SourceEID          EndpointID
	CreationTimestamp  int64 // Unix epoch seconds
	SequenceNumber     uint64
}

func (b BundleID) String() string {
	return fmt.Sprintf("%s-%d-%d", b.SourceEID.String(), b.CreationTimestamp, b.SequenceNumber)
}

// Hash returns a hash of the bundle ID for use as a map key
func (b BundleID) Hash() string {
	h := sha256.New()
	h.Write([]byte(b.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// Bundle represents a BPv7 bundle
type Bundle struct {
	ID          BundleID
	Destination EndpointID
	Payload     []byte
	Priority    Priority
	Lifetime    int64      // seconds
	CreatedAt   int64      // Unix epoch seconds
	BundleType  BundleType
}

// IsExpired checks if the bundle has expired at the given time
func (b *Bundle) IsExpired(currentTime int64) bool {
	return b.CreatedAt+b.Lifetime <= currentTime
}

// Size returns the approximate size of the bundle in bytes
func (b *Bundle) Size() int {
	// Approximate: payload + overhead for headers
	return len(b.Payload) + 256 // 256 bytes estimated header overhead
}
