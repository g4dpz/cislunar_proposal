# Tasks: Cislunar Mission (Phase 4)

## Task 1: LDPC/Turbo FEC Codec
- [ ] 1.1 Implement `fec_init`: initialize LDPC or Turbo codec with configuration (code rate, block size, max iterations), allocate state from `POOL_FEC_STATE`
- [ ] 1.2 Implement `fec_encode`: LDPC or Turbo encoding of data blocks before BPSK modulation
- [ ] 1.3 Implement `fec_decode`: hard-decision LDPC or Turbo decoding of received blocks after BPSK demodulation
- [ ] 1.4 Implement `fec_decode_soft`: soft-decision decoding using log-likelihood ratios from BPSK demodulator
- [ ] 1.5 Implement `fec_get_stats`, `fec_reset_stats`, `fec_get_memory_usage`: statistics tracking (blocks encoded/decoded, failures, corrected errors, average iterations)
- [ ] 1.6 Write unit tests for FEC encode/decode with known test vectors, decode failure on uncorrectable errors
- [ ] 1.7 Write property test: FEC codec round-trip (Property 27) — generate random data blocks, encode then decode without channel errors, assert output equals input
- [ ] 1.8 Write property test: FEC decode detects uncorrectable errors (Property 28) — generate encoded blocks, inject errors beyond correction capability, verify decode returns error

## Task 2: S-Band IQ Baseband DSP Adaptation
- [ ] 2.1 Adapt IQ DSP from Phase 3 GMSK/BPSK at 9.6 kbps to BPSK at 500 bps for S-band 2.2 GHz
- [ ] 2.2 Integrate FEC codec into DSP pipeline: TX path (data → FEC encode → BPSK modulate → IQ), RX path (IQ → BPSK demodulate → soft symbols → FEC decode → data)
- [ ] 2.3 Configure DMA double-buffered ping-pong for S-band IQ sample rates (smaller buffers due to lower data rate)
- [ ] 2.4 Implement S-band flight transceiver initialization, power control, and 2.2 GHz configuration
- [ ] 2.5 Write unit tests for BPSK modulation/demodulation at 500 bps, FEC integration in DSP pipeline
- [ ] 2.6 Write property test: S-band radio path round-trip (Property 14) — generate random bundles, full stack BPv7 → LTP → AX.25 → FEC encode → BPSK mod → BPSK demod → FEC decode → AX.25 → LTP → BPv7, assert equality

## Task 3: CLA Adaptation for S-Band and Cislunar LTP
- [ ] 3.1 Adapt CLA from Phase 3 UHF 437 MHz to S-band 2.2 GHz at 500 bps BPSK
- [ ] 3.2 Implement `cla_activate_link_cislunar`: activate S-band link with Doppler profile, FEC config, and cislunar LTP parameters
- [ ] 3.3 Configure LTP retransmission timer for cislunar delay (default 10 seconds), session timeout (10 seconds), max retries (5), concurrent sessions (4)
- [ ] 3.4 Integrate FEC codec into CLA TX/RX data paths (CLA → FEC → DSP → transceiver)
- [ ] 3.5 Implement transceiver health check and 3-retry reinitialization logic on failure
- [ ] 3.6 Write unit tests for CLA activation/deactivation, LTP cislunar timer configuration, FEC integration, transceiver failure handling
- [ ] 3.7 Write property test: AX.25 callsign framing (Property 15) — generate random bundles, verify frames carry valid callsigns
- [ ] 3.8 Write property test: LTP segmentation/reassembly round-trip (Property 16) — generate random large bundles, segment, reassemble, assert equality
- [ ] 3.9 Write property test: LTP cislunar timer correctness (Property 29) — generate random session parameters, verify retransmission timer ≥ 2 seconds, concurrent sessions independent

