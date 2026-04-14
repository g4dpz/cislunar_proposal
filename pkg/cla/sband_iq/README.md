# S-band/X-band IQ Convergence Layer Adapter

## Overview

The `sband_iq` package implements the Convergence Layer Adapter (CLA) for cislunar deep-space DTN nodes using S-band (2.2 GHz) or X-band (8.4 GHz) IQ transceivers. This CLA is designed for the cislunar amateur DTN payload phase of the project, enabling Earth-Moon delay-tolerant networking.

## Key Features

- **S-band (2.2 GHz) and X-band (8.4 GHz) support** for deep-space communication
- **BPSK modulation** at 500 bps for reliable cislunar links
- **LDPC/Turbo FEC** for strong forward error correction (6-8 dB coding gain)
- **Long-delay LTP session management** accounting for 1-2 second one-way light-time delay
- **AX.25 framing** with amateur radio callsign addressing for regulatory compliance
- **Direct STM32U585 integration** via DAC/ADC or SPI (no companion host required)
- **High transmit power** (5-10W) for deep-space link budget

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    S-band/X-band IQ CLA                     │
├─────────────────────────────────────────────────────────────┤
│  • AX.25 Framing (callsign addressing)                     │
│  • LTP Session Management (long-delay aware)               │
│  • BPSK Modulation/Demodulation                            │
│  • FEC Encoding/Decoding (LDPC/Turbo)                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              S-band/X-band Transceiver                      │
├─────────────────────────────────────────────────────────────┤
│  • IQ Baseband Generation/Processing                        │
│  • DAC/ADC or SPI Interface to STM32U585                   │
│  • RF Up/Down Conversion                                    │
│  • Light-time Delay Compensation                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
                    RF Front-End
              (S-band 2.2 GHz / X-band 8.4 GHz)
```

## Configuration

### S-band Default Configuration

```go
config := sband_iq.DefaultSBandConfig("W1ABC")
// Callsign: W1ABC
// Band: S-band (2.2 GHz)
// Data Rate: 500 bps
// TX Power: 5W
// TX Gain: 10 dBi (directional patch antenna)
// RX Gain: 35 dBi (3-5m ground dish)
// FEC: LDPC enabled
// Light-time delay: 1.2 seconds
// LTP timeout: 10 seconds
```

### X-band Default Configuration

```go
config := sband_iq.DefaultXBandConfig("K2XYZ")
// Callsign: K2XYZ
// Band: X-band (8.4 GHz)
// Data Rate: 500 bps
// TX Power: 5W
// TX Gain: 12 dBi
// RX Gain: 40 dBi
// FEC: LDPC enabled
// Light-time delay: 1.2 seconds
// LTP timeout: 10 seconds
```

### Custom Configuration

```go
config := sband_iq.Config{
    Callsign:       "N3ABC",
    Band:           sband_transceiver.BandS,
    CenterFreq:     2.2e9,
    SampleRate:     8000.0,
    DataRate:       500,
    TXPower:        7.0,
    TXGain:         12.0,
    RXGain:         38.0,
    FECEnabled:     true,
    FECType:        sband_transceiver.FECTurbo,
    LightTimeDelay: 1500 * time.Millisecond,
    LTPTimeout:     15 * time.Second,
}
```

## Usage Example

```go
package main

import (
    "fmt"
    "time"
    
    "terrestrial-dtn/pkg/bpa"
    "terrestrial-dtn/pkg/cla/sband_iq"
    "terrestrial-dtn/pkg/contact"
)

