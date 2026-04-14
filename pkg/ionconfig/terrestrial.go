package ionconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// TerrestrialNodeConfig represents configuration for a terrestrial DTN node
type TerrestrialNodeConfig struct {
	NodeNumber      int
	NodeName        string
	Callsign        string
	TNCDevice       string
	ContactPlan     []ContactEntry
	StorageQuota    int64 // bytes
	HeapWords       int64
}

// ContactEntry represents a contact window in the contact plan
type ContactEntry struct {
	StartTime    int64  // Unix timestamp
	EndTime      int64  // Unix timestamp
	FromNode     int
	ToNode       int
	DataRate     int64 // bits per second
}

// ionrcTemplate is the template for ionrc (ION initialization)
const ionrcTemplate = `## ionrc configuration for {{.NodeName}}
## Node {{.NodeNumber}} - Terrestrial DTN Node

# Initialize ION node
1 {{.NodeNumber}} {{.Callsign}} {{.HeapWords}}

# Start ION
s
`

// ltprcTemplate is the template for ltprc (LTP configuration)
const ltprcTemplate = `## ltprc configuration for {{.NodeName}}
## LTP (Licklider Transmission Protocol) over AX.25

# Initialize LTP engine
1 {{.StorageQuota}}

# Add LTP span for each contact
{{range .ContactPlan}}# Contact: Node {{.FromNode}} -> Node {{.ToNode}}
a span {{.ToNode}} 32 32 1400 10000 1 'udplso localhost:{{.ToNode}}1113'
{{end}}

# Start LTP
s
`

// bprcTemplate is the template for bprc (Bundle Protocol configuration)
const bprcTemplate = `## bprc configuration for {{.NodeName}}
## Bundle Protocol v7 configuration

# Initialize BP
1

# Add endpoint for this node
a endpoint ipn:{{.NodeNumber}}.0 q
a endpoint ipn:{{.NodeNumber}}.1 q
a endpoint ipn:{{.NodeNumber}}.2 q

# Add protocol for LTP
a protocol ltp 1400 100

# Add induct for receiving bundles
a induct ltp {{.NodeNumber}} ltpcli

# Add outduct for each contact
{{range .ContactPlan}}{{if eq .FromNode $.NodeNumber}}a outduct ltp {{.ToNode}} ltpclo
{{end}}{{end}}

# Start BP
s

# Watch characters (for debugging)
## w 1
`

// ipnrcTemplate is the template for ipnrc (IPN scheme configuration)
const ipnrcTemplate = `## ipnrc configuration for {{.NodeName}}
## IPN (Interplanetary Network) scheme routing

# Add plan for each destination node
{{range .ContactPlan}}{{if eq .FromNode $.NodeNumber}}a plan {{.ToNode}} ltp/{{.ToNode}}
{{end}}{{end}}
`

// kissIonConfigTemplate is the template for KISS interface configuration
const kissIonConfigTemplate = `## KISS interface configuration for {{.NodeName}}
## Mobilinkd TNC4 via USB serial

# TNC device path
device={{.TNCDevice}}

# Baud rate (9600 for G3RUH-compatible GFSK)
baudrate=9600

# KISS mode
kiss=1

# Callsign
callsign={{.Callsign}}
`

// GenerateTerrestrialConfig generates ION-DTN configuration files for a terrestrial node
func GenerateTerrestrialConfig(config TerrestrialNodeConfig, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate ionrc
	if err := generateFile(outputDir, "node.ionrc", ionrcTemplate, config); err != nil {
		return fmt.Errorf("failed to generate ionrc: %w", err)
	}

	// Generate ltprc
	if err := generateFile(outputDir, "node.ltprc", ltprcTemplate, config); err != nil {
		return fmt.Errorf("failed to generate ltprc: %w", err)
	}

	// Generate bprc
	if err := generateFile(outputDir, "node.bprc", bprcTemplate, config); err != nil {
		return fmt.Errorf("failed to generate bprc: %w", err)
	}

	// Generate ipnrc
	if err := generateFile(outputDir, "node.ipnrc", ipnrcTemplate, config); err != nil {
		return fmt.Errorf("failed to generate ipnrc: %w", err)
	}

	// Generate KISS interface config
	if err := generateFile(outputDir, "kiss.ionconfig", kissIonConfigTemplate, config); err != nil {
		return fmt.Errorf("failed to generate kiss.ionconfig: %w", err)
	}

	return nil
}

// generateFile generates a configuration file from a template
func generateFile(outputDir, filename, templateStr string, config TerrestrialNodeConfig) error {
	tmpl, err := template.New(filename).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	filePath := filepath.Join(outputDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateTwoNodeSetup generates configuration for a two-node terrestrial setup
func GenerateTwoNodeSetup(nodeADir, nodeBDir string) error {
	// Node A configuration
	nodeA := TerrestrialNodeConfig{
		NodeNumber:   1,
		NodeName:     "node-a",
		Callsign:     "KA1ABC",
		TNCDevice:    "/dev/ttyUSB0",
		StorageQuota: 512 * 1024 * 1024, // 512 MB
		HeapWords:    100000,
		ContactPlan: []ContactEntry{
			{
				StartTime: 0,
				EndTime:   2147483647, // Max int32 (always available for terrestrial)
				FromNode:  1,
				ToNode:    2,
				DataRate:  9600,
			},
		},
	}

	// Node B configuration
	nodeB := TerrestrialNodeConfig{
		NodeNumber:   2,
		NodeName:     "node-b",
		Callsign:     "KB2XYZ",
		TNCDevice:    "/dev/ttyUSB0",
		StorageQuota: 512 * 1024 * 1024, // 512 MB
		HeapWords:    100000,
		ContactPlan: []ContactEntry{
			{
				StartTime: 0,
				EndTime:   2147483647,
				FromNode:  2,
				ToNode:    1,
				DataRate:  9600,
			},
		},
	}

	// Generate configurations
	if err := GenerateTerrestrialConfig(nodeA, nodeADir); err != nil {
		return fmt.Errorf("failed to generate node A config: %w", err)
	}

	if err := GenerateTerrestrialConfig(nodeB, nodeBDir); err != nil {
		return fmt.Errorf("failed to generate node B config: %w", err)
	}

	return nil
}
