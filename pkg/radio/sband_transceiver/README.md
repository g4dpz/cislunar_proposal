# S-band/X-band IQ Transceiver

This package provides an interface to the S-band/X-band IQ transceiver for cislunar deep-space communication nodes.

## Overview

The S-band/X-band transceiver supports:
- **S-band**: 2.2 GHz carrier frequency (2.0-2.3 GHz range)
- **X-band**: 8.4 GHz carrier frequency (8.0-8.5 GHz range)
- **Data rate**: 500 bps with BPSK modulation
- **FEC**: LDPC or Turbo forward error correction coding
- **Power**: 5-10W transmit power for deep-space link
- **Delay compensation**: 1-2 second one-way light-time delay handling

## Architecture

The transceiver interfaces directly with the STM32U585 (or higher-capability processor) via:
- **DAC/ADC interface**: Analog IQ sample streaming
- **SPI interface**: Digital IQ sample streaming

The processor generates TX IQ samples and processes RX IQ samples via DMA, providing full software-defined control over modulation/demodulation.

## Link Budget

The default S-band configuration supports cislunar link budget requirements:
- **TX power**: 5W
- **TX antenna gain**: 10 dBi (directional patch)
- **RX antenna gain**: 35 dBi (3-5m ground dish)
- **Link margin**: 5-7 dB at 500 bps with BPSK + LDPC/Turbo FEC
- **Distance**: 384,400 km (Earth-Moon)

## Usage

### Basic Setup

```go
import "terrestrial-dtn/pkg/radio/sband_transceiver"

// Create S-band transceiver with default configuration
config := sband_transceiver.DefaultSBandConfig()
transceiver, err := sband_transceiver.New(config)
if err != nil {
    log.Fatal(err)
}

// Open the transceiver
if err := transceiver.Open(); err != nil {
    log.Fatal(err)
}
defer transceiver.Close()
```

### X-band Configuration

```go
// Create X-band transceiver
config := sband_transceiver.DefaultXBandConfig()
transceiver, err := sband_transceiver.New(config)
```

### Streaming IQ Samples

```go
// Start streaming
if err := transceiver.StartStreaming(); err != nil {
    log.Fatal(err)
}
defer transceiver.StopStreaming()

// Transmit IQ buffer
buffer := iq.NewIQBuffer(1024, config.SampleRate)
// ... populate buffer with IQ samples ...
if err := transceiver.Transmit(buffer); err != nil {
    log.Printf("TX error: %v", err)
}

// Receive IQ buffer (accounts for light-time delay)
rxBuffer, err := transceiver.Receive()
if err != nil {
    log.Printf("RX error: %v", err)
} else {
    // Process received IQ samples
}
```

### Configuration

```go
// Set center frequency (S-band: 2.0-2.3 GHz, X-band: 8.0-8.5 GHz)
transceiver.SetCenterFreq(2.2e9) // 2.2 GHz

// Set sample rate (4-16 kHz for 500 bps)
transceiver.SetSampleRate(8000.0) // 8 kHz

// Set transmit power (1-10W)
transceiver.SetTXPower(5.0) // 5W

// Set gains
transceiver.SetTXGain(10.0) // 10 dBi
transceiver.SetRXGain(35.0) // 35 dBi

// Set light-time delay (1-3 seconds)
transceiver.SetLightTimeDelay(1200 * time.Millisecond) // 1.2s
```

### Forward Error Correction

```go
// Enable LDPC FEC
transceiver.EnableFEC(sband_transceiver.FECLDPC)

// Enable Turbo FEC
transceiver.EnableFEC(sband_transceiver.FECTurbo)

// Disable FEC
transceiver.DisableFEC()

// Check FEC status
if transceiver.IsFECEnabled() {
    fmt.Printf("FEC type: %s\n", transceiver.GetFECType())
}
```

### Link Quality Metrics

```go
// Get link metrics
metrics := transceiver.GetLinkMetrics()
fmt.Printf("RSSI: %.1f dBm\n", metrics.RSSI)
fmt.Printf("SNR: %.1f dB\n", metrics.SNR)
fmt.Printf("EVM: %.1f%%\n", metrics.EVM)
fmt.Printf("Frequency error: %.1f Hz\n", metrics.FreqError)
```

## Light-Time Delay

The transceiver accounts for 1-2 second one-way light-time delay in cislunar operations:

- **Earth-Moon distance**: ~384,400 km
- **One-way delay**: ~1.2 seconds
- **Round-trip time**: ~2.4 seconds

The `Receive()` method automatically extends its timeout to account for the configured light-time delay, ensuring proper reception of delayed signals.

## Comparison with LEO IQ Transceiver

| Feature | LEO IQ Transceiver | S-band/X-band Transceiver |
|---------|-------------------|---------------------------|
| Frequency | UHF 437 MHz | S-band 2.2 GHz / X-band 8.4 GHz |
| Data rate | 9.6 kbps | 500 bps |
| Modulation | GMSK/BPSK | BPSK |
| FEC | Optional | Required (LDPC/Turbo) |
| TX power | 2W | 5-10W |
| Distance | 500 km (LEO) | 384,400 km (cislunar) |
| Delay | <1ms | 1-2 seconds |
| Link margin | 31 dB | 5-7 dB |

## Requirements

This implementation satisfies the following requirements from the cislunar DTN payload specification:

- **Requirement 14.1**: S-band 2.2 GHz (or X-band 8.4 GHz) operation with BPSK + LDPC/Turbo FEC at 500 bps
- **Requirement 14.2**: 1-2 second one-way light-time delay compensation

## Testing

Run the test suite:

```bash
go test -v ./pkg/radio/sband_transceiver/
```

The test suite validates:
- Configuration and initialization
- Streaming operations
- Parameter validation and range checking
- FEC control
- Link budget requirements
- Light-time delay handling

## Implementation Notes

This is a simulation implementation for development and testing. In a real flight implementation:

1. **Hardware interface**: Initialize DAC/ADC or SPI interface to the transceiver IC
2. **Register configuration**: Configure transceiver IC registers for frequency, gain, power
3. **FEC encoding/decoding**: Implement LDPC or Turbo coding algorithms
4. **DMA streaming**: Set up DMA channels for continuous IQ sample streaming
5. **Power management**: Implement power control for 5-10W transmit power
6. **Timing compensation**: Account for light-time delay in protocol timing

## See Also

- `pkg/radio/iq_transceiver/` - LEO UHF IQ transceiver (437 MHz, 9.6 kbps)
- `pkg/iq/` - IQ baseband sample types and modulation
- `pkg/cla/sband_iq/` - S-band convergence layer adapter (to be implemented)
