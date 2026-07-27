# G4DPZ Ground Station — TCPCLv3 + LTP Configuration

ION-DTN configuration for the G4DPZ ground station with dual convergence layers:

- **LTP/UDP** — satellite link to node 20 (port 2113 ↔ 1113)
- **TCPCLv3** — terrestrial input from any DTN peer (listening on port 4556)

## Network Topology

```
Terrestrial Peers (internet)         Satellite (node 20)
┌──────────────────┐                ┌──────────────────┐
│  ipn:30.x        │                │  ipn:20.x        │
│  TCP → :4556     │                │  LTP/UDP :1113   │
└────────┬─────────┘                └────────▲─────────┘
         │ TCPCLv3                           │ LTP/UDP
         ▼                                   │
┌─────────────────────────────────────────────────────┐
│           G4DPZ Ground Station (Node 10)            │
│                                                     │
│  TCP induct:  0.0.0.0:4556 (tcpcli)               │
│  LTP induct:  0.0.0.0:2113 (udplsi → ltpcli)      │
│  LTP outduct: 192.168.1.20:1113 (udplso → ltpclo) │
│                                                     │
│  Endpoints: ipn:10.0, ipn:10.1, ipn:10.2          │
└─────────────────────────────────────────────────────┘
```

## Usage

```bash
# Start ION with this config
./scripts/ion-ctl.sh start configs/g4dpz-tcpcl/

# Verify TCP listener is active
netstat -tlnp | grep 4556

# Check status
./scripts/ion-ctl.sh status
```

## Terrestrial Client Connection

Any DTN node speaking TCPCLv3 can connect to port 4556 and deliver bundles.
ION will accept the connection, receive bundles, and route them based on the
destination EID:

- `ipn:10.x` → delivered locally
- `ipn:20.x` → forwarded to satellite via LTP

## Amateur Radio Compliance

- No encryption configured
- No BPSec
- All payloads transmitted in the clear
