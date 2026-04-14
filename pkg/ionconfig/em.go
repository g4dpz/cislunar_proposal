package ionconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EMNodeConfig holds configuration for an Engineering Model node
type EMNodeConfig struct {
	NodeID       string
	NodeNumber   int
	Callsign     string
	StorageBytes uint64
	SRAMBytes    uint64
	ContactPlan  []EMContact
}

// EMContact represents a simulated orbital pass contact window
type EMContact struct {
	RemoteNodeNumber int
	RemoteCallsign   string
	StartTime        time.Time
	Duration         time.Duration // 8 minutes typical for LEO pass
	DataRate         int           // 9600 bps for UHF
}

// GenerateEMConfig generates ION-DTN configuration files for an EM node
func GenerateEMConfig(config EMNodeConfig, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each config file
	if err := generateEMIonrc(config, outputDir); err != nil {
		return err
	}
	if err := generateEMLtprc(config, outputDir); err != nil {
		return err
	}
	if err := generateEMBprc(config, outputDir); err != nil {
		return err
	}
	if err := generateEMIpnrc(config, outputDir); err != nil {
		return err
	}
	if err := generateEMIonconfig(config, outputDir); err != nil {
		return err
	}

	fmt.Printf("Generated EM config for node %s in %s\n", config.NodeID, outputDir)
	return nil
}

// generateEMIonrc generates the ionrc file for EM node
func generateEMIonrc(config EMNodeConfig, outputDir string) error {
	tmpl := `## ionrc configuration for EM node {{.NodeID}}
## STM32U585 OBC + Ettus B200mini SDR
## Generated: {{.Timestamp}}

## Initialize ION node
1 {{.NodeNumber}} ''

## Configure SDR memory (SRAM: {{.SRAMBytes}} bytes, NVM: {{.StorageBytes}} bytes)
## SRAM for in-flight bundles and IQ buffers
## External NVM for persistent bundle storage
s

## Start ION
`

	data := struct {
		NodeID       string
		NodeNumber   int
		SRAMBytes    uint64
		StorageBytes uint64
		Timestamp    string
	}{
		NodeID:       config.NodeID,
		NodeNumber:   config.NodeNumber,
		SRAMBytes:    config.SRAMBytes,
		StorageBytes: config.StorageBytes,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	return writeTemplate(filepath.Join(outputDir, "node.ionrc"), tmpl, data)
}

// generateEMLtprc generates the ltprc file for EM node
func generateEMLtprc(config EMNodeConfig, outputDir string) error {
	tmpl := `## ltprc configuration for EM node {{.NodeID}}
## LTP over AX.25 via UHF IQ baseband (B200mini)

## Initialize LTP engine
1 {{.SegmentSize}}

## Add LTP spans for each contact
{{range .Contacts}}
## Contact with node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Start: {{.StartTime}}, Duration: {{.Duration}}, Rate: {{.DataRate}} bps
a span {{.RemoteNodeNumber}} {{$.MaxBlockSize}} {{$.MaxSessions}} {{$.MaxSegments}} {{.DataRate}} 100 {{$.AggregationSize}} 'udplso localhost:{{$.LTPPort}}'
{{end}}

## Start LTP
s
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		StartTime        string
		Duration         string
		DataRate         int
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format("15:04:05"),
			Duration:         c.Duration.String(),
			DataRate:         c.DataRate,
		}
	}

	data := struct {
		NodeID          string
		SegmentSize     int
		MaxBlockSize    int
		MaxSessions     int
		MaxSegments     int
		AggregationSize int
		LTPPort         int
		Contacts        []ContactData
	}{
		NodeID:          config.NodeID,
		SegmentSize:     1400,
		MaxBlockSize:    10000,
		MaxSessions:     10,
		MaxSegments:     100,
		AggregationSize: 1000,
		LTPPort:         1113,
		Contacts:        contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "node.ltprc"), tmpl, data)
}

// generateEMBprc generates the bprc file for EM node
func generateEMBprc(config EMNodeConfig, outputDir string) error {
	tmpl := `## bprc configuration for EM node {{.NodeID}}
## Bundle Protocol v7 configuration

## Initialize BP
1

## Add endpoint for this node
a endpoint ipn:{{.NodeNumber}}.0 q
a endpoint ipn:{{.NodeNumber}}.1 q
a endpoint ipn:{{.NodeNumber}}.2 q

## Add protocols
a protocol ipn:{{.NodeNumber}} ltp/{{.NodeNumber}}

## Add outducts for each contact
{{range .Contacts}}
## Outduct to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
a outduct ltp/{{.RemoteNodeNumber}} ltp/{{.RemoteNodeNumber}}
{{end}}

## Start BP
s
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
		}
	}

	data := struct {
		NodeID     string
		NodeNumber int
		Contacts   []ContactData
	}{
		NodeID:     config.NodeID,
		NodeNumber: config.NodeNumber,
		Contacts:   contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "node.bprc"), tmpl, data)
}

