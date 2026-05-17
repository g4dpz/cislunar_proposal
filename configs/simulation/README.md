# Cislunar DTN Simulation (HDTN)

Three-node loopback simulation of Earth-Moon DTN communication using HDTN.

## Topology

```
bping/bpsendfile (ipn:10.1)
  → [STCP :4556] → Ground Station HDTN (nodeId=10)
  → [LTP/UDP :2113 → :1113, 1300ms OWLT, 500 bps] → Orbiter HDTN (nodeId=20)
  → [LTP/UDP :2123 → :3113, 10ms OWLT, 9600 bps] → Lander HDTN (nodeId=30)
  → [STCP :4558] → bprecvfile (ipn:30.1)

Return path:
  Lander → [LTP/UDP :2143 → :3133] → Orbiter → [LTP/UDP :2133 → :1133] → Ground
```

## Link Characteristics

| Link | Data Rate | One-Way Light Time | RTT | Protocol | Ports |
|------|-----------|-------------------|-----|----------|-------|
| Ground → Orbiter | 500 bps | 1.3 seconds | 2.6 seconds | LTP/UDP | 2113 → 1113 |
| Orbiter → Ground | 500 bps | 1.3 seconds | 2.6 seconds | LTP/UDP | 2133 → 1133 |
| Orbiter → Lander | 9600 bps | 10 ms | 20 ms | LTP/UDP | 2123 → 3113 |
| Lander → Orbiter | 9600 bps | 10 ms | 20 ms | LTP/UDP | 2143 → 3133 |

## Configuration Files

| File | Purpose |
|------|---------|
| `cislunar-ground-station.json` | HDTN config for Ground Station (nodeId=10) |
| `cislunar-orbiter.json` | HDTN config for Lunar Orbiter relay (nodeId=20) |
| `cislunar-lander.json` | HDTN config for Lunar Lander (nodeId=30) |
| `cislunar-contact-plan.json` | Contact plan (separate file, shared by all nodes) |
| `bping-outducts.json` | Outduct config for bping/bpsendfile → Ground Station |
| `bpsink-inducts.json` | Induct config for bprecvfile/bpsink at Lander |

## Running

```bash
./scripts/run-cislunar-sim.sh
```

Requires HDTN installed (`hdtn-one-process` in PATH or set `HDTN_BIN`).

The script starts nodes in reverse order (lander first, then orbiter, then ground) with 5s sleep between each to allow LTP listeners to bind before senders connect.

## Sending Test Data

```bash
# Send a file from Ground Station to Lander
bpsendfile --my-uri-eid=ipn:10.1 --dest-uri-eid=ipn:30.1 \
    --outducts-config-file=configs/simulation/bping-outducts.json \
    --file-or-folder-path=test.txt

# Ping the Lander from Ground Station
bping --my-uri-eid=ipn:10.1 --dest-uri-eid=ipn:30.2047 \
    --outducts-config-file=configs/simulation/bping-outducts.json \
    --bundle-lifetime=300

# Receive at the Lander (already started by run script)
bprecvfile --my-uri-eid=ipn:30.1 \
    --inducts-config-file=configs/simulation/bpsink-inducts.json \
    --save-directory=/tmp/hdtn-sim/lander/received \
    --max-rx-bundle-size-bytes=10000000
```

## HDTN Command Format

Each HDTN node is started with:
```bash
hdtn-one-process \
    --hdtn-config-file=<node-config.json> \
    --contact-plan-file=cislunar-contact-plan.json
```

The contact plan is a separate file (not embedded in the node config).

## What This Demonstrates

- **LTP deferred acknowledgment**: LTP retransmission timers account for 1.3s propagation delay
- **Store-and-forward relay**: Orbiter stores bundles from Earth, forwards to Lander when link available
- **CGR routing**: Ground Station routes to Lander via Orbiter (next-hop routing via contact plan)
- **Low data rate operation**: 500 bps S-band link shows LTP segmentation with small segments
- **Bidirectional links**: Return path from Lander → Orbiter → Ground for acknowledgments

## Expected Behavior

1. bping/bpsendfile connects to Ground Station via STCP on port 4556
2. Ground Station's CGR routes bundle to Orbiter (next hop nodeId=20)
3. LTP segments the bundle, transmits at 500 bps to Orbiter via UDP (port 2113→1113)
4. LTP waits 2.6s+ for report segment (2 × 1.3s OWLT + margin)
5. Orbiter receives, stores, then forwards to Lander at 9600 bps (port 2123→3113)
6. Lander receives bundle, delivers to bprecvfile via STCP on port 4558
7. End-to-end delivery confirmed

## Storage Directories

- `/tmp/hdtn-sim/ground-station/` — Ground station bundle storage
- `/tmp/hdtn-sim/orbiter/` — Orbiter relay storage
- `/tmp/hdtn-sim/lander/` — Lander bundle storage
- `/tmp/hdtn-sim/lander/received/` — Files received by bprecvfile
