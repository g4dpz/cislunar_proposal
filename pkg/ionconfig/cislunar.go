package ionconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CislunarNodeConfig holds configuration for a cislunar deep-space node
type CislunarNodeConfig struct {
	NodeID           string
	NodeNumber       int
	Callsign         string
	StorageBytes     uint64
	SRAMBytes        uint64
	ContactPlan      []CislunarContact
	OrbitalParams    *OrbitalParameters
	TelemetryEnabled bool
	Band             CislunarBand // S-band or X-band
}

// CislunarBand represents the RF band for cislunar communication
type CislunarBand string

const (
	SBand CislunarBand = "sband" // 2.2 GHz
	XBand CislunarBand = "xband" // 8.4 GHz
)

// CislunarContact represents a CGR-predicted cislunar contact window
type CislunarContact struct {
	RemoteNodeNumber int
	RemoteCallsign   string
	StartTime        time.Time
	Duration         time.Duration // Hours typical for cislunar passes
	DataRate         int           // 500 bps for S-band with BPSK+LDPC
	MaxElevationDeg  float64       // Peak elevation during pass
	Confidence       float64       // Prediction confidence (0.0-1.0)
	LightTimeDelay   float64       // One-way light-time delay in seconds (1-2s)
}

// GenerateCislunarConfig generates ION-DTN configuration files for a cislunar node
func GenerateCislunarConfig(config CislunarNodeConfig, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each config file
	if err := generateCislunarIonrc(config, outputDir); err != nil {
		return err
	}
	if err := generateCislunarLtprc(config, outputDir); err != nil {
		return err
	}
	if err := generateCislunarBprc(config, outputDir); err != nil {
		return err
	}
	if err := generateCislunarIpnrc(config, outputDir); err != nil {
		return err
	}
	if err := generateCislunarIonconfig(config, outputDir); err != nil {
		return err
	}

	fmt.Printf("Generated cislunar config for node %s in %s\n", config.NodeID, outputDir)
	return nil
}

// generateCislunarIonrc generates the ionrc file for cislunar node
func generateCislunarIonrc(config CislunarNodeConfig, outputDir string) error {
	tmpl := `## ionrc configuration for cislunar node {{.NodeID}}
## STM32U585 OBC (or higher-capability processor) + IQ Transceiver
## Deep-space DTN with 1-2 second light-time delay
## Generated: {{.Timestamp}}

## Initialize ION node
1 {{.NodeNumber}} ''

## Configure memory for long-duration message storage
## SRAM: {{.SRAMBytes}} bytes (in-flight bundles and IQ buffers)
## NVM: {{.StorageBytes}} bytes (persistent bundle storage for extended contact gaps)
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

// generateCislunarLtprc generates the ltprc file for cislunar node
func generateCislunarLtprc(config CislunarNodeConfig, outputDir string) error {
	tmpl := `## ltprc configuration for cislunar node {{.NodeID}}
## LTP over AX.25 via {{.Band}} IQ baseband
## CGR-predicted contact windows with 1-2 second light-time delay
## Long-delay session management for deep-space links

## Initialize LTP engine with larger segment size for deep-space
1 {{.SegmentSize}}

## Add LTP spans for CGR-predicted cislunar contacts
{{range .Contacts}}
## Contact with node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Start: {{.StartTime}}, Duration: {{.Duration}}, Elevation: {{.MaxElevationDeg}}°
## Light-time delay: {{.LightTimeDelay}}s one-way, RTT: {{.RoundTripTime}}s
## Confidence: {{.Confidence}}
## LTP timeout configured for deep-space delay ({{.LTPTimeout}}s)
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
		LightTimeDelay   float64
		RoundTripTime    float64
		LTPTimeout       float64
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
			LightTimeDelay:   c.LightTimeDelay,
			RoundTripTime:    c.LightTimeDelay * 2,
			LTPTimeout:       c.LightTimeDelay*2 + 10.0, // RTT + 10s processing margin
		}
	}

	bandStr := "S-band 2.2 GHz"
	if config.Band == XBand {
		bandStr = "X-band 8.4 GHz"
	}

	data := struct {
		NodeID          string
		Band            string
		SegmentSize     int
		MaxBlockSize    int
		MaxSessions     int
		MaxSegments     int
		AggregationSize int
		LTPPort         int
		Contacts        []ContactData
	}{
		NodeID:          config.NodeID,
		Band:            bandStr,
		SegmentSize:     1400,
		MaxBlockSize:    50000,  // Larger blocks for long-duration contacts
		MaxSessions:     20,     // More sessions for extended contact windows
		MaxSegments:     500,    // More segments for deep-space reliability
		AggregationSize: 5000,   // Larger aggregation for efficiency
		LTPPort:         1113,
		Contacts:        contacts,
	}

	return writeTemplate(filepath.Join(outputDir, "node.ltprc"), tmpl, data)
}

