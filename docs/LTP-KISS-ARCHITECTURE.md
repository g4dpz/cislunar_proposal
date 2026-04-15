# LTP-over-KISS Architecture for Amateur Radio DTN

**Version:** 1.0  
**Date:** April 15, 2026  
**Status:** Proposed for Phase 1.5 (QO-100 GEO Satellite)

---

## Executive Summary

This document proposes a simplified DTN architecture for amateur radio that eliminates the AX.25 protocol layer and wraps LTP (Licklider Transmission Protocol) frames directly in KISS framing. Station identification is achieved through DTN Endpoint Identifiers (EIDs) containing amateur radio callsigns, ensuring regulatory compliance while reducing protocol overhead and implementation complexity.

---

## Motivation

### Current Architecture (Phase 1)

```
┌─────────────────────────────────────┐
│   Application (bping, bpsendfile)   │
├─────────────────────────────────────┤
│   BPv7 (Bundle Protocol)            │
│   EID: ipn:1.1                      │
├─────────────────────────────────────┤
│   BPSec (Integrity - HMAC-SHA-256)  │
├─────────────────────────────────────┤
│   LTP (Licklider Transmission)      │
├─────────────────────────────────────┤
│   AX.25 (Amateur Radio Link Layer)  │  ← 17+ byte overhead
│   Callsign: G4DPZ-1                 │
├─────────────────────────────────────┤
│   KISS (TNC Serial Protocol)        │
├─────────────────────────────────────┤
│   USB Serial (TNC4)                 │
├─────────────────────────────────────┤
│   G3RUH GFSK (9600 baud)            │
└─────────────────────────────────────┘
```

### Problems with Current Approach

1. **Redundant Addressing**: Both AX.25 and DTN provide addressing mechanisms
2. **Protocol Overhead**: AX.25 header adds 17+ bytes per frame
3. **Implementation Complexity**: AX.25 frame construction, FCS calculation, address parsing
4. **Semantic Mismatch**: AX.25 is designed for packet radio, not DTN bundles
5. **Limited Interoperability**: Not compatible with traditional packet radio anyway (DTN bundles are opaque)

---

## Proposed Architecture

### Simplified Stack

```
┌─────────────────────────────────────┐
│   Application (bping, bpsendfile)   │
├─────────────────────────────────────┤
│   BPv7 (Bundle Protocol)            │
│   EID: dtn://g4dpz-1                │  ← Callsign in EID
├─────────────────────────────────────┤
│   BPSec (Integrity - HMAC-SHA-256)  │
├─────────────────────────────────────┤
│   LTP (Licklider Transmission)      │
├─────────────────────────────────────┤
│   KISS (TNC Serial Framing)         │  ← LTP directly in KISS
├─────────────────────────────────────┤
│   USB Serial (TNC4)                 │
├─────────────────────────────────────┤
│   G3RUH GFSK (9600 baud)            │
└─────────────────────────────────────┘
```

### Key Changes

1. **Remove AX.25 Layer**: Eliminate protocol layer entirely
2. **LTP in KISS**: Wrap LTP segments directly in KISS frames
3. **DTN EID Addressing**: Use `dtn://callsign-ssid` format for station identification
4. **Native DTN Semantics**: No impedance mismatch between protocols

---

## Technical Specification

### 1. KISS Framing

KISS (Keep It Simple, Stupid) is a minimal framing protocol for serial communication with TNCs.

**Frame Structure:**
```
[FEND] [CMD] [DATA...] [FEND]

FEND = 0xC0 (frame boundary)
CMD  = 0x00 (data frame)
DATA = LTP segment bytes
```

**Byte Stuffing:**
- `0xC0` → `0xDB 0xDC` (FESC + TFEND)
- `0xDB` → `0xDB 0xDD` (FESC + TFESC)

**Example:**
```
KISS Frame containing LTP segment:
[C0] [00] [12 34 56 78 ... LTP data ...] [C0]
```

### 2. DTN Endpoint Identifiers (EIDs)

**Format:**
```
dtn://<callsign>-<ssid>[/<service>]
```