// generateEMIpnrc generates the ipnrc file for EM node
func generateEMIpnrc(config EMNodeConfig, outputDir string) error {
	tmpl := `## ipnrc configuration for EM node {{.NodeID}}
## IPN scheme configuration

## Add plans for each contact
{{range .Contacts}}
## Plan to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
a plan {{.RemoteNodeNumber}} ltp/{{.RemoteNodeNumber}}
{{end}}
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
		}
	}

	data := struct {
		NodeID   string
		Contacts []ContactData
	}{
		NodeID:   config.NodeID,
		Contacts: contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "node.ipnrc"), tmpl, data)
}

// generateEMIonconfig generates the ionconfig file for EM node
func generateEMIonconfig(config EMNodeConfig, outputDir string) error {
	tmpl := `## ionconfig for EM node {{.NodeID}}
## Simulated orbital pass configuration
## STM32U585 + B200mini SDR (UHF 437 MHz, 9.6 kbps)

## Node configuration
NODE_NUMBER={{.NodeNumber}}
CALLSIGN={{.Callsign}}
STORAGE_BYTES={{.StorageBytes}}
SRAM_BYTES={{.SRAMBytes}}

## B200mini SDR configuration
SDR_TYPE=b200mini
CENTER_FREQ=437000000
SAMPLE_RATE=1000000
TX_GAIN=50.0
RX_GAIN=40.0

## Simulated orbital passes (8 min duration, 9.6 kbps)
{{range $i, $c := .Contacts}}
CONTACT_{{$i}}_REMOTE={{$c.RemoteNodeNumber}}
CONTACT_{{$i}}_CALLSIGN={{$c.RemoteCallsign}}
CONTACT_{{$i}}_START={{$c.StartTime}}
CONTACT_{{$i}}_DURATION={{$c.Duration}}
CONTACT_{{$i}}_RATE={{$c.DataRate}}
{{end}}
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		StartTime        string
		Duration         int
		DataRate         int
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format(time.RFC3339),
			Duration:         int(c.Duration.Seconds()),
			DataRate:         c.DataRate,
		}
	}

	data := struct {
		NodeID       string
		NodeNumber   int
		Callsign     string
		StorageBytes uint64
		SRAMBytes    uint64
		Contacts     []ContactData
	}{
		NodeID:       config.NodeID,
		NodeNumber:   config.NodeNumber,
		Callsign:     config.Callsign,
		StorageBytes: config.StorageBytes,
		SRAMBytes:    config.SRAMBytes,
		Contacts:     contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "em.ionconfig"), tmpl, data)
}

// GenerateEMTwoNodeSetup generates configuration for a two-node EM test setup
func GenerateEMTwoNodeSetup(nodeACallsign, nodeBCallsign, outputDir string) error {
	// Node A configuration (EM node)
	now := time.Now()
	nodeA := EMNodeConfig{
		NodeID:       "em-node-a",
		NodeNumber:   1,
		Callsign:     nodeACallsign,
		StorageBytes: 128 * 1024 * 1024, // 128 MB external NVM
		SRAMBytes:    786 * 1024,         // 786 KB SRAM
		ContactPlan: []EMContact{
			{
				RemoteNodeNumber: 2,
				RemoteCallsign:   nodeBCallsign,
				StartTime:        now.Add(5 * time.Minute),
				Duration:         8 * time.Minute, // 8-minute simulated pass
				DataRate:         9600,             // 9.6 kbps UHF
			},
			{
				RemoteNodeNumber: 2,
				RemoteCallsign:   nodeBCallsign,
				StartTime:        now.Add(95 * time.Minute), // 90 min orbit + 5 min
				Duration:         8 * time.Minute,
				DataRate:         9600,
			},
		},
	}

	// Node B configuration (ground station)
	nodeB := EMNodeConfig{
		NodeID:       "em-node-b",
		NodeNumber:   2,
		Callsign:     nodeBCallsign,
		StorageBytes: 512 * 1024 * 1024, // 512 MB (ground station)
		SRAMBytes:    0,                  // N/A for ground station
		ContactPlan: []EMContact{
			{
				RemoteNodeNumber: 1,
				RemoteCallsign:   nodeACallsign,
				StartTime:        now.Add(5 * time.Minute),
				Duration:         8 * time.Minute,
				DataRate:         9600,
			},
			{
				RemoteNodeNumber: 1,
				RemoteCallsign:   nodeACallsign,
				StartTime:        now.Add(95 * time.Minute),
				Duration:         8 * time.Minute,
				DataRate:         9600,
			},
		},
	}

	// Generate configs
	if err := GenerateEMConfig(nodeA, filepath.Join(outputDir, "node-a")); err != nil {
		return err
	}
	if err := GenerateEMConfig(nodeB, filepath.Join(outputDir, "node-b")); err != nil {
		return err
	}

	fmt.Printf("Generated two-node EM setup in %s\n", outputDir)
	return nil
}
