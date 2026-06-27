# Abstraction Layer: ION-DTN ↔ Hardy Interoperability

**Date:** June 27, 2026  
**Status:** Working (tested up to 1MB payload)

---

## Overview

The `radiant-dtn-abstraction` crate provides a single canonical data model (`NetworkConfiguration`) that describes a DTN node's identity, neighbors, links, contact plan, and routing strategy. From this model, backend-specific adapters generate the configuration files and manage the lifecycle of each DTN engine.

For ION-DTN ↔ Hardy interoperability over LTP/UDP, the abstraction layer orchestrates three processes:

```
┌────────────────────────────────────────────────────────────────┐
│              radiant-dtn-abstraction                            │
│                                                                │
│  NetworkConfiguration (canonical model)                        │
│       │                        │                               │
│       ▼                        ▼                               │
│  ┌──────────────┐     ┌──────────────────┐                    │
│  │ ION Adapter  │     │  Hardy Adapter   │                    │
│  │              │     │                  │                    │
│  │ config_gen() │     │ config_gen()     │                    │
│  │ → node10.*rc │     │ → hardy.yaml     │                    │
│  │              │     │ → ltp-cla.yaml   │                    │
│  │ lifecycle    │     │                  │                    │
│  │ → start()   │     │ (process spawn)  │                    │
│  │ → stop()    │     │                  │                    │
│  │ → health()  │     │                  │                    │
│  └──────────────┘     └──────────────────┘                    │
└────────────────────────────────────────────────────────────────┘
         │                        │
         ▼                        ▼
┌─────────────────┐    ┌───────────────────┐    ┌──────────────┐
│    ION-DTN      │    │ hardy-ltp-server  │    │  Hardy BPA   │
│    (node 10)    │    │   (LTP CLA)      │    │  (node 20)   │
│                 │    │                   │    │              │
│  LTP engine 10  │    │  LTP engine 20   │    │ gRPC :50051  │
│  UDP :2113      │◄──►│  UDP :1113       │───►│              │
│                 │LTP │                   │gRPC│              │
└─────────────────┘    └───────────────────┘    └──────────────┘
```

## Canonical Model

Both nodes are described using the same `NetworkConfiguration` struct:

```rust
// ION node (sender)
NetworkConfiguration {
    backend: "ion-dtn",
    local_node: NodeDefinition { node_number: 10, ... },
    neighbors: vec![Neighbor {
        node_number: 20,
        links: vec![ConvergenceLayerLink::LtpUdp {
            local_engine_id: 10,
            remote_engine_id: 20,
            remote_host: "127.0.0.1",
            remote_port: 1113,
            local_port: 2113,
            mtu: Some(1400),
        }],
    }],
    contact_plan: ContactPlan { ... },
    ...
}

// Hardy node (receiver)
NetworkConfiguration {
    backend: "hardy",
    local_node: NodeDefinition { node_number: 20, ... },
    neighbors: vec![Neighbor {
        node_number: 10,
        links: vec![ConvergenceLayerLink::LtpUdp {
            local_engine_id: 20,
            remote_engine_id: 10,
            remote_host: "127.0.0.1",
            remote_port: 2113,
            local_port: 1113,
            mtu: Some(1400),
        }],
    }],
    ...
}
```

The `ConvergenceLayerLink::LtpUdp` variant carries all the information needed to configure both sides of the link.

## Config Generation

### ION Adapter (`generate_ion_config`)

Produces four files from the canonical model:

| File | Purpose | Key content |
|------|---------|-------------|
| `node10.ionrc` | ION node init, contacts, ranges | `1 10 ''`, contact/range commands |
| `node10.ltprc` | LTP spans and listener | `a span 10 ...`, `a span 20 ...`, `s 'udplsi ...'` |
| `node10.bprc` | BP scheme, endpoints, protocol, ducts | `a induct ltp 10 ltpcli`, `a outduct ltp 20 ltpclo` |
| `node10.ipnrc` | Routing plan | `a plan 20 ltp/20` |

**Note:** The ION adapter currently requires a manual loopback span (`a span <local_engine_id>`) which is not yet auto-generated from the canonical model. This is a known limitation being tracked.

### Hardy Adapter (`generate_hardy_config`)

Produces two files when LTP/UDP links are present:

| File | Purpose | Key content |
|------|---------|-------------|
| `hardy.yaml` | BPA server config | node-ids, gRPC (with `"ltp"` service), storage, services |
| `ltp-cla.yaml` | hardy-ltp-server config | gRPC endpoint, LTP bind/engine-id/spans |

The Hardy adapter automatically:
- Adds `"ltp"` to the gRPC services list when LTP links exist
- Sets `engine-id` from `local_engine_id` in the link definition
- Sets `framing: none` for standard BPv7-over-LTP interoperability
- Includes remote node-ids in each span for peer registration

#### Generated `ltp-cla.yaml` structure

```yaml
grpc:
  endpoint: "http://[::1]:50051"
  reconnect-initial-secs: 1.0
  reconnect-max-secs: 10.0

ltp:
  bind: "0.0.0.0:1113"        # from link.local_port
  engine-id: 20                # from link.local_engine_id
  client-service-id: 1
  spans:
    - engine-id: 10            # from link.remote_engine_id
      address: "127.0.0.1:2113"  # from link.remote_host:remote_port
      max-segment-size: 1400   # from link.mtu
      framing: none            # standard for ION interop
      node-ids:
        - "ipn:10.0"          # from neighbor.node_number
```

