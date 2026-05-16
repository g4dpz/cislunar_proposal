package hdtnconfig

import "fmt"

// GenerateTerrestrialConfig creates a complete HDTN config for a terrestrial node.
// It includes LTP-over-UDP and KISS CLA convergence layers, plus a contact plan entry.
func GenerateTerrestrialConfig(opts TerrestrialOpts) (*HDTNConfig, error) {
	cfg := &HDTNConfig{
		HDTNConfigName: opts.NodeName,
		MyNodeID:       opts.NodeNumber,
		MySchemeStr:    "ipn",
		MyDtnEidStr:    fmt.Sprintf("ipn:%d.0", opts.NodeNumber),
		StoragePath:    opts.StoragePath,
		InductsConfig: InductsConfig{
			InductVector: []Induct{
				{
					ConvergenceLayer:  "ltp_over_udp",
					Name:              fmt.Sprintf("%s-udp-induct", opts.NodeName),
					BoundPort:         opts.UDPLocalPort,
					ThisLtpEngineID:   uint64(opts.NodeNumber),
					RemoteLtpEngineID: uint64(opts.RemoteNodeNumber),
				},
				{
					ConvergenceLayer:  "kiss",
					Name:              fmt.Sprintf("%s-kiss-induct", opts.NodeName),
					KissTncDevice:     opts.TNCDevice,
					KissBaudRate:      opts.TNCBaudRate,
					ThisLtpEngineID:   uint64(opts.NodeNumber),
					RemoteLtpEngineID: uint64(opts.RemoteNodeNumber),
				},
			},
		},
		OutductsConfig: OutductsConfig{
			OutductVector: []Outduct{
				{
					ConvergenceLayer:  "ltp_over_udp",
					Name:              fmt.Sprintf("%s-udp-outduct", opts.NodeName),
					NextHopNodeID:     opts.RemoteNodeNumber,
					RemoteHostname:    opts.UDPRemoteHost,
					RemotePort:        opts.UDPRemotePort,
					ThisLtpEngineID:   uint64(opts.NodeNumber),
					RemoteLtpEngineID: uint64(opts.RemoteNodeNumber),
				},
				{
					ConvergenceLayer:  "kiss",
					Name:              fmt.Sprintf("%s-kiss-outduct", opts.NodeName),
					NextHopNodeID:     opts.RemoteNodeNumber,
					KissTncDevice:     opts.TNCDevice,
					KissBaudRate:      opts.TNCBaudRate,
					ThisLtpEngineID:   uint64(opts.NodeNumber),
					RemoteLtpEngineID: uint64(opts.RemoteNodeNumber),
				},
			},
		},
		ContactPlanJSON: ContactPlanJSON{
			Contacts: []ContactEntry{
				{
					Source:         opts.NodeNumber,
					Dest:           opts.RemoteNodeNumber,
					StartTime:      0,
					EndTime:        86400,
					RateBitsPerSec: opts.ContactDataRate,
				},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("generated config validation failed: %w", err)
	}

	return cfg, nil
}