// generateCislunarBprc generates the bprc file for cislunar node
func generateCislunarBprc(config CislunarNodeConfig, outputDir string) error {
	tmpl := `## bprc configuration for cislunar node {{.NodeID}}
## Bundle Protocol v7 configuration
## CGR-predicted contact windows for deep-space store-and-forward
## Long-duration message storage for extended contact gaps

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

## Add outducts for CGR-predicted cislunar contacts
{{range .Contacts}}
## Outduct to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Pass window: {{.StartTime}} for {{.Duration}}
## Light-time delay: {{.LightTimeDelay}}s one-way
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
		LightTimeDelay   float64
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format("15:04:05"),
			Duration:         c.Duration.String(),
			LightTimeDelay:   c.LightTimeDelay,
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

// generateCislunarIpnrc generates the ipnrc file for cislunar node
func generateCislunarIpnrc(config CislunarNodeConfig, outputDir string) error {
	tmpl := `## ipnrc configuration for cislunar node {{.NodeID}}
## IPN scheme configuration
## CGR-predicted contact plans for deep-space communication

## Add plans for CGR-predicted cislunar contacts
{{range .Contacts}}
## Plan to node {{.RemoteNodeNumber}} ({{.RemoteCallsign}})
## Confidence: {{.Confidence}} (degrades faster for cislunar)
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

// generateCislunarIonconfig generates the ionconfig file for cislunar node
func generateCislunarIonconfig(config CislunarNodeConfig, outputDir string) error {
	tmpl := `## ionconfig for cislunar node {{.NodeID}}
## CGR-predicted cislunar pass configuration
## STM32U585 (or higher-capability processor) + IQ Transceiver
## {{.Band}} with BPSK + LDPC/Turbo FEC

## Node configuration
NODE_NUMBER={{.NodeNumber}}
CALLSIGN={{.Callsign}}
STORAGE_BYTES={{.StorageBytes}}
SRAM_BYTES={{.SRAMBytes}}

## {{.Band}} IQ transceiver configuration
RADIO_TYPE=cislunar_iq
CENTER_FREQ={{.CenterFreq}}
SAMPLE_RATE={{.SampleRate}}
TX_GAIN={{.TXGain}}
RX_GAIN={{.RXGain}}
MODULATION=BPSK
FEC_TYPE={{.FECType}}
FEC_ENABLED=true

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

## CGR-predicted cislunar passes (hours duration, 500 bps)
{{range $i, $c := .Contacts}}
CONTACT_{{$i}}_REMOTE={{$c.RemoteNodeNumber}}
CONTACT_{{$i}}_CALLSIGN={{$c.RemoteCallsign}}
CONTACT_{{$i}}_START={{$c.StartTime}}
CONTACT_{{$i}}_DURATION={{$c.Duration}}
CONTACT_{{$i}}_RATE={{$c.DataRate}}
CONTACT_{{$i}}_MAX_ELEVATION={{$c.MaxElevationDeg}}
CONTACT_{{$i}}_CONFIDENCE={{$c.Confidence}}
CONTACT_{{$i}}_LIGHT_TIME_DELAY={{$c.LightTimeDelay}}
{{end}}

## Telemetry configuration
TELEMETRY_ENABLED={{.TelemetryEnabled}}
{{if .TelemetryEnabled}}
TELEMETRY_ENDPOINT=ipn:{{.NodeNumber}}.10
TELEMETRY_INTERVAL=300  ## seconds (5 min for cislunar)
{{end}}

## Deep-space configuration
LIGHT_TIME_DELAY={{.LightTimeDelay}}  ## seconds (1-2s for Earth-Moon)
LONG_DURATION_STORAGE=true
EXTENDED_CONTACT_GAPS=true
`

	type ContactData struct {
		RemoteNodeNumber int
		RemoteCallsign   string
		StartTime        string
		Duration         int
		DataRate         int
		MaxElevationDeg  float64
		Confidence       float64
		LightTimeDelay   float64
	}

	contacts := make([]ContactData, len(config.ContactPlan))
	avgLightTimeDelay := 0.0
	for i, c := range config.ContactPlan {
		contacts[i] = ContactData{
			RemoteNodeNumber: c.RemoteNodeNumber,
			RemoteCallsign:   c.RemoteCallsign,
			StartTime:        c.StartTime.Format(time.RFC3339),
			Duration:         int(c.Duration.Seconds()),
			DataRate:         c.DataRate,
			MaxElevationDeg:  c.MaxElevationDeg,
			Confidence:       c.Confidence,
			LightTimeDelay:   c.LightTimeDelay,
		}
		avgLightTimeDelay += c.LightTimeDelay
	}
	if len(contacts) > 0 {
		avgLightTimeDelay /= float64(len(contacts))
	} else {
		avgLightTimeDelay = 1.28 // Default Earth-Moon delay
	}

	// Band-specific parameters
	centerFreq := uint64(2200000000) // 2.2 GHz S-band
	sampleRate := uint64(2000000)    // 2 MHz
	txGain := 10.0                   // 10 dBi directional patch
	rxGain := 35.0                   // 35 dBi ground dish (3-5m)
	bandStr := "S-band 2.2 GHz"
	fecType := "LDPC"

	if config.Band == XBand {
		centerFreq = 8400000000 // 8.4 GHz X-band
		sampleRate = 10000000   // 10 MHz
		txGain = 15.0           // 15 dBi
		rxGain = 40.0           // 40 dBi
		bandStr = "X-band 8.4 GHz"
		fecType = "Turbo"
	}

	data := struct {
		NodeID           string
		NodeNumber       int
		Callsign         string
		StorageBytes     uint64
		SRAMBytes        uint64
		Band             string
		CenterFreq       uint64
		SampleRate       uint64
		TXGain           float64
		RXGain           float64
		FECType          string
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
		LightTimeDelay   float64
	}{
		NodeID:           config.NodeID,
		NodeNumber:       config.NodeNumber,
		Callsign:         config.Callsign,
		StorageBytes:     config.StorageBytes,
		SRAMBytes:        config.SRAMBytes,
		Band:             bandStr,
		CenterFreq:       centerFreq,
		SampleRate:       sampleRate,
		TXGain:           txGain,
		RXGain:           rxGain,
		FECType:          fecType,
		OrbitalParams:    config.OrbitalParams != nil,
		TelemetryEnabled: config.TelemetryEnabled,
		Contacts:         contacts,
		LightTimeDelay:   avgLightTimeDelay,
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

	filename := "cislunar.ionconfig"
	if config.Band == SBand {
		filename = "sband.ionconfig"
	} else if config.Band == XBand {
		filename = "xband.ionconfig"
	}

	return writeTemplate(filepath.Join(outputDir, filename), tmpl, data)
}

// GenerateCislunarWithCGRPrediction generates cislunar config with CGR-predicted contact windows
func GenerateCislunarWithCGRPrediction(
	nodeID string,
	nodeNumber int,
	callsign string,
	band CislunarBand,
	orbitalParams *OrbitalParameters,
	predictedContacts []CislunarContact,
	outputDir string,
) error {
	config := CislunarNodeConfig{
		NodeID:           nodeID,
		NodeNumber:       nodeNumber,
		Callsign:         callsign,
		StorageBytes:     512 * 1024 * 1024, // 512 MB for long-duration storage
		SRAMBytes:        786 * 1024,         // 786 KB SRAM (STM32U585) or more for higher-capability processor
		Band:             band,
		ContactPlan:      predictedContacts,
		OrbitalParams:    orbitalParams,
		TelemetryEnabled: true,
	}

	return GenerateCislunarConfig(config, outputDir)
}
