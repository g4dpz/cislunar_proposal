// Package hdtnconfig generates HDTN JSON configuration files from Go structs.
package hdtnconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HDTNConfig is the top-level HDTN configuration structure.
type HDTNConfig struct {
	HDTNConfigName     string          `json:"hdtnConfigName"`
	MyNodeID           int             `json:"myNodeId"`
	MySchemeStr        string          `json:"mySchemeStr"`
	MyDtnEidStr        string          `json:"myDtnEidStr"`
	MyDtnDemuxServices []string        `json:"myDtnDemuxServices,omitempty"`
	StoragePath        string          `json:"storagePath"`
	InductsConfig      InductsConfig   `json:"inductsConfig"`
	OutductsConfig     OutductsConfig  `json:"outductsConfig"`
	ContactPlanJSON    ContactPlanJSON `json:"contactPlanJson"`
}

// InductsConfig holds all induct (receiver) definitions.
type InductsConfig struct {
	InductVector []Induct `json:"inductVector"`
}

// Induct defines a single convergence layer induct.
type Induct struct {
	ConvergenceLayer  string `json:"convergenceLayer"`
	Name              string `json:"name"`
	BoundPort         int    `json:"boundPort,omitempty"`
	ThisLtpEngineID   uint64 `json:"thisLtpEngineId,omitempty"`
	RemoteLtpEngineID uint64 `json:"remoteLtpEngineId,omitempty"`
	KissTncDevice     string `json:"kissTncDevice,omitempty"`
	KissBaudRate      int    `json:"kissBaudRate,omitempty"`
	KissPortNumber    int    `json:"kissPortNumber,omitempty"`
	LtpMtu            int    `json:"ltpMtu,omitempty"`
}

// OutductsConfig holds all outduct (sender) definitions.
type OutductsConfig struct {
	OutductVector []Outduct `json:"outductVector"`
}

// Outduct defines a single convergence layer outduct.
type Outduct struct {
	ConvergenceLayer   string `json:"convergenceLayer"`
	Name               string `json:"name"`
	NextHopNodeID      int    `json:"nextHopNodeId"`
	RemoteHostname     string `json:"remoteHostname,omitempty"`
	RemotePort         int    `json:"remotePort,omitempty"`
	ThisLtpEngineID    uint64 `json:"thisLtpEngineId,omitempty"`
	RemoteLtpEngineID  uint64 `json:"remoteLtpEngineId,omitempty"`
	KissTncDevice      string `json:"kissTncDevice,omitempty"`
	KissBaudRate       int    `json:"kissBaudRate,omitempty"`
	KissPortNumber     int    `json:"kissPortNumber,omitempty"`
	LtpMtu             int    `json:"ltpMtu,omitempty"`
	LtpDataSegmentRate int    `json:"ltpDataSegmentRate,omitempty"`
}

// ContactPlanJSON holds the embedded contact plan.
type ContactPlanJSON struct {
	Contacts []ContactEntry `json:"contacts"`
}

// ContactEntry defines a single contact in the plan.
type ContactEntry struct {
	Source         int   `json:"source"`
	Dest           int   `json:"dest"`
	StartTime      int64 `json:"startTime"`
	EndTime        int64 `json:"endTime"`
	RateBitsPerSec int64 `json:"rateBitsPerSec"`
}

// TerrestrialOpts holds parameters for generating a terrestrial node config.
type TerrestrialOpts struct {
	NodeNumber       int
	NodeName         string
	Callsign         string
	StoragePath      string
	TNCDevice        string
	TNCBaudRate      int
	UDPLocalPort     int
	UDPRemoteHost    string
	UDPRemotePort    int
	RemoteNodeNumber int
	ContactDataRate  int64
}

// Validate checks all required fields and returns an error if invalid.
func (c *HDTNConfig) Validate() error {
	if c.MyNodeID <= 0 {
		return fmt.Errorf("invalid field myNodeId: must be greater than 0")
	}
	if c.StoragePath == "" {
		return fmt.Errorf("invalid field storagePath: must not be empty")
	}
	if len(c.InductsConfig.InductVector) == 0 {
		return fmt.Errorf("invalid field inductsConfig.inductVector: must not be empty")
	}
	if len(c.OutductsConfig.OutductVector) == 0 {
		return fmt.Errorf("invalid field outductsConfig.outductVector: must not be empty")
	}
	if len(c.ContactPlanJSON.Contacts) == 0 {
		return fmt.Errorf("invalid field contactPlanJson.contacts: must not be empty")
	}
	return nil
}

// WriteToFile serializes the config to JSON and writes to the given path.
func (c *HDTNConfig) WriteToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}
