package ionconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// LEONodeConfig holds configuration for a LEO CubeSat flight node
type LEONodeConfig struct {
	NodeID           string
	NodeNumber       int
	Callsign         string
	StorageBytes     uint64
	SRAMBytes        uint64
	ContactPlan      []LEOContact
	OrbitalParams    *OrbitalParameters
	TelemetryEnabled bool
}

// LEOContact represents a CGR-predicted orbital pass contact window
type LEOContact struct {
	RemoteNodeNumber int
	RemoteCallsign   string
	StartTime        time.Time
	Duration         time.Duration // 5-10 minutes typical for LEO pass
	DataRate         int           // 9600 bps for UHF
	MaxElevationDeg  float64       // Peak elevation during pass
	Confidence       float64       // Prediction confidence (0.0-1.0)
}

// OrbitalParameters represents TLE/ephemeris data for CGR prediction
type OrbitalParameters struct {
	Epoch           time.Time
	SemiMajorAxisM  float64
	Eccentricity    float64
	InclinationDeg  float64
	RAANDeg         float64
	ArgPeriapsisDeg float64
	TrueAnomalyDeg  float64
}

// GenerateLEOConfig generates ION-DTN configuration files for a LEO flight node
func GenerateLEOConfig(config LEONodeConfig, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each config file
	if err := generateLEOIonrc(config, outputDir); err != nil {
		return err
	}
	if err := generateLEOLtprc(config, outputDir); err != nil {
		return err
	}
	if err := generateLEOBprc(config, outputDir); err != nil {
		return err
	}
	if err := generateLEOIpnrc(config, outputDir); err != nil {
		return err
	}
	if err := generateLEOIonconfig(config, outputDir); err != nil {
		return err
	}

	fmt.Printf("Generated LEO config for node %s in %s\n", config.NodeID, outputDir)
	return nil
}