**Examples:**
```
dtn://g4dpz-1              # Primary station
dtn://g4dpz-2              # Secondary station  
dtn://w1abc-1              # Another station
dtn://g4dpz-1/ping         # Ping service endpoint
dtn://g4dpz-1/file         # File transfer endpoint
dtn://g4dpz-1/beacon       # Beacon service
```

**SSID (Secondary Station Identifier):**
- Range: 0-15 (following AX.25 convention)
- Allows multiple logical stations per callsign
- Example: `-1` for primary, `-2` for secondary

### 3. LTP Configuration

**LTP Segment Structure:**
```
┌──────────────────────────────────────┐
│ LTP Header                           │
│ - Version, Type, Flags               │
│ - Session ID                         │
│ - Serial Number                      │
├──────────────────────────────────────┤
│ LTP Data Segment                     │
│ - Client Service Data (Bundle)       │
├──────────────────────────────────────┤
│ LTP Trailer (optional)               │
│ - Checksum                           │
└──────────────────────────────────────┘
```

**ION-DTN Configuration:**
```
# ltprc configuration
a span 1 10 10 1400 10000 1 'dtn://g4dpz-1'
s 'dtn://w1abc-1'
```

### 4. Station Identification

**Primary Mechanism: EID in Every Bundle**

Every DTN bundle contains source and destination EIDs:
```
Bundle Primary Block:
- Source: dtn://g4dpz-1
- Destination: dtn://w1abc-1
- Creation Timestamp
- Lifetime
```

**Secondary Mechanism: Periodic Beacons**

Send identification beacon every 10 minutes (amateur radio requirement):
```
Source: dtn://g4dpz-1
Destination: dtn://beacon
Payload: "G4DPZ amateur radio DTN experimental station"
```

---

## Regulatory Compliance

### Amateur Radio Requirements

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **Station Identification** | Callsign in DTN EID (every bundle) | ✅ Compliant |
| **No Encryption** | BPSec integrity only (HMAC-SHA256) | ✅ Compliant |
| **Published Protocol** | RFCs 9171, 9172, 5326 + open source | ✅ Compliant |
| **Unobstructed Communication** | All protocols publicly documented | ✅ Compliant |
| **Periodic ID** | Beacon bundles every 10 minutes | ✅ Compliant |

### Protocol Documentation

All protocols are publicly available:

