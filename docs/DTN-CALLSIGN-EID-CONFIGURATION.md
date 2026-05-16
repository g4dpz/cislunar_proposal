# DTN Callsign EID Configuration for ION-DTN

**Version:** 1.0  
**Date:** May 16, 2026  
**Status:** Reference Documentation

---

## Overview

ION-DTN supports two Endpoint Identifier (EID) schemes:

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

## ION-DTN Configuration

ION uses separate configuration files for each protocol layer. To add `dtn://` callsign EIDs, you need to configure three files:

### 1. Bundle Protocol (`node.bprc`)

Register the `dtn` scheme and declare local endpoints.

```
## ─── IPN Scheme (numeric, for CGR routing) ───
a scheme ipn 'ipnfw' 'ipnadminep'
a endpoint ipn:1.0 q
a endpoint ipn:1.1 q
a endpoint ipn:1.2 q

## ─── DTN Scheme (callsign EIDs) ───
a scheme dtn 'dtn2fw' 'dtn2adminep'

## Local endpoints for this station
a endpoint dtn://g4dpz-1 q
a endpoint dtn://g4dpz-1/mail q
a endpoint dtn://g4dpz-1/file q
a endpoint dtn://g4dpz-1/telemetry q
a endpoint dtn://g4dpz-1/beacon q
```

