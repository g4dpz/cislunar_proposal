# Implementation Plan: Terrestrial DTN Phase 1

## Overview

Phase 1 validates ION-DTN over amateur radio using the existing ION-DTN KISS CLA (`ltpkisscli`/`ltpkissclo`) with Mobilinkd TNC4 and Yaesu FT-817 at 9600 baud. ION-DTN provides BPv7, LTP, bundle storage, BPSec, priority handling, and lifetime enforcement out of the box. Our code is limited to: AX.25 frame validation tools (already done), ION-DTN configuration, a thin Go orchestration wrapper for node lifecycle and telemetry, and integration testing.

The ION-DTN KISS CLA is located at `ION-DTN/ltp/kiss/` and provides `ltpkissclo` (transmit) and `ltpkisscli` (receive) programs that wrap LTP segments in KISS frames for serial TNCs.

## Tasks

- [x] 1. Half-Duplex AX.25 Transfer Validation (Pre-ION-DTN)
  - [x] 1.1 Implement basic AX.25 frame construction and parsing
    - Created `ax25/` Go package with `BuildUIFrame`, `ParseFrame`, `ParseCallsign`
    - _Requirements: 9.1, 9.4, 9.5_

  - [x] 1.2 Implement TNC4 USB serial interface
    - Created `kiss/` Go package with KISS encode/decode and TNC serial interface
    - _Requirements: 9.4, 9.5_

  - [x] 1.3 Implement half-duplex AX.25 send/receive test harness
    - Created `cmd/ax25send/` and `cmd/ax25recv/` CLI tools
    - _Requirements: 9.1, 9.5_

  - [x] 1.4 Write AX.25 frame round-trip test (loopback or two-node)
    - Validated G4DPZ-1 ↔ G4DPZ-2 over TNC4 + FT-817 at 9600 baud in both directions
    - _Requirements: 9.1, 9.5, 9.6_

- [x] 2. Checkpoint — Half-duplex AX.25 link validated. AX.25 frames sent and received between two nodes over TNC4 + FT-817 at 9600 baud.

- [x] 3. Build ION-DTN with KISS CLA
  - [x] 3.1 Build ION-DTN from source with KISS CLA enabled
    - Configure and compile ION-DTN (`./configure && make`) targeting macOS/Linux
    - Verify `ltpkisscli` and `ltpkissclo` binaries are built
    - Verify `ionadmin`, `ltpadmin`, `bpadmin`, `bping`, `bpsink`, `bpsendfile`, `bprecvfile` are available
    - _Requirements: all_

  - [x] 3.2 Verify ION-DTN KISS CLA compiles and links correctly
    - Run ION-DTN test suite (if available) to confirm build integrity
    - Verify KISS CLA can open a serial device (loopback test with virtual serial port if no hardware)
    - _Requirements: 9.1, 9.4_

- [x] 4. Checkpoint — ION-DTN built with KISS CLA

- [x] 5. Create ION-DTN configuration for two-node terrestrial setup
  - [x] 5.1 Create Node A (Engine 1) configuration files
    - Create `configs/node-a/node.ionrc` — ION initialization, contacts, ranges
    - Create `configs/node-a/node.ltprc` — LTP spans using `ltpkissclo`/`ltpkisscli` with KISS CLA
    - Create `configs/node-a/node.bprc` — BP scheme, endpoints, protocol, inducts/outducts
    - Create `configs/node-a/kiss.ionconfig` — KISS serial device path (TNC4), 9600 baud, MTU 512, rate 960
    - Configure for Mobilinkd TNC4 device path (e.g., `/dev/tty.usbmodem2086327235531`)
    - _Requirements: 7.1, 9.1, 9.4, 9.5_

  - [x] 5.2 Create Node B (Engine 2) configuration files
    - Create `configs/node-b/node.ionrc` — ION initialization, contacts, ranges
    - Create `configs/node-b/node.ltprc` — LTP spans using `ltpkissclo`/`ltpkisscli` with KISS CLA
    - Create `configs/node-b/node.bprc` — BP scheme, endpoints, protocol, inducts/outducts
    - Create `configs/node-b/kiss.ionconfig` — KISS serial device path (TNC4), 9600 baud, MTU 512, rate 960
    - Configure for Mobilinkd TNC4 device path (e.g., `/dev/tty.usbmodem20A5329335531`)
    - _Requirements: 7.1, 9.1, 9.4, 9.5_

  - [x] 5.3 Create startup and shutdown scripts for each node
    - Create `scripts/start-node-a.sh` — runs `ionadmin`, `ltpadmin`, `bpadmin` with config files
    - Create `scripts/start-node-b.sh` — same for Node B
    - Create `scripts/stop-node.sh` — runs `ionstop` for clean shutdown
    - _Requirements: all_

  - [x] 5.4 Document configuration parameters and device mapping
    - Document which TNC4 device maps to which node
    - Document contact windows, engine IDs, endpoint IDs
    - _Requirements: 7.1_

- [x] 6. Checkpoint — ION-DTN configuration files created for two-node setup

- [x] 7. Test ION-DTN bping over KISS CLA
  - [x] 7.1 Start ION-DTN on both nodes
    - Run startup scripts on Node A and Node B
    - Verify ION-DTN initializes without errors (`ion.log`)
    - Verify `ltpkisscli` and `ltpkissclo` processes are running
    - _Requirements: all_

  - [x] 7.2 Run bping from Node A to Node B
    - Execute `bping ipn:1.1 ipn:2.1 -c 5` on Node A
    - Verify ping responses received from Node B
    - Record round-trip times
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 7.3 Run bping from Node B to Node A
    - Execute `bping ipn:2.1 ipn:1.1 -c 5` on Node B
    - Verify ping responses received from Node A
    - Confirm half-duplex DTN ping works in both directions
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 8. Checkpoint — DTN ping validated over KISS CLA + TNC4 + FT-817