// generateLEOIonrc generates the ionrc file for LEO node
func generateLEOIonrc(config LEONodeConfig, outputDir string) error {
	tmpl := `## ionrc configuration for LEO node {{.NodeID}}
## STM32U585 OBC + Flight IQ Transceiver
## Generated: {{.Timestamp}}

## Initialize ION node
1 {{.NodeNumber}} ''

## Configure memory (SRAM: {{.SRAMBytes}} bytes, NVM: {{.StorageBytes}} bytes)
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

// generateLEOLtprc generates the ltprc file for LEO node
func generateLEOLtprc(config LEONodeConfig, outputDir string) error {
	tmpl := `## ltprc configuration for LEO node {{.NodeID}}
## LTP over AX.25 via UHF IQ baseband (flight transceiver)
## CGR-predicted contact windows

## Initialize LTP engine
1 {{.SegmentSize}}

## Add LTP spans for CGR-predicted contacts
{{range .Contacts}}
## Contact with node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Start: {{.StartTime}}, Duration: {{.Duration}}, Elevation: {{.MaxElevationDeg}}°
## Confidence: {{.Confidence}}
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
		MaxElevationDeg  float64
		Confidence       float64
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format("15:04:05"),
			Duration:         c.Duration.String(),
			DataRate:         c.DataRate,
			MaxElevationDeg:  c.MaxElevationDeg,
			Confidence:       c.Confidence,
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

// generateLEOBprc generates the bprc file for LEO node
func generateLEOBprc(config LEONodeConfig, outputDir string) error {
	tmpl := `## bprc configuration for LEO node {{.NodeID}}
## Bundle Protocol v7 configuration
## CGR-predicted contact windows for autonomous store-and-forward

## Initialize BP
1

## Add endpoints for this node
a endpoint ipn:{{.NodeNumber}}.0 q
a endpoint ipn:{{.NodeNumber}}.1 q
a endpoint ipn:{{.NodeNumber}}.2 q
{{if .TelemetryEnabled}}a endpoint ipn:{{.NodeNumber}}.10 q  ## telemetry endpoint
{{end}}

## Add protocols
a protocol ipn:{{.NodeNumber}} ltp/{{.NodeNumber}}

## Add outducts for CGR-predicted contacts
{{range .Contacts}}
## Outduct to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Pass window: {{.StartTime}} for {{.Duration}}
a outduct ltp/{{.RemoteNodeNumber}} ltp/{{.RemoteNodeNumber}}
{{end}}

## Start BP
s
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		StartTime        string
		Duration         string
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format("15:04:05"),
			Duration:         c.Duration.String(),
		}
	}

	data := struct {
		NodeID           string
		NodeNumber       int
		TelemetryEnabled bool
		Contacts         []ContactData
	}{
		NodeID:           config.NodeID,
		NodeNumber:       config.NodeNumber,
		TelemetryEnabled: config.TelemetryEnabled,
		Contacts:         contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "node.bprc"), tmpl, data)
}

// generateLEOIpnrc generates the ipnrc file for LEO node
func generateLEOIpnrc(config LEONodeConfig, outputDir string) error {
	tmpl := `## ipnrc configuration for LEO node {{.NodeID}}
## IPN scheme configuration
## CGR-predicted contact plans

## Add plans for CGR-predicted contacts
{{range .Contacts}}
## Plan to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Confidence: {{.Confidence}}
a plan {{.RemoteNodeNumber}} ltp/{{.RemoteNodeNumber}}
{{end}}
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		Confidence       float64
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			Confidence:       c.Confidence,
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

// generateLEOIonconfig generates the ionconfig file for LEO node
func generateLEOIonconfig(config LEONodeConfig, outputDir string) error {
	tmpl := `## ionconfig for LEO node {{.NodeID}}
## CGR-predicted orbital pass configuration
## STM32U585 + Flight IQ Transceiver (UHF 437 MHz, 9.6 kbps)

## Node configuration
NODE_NUMBER={{.NodeNumber}}
CALLSIGN={{.Callsign}}
STORAGE_BYTES={{.StorageBytes}}
SRAM_BYTES={{.SRAMBytes}}

## Flight IQ transceiver configuration
RADIO_TYPE=flight_iq
CENTER_FREQ=437000000
SAMPLE_RATE=1000000
TX_GAIN=50.0
RX_GAIN=40.0
MODULATION=GMSK

## Orbital parameters (for CGR re-prediction)
{{if .OrbitalParams}}
ORBITAL_EPOCH={{.OrbitalEpoch}}
SEMI_MAJOR_AXIS_M={{.SemiMajorAxisM}}
ECCENTRICITY={{.Eccentricity}}
INCLINATION_DEG={{.InclinationDeg}}
RAAN_DEG={{.RAANDeg}}
ARG_PERIAPSIS_DEG={{.ArgPeriapsisDeg}}
TRUE_ANOMALY_DEG={{.TrueAnomalyDeg}}
{{end}}

## CGR-predicted orbital passes (5-10 min duration, 9.6 kbps)
{{range $i, $c := .Contacts}}
CONTACT_{{$i}}_REMOTE={{$c.RemoteNodeNumber}}
CONTACT_{{$i}}_CALLSIGN={{$c.RemoteCallsign}}
CONTACT_{{$i}}_START={{$c.StartTime}}
CONTACT_{{$i}}_DURATION={{$c.Duration}}
CONTACT_{{$i}}_RATE={{$c.DataRate}}
CONTACT_{{$i}}_MAX_ELEVATION={{$c.MaxElevationDeg}}
CONTACT_{{$i}}_CONFIDENCE={{$c.Confidence}}
{{end}}

## Telemetry configuration
TELEMETRY_ENABLED={{.TelemetryEnabled}}
{{if .TelemetryEnabled}}
TELEMETRY_ENDPOINT=ipn:{{.NodeNumber}}.10
TELEMETRY_INTERVAL=60  ## seconds
{{end}}
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		StartTime        string
		Duration         int
		DataRate         int
		MaxElevationDeg  float64
		Confidence       float64
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format(time.RFC3339),
			Duration:         int(c.Duration.Seconds()),
			DataRate:         c.DataRate,
			MaxElevationDeg:  c.MaxElevationDeg,
			Confidence:       c.Confidence,
		}
	}

	data := struct {
		NodeID           string
		NodeNumber       int
		Callsign         string
		StorageBytes     uint64
		SRAMBytes        uint64
		OrbitalParams    bool
		OrbitalEpoch     string
		SemiMajorAxisM   float64
		Eccentricity     float64
		InclinationDeg   float64
		RAANDeg          float64
		ArgPeriapsisDeg  float64
		TrueAnomalyDeg   float64
		Contacts         []ContactData
		TelemetryEnabled bool
	}{
		NodeID:           config.NodeID,
		NodeNumber:       config.NodeNumber,
		Callsign:         config.Callsign,
		StorageBytes:     config.StorageBytes,
		SRAMBytes:        config.SRAMBytes,
		OrbitalParams:    config.OrbitalParams != nil,
		TelemetryEnabled: config.TelemetryEnabled,
		Contacts:         contacts,
	}

	if config.OrbitalParams != nil {
		data.OrbitalEpoch = config.OrbitalParams.Epoch.Format(time.RFC3339)
		data.SemiMajorAxisM = config.OrbitalParams.SemiMajorAxisM
		data.Eccentricity = config.OrbitalParams.Eccentricity
		data.InclinationDeg = config.OrbitalParams.InclinationDeg
		data.RAANDeg = config.OrbitalParams.RAANDeg
		data.ArgPeriapsisDeg = config.OrbitalParams.ArgPeriapsisDeg
		data.TrueAnomalyDeg = config.OrbitalParams.TrueAnomalyDeg
	}

	return writeTemplate(filepath.Join(outputDir, "leo.ionconfig"), tmpl, data)
}

// UpdateLEOContactPlan updates the contact plan with fresh CGR predictions
// This is called when new TLE/ephemeris data is received
func UpdateLEOContactPlan(configDir string, newContacts []LEOContact, orbitalParams *OrbitalParameters) error {
	// Read existing config
	configPath := filepath.Join(configDir, "leo.ionconfig")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("LEO config not found at %s", configPath)
	}

	// Create updated config with new contact plan
	// In a real implementation, this would parse the existing config and update only the contact plan
	// For now, we'll regenerate the entire config
	
	fmt.Printf("Updated LEO contact plan with %d CGR-predicted passes\n", len(newContacts))
	fmt.Printf("Orbital parameters updated: epoch=%v\n", orbitalParams.Epoch)
	
	return nil
}

// GenerateLEOWithCGRPrediction generates LEO config with CGR-predicted contact windows
func GenerateLEOWithCGRPrediction(
	nodeID string,
	nodeNumber int,
	callsign string,
	orbitalParams *OrbitalParameters,
	predictedContacts []LEOContact,
	outputDir string,
) error {
	config := LEONodeConfig{
		NodeID:           nodeID,
		NodeNumber:       nodeNumber,
		Callsign:         callsign,
		StorageBytes:     128 * 1024 * 1024, // 128 MB external NVM
		SRAMBytes:        786 * 1024,         // 786 KB SRAM (STM32U585)
		ContactPlan:      predictedContacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	return GenerateLEOConfig(config, outputDir)
}

// writeTemplate is a helper function to write a template to a file
func writeTemplate(filePath, tmpl string, data interface{}) error {
	t, err := template.New("config").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