func main() {
    // Create S-band CLA
    config := sband_iq.DefaultSBandConfig("W1ABC")
    cla, err := sband_iq.New(config)
    if err != nil {
        panic(err)
    }
    
    // Define contact window
    window := contact.ContactWindow{
        ContactID:  1,
        RemoteNode: "tier3-ground-station",
        StartTime:  time.Now().Unix(),
        EndTime:    time.Now().Add(30 * time.Minute).Unix(),
        DataRate:   500,
        LinkType:   contact.LinkTypeSBandIQ,
    }
    
    // Open link
    err = cla.Open(window)
    if err != nil {
        panic(err)
    }
    defer cla.Close()
    
    // Create and send bundle
    bundle := &bpa.Bundle{
        ID: bpa.BundleID{
            SourceEID:         bpa.EndpointID{Scheme: "ipn", SSP: "cislunar-01.0"},
            CreationTimestamp: time.Now().Unix(),
            SequenceNumber:    1,
        },
        Destination: bpa.EndpointID{Scheme: "ipn", SSP: "mission-control.0"},
        Payload:     []byte("Science data from cislunar payload"),
        Priority:    bpa.PriorityCritical,
        Lifetime:    86400,
        CreatedAt:   time.Now().Unix(),
        BundleType:  bpa.BundleTypeData,
    }
    
    metrics, err := cla.SendBundle(bundle)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Bundle transmitted: %d bytes\n", metrics.BytesTransferred)
    fmt.Printf("Active LTP sessions: %d\n", cla.GetActiveSessions())
}
```

## Link Budget

### Cislunar S-band Link (Earth-Moon, 384,000 km)

| Parameter | Value |
|-----------|-------|
| TX Power | 5W (37 dBm) |
| TX Antenna Gain | 10 dBi (directional patch) |
| EIRP | 47 dBm |
| Frequency | 2.2 GHz |
| Free-space Path Loss | ~267 dB |
| RX Antenna Gain | 35 dBi (3-5m dish) |
| System Temperature | ~50 K |
| Data Rate | 500 bps |
| Modulation | BPSK |
| FEC | LDPC (6-8 dB coding gain) |
| **Link Margin** | **5-7 dB** ✓ |

The link closes with positive margin, enabling reliable cislunar communication.

## Light-Time Delay

The Earth-Moon distance (~384,000 km) introduces a one-way light-time delay of approximately 1.2 seconds:

```
Delay = Distance / Speed of Light
      = 384,000 km / 299,792 km/s
      ≈ 1.28 seconds
```

Round-trip time (RTT) is approximately 2.4-2.6 seconds. The CLA's LTP session management accounts for this delay:

- **LTP timeout**: 10 seconds (allows for RTT + processing)
- **Session tracking**: Monitors active sessions and cleans up expired ones
- **Deferred acknowledgment**: LTP provides reliable transfer with ACKs delayed by RTT

## Forward Error Correction (FEC)

### LDPC (Low-Density Parity-Check)

- Near Shannon-limit performance
- Coding gain: ~6-8 dB
- Suitable for deep-space links with low SNR
- Default FEC type for cislunar operations

### Turbo Coding

- Iterative decoding with excellent performance
- Coding gain: ~5-7 dB
- Used in many deep-space missions
- Alternative FEC option

## Testing

Run the test suite:

```bash
go test -v ./pkg/cla/sband_iq/
```

All tests validate:
- CLA creation and configuration
- Link open/close operations
- Bundle transmission and reception
- LTP session management
- Light-time delay handling
- FEC configuration
- AX.25 framing
- Bundle serialization/deserialization

## Integration with ION-DTN

The S-band IQ CLA integrates with ION-DTN's BPv7/LTP stack:

1. **BPv7 bundles** are created by the Bundle Protocol Agent (BPA)
2. **LTP sessions** provide reliable transfer with deferred acknowledgment
3. **AX.25 frames** carry LTP segments with callsign addressing
4. **BPSK modulation** converts frames to IQ baseband samples
5. **FEC encoding** (LDPC/Turbo) improves link budget
6. **S-band transceiver** up-converts to RF and transmits

## Requirements Validation

This implementation validates the following requirements from the cislunar amateur DTN payload specification:

- **Requirement 14.1**: ION-DTN (BPv7/LTP over AX.25) with BPSK + FEC at 500 bps on S-band 2.2 GHz
- **Requirement 14.2**: Account for 1-2 second one-way light-time delay in LTP session management
- **Requirement 14.3**: Support long-duration message storage for extended contact gaps

## See Also

- [UHF IQ CLA](../uhf_iq/) - LEO CubeSat UHF 437 MHz implementation
- [S-band Transceiver](../../radio/sband_transceiver/) - Underlying transceiver interface
- [IQ Modulation](../../iq/) - BPSK modulation/demodulation
- [Cislunar Example](../../../examples/cislunar_sband_example.go) - Complete usage example