- [x] 9. Test ION-DTN store-and-forward over KISS CLA
  - [x] 9.1 Test bpsendfile / bprecvfile
    - Start `bprecvfile ipn:2.1 1` on Node B
    - Send a test file from Node A: `bpsendfile ipn:1.1 ipn:2.1 testfile.txt`
    - Verify file received intact on Node B (checksum comparison)
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 9.2 Test store-and-forward with delayed contact
    - Send a bundle from Node A while Node B is offline (no contact window active)
    - Verify bundle is stored by ION-DTN on Node A
    - Start Node B and establish contact
    - Verify bundle is delivered to Node B when contact opens
    - _Requirements: 2.1, 2.2, 5.2, 5.5_

  - [x] 9.3 Test priority-based delivery
    - Send multiple bundles with different priorities (bulk, normal, expedited, critical)
    - Verify ION-DTN delivers them in priority order during the contact window
    - _Requirements: 5.3, 11.1, 11.2_

  - [x] 9.4 Test bundle lifetime expiry
    - Send a bundle with a short lifetime (e.g., 30 seconds)
    - Wait for the lifetime to expire before establishing contact
    - Verify the bundle is not delivered (expired and removed by ION-DTN)
    - _Requirements: 3.1, 3.2_

- [x] 10. Checkpoint — Store-and-forward validated over KISS CLA

- [x] 11. Test BPSec integrity over KISS CLA
  - [x] 11.1 Configure BPSec on both nodes
    - Create `configs/node-a/node.bpsecrc` and `configs/node-b/node.bpsecrc`
    - Configure HMAC-SHA-256 integrity rules with pre-shared keys
    - No encryption (amateur radio compliance)
    - _Requirements: 10.1, 10.2, 10.4_

  - [x] 11.2 Test bundle integrity verification
    - Send a bundle with BPSec integrity from Node A to Node B
    - Verify Node B accepts the bundle (integrity check passes)
    - _Requirements: 10.1, 10.3_

  - [x]* 11.3 Test integrity failure detection
    - Attempt to send a bundle with an incorrect integrity key
    - Verify Node B rejects the bundle (integrity check fails)
    - _Requirements: 10.3_

- [x] 12. Checkpoint — BPSec integrity validated

- [x] 13. Build Go orchestration wrapper
  - [x] 13.1 Create Go wrapper for ION-DTN node lifecycle
    - Implement `Start()` — execute ionadmin/ltpadmin/bpadmin with config files
    - Implement `Stop()` — execute ionstop for clean shutdown
    - Implement `IsRunning()` — check if ION-DTN processes are alive
    - Handle Ctrl+C for graceful shutdown
    - _Requirements: 14.3_

  - [x] 13.2 Create Go wrapper for telemetry collection
    - Query ION-DTN status via `bpadmin`/`ltpadmin` commands
    - Parse output to extract: bundles stored, bundles sent, bundles received, contacts completed/missed
    - Expose telemetry via local interface (JSON file or HTTP endpoint)
    - _Requirements: 13.1, 13.2, 13.3, 13.4_

  - [x] 13.3 Create Go wrapper for contact plan management
    - Load contact plan from a YAML/JSON config file
    - Generate ION-DTN `ionadmin` contact/range commands
    - Support adding/removing contacts at runtime via `ionadmin`
    - _Requirements: 7.1, 7.2, 7.3, 7.6, 7.7_

  - [x] 13.4 Create unified CLI for node operation
    - Create `cmd/dtn-node/main.go` — single entry point for starting a terrestrial DTN node
    - Parse config file (node ID, callsign, TNC device, contact plan, ION-DTN config paths)
    - Start ION-DTN, monitor health, expose telemetry, handle shutdown
    - _Requirements: all_

  - [ ]* 13.5 Write unit tests for Go orchestration wrapper
    - Test config file parsing
    - Test ION-DTN command generation
    - Test telemetry parsing
    - _Requirements: 13.1, 13.2_

- [x] 14. Checkpoint — Go orchestration wrapper complete

- [x] 15. End-to-end integration validation
  - [x] 15.1 Run full end-to-end test using Go wrapper
    - Start both nodes using `cmd/dtn-node`
    - Run bping in both directions
    - Send files in both directions
    - Verify telemetry reports correct statistics
    - _Requirements: all_

  - [x] 15.2 Run extended duration test
    - Run both nodes for 1+ hours with periodic bundle exchanges
    - Verify no memory leaks, no process crashes, telemetry remains accurate
    - _Requirements: 15.1, 13.1, 13.2_

  - [x]* 15.3 Document operational procedures
    - Write a README with setup instructions, configuration guide, and troubleshooting
    - Include device mapping for the two Mobilinkd TNC4 devices
    - _Requirements: all_

- [x] 16. Final checkpoint — Phase 1 terrestrial DTN validated

## Notes

- Tasks marked with `*` are optional
- ION-DTN provides: BPv7, LTP, bundle storage, priority queuing, lifetime enforcement, BPSec, eviction — we do NOT reimplement these
- ION-DTN's KISS CLA (`ltpkisscli`/`ltpkissclo`) handles KISS framing and serial I/O — we do NOT reimplement this
- Our Go code is a thin orchestration layer: node lifecycle, telemetry collection, contact plan management, and CLI
- The `ax25/` and `kiss/` Go packages from tasks 1.1-1.4 remain useful for standalone AX.25 testing and debugging
- Mobilinkd TNC4 devices: `/dev/tty.usbmodem2086327235531` (Node A) and `/dev/tty.usbmodem20A5329335531` (Node B)
