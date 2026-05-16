# DTN Callsign EID Configuration for HDTN

**Version:** 2.0  
**Date:** May 16, 2026  
**Status:** Reference Documentation

---

## Overview

HDTN (NASA Glenn's High-rate Delay Tolerant Networking) supports two Endpoint Identifier (EID) schemes:

| Scheme | Format | Routing | Wire Size | Use Case |
|--------|--------|---------|-----------|----------|
| `ipn://` | `ipn:<nodeNbr>.<serviceNbr>` | CBHE (numeric, compact) | Small (CBOR integers) | Space links, bandwidth-constrained |
| `dtn://` | `dtn://<node_name>/<demux>` | Non-CBHE (string-based) | Larger (full URI) | Human-readable, terrestrial |

For amateur radio DTN, the `dtn://` scheme allows embedding callsigns directly in endpoint identifiers, providing regulatory compliance and human readability.

---

## EID Format

```
dtn://<callsign>-<ssid>/<service>
```

**Examples:**
```
dtn://g4dpz-1              Node endpoint (primary station)
dtn://g4dpz-1/mail         Mail service
dtn://g4dpz-1/file         File transfer service
dtn://g4dpz-1/telemetry    Telemetry service
dtn://g4dpz-1/beacon       Beacon (station identification)
dtn://m0xer-1              Another station
```

**SSID Convention:**
- Range 0–15 (follows AX.25 convention)
- `-1` = primary station
- `-2` = secondary station or relay

---

## HDTN JSON Configuration

HDTN uses a single JSON configuration file per node. EID configuration, routing, induct/outduct definitions, and contact plans are all specified in this JSON file. The main configuration sections are:

- `hdtnConfigName` — node identity and EID configuration
- `inductsConfig` — convergence layer inbound adapters
- `outductsConfig` — convergence layer outbound adapters
- `storageConfig` — bundle storage parameters
- `contactPlanJson` — contact schedule for CGR

### Node Configuration (`hdtn-config.json`)

A single JSON file configures the node identity, local endpoints, routing, and convergence layer adapters.

**Node A (G4DPZ, node 1) — `configs/node-a/hdtn-config.json`:**
```json
{
  "hdtnConfigName": "node-a-g4dpz",
  "userRecycledFilePath": "/tmp/hdtn-node-a/",
  "myNodeId": 1,
  "myBpEchoServiceId": 2047,
  "mySchemeStr": "dtn",
  "myDtnEidStr": "dtn://g4dpz-1",
  "myDtnDemuxServices": ["mail", "file", "telemetry", "beacon"],
  "isAcsAware": false,
  "inductsConfig": {
    "inductVector": [
      {
        "convergenceLayer": "kiss_ltp",
        "name": "kissInduct",
        "boundPort": 0,
        "kissTncDevice": "/dev/tty.usbmodem2086327235531",
        "kissBaudRate": 9600,
        "thisLtpEngineId": 1,
        "remoteLtpEngineId": 2
      }
    ]
  },
  "outductsConfig": {
    "outductVector": [
      {
        "convergenceLayer": "kiss_ltp",
        "name": "kissOutduct",
        "nextHopNodeId": 2,
        "kissTncDevice": "/dev/tty.usbmodem2086327235531",
        "kissBaudRate": 9600,
        "thisLtpEngineId": 1,
        "remoteLtpEngineId": 2,
        "ltpMtu": 512,
        "ltpDataSegmentRate": 960
      }
    ]
  },
  "storageConfig": {
    "storageImplementation": "stdio_multi_threaded",
    "storageDiskConfigVector": [
      {
        "name": "bundleStore",
        "storeFilePath": "/tmp/hdtn-node-a/bundles/"
      }
    ]
  },
  "contactPlanJson": {
    "contacts": [
      {
        "source": 1,
        "dest": 2,
        "startTime": 0,
        "endTime": 86400,
        "rateBitsPerSec": 9600
      },
      {
        "source": 2,
        "dest": 1,
        "startTime": 0,
        "endTime": 86400,
        "rateBitsPerSec": 9600
      }
    ]
  }
}
```

**Node B (M0XER, node 2) — `configs/node-b/hdtn-config.json`:**
```json
{
  "hdtnConfigName": "node-b-m0xer",
  "userRecycledFilePath": "/tmp/hdtn-node-b/",
  "myNodeId": 2,
  "myBpEchoServiceId": 2047,
  "mySchemeStr": "dtn",
  "myDtnEidStr": "dtn://m0xer-1",
  "myDtnDemuxServices": ["mail", "file", "telemetry", "beacon"],
  "isAcsAware": false,
  "inductsConfig": {
    "inductVector": [
      {
        "convergenceLayer": "kiss_ltp",
        "name": "kissInduct",
        "boundPort": 0,
        "kissTncDevice": "/dev/tty.usbmodem20A5329335531",
        "kissBaudRate": 9600,
        "thisLtpEngineId": 2,
        "remoteLtpEngineId": 1
      }
    ]
  },
  "outductsConfig": {
    "outductVector": [
      {
        "convergenceLayer": "kiss_ltp",
        "name": "kissOutduct",
        "nextHopNodeId": 1,
        "kissTncDevice": "/dev/tty.usbmodem20A5329335531",
        "kissBaudRate": 9600,
        "thisLtpEngineId": 2,
        "remoteLtpEngineId": 1,
        "ltpMtu": 512,
        "ltpDataSegmentRate": 960
      }
    ]
  },
  "storageConfig": {
    "storageImplementation": "stdio_multi_threaded",
    "storageDiskConfigVector": [
      {
        "name": "bundleStore",
        "storeFilePath": "/tmp/hdtn-node-b/bundles/"
      }
    ]
  },
  "contactPlanJson": {
    "contacts": [
      {
        "source": 1,
        "dest": 2,
        "startTime": 0,
        "endTime": 86400,
        "rateBitsPerSec": 9600
      },
      {
        "source": 2,
        "dest": 1,
        "startTime": 0,
        "endTime": 86400,
        "rateBitsPerSec": 9600
      }
    ]
  }
}
```

**Key points:**
- `myDtnEidStr` sets the node's primary dtn:// EID with the callsign
- `myDtnDemuxServices` registers local service endpoints (mail, file, telemetry, beacon)
- Both `ipn` and `dtn` schemes can coexist — set `mySchemeStr` to `dtn` for callsign-based addressing
- HDTN's modular CLA plugin architecture handles KISS/LTP convergence layer via `inductsConfig` and `outductsConfig`
- Contact plans are embedded in the JSON or loaded from a separate contact plan JSON file

### HDTN Startup

Start the HDTN node with the JSON configuration:

```bash
#!/bin/bash
# Start HDTN node with callsign EID configuration

hdtn-one-process --hdtn-config-file=hdtn-config.json
```

Or using separate HDTN modules:

```bash
# Start HDTN ingress/egress/storage/scheduler modules
hdtn-ingress --hdtn-config-file=hdtn-config.json &
hdtn-egress --hdtn-config-file=hdtn-config.json &
hdtn-storage --hdtn-config-file=hdtn-config.json &
hdtn-scheduler --contact-plan-file=contact-plan.json &
```

---

## How the Routing Works

The `dtn://` scheme routing in HDTN:

```
Application sends bundle to: dtn://m0xer-1/mail
                                    │
                                    ▼
              HDTN looks up outduct for node "m0xer-1"
                                    │
                                    ▼
              Outduct config says: kiss_ltp to node 2
                                    │
                                    ▼
              LTP transmits to node 2 (Node B) via KISS CLA
                                    │
                                    ▼
              Node B receives, checks destination EID
              dtn://m0xer-1/mail matches local endpoint
                                    │
                                    ▼
              Bundle delivered to "mail" application
```

The numeric node IDs in the contact plan handle routing and scheduling. The `dtn://` EIDs provide the human-readable addressing layer on top.

---

## Complete Two-Node Example

### Network Topology

```
┌─────────────────────┐         ┌─────────────────────┐
│  Node A             │   LTP   │  Node B             │
│  Callsign: G4DPZ   │◄───────►│  Callsign: M0XER   │
│  Node ID: 1        │  KISS   │  Node ID: 2        │
│  EID: dtn://g4dpz-1 │         │  EID: dtn://m0xer-1 │
└─────────────────────┘         └─────────────────────┘
```

### Configuration

Each node uses a single `hdtn-config.json` file (shown above) that contains:
- Node identity (`myNodeId`, `myDtnEidStr`)
- Local service endpoints (`myDtnDemuxServices`)
- Induct configuration (KISS CLA receive)
- Outduct configuration (KISS CLA transmit with next-hop node ID)
- Contact plan (time-tagged communication windows)
- Storage configuration (bundle persistence)

---

## Application Usage

### Sending Bundles with Callsign EIDs

```bash
# Ping using callsign EIDs
bping --my-uri-eid=dtn://g4dpz-1 --dest-uri-eid=dtn://m0xer-1

# Send a file
bpsendfile --my-uri-eid=dtn://g4dpz-1 --dest-uri-eid=dtn://m0xer-1/file --file-or-folder-path=message.txt

# Send to a specific service
bpsendfile --my-uri-eid=dtn://g4dpz-1/beacon --dest-uri-eid=dtn://m0xer-1 --file-or-folder-path=beacon.txt
```

### Receiving Bundles

```bash
# Listen on the mail endpoint
bpreceivefile --my-uri-eid=dtn://m0xer-1/mail --save-directory=./received/

# Receive files
bpreceivefile --my-uri-eid=dtn://m0xer-1/file --save-directory=./received/
```

---

## Multi-Node Network Example

For a larger network with multiple stations, each node's `hdtn-config.json` includes outducts for each reachable peer:

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ G4DPZ    │     │ M0XER    │     │ W5DJT    │
│ Node 1   │◄───►│ Node 2   │◄───►│ Node 3   │
└──────────┘     └──────────┘     └──────────┘
```

**G4DPZ outducts:**
```json
{
  "outductVector": [
    { "convergenceLayer": "kiss_ltp", "nextHopNodeId": 2, "..." : "..." }
  ]
}
```

**M0XER outducts:**
```json
{
  "outductVector": [
    { "convergenceLayer": "kiss_ltp", "nextHopNodeId": 1, "..." : "..." },
    { "convergenceLayer": "kiss_ltp", "nextHopNodeId": 3, "..." : "..." }
  ]
}
```

**W5DJT outducts:**
```json
{
  "outductVector": [
    { "convergenceLayer": "kiss_ltp", "nextHopNodeId": 2, "..." : "..." }
  ]
}
```

---

## Considerations

### Case Sensitivity

HDTN treats node names as case-sensitive. Convention: use **lowercase** for callsigns in EIDs to avoid mismatches.

```
dtn://g4dpz-1    ✅ Preferred
dtn://G4DPZ-1    ⚠️  Works but must be consistent everywhere
```

### CGR Compatibility

HDTN's Contact Graph Routing (CGR) operates on numeric node IDs from the contact plan JSON. The `dtn://` scheme provides human-readable addressing that maps to these numeric IDs. For scheduled contacts (e.g., LEO passes), you can:

1. Use `ipn://` for CGR-routed traffic (space links)
2. Use `dtn://` for statically-routed traffic (terrestrial, GEO)
3. Or configure both — a node can respond to both `ipn:1.1` and `dtn://g4dpz-1`

### Bandwidth Overhead

The `dtn://` URI is encoded as a UTF-8 string in the bundle primary block, while `ipn://` uses compact CBOR integers. For bandwidth-constrained links:

| EID | Wire Size (approx) |
|-----|---------------------|
| `ipn:1.1` | 4 bytes |
| `dtn://g4dpz-1` | 16 bytes |
| `dtn://g4dpz-1/telemetry` | 26 bytes |

For 9600 baud terrestrial links this is negligible. For deep-space links, prefer `ipn://`.

---

## References

1. **RFC 9171** — Bundle Protocol Version 7 (BPv7)
2. **RFC 5326** — Licklider Transmission Protocol (LTP)
3. **HDTN** — https://github.com/nasa/HDTN
4. **HDTN Configuration Guide** — HDTN documentation
5. **LTP-KISS Architecture** — `docs/LTP-KISS-ARCHITECTURE.md`

---

**Contact:** G4DPZ
