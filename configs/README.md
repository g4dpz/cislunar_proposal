# ION-DTN Two-Node Terrestrial Configuration

## Node Mapping

| Node | Engine ID | Endpoints | TNC4 Device | Callsign |
|------|-----------|-----------|-------------|----------|
| A | 1 | ipn:1.0, ipn:1.1, ipn:1.2 | /dev/tty.usbmodem2086327235531 | G4DPZ-1 |
| B | 2 | ipn:2.0, ipn:2.1, ipn:2.2 | /dev/tty.usbmodem20A5329335531 | G4DPZ-2 |

## Radio Configuration

- Radio: Yaesu FT-817 (both nodes)
- TNC: Mobilinkd TNC4 (USB connection)
- Baud rate: 9600 (G3RUH GFSK)
- Band: UHF or VHF (operator's choice)
- MTU: 512 bytes
- Max data rate: 960 bytes/sec

## Configuration Files Per Node

Each node directory (`node-a/`, `node-b/`) contains:

| File | Purpose |
|------|---------|
| `node.ionrc` | ION initialization, contacts, ranges |
| `node.ltprc` | LTP spans using KISS CLA |
| `node.bprc` | BP scheme, endpoints, protocol, inducts/outducts |
| `node.ipnrc` | IPN static routing |
| `kiss.ionconfig` | KISS serial device, baud rate, MTU, rate limit |

## Contact Window

The default configuration sets a 24-hour contact window (`+0` to `+86400`).
For testing with scheduled windows, modify the `a contact` lines in `node.ionrc`.

## Quick Start

From the project root, on two separate terminals (or machines):

```bash
# Terminal 1 — Node A
./scripts/start-node-a.sh

# Terminal 2 — Node B
./scripts/start-node-b.sh
```

Test with bping:
```bash
# On Node A
bping ipn:1.1 ipn:2.1 -c 5
```

Stop:
```bash
./scripts/stop-node.sh
```

## Customization

- Update TNC4 device paths in `kiss.ionconfig` if your devices differ
- Adjust contact windows in `node.ionrc` for scheduled pass simulation
- Adjust MTU/rate in `kiss.ionconfig` for different baud rates
