package hdtnconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// --- Generators ---

// genNonEmptyString generates a non-empty ASCII string suitable for config fields.
func genNonEmptyString() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9/_\-\.]{1,50}`)
}

// genPositiveInt generates a positive integer (1 to 10000).
func genPositiveInt() *rapid.Generator[int] {
	return rapid.IntRange(1, 10000)
}

// genPort generates a valid port number.
func genPort() *rapid.Generator[int] {
	return rapid.IntRange(1024, 65535)
}

// genBaudRate generates a valid baud rate.
func genBaudRate() *rapid.Generator[int] {
	return rapid.SampledFrom([]int{1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200})
}

// genContactDataRate generates a positive data rate.
func genContactDataRate() *rapid.Generator[int64] {
	return rapid.Int64Range(1, 1000000)
}

// genInduct generates a valid Induct struct.
func genInduct(t *rapid.T) Induct {
	cl := rapid.SampledFrom([]string{"ltp_over_udp", "kiss"}).Draw(t, "convergenceLayer")
	induct := Induct{
		ConvergenceLayer: cl,
		Name:             genNonEmptyString().Draw(t, "inductName"),
	}
	if cl == "ltp_over_udp" {
		induct.BoundPort = genPort().Draw(t, "boundPort")
		induct.ThisLtpEngineID = uint64(genPositiveInt().Draw(t, "thisLtpEngine"))
		induct.RemoteLtpEngineID = uint64(genPositiveInt().Draw(t, "remoteLtpEngine"))
	} else {
		induct.KissTncDevice = genNonEmptyString().Draw(t, "kissTncDevice")
		induct.KissBaudRate = genBaudRate().Draw(t, "kissBaudRate")
		induct.ThisLtpEngineID = uint64(genPositiveInt().Draw(t, "thisLtpEngine"))
		induct.RemoteLtpEngineID = uint64(genPositiveInt().Draw(t, "remoteLtpEngine"))
	}
	return induct
}

// genOutduct generates a valid Outduct struct.
func genOutduct(t *rapid.T) Outduct {
	cl := rapid.SampledFrom([]string{"ltp_over_udp", "kiss"}).Draw(t, "convergenceLayer")
	outduct := Outduct{
		ConvergenceLayer: cl,
		Name:             genNonEmptyString().Draw(t, "outductName"),
		NextHopNodeID:    genPositiveInt().Draw(t, "nextHopNodeId"),
	}
	if cl == "ltp_over_udp" {
		outduct.RemoteHostname = genNonEmptyString().Draw(t, "remoteHostname")
		outduct.RemotePort = genPort().Draw(t, "remotePort")
		outduct.ThisLtpEngineID = uint64(genPositiveInt().Draw(t, "thisLtpEngine"))
		outduct.RemoteLtpEngineID = uint64(genPositiveInt().Draw(t, "remoteLtpEngine"))
	} else {
		outduct.KissTncDevice = genNonEmptyString().Draw(t, "kissTncDevice")
		outduct.KissBaudRate = genBaudRate().Draw(t, "kissBaudRate")
		outduct.ThisLtpEngineID = uint64(genPositiveInt().Draw(t, "thisLtpEngine"))
		outduct.RemoteLtpEngineID = uint64(genPositiveInt().Draw(t, "remoteLtpEngine"))
	}
	return outduct
}

// genContactEntry generates a valid ContactEntry.
func genContactEntry(t *rapid.T) ContactEntry {
	start := rapid.Int64Range(0, 100000).Draw(t, "startTime")
	end := rapid.Int64Range(start+1, start+86400).Draw(t, "endTime")
	return ContactEntry{
		Source:         genPositiveInt().Draw(t, "source"),
		Dest:           genPositiveInt().Draw(t, "dest"),
		StartTime:      start,
		EndTime:        end,
		RateBitsPerSec: genContactDataRate().Draw(t, "rateBitsPerSec"),
	}
}

// genValidHDTNConfig generates an arbitrary valid HDTNConfig.
func genValidHDTNConfig(t *rapid.T) *HDTNConfig {
	nodeID := genPositiveInt().Draw(t, "nodeId")
	numInducts := rapid.IntRange(1, 5).Draw(t, "numInducts")
	numOutducts := rapid.IntRange(1, 5).Draw(t, "numOutducts")
	numContacts := rapid.IntRange(1, 5).Draw(t, "numContacts")

	inducts := make([]Induct, numInducts)
	for i := range inducts {
		inducts[i] = genInduct(t)
	}

	outducts := make([]Outduct, numOutducts)
	for i := range outducts {
		outducts[i] = genOutduct(t)
	}

	contacts := make([]ContactEntry, numContacts)
	for i := range contacts {
		contacts[i] = genContactEntry(t)
	}

	// Optionally generate demux services
	var demuxServices []string
	if rapid.Bool().Draw(t, "hasDemuxServices") {
		n := rapid.IntRange(1, 3).Draw(t, "numDemuxServices")
		demuxServices = make([]string, n)
		for i := range demuxServices {
			demuxServices[i] = genNonEmptyString().Draw(t, "demuxService")
		}
	}

	return &HDTNConfig{
		HDTNConfigName:     genNonEmptyString().Draw(t, "configName"),
		MyNodeID:           nodeID,
		MySchemeStr:        "ipn",
		MyDtnEidStr:        fmt.Sprintf("ipn:%d.0", nodeID),
		MyDtnDemuxServices: demuxServices,
		StoragePath:        genNonEmptyString().Draw(t, "storagePath"),
		InductsConfig:      InductsConfig{InductVector: inducts},
		OutductsConfig:     OutductsConfig{OutductVector: outducts},
		ContactPlanJSON:    ContactPlanJSON{Contacts: contacts},
	}
}

// genValidTerrestrialOpts generates arbitrary valid TerrestrialOpts.
func genValidTerrestrialOpts(t *rapid.T) TerrestrialOpts {
	nodeNum := genPositiveInt().Draw(t, "nodeNumber")
	remoteNum := genPositiveInt().Draw(t, "remoteNodeNumber")
	// Ensure remote != local
	for remoteNum == nodeNum {
		remoteNum = genPositiveInt().Draw(t, "remoteNodeNumber2")
	}
	return TerrestrialOpts{
		NodeNumber:       nodeNum,
		NodeName:         genNonEmptyString().Draw(t, "nodeName"),
		Callsign:         genNonEmptyString().Draw(t, "callsign"),
		StoragePath:      genNonEmptyString().Draw(t, "storagePath"),
		TNCDevice:        genNonEmptyString().Draw(t, "tncDevice"),
		TNCBaudRate:      genBaudRate().Draw(t, "tncBaudRate"),
		UDPLocalPort:     genPort().Draw(t, "udpLocalPort"),
		UDPRemoteHost:    genNonEmptyString().Draw(t, "udpRemoteHost"),
		UDPRemotePort:    genPort().Draw(t, "udpRemotePort"),
		RemoteNodeNumber: remoteNum,
		ContactDataRate:  genContactDataRate().Draw(t, "contactDataRate"),
	}
}

// --- Property Tests ---

// TestProperty1_SerializationRoundTrip verifies that any valid HDTNConfig
// survives a JSON serialize/deserialize round-trip with deep equality.
// Feature: hdtn-migration, Property 1: Configuration serialization round-trip
// **Validates: Requirements 4.5, 10.3**
func TestProperty1_SerializationRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := genValidHDTNConfig(t)

		// Serialize to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}

		// Deserialize back
		var restored HDTNConfig
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}

		// Assert deep equality
		if !reflect.DeepEqual(*original, restored) {
			t.Fatalf("round-trip failed:\noriginal: %+v\nrestored: %+v", *original, restored)
		}
	})
}

// TestProperty2_TerrestrialConfigStructuralCompleteness verifies that
// GenerateTerrestrialConfig produces structurally complete configs.
// Feature: hdtn-migration, Property 2: Generated terrestrial config structural completeness
// **Validates: Requirements 4.1, 4.2, 4.3**
func TestProperty2_TerrestrialConfigStructuralCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		opts := genValidTerrestrialOpts(t)

		cfg, err := GenerateTerrestrialConfig(opts)
		if err != nil {
			t.Fatalf("GenerateTerrestrialConfig failed: %v", err)
		}

		// Non-empty EID
		if cfg.MyDtnEidStr == "" {
			t.Fatal("myDtnEidStr must not be empty")
		}

		// Non-empty storage path
		if cfg.StoragePath == "" {
			t.Fatal("storagePath must not be empty")
		}

		// At least one LTP-over-UDP induct
		hasLtpInduct := false
		for _, ind := range cfg.InductsConfig.InductVector {
			if ind.ConvergenceLayer == "ltp_over_udp" {
				hasLtpInduct = true
				break
			}
		}
		if !hasLtpInduct {
			t.Fatal("config must contain at least one LTP-over-UDP induct")
		}

		// At least one KISS CLA induct
		hasKissInduct := false
		for _, ind := range cfg.InductsConfig.InductVector {
			if ind.ConvergenceLayer == "kiss" {
				hasKissInduct = true
				break
			}
		}
		if !hasKissInduct {
			t.Fatal("config must contain at least one KISS CLA induct")
		}

		// At least one LTP-over-UDP outduct
		hasLtpOutduct := false
		for _, out := range cfg.OutductsConfig.OutductVector {
			if out.ConvergenceLayer == "ltp_over_udp" {
				hasLtpOutduct = true
				break
			}
		}
		if !hasLtpOutduct {
			t.Fatal("config must contain at least one LTP-over-UDP outduct")
		}

		// At least one KISS CLA outduct
		hasKissOutduct := false
		for _, out := range cfg.OutductsConfig.OutductVector {
			if out.ConvergenceLayer == "kiss" {
				hasKissOutduct = true
				break
			}
		}
		if !hasKissOutduct {
			t.Fatal("config must contain at least one KISS CLA outduct")
		}

		// At least one contact plan entry
		if len(cfg.ContactPlanJSON.Contacts) == 0 {
			t.Fatal("config must contain at least one contact plan entry")
		}
	})
}

// TestProperty3_ValidationRejectsInvalidConfigs verifies that Validate()
// returns an error for configs with at least one invalid field.
// Feature: hdtn-migration, Property 3: Configuration validation rejects invalid configs
// **Validates: Requirements 4.6**
func TestProperty3_ValidationRejectsInvalidConfigs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Start with a valid config
		cfg := genValidHDTNConfig(t)

		// Choose which field to invalidate
		invalidField := rapid.IntRange(0, 4).Draw(t, "invalidField")

		var expectedFieldName string
		switch invalidField {
		case 0:
			// Invalid node ID (≤ 0)
			cfg.MyNodeID = rapid.IntRange(-100, 0).Draw(t, "badNodeId")
			expectedFieldName = "myNodeId"
		case 1:
			// Empty storage path
			cfg.StoragePath = ""
			expectedFieldName = "storagePath"
		case 2:
			// Empty inducts
			cfg.InductsConfig.InductVector = nil
			expectedFieldName = "inductVector"
		case 3:
			// Empty outducts
			cfg.OutductsConfig.OutductVector = nil
			expectedFieldName = "outductVector"
		case 4:
			// Empty contacts
			cfg.ContactPlanJSON.Contacts = nil
			expectedFieldName = "contacts"
		}

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("expected validation error for invalid field %q, got nil", expectedFieldName)
		}

		if !strings.Contains(err.Error(), expectedFieldName) {
			t.Fatalf("expected error to contain %q, got: %v", expectedFieldName, err)
		}
	})
}

// TestProperty4_NodeEIDFormatCorrectness verifies that for any positive
// integer N, the generated config has myDtnEidStr == "ipn:<N>.0".
// Feature: hdtn-migration, Property 4: Node EID format correctness
// **Validates: Requirements 4.4**
func TestProperty4_NodeEIDFormatCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 10000).Draw(t, "nodeNumber")

		opts := TerrestrialOpts{
			NodeNumber:       n,
			NodeName:         "test-node",
			Callsign:         "W1AW",
			StoragePath:      "/tmp/storage",
			TNCDevice:        "/dev/ttyUSB0",
			TNCBaudRate:      9600,
			UDPLocalPort:     4556,
			UDPRemoteHost:    "192.168.1.2",
			UDPRemotePort:    4557,
			RemoteNodeNumber: n + 1, // ensure different from node number
			ContactDataRate:  9600,
		}

		cfg, err := GenerateTerrestrialConfig(opts)
		if err != nil {
			t.Fatalf("GenerateTerrestrialConfig failed for node %d: %v", n, err)
		}

		expected := fmt.Sprintf("ipn:%d.0", n)
		if cfg.MyDtnEidStr != expected {
			t.Fatalf("expected EID %q, got %q", expected, cfg.MyDtnEidStr)
		}
	})
}