1. **Bundle Protocol v7**: [RFC 9171](https://www.rfc-editor.org/rfc/rfc9171.html)
2. **BPSec**: [RFC 9172](https://www.rfc-editor.org/rfc/rfc9172.html)
3. **LTP**: [RFC 5326](https://www.rfc-editor.org/rfc/rfc5326.html)
4. **KISS**: [KISS Protocol Specification](http://www.ax25.net/kiss.aspx)
5. **Implementation**: MIT licensed, public GitHub repository

### Precedent

This approach follows established amateur radio digital modes:

- **APRS**: Published protocol, callsign in packets
- **FT8/WSPR**: Published protocol, callsign in messages  
- **D-STAR**: Published protocol, callsign in headers
- **Winlink**: Published protocol, callsign in message headers

---

## Benefits

### 1. Reduced Overhead

**AX.25 Frame Overhead:**
```
Destination Address: 7 bytes
Source Address:      7 bytes
Control Field:       1 byte
PID:                 1 byte
FCS:                 2 bytes
Total:              18 bytes minimum
```

**KISS Frame Overhead:**
```
FEND (start):        1 byte
Command:             1 byte
FEND (end):          1 byte
Total:               3 bytes
```

**Savings:** 15 bytes per frame (~0.8% at 9600 baud)

### 2. Implementation Simplicity

**Removed Complexity:**
- AX.25 frame construction
- Address encoding (callsign → 7-byte format)
- FCS calculation (CRC-16)
- Control field handling
- PID field management

**Added Simplicity:**
- KISS framing is trivial (3 bytes + escaping)
- LTP segments pass through unchanged
- No protocol translation needed

### 3. Native DTN Semantics

- No impedance mismatch between AX.25 and DTN
- DTN addressing used end-to-end
- Cleaner architecture
- Easier to reason about

### 4. Bandwidth Efficiency

At 9600 baud (1200 bytes/sec):
- **Current**: ~18 bytes overhead per frame
- **Proposed**: ~3 bytes overhead per frame
- **Improvement**: 15 bytes × 8 bits = 120 bits saved per frame
- **At 10 frames/sec**: 1200 bits/sec saved = 10% throughput improvement

---

## Tradeoffs

### Advantages

✅ Simpler implementation  
✅ Reduced protocol overhead  
✅ Native DTN addressing  
✅ Cleaner architecture  
✅ Easier to debug  
✅ Better bandwidth efficiency  
✅ Regulatory compliant  

### Disadvantages

❌ No interoperability with traditional packet radio (but we didn't have this anyway)  
❌ No digipeater support (not needed for point-to-point satellite links)  
❌ Less familiar to amateur radio community (but DTN is novel regardless)  
❌ Requires custom tooling (but we're building that anyway)  

---

## Implementation Plan

### Phase 1: Terrestrial DTN (Refactoring)

Implement this architecture in Phase 1 first to validate the approach in a controlled terrestrial environment:

**Benefits of Phase 1 Implementation:**
1. **Controlled Environment**: Ground-based testing with direct access to hardware
2. **Easier Debugging**: Can use oscilloscopes, logic analyzers, serial monitors
3. **Rapid Iteration**: Quick feedback loop for testing and refinement
4. **Proven Hardware**: Existing Raspberry Pi + TNC4 + FT-817 setup
5. **Foundation for Phase 1.5**: Validates architecture before satellite deployment

**Phase 1 Validation Goals:**
- Verify KISS framing works correctly with LTP segments
- Confirm DTN EID addressing functions properly
- Validate regulatory compliance (callsign identification)
- Measure bandwidth improvement vs. AX.25 approach
- Test interoperability between two terrestrial nodes

### Phase 1.5: QO-100 GEO Satellite

After Phase 1 validation, deploy to Phase 1.5:

1. **Point-to-Point Link**: No digipeaters needed (direct to satellite)
2. **Narrowband Transponder**: Reduced overhead is valuable
3. **Geostationary**: No Doppler shift, simpler link management
4. **Proven Architecture**: Already validated in Phase 1 terrestrial testing
5. **Real Space Link**: First space-based DTN demonstration

### Implementation Steps (Phase 1)

1. **ION-DTN Configuration**
   - Configure LTP convergence layer for KISS serial device
   - Map DTN EIDs to serial ports
   - Example: `dtn://g4dpz-1` → `/dev/ttyUSB0`
   - Update node-a and node-b configurations

2. **KISS Interface**
   - Remove AX.25 frame construction from existing code
   - Implement simple KISS framing (FEND + CMD + DATA + FEND)
   - Add byte stuffing for 0xC0 and 0xDB
   - Modify TNC4 interface to use KISS directly

3. **EID Management**
   - Configure source EID: `dtn://g4dpz-1` (node-a)
   - Configure destination EID: `dtn://w1abc-1` (node-b)
   - Implement beacon service for periodic identification
   - Update application layer to use dtn:// URIs

4. **Testing**
   - Loopback test: Verify KISS framing with single TNC4
   - Two-node test: Validate over VHF/UHF between node-a and node-b
   - Performance test: Measure throughput improvement vs. AX.25
   - Compliance test: Verify callsign identification in all bundles
   - Extended duration test: 24-hour stability test

5. **Documentation**
   - Update Phase 1 documentation with new architecture
   - Document configuration changes
   - Update quick start guide
   - Add troubleshooting section

### Code Changes (Phase 1)

**Remove:**
- `ax25/ax25.go` (AX.25 frame handling) - 200+ lines
- `ax25/ax25_test.go` (AX.25 tests) - 100+ lines

**Modify:**
- `pkg/ion/ltp.go` (LTP → KISS instead of LTP → AX.25 → KISS)
- `cmd/dtn-node/main.go` (EID configuration, remove AX.25 setup)
- `configs/node-a/*.ionrc` (Update to use dtn:// EIDs)
- `configs/node-b/*.ionrc` (Update to use dtn:// EIDs)

**Add:**
- `pkg/kiss/kiss.go` (Simple KISS framing - ~100 lines)
- `pkg/kiss/kiss_test.go` (KISS tests)
- `pkg/eid/eid.go` (EID parsing and validation - ~50 lines)
- `pkg/eid/eid_test.go` (EID tests)

**Net Result:** Simpler codebase (~150 lines removed, ~150 lines added, but much simpler)

---

## Example Configuration

### ION-DTN Configuration Files

**node.ionrc:**
```
1 1 ''
a contact +0 +3600 1 2 100000
a range +0 +3600 1 2 1
m production 1000000
m consumption 1000000
```

**node.ltprc:**
```
1 32
a span 1 10 10 1400 10000 1 'dtn://g4dpz-1'
s 'dtn://w1abc-1'
```

**node.bprc:**
```
1
a scheme dtn 'dtnpn' 'dtndeliver'
a endpoint dtn://g4dpz-1 q
a protocol ltp 1400 100
a induct ltp 1 ltpcli
a outduct ltp 1 ltpclo
s
```

### Application Usage

**Ping:**
```bash
bping dtn://g4dpz-1 dtn://w1abc-1 -c 5
```

**File Transfer:**
```bash
bpsendfile dtn://g4dpz-1 dtn://w1abc-1/file message.txt
```

**Beacon:**
```bash
# Automated beacon every 10 minutes
bpsendfile dtn://g4dpz-1 dtn://beacon beacon.txt
```

---

## Future Considerations

### Phase 1: Terrestrial Validation

**Success Criteria:**
- ✅ KISS framing works reliably with LTP segments
- ✅ DTN EID addressing functions correctly
- ✅ Throughput improvement measured (target: 10% increase)
- ✅ Regulatory compliance verified (callsign in all bundles)
- ✅ 24-hour stability test passes

**If Successful:** Proceed to Phase 1.5 (QO-100) with confidence

**If Issues Found:** Iterate in Phase 1 terrestrial environment before satellite deployment

### Phase 1.5-4: Deployment

This architecture can be used for:

- ✅ **Phase 1**: Terrestrial validation (CURRENT)
- ✅ **Phase 1.5 (QO-100)**: Geostationary satellite demonstration
- ✅ **Phase 2 (EM)**: Ground-based flatsat testing
- ✅ **Phase 3 (LEO)**: Orbital CubeSat (point-to-point)
- ✅ **Phase 4 (Cislunar)**: Deep space links (point-to-point)

### Potential Extensions

1. **Multi-hop DTN**: Use DTN routing instead of AX.25 digipeaters
2. **Service Discovery**: Advertise available services via beacon bundles
3. **QoS Mapping**: Map DTN priority classes to link-layer priorities
4. **Adaptive Coding**: Adjust LTP parameters based on link quality

---

## References

1. **RFC 9171**: Bundle Protocol Version 7 (BPv7)
2. **RFC 9172**: Bundle Protocol Security (BPSec)
3. **RFC 5326**: Licklider Transmission Protocol (LTP)
4. **KISS Protocol**: http://www.ax25.net/kiss.aspx
5. **ION-DTN**: https://sourceforge.net/projects/ion-dtn/
6. **Amateur Radio Regulations**: FCC Part 97 (US), OfCom (UK)

---

## Conclusion

Wrapping LTP directly in KISS frames and using DTN EIDs for station identification provides a simpler, more efficient, and regulatory-compliant architecture for amateur radio DTN. This approach eliminates unnecessary protocol layers, reduces overhead, and aligns with DTN's native semantics while maintaining full amateur radio compliance through callsign-embedded EIDs and published protocols.

**Recommendation:** 
1. **Phase 1 (Terrestrial)**: Implement and validate this architecture in the existing terrestrial setup to prove the concept in a controlled environment
2. **Phase 1.5 (QO-100)**: Deploy to geostationary satellite after successful Phase 1 validation
3. **Phase 2-4**: Continue using this architecture for all subsequent phases

**Implementation Priority:** HIGH - This refactoring simplifies the codebase and improves efficiency across all phases.

---

**Document Status:** Proposed for Phase 1 Implementation  
**Next Steps:** 
1. Create Phase 1 refactoring spec/tasks
2. Implement KISS and EID packages
3. Remove AX.25 dependencies
4. Test and validate in terrestrial environment
5. Deploy to Phase 1.5 after validation

**Contact:** G4DPZ
