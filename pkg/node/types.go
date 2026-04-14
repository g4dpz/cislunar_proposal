package node

import (
	"terrestrial-dtn/pkg/bpa"
)

// NodeType represents the type of DTN node
type NodeType int

const (
	NodeTypeTerrestrial NodeType = iota
	NodeTypeEngineeringModel
	NodeTypeLEOCubesat
	NodeTypeCislunar
)

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeTerrestrial:
		return "terrestrial"
	case NodeTypeEngineeringModel:
		return "engineering_model"
	case NodeTypeLEOCubesat:
		return "leo_cubesat"
	case NodeTypeCislunar:
		return "cislunar"
	default:
		return "unknown"
	}
}

// NodeConfig represents the configuration for a DTN node
type NodeConfig struct {
	NodeID          string
	NodeType        NodeType
	Endpoints       []bpa.EndpointID
	MaxStorageBytes int64
	SRAMBytes       int64
	DefaultPriority bpa.Priority
}

// NodeHealth represents the health status of a node
type NodeHealth struct {
	UptimeSeconds       int64
	StorageUsedPercent  float64
	BundlesStored       int
	BundlesForwarded    int
	BundlesDropped      int
	LastContactTime     *int64   // optional
	Temperature         *float64 // optional, for space nodes
	BatteryPercent      *float64 // optional, for space nodes
}

// NodeStatistics represents cumulative statistics for a node
type NodeStatistics struct {
	TotalBundlesReceived int64
	TotalBundlesSent     int64
	TotalBytesReceived   int64
	TotalBytesSent       int64
	AverageLatency       float64 // seconds
	ContactsCompleted    int64
	ContactsMissed       int64
	BundlesForwarded     int64
}