## Task 4: Cislunar CGR Engine Adaptation
- [ ] 4.1 Implement ephemeris table interpolation (`cgr_interpolate`): Hermite or Lagrange interpolation of satellite position/velocity from ephemeris points
- [ ] 4.2 Adapt contact prediction (`cgr_predict_contacts`) from SGP4/SDP4 to ephemeris-based propagation: iterate over time steps, find AOS/LOS for each ground station, compute max elevation, Doppler at 2.2 GHz, light-time delay, and range per arc
- [ ] 4.3 Implement `cgr_load_ephemeris`, `cgr_get_ephemeris`, `cgr_ephemeris_age`: ephemeris lifecycle management
- [ ] 4.4 Implement `cgr_next_contact_for_station` and `cgr_next_contact_any` for cislunar arcs
- [ ] 4.5 Extend prediction horizon to 48 hours (from Phase 3's 24 hours)
- [ ] 4.6 Implement light-time delay computation (range / speed_of_light) for each predicted contact
- [ ] 4.7 Implement `cgr_get_memory_usage` and pool-based allocation from `POOL_CGR_STATE` (~64 KB budget)
- [ ] 4.8 Write unit tests for ephemeris interpolation against known reference positions, contact prediction with known cislunar geometry
- [ ] 4.9 Write property test: CGR cislunar prediction invariants (Property 30) — generate random valid ephemeris + catalog, verify start < end, elevation ≥ min, Doppler within ±5 kHz, horizon ≥ 48h, light-time > 0, all direct contacts

## Task 5: Ephemeris Manager
- [ ] 5.1 Implement `eph_init`: load ephemeris from NVM if available, initialize staleness tracking
- [ ] 5.2 Implement `eph_update`: validate ephemeris format and epoch, reject if invalid or older, persist to NVM, trigger CGR re-prediction
- [ ] 5.3 Implement `eph_get_age`, `eph_is_stale`, `eph_get_margin_factor`: staleness computation and margin widening
- [ ] 5.4 Implement `eph_persist` and `eph_reload`: NVM round-trip with CRC protection (`nvm_ephemeris_sector_t`)
- [ ] 5.5 Write unit tests for ephemeris validation (valid tables, bad CRCs, old epochs, format errors)
- [ ] 5.6 Write property test: Ephemeris update validation (Property 31) — generate random ephemeris tables, verify acceptance/rejection based on validity and epoch
- [ ] 5.7 Write property test: Ephemeris NVM round-trip (Property 32) — generate random valid ephemeris, persist, reload, assert equality
- [ ] 5.8 Write property test: Ephemeris staleness tracking (Property 33) — generate random epochs/times/thresholds, verify age, stale flag, margin factor

## Task 6: Cislunar Doppler Compensator Adaptation
- [ ] 6.1 Adapt Doppler compensator from Phase 3 UHF 437 MHz (±10 kHz) to S-band 2.2 GHz (±5 kHz)
- [ ] 6.2 Implement Doppler profile computation from CGR ephemeris-based pass geometry at 2.2 GHz
- [ ] 6.3 Support profiles with hundreds of points for hours-long arcs (10-second intervals)
- [ ] 6.4 Relax update rate to once per 10 seconds (from Phase 3's 1 second)
- [ ] 6.5 Write unit tests for Doppler interpolation at S-band, boundary values (±5 kHz, 0 Hz), long-arc profiles
- [ ] 6.6 Write property test: Doppler computation bounds at S-band (Property 34) — generate random positions/velocities, verify offset within ±5 kHz and consistent with v_radial formula at 2.2 GHz
- [ ] 6.7 Write property test: Doppler correction round-trip at S-band (Property 35) — generate random IQ samples + offsets within ±5 kHz, apply then correct, verify restoration within precision

## Task 7: Enhanced Radiation Monitor (TMR)
- [ ] 7.1 Implement `tmr_write`, `tmr_read`, `tmr_validate`: triple modular redundancy for critical control variables (contact plan active flag, current contact index, power state)
- [ ] 7.2 Implement enhanced `rad_init` with TMR initialization for critical variables
- [ ] 7.3 Implement `rad_get_validation_interval` returning 60 seconds for cislunar
- [ ] 7.4 Integrate TMR validation into periodic SRAM validation cycle (every 60 seconds active, every wake cycle)
- [ ] 7.5 Track TMR corrections separately in SEU count (`radiation_tmr_corrections`)
- [ ] 7.6 Register FEC codec state as a CRC-protected SRAM region
- [ ] 7.7 Write unit tests for TMR write/read/validate, single-copy corruption detection and repair, majority vote
- [ ] 7.8 Write property test: SRAM radiation protection (Property 36) — generate random data, flip random bits in primary, verify detection and recovery from redundant, SEU counter increment
- [ ] 7.9 Write property test: TMR correctness (Property 37) — generate random values, corrupt one of three copies, verify majority vote returns correct value and validate repairs
- [ ] 7.10 Write property test: NVM read CRC validation (Property 38) — generate random NVM data, corrupt, verify CRC failure detected

## Task 8: Ground Station Catalog Adaptation (Tier 3/4)
- [ ] 8.1 Extend `ground_station_phase4_t` with `antenna_gain_dbi` field
- [ ] 8.2 Adapt `catalog_add_station` for Phase 4 validation (including antenna gain ≥ 0)
- [ ] 8.3 Update NVM layout to `nvm_catalog_phase4_sector_t` with magic `0x47534334`
- [ ] 8.4 Implement `catalog_persist` and `catalog_reload` for Phase 4 format
- [ ] 8.5 Write unit tests for catalog operations with antenna gain field, NVM round-trip
- [ ] 8.6 Write property test: Catalog NVM round-trip (Property 41) — generate random valid entries (up to 32, with antenna gain), persist, reload, assert equality

## Task 9: Autonomous Node Controller Adaptation
- [ ] 9.1 Adapt `node_init` for Phase 4 configuration (`node_config_phase4_t`): FEC config, LTP cislunar params, radiation validation interval, sleep threshold
- [ ] 9.2 Adapt `node_run_cycle` for 2-second cycle time (from Phase 3's 1 second) to accommodate FEC processing
- [ ] 9.3 Implement hours-long contact arc management (vs. Phase 3's 5–10 minute passes)
- [ ] 9.4 Implement 5-minute sleep threshold between arcs (from Phase 3's 60 seconds)
- [ ] 9.5 Implement periodic radiation validation every 60 seconds during active contact arcs
- [ ] 9.6 Implement `node_generate_telemetry` with Phase 4 telemetry fields (FEC stats, light-time, TMR corrections, ephemeris age)
- [ ] 9.7 Implement administrative bundle dispatch for ephemeris updates (replacing TLE updates)
- [ ] 9.8 Implement `node_cold_boot` with ephemeris reload (replacing TLE reload)
- [ ] 9.9 Write unit tests for single cycle execution, contact arc scheduling, telemetry generation, admin bundle dispatch
- [ ] 9.10 Write property test: Sleep decision correctness (Property 21) — generate random system states, verify sleep/wake decision with 5-minute threshold and RTC alarm
- [ ] 9.11 Write property test: No transmission after arc end (Property 17) — generate random arcs and time sequences, verify no transmission after end
- [ ] 9.12 Write property test: Missed contact retains bundles (Property 18) — generate random failed contacts, verify bundles retained and counter incremented

## Task 10: Phase 3 Carry-Forward Components — Adaptation and Property Tests
- [ ] 10.1 Adapt BPA for Phase 4 (ephemeris admin bundle type replacing TLE, FEC-aware bundle size limits)
- [ ] 10.2 Adapt NVM Bundle Store for 256 MB–1 GB capacity (extended addressing, sector management)
- [ ] 10.3 Adapt NVM Bundle Store SRAM index for enhanced radiation protection (CRC + redundant + TMR flags)
- [ ] 10.4 Add `POOL_FEC_STATE` to pool allocator
- [ ] 10.5 Adapt power manager for 10–20 W budget and 5-minute sleep threshold
- [ ] 10.6 Write property test: Bundle serialization round-trip (Property 1)
- [ ] 10.7 Write property test: Bundle creation and validation (Property 2)
- [ ] 10.8 Write property test: Store/retrieve round-trip (Property 3)
- [ ] 10.9 Write property test: Priority ordering (Property 4)
- [ ] 10.10 Write property test: Eviction ordering (Property 5)
- [ ] 10.11 Write property test: Capacity bound (Property 6)
- [ ] 10.12 Write property test: Store reload with CRC (Property 7)
- [ ] 10.13 Write property test: Lifetime enforcement (Property 8)
- [ ] 10.14 Write property test: Ping echo correctness (Property 9)
- [ ] 10.15 Write property test: Local vs remote routing (Property 10)
- [ ] 10.16 Write property test: ACK/no-ACK behavior — cislunar delay (Property 11)
- [ ] 10.17 Write property test: Bundle retention without contact (Property 12)
- [ ] 10.18 Write property test: No relay (Property 13)
- [ ] 10.19 Write property test: BPSec integrity round-trip (Property 19)
- [ ] 10.20 Write property test: No encryption (Property 20)
- [ ] 10.21 Write property test: Pool exhaustion safety (Property 23)
- [ ] 10.22 Write property test: Rate limiting (Property 24)
- [ ] 10.23 Write property test: Bundle size limit (Property 25)
- [ ] 10.24 Write property test: Statistics monotonicity (Property 26)
- [ ] 10.25 Write property test: Power state transition logging (Property 22)
- [ ] 10.26 Write property test: Time synchronization threshold (Property 39)
- [ ] 10.27 Write property test: Reset recovery completeness (Property 40)

## Task 11: Integration Testing
- [ ] 11.1 End-to-end store-and-forward test: Tier 3 ground station → S-band transceiver → OBC store (with LDPC/Turbo FEC) → retrieve → S-band transceiver → destination Tier 3 station
- [ ] 11.2 End-to-end cislunar ping test through full S-band RF path with RTT measurement (expected 2–4 seconds)
- [ ] 11.3 Autonomous contact arc sequence test: multiple CGR-predicted arcs (hours-long) with Doppler compensation and FEC, verify all bundles delivered
- [ ] 11.4 Power cycle recovery test: populate NVM (256 MB–1 GB), power cycle, verify state restored within 5 seconds
- [ ] 11.5 Watchdog reset recovery test: simulate hang, verify watchdog reset, verify state recovery
- [ ] 11.6 Transceiver failure test: mock unresponsive S-band transceiver, verify 3 retries, contact missed, bundles retained
- [ ] 11.7 Ephemeris update flow test: send ephemeris bundle during arc, verify acceptance, persistence, CGR re-prediction with 48h horizon
- [ ] 11.8 Catalog update flow test: send catalog bundle (with antenna gain), verify station added, CGR re-prediction
- [ ] 11.9 Time sync flow test: send time sync bundle, verify RTC update based on threshold
- [ ] 11.10 Radiation simulation test (enhanced): inject bit flips into CRC-protected SRAM and TMR variables, verify detection, recovery, SEU counting, TMR majority vote
- [ ] 11.11 SRAM budget validation test: run all subsystems concurrently (ION-DTN, IQ DSP, FEC codec, CGR, radiation monitor), verify total ≤ OBC SRAM budget
- [ ] 11.12 Doppler tracking test: simulated cislunar arc with realistic S-band Doppler profile (±5 kHz, hours-long), verify demodulator lock
- [ ] 11.13 Stale ephemeris operation test: operate with ephemeris > 7 days old, verify warning and widened margins
- [ ] 11.14 FEC performance test: inject channel errors at Eb/N0 ≈ 2 dB, verify BER ≤ 1e-5 after decoding
- [ ] 11.15 LTP cislunar delay test: verify LTP sessions operate correctly with 2–4 second RTT, concurrent sessions, retransmission after timeout
- [ ] 11.16 Long-duration contact arc test: simulate a 3-hour contact arc at 500 bps, verify continuous operation, Doppler tracking, FEC decode, bundle delivery, and power budget compliance