**Key points:**
- `dtn2fw` is the forwarder daemon for the dtn scheme
- `dtn2adminep` handles administrative bundles for the dtn scheme
- `q` means queue bundles for this endpoint (don't discard if no app is listening)
- Both `ipn` and `dtn` schemes can coexist on the same node

### 2. DTN Scheme Routing (`node.dtn2rc`)

This file maps remote callsign node names to transport directives (how to reach them).

**Node A (G4DPZ, engine 1) — `configs/node-a/node.dtn2rc`:**
```
## DTN scheme routing for Node A (G4DPZ)
## Maps remote callsign EIDs to LTP transport

## Route bundles for dtn://m0xer-1/* via LTP to engine 2
a plan m0xer-1 ltp/2

## Optional: per-service routing rules
## Format: a rule <node_name> <demux_name> <directive>
## a rule m0xer-1 telemetry ltp/2
```

**Node B (M0XER, engine 2) — `configs/node-b/node.dtn2rc`:**
```
## DTN scheme routing for Node B (M0XER)
## Maps remote callsign EIDs to LTP transport

## Route bundles for dtn://g4dpz-1/* via LTP to engine 1
a plan g4dpz-1 ltp/1
```

**Directive format:**
```
a plan <node_name> <protocol>/<endpoint_id>
```

Where:
- `node_name` = the callsign-ssid portion of the remote EID (e.g., `m0xer-1`)
- `protocol` = transport protocol (`ltp`, `tcp`, `udp`, `stcp`)
- `endpoint_id` = the remote engine/node number for that protocol

### 3. ION Startup

Load the dtn2rc file during node initialization:

```bash
#!/bin/bash
# Start ION node with both ipn and dtn schemes

ionadmin node.ionrc
ltpadmin node.ltprc
bpadmin node.bprc
ipnadmin node.ipnrc
dtn2admin node.dtn2rc    # ← Load callsign routing
```

Or using `ionstart`:
```bash
ionstart -i node.ionrc -l node.ltprc -b node.bprc -p node.ipnrc -d node.dtn2rc
```

---

## How the Alias Works

The `dtn://` scheme acts as an overlay on top of the numeric transport layer:

```
Application sends bundle to: dtn://m0xer-1/mail
                                    │
                                    ▼
              dtn2fw looks up plan for "m0xer-1"
                                    │
                                    ▼
              Plan says: ltp/2 (use LTP engine 2)
                                    │
                                    ▼
              LTP transmits to engine 2 (Node B)
                                    │
                                    ▼
              Node B receives, checks destination EID
              dtn://m0xer-1/mail matches local endpoint
                                    │
                                    ▼
              Bundle delivered to "mail" application
```

The numeric engine IDs (`ionrc`) handle routing and transport. The `dtn://` EIDs provide the human-readable addressing layer on top.

---

## Complete Two-Node Example

### Network Topology

```
┌─────────────────────┐         ┌─────────────────────┐
│  Node A             │   LTP   │  Node B             │
│  Callsign: G4DPZ   │◄───────►│  Callsign: M0XER   │
│  Engine: 1          │  KISS   │  Engine: 2          │
│  EID: dtn://g4dpz-1 │         │  EID: dtn://m0xer-1 │
└─────────────────────┘         └─────────────────────┘
```

### Node A Configuration Files

**`node.ionrc`** (unchanged — numeric engine ID):
```
1 1 ''
s
a contact +0 +86400 1 2 120
a contact +0 +86400 2 1 120
a range +0 +86400 1 2 1
a range +0 +86400 2 1 1
m production 1000000
m consumption 1000000
```

**`node.bprc`** (add dtn scheme):
```
1

## IPN scheme (numeric routing)
a scheme ipn 'ipnfw' 'ipnadminep'
a endpoint ipn:1.0 q
a endpoint ipn:1.1 q
a endpoint ipn:1.2 q

## DTN scheme (callsign EIDs)
a scheme dtn 'dtn2fw' 'dtn2adminep'
a endpoint dtn://g4dpz-1 q
a endpoint dtn://g4dpz-1/mail q
a endpoint dtn://g4dpz-1/file q
a endpoint dtn://g4dpz-1/beacon q

## LTP transport
a protocol ltp 1400 100
a induct ltp 1 ltpcli
a outduct ltp 2 ltpclo

s
```

**`node.ipnrc`** (unchanged):
```
a plan 2 ltp/2
```

**`node.dtn2rc`** (new file):
```
## Route to Node B (M0XER) via LTP engine 2
a plan m0xer-1 ltp/2
```

### Node B Configuration Files

**`node.ionrc`**:
```
1 2 ''
s
a contact +0 +86400 1 2 120
a contact +0 +86400 2 1 120
a range +0 +86400 1 2 1
a range +0 +86400 2 1 1
m production 1000000
m consumption 1000000
```

**`node.bprc`**:
```
1

## IPN scheme
a scheme ipn 'ipnfw' 'ipnadminep'
a endpoint ipn:2.0 q
a endpoint ipn:2.1 q
a endpoint ipn:2.2 q

## DTN scheme (callsign EIDs)
a scheme dtn 'dtn2fw' 'dtn2adminep'
a endpoint dtn://m0xer-1 q
a endpoint dtn://m0xer-1/mail q
a endpoint dtn://m0xer-1/file q
a endpoint dtn://m0xer-1/beacon q

## LTP transport
a protocol ltp 1400 100
a induct ltp 2 ltpcli
a outduct ltp 1 ltpclo

s
```

**`node.ipnrc`**:
```
a plan 1 ltp/1
```

**`node.dtn2rc`** (new file):
```
## Route to Node A (G4DPZ) via LTP engine 1
a plan g4dpz-1 ltp/1
```

---

## Application Usage

### Sending Bundles with Callsign EIDs

```bash
# Ping using callsign EIDs
bping dtn://g4dpz-1 dtn://m0xer-1

# Send a file
bpsendfile dtn://g4dpz-1 dtn://m0xer-1/file message.txt

# Send to a specific service
bpsource dtn://g4dpz-1/beacon "G4DPZ amateur radio DTN station"
```

### Receiving Bundles

```bash
# Listen on the mail endpoint
bpsink dtn://m0xer-1/mail

# Receive files
bprecvfile dtn://m0xer-1/file
```

---

## dtn2admin Commands Reference

| Command | Description |
|---------|-------------|
| `a plan <node> <directive>` | Add routing plan for a remote node |
| `d plan <node>` | Delete a plan |
| `i plan <node>` | Show plan info |
| `l plan` | List all plans |
| `a rule <node> <demux> <directive>` | Add per-service routing rule |
| `d rule <node> <demux>` | Delete a rule |
| `i rule <node> <demux>` | Show rule info |
| `l rule` | List all rules |

**Directive format:** `<protocol>/<endpoint>`
- `ltp/2` — send via LTP to engine 2
- `tcp/192.168.1.100:4556` — send via TCP
- `udp/192.168.1.100:4556` — send via UDP

---

## Multi-Node Network Example

For a larger network with multiple stations:

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ G4DPZ    │     │ M0XER    │     │ W5DJT    │
│ Engine 1 │◄───►│ Engine 2 │◄───►│ Engine 3 │
└──────────┘     └──────────┘     └──────────┘
```

**G4DPZ `node.dtn2rc`:**
```
a plan m0xer-1 ltp/2
a plan w5djt-1 ltp/2    # Route via M0XER (next hop)
```

**M0XER `node.dtn2rc`:**
```
a plan g4dpz-1 ltp/1
a plan w5djt-1 ltp/3
```

**W5DJT `node.dtn2rc`:**
```
a plan g4dpz-1 ltp/2    # Route via M0XER (next hop)
a plan m0xer-1 ltp/2
```

---

## Considerations

### Case Sensitivity

ION treats node names as case-sensitive. Convention: use **lowercase** for callsigns in EIDs to avoid mismatches.

```
dtn://g4dpz-1    ✅ Preferred
dtn://G4DPZ-1    ⚠️  Works but must be consistent everywhere
```

### CGR Compatibility

ION's Contact Graph Routing (CGR) operates on numeric node IDs from `ionrc`. The `dtn://` scheme uses static routing via `dtn2rc` plans. For scheduled contacts (e.g., LEO passes), you can:

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
3. **ION-DTN dtn2admin** — `man dtn2admin` or ION documentation
4. **ION-DTN dtn2rc** — `man dtn2rc` or ION documentation
5. **LTP-KISS Architecture** — `docs/LTP-KISS-ARCHITECTURE.md`

---

**Contact:** G4DPZ