## Lifecycle Management

### ION (`IonLifecycle`)

```rust
let mut ion = IonLifecycle::new(None);  // None = use $PATH for binaries
ion.start(&config_dir).await?;          // Runs ionadmin, ltpadmin, bpadmin, ipnadmin
let health = ion.health().await?;       // Checks rfxclock is running
ion.stop().await?;                      // Runs ionstop + killm
```

Key implementation details:
- Admin tools are run with `stdin(Stdio::null())` to prevent blocking on stdin
- `current_dir` is set to the config file's parent directory (ION writes SDR files to cwd)
- A 10-second timeout prevents hangs if an admin tool stalls
- `find_config_file()` locates files by extension (`.ionrc`, `.ltprc`, etc.)

### Hardy BPA + LTP CLA

Hardy processes are spawned directly (no lifecycle manager wrapper yet):

```rust
// Hardy BPA
Command::new("hardy-bpa-server")
    .arg("-c").arg("hardy.yaml")
    .spawn()?;

// hardy-ltp-server
Command::new("hardy-ltp-server")
    .arg("ltp-cla.yaml")
    .env("RUST_LOG", "hardy_ltp_cla=trace")
    .spawn()?;
```

The hardy-ltp-server automatically:
1. Connects to the Hardy BPA via gRPC
2. Registers as an LTP CLA (sends span descriptors, receives node-ids)
3. Binds the UDP socket for LTP segment transport
4. Begins receiving and reassembling LTP blocks
5. Dispatches completed bundles to the BPA

## Data Flow (1MB bundle transfer)

```
1. bpsendfile ipn:10.1 ipn:20.1 payload.dat
   │
2. ION BP agent segments the bundle into an LTP block
   │
3. ION LTP engine creates ~756 UDP segments (1400 bytes each)
   │ udplso sends to 127.0.0.1:1113
   ▼
4. hardy-ltp-server receives datagrams on port 1113
   │ Decodes LTP segment headers (SDNV engine_id, session_number)
   │ Routes to import session via spans HashMap
   │
5. Import session state machine:
   │ - Records each segment's byte range in ExtentMap
   │ - On RedEOB checkpoint: generates Report Segment
   │ - When all bytes received: emits DeliverBlock action
   │
6. Report Segment sent back to ION (port 2113)
   │ ION responds with ReportAck
   │
7. Block unpacked (framing: none → raw bundle)
   │
8. sink.dispatch(bundle, engine_id=10) → gRPC → Hardy BPA
   │
9. Hardy BPA stores/delivers the bundle
```

## Running the Example

```bash
cd radiant-dtn-abstraction

# Build (requires rustc ≥ 1.88)
cargo build --example interop_20k --features interop-network

# Run (ION-DTN must be installed on $PATH)
cargo run --example interop_20k --features interop-network
```

Output:
```
=== ION→Hardy 1MB Interop Test (abstraction layer) ===

[1/6] Generating configs via abstraction layer...
  ION: 4 files (generated from canonical config)
  Hardy: 2 files (["hardy.yaml", "ltp-cla.yaml"])

[2/6] Starting Hardy BPA...
  PID 89172
[3/6] Starting hardy-ltp-server...
  PID 89177
[4/6] Starting ION via IonLifecycle...
  ION running: true
[5/6] Sending 1MB bundle (ipn:10.1 → ipn:20.1)...
  Sent!
[6/6] Waiting for LTP delivery...

=== SUCCESS ===
  LTP span: bundle dispatched to BPA session_number=1 bundle_len=1048660

Tearing down...
Done.
```

## Critical Configuration Parameters

| Parameter | Where | Why |
|-----------|-------|-----|
| `engine-id: 20` | ltp-cla.yaml | Must match Hardy's node number for BPA registration |
| `"ltp"` in services | hardy.yaml | BPA must host the LTP gRPC service or returns UNIMPLEMENTED |
| `framing: none` | ltp-cla.yaml span | ION sends one raw bundle per LTP block (no length prefix) |
| `a span <local>` | node10.ltprc | ION needs a loopback span for its own engine ID |
| `stdin: null` | IonLifecycle | ION admin tools block if stdin is open |
| `current_dir` | IonLifecycle | ION writes SDR/log files to the working directory |

## Known Limitations

1. **ION config generator missing loopback span** — The `generate_ion_config()` function doesn't yet emit `a span <local_engine_id>` which ION requires. Currently worked around by writing ION configs directly from the canonical model parameters.

2. **No HardyLifecycle manager** — Hardy BPA and LTP CLA are spawned as child processes without a formal lifecycle wrapper. The `HardyLifecycle` struct exists but isn't used in the interop path.

3. **ION admin comments** — The ION config generator emits `##` comment lines which some ION builds reject. Comments should be stripped or the generator should omit them.

4. **No readiness polling** — The example uses fixed `sleep()` durations rather than polling for port availability or process health.

## File Locations

```
radiant-dtn-abstraction/
├── src/adapter/hardy/config_gen.rs   # generates hardy.yaml + ltp-cla.yaml
├── src/adapter/ion/config_gen.rs     # generates node*.ionrc/ltprc/bprc/ipnrc
├── src/adapter/ion/lifecycle.rs      # IonLifecycle (start/stop/health)
├── src/model/neighbor.rs             # ConvergenceLayerLink::LtpUdp definition
├── examples/interop_20k.rs           # Full working interop example
└── docs/INTEROP-ABSTRACTION-LAYER.md # This document
```
