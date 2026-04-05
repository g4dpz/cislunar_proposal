# Tasks: LEO CubeSat Flight (Phase 3)

## Task 1: CGR Engine — SGP4/SDP4 Orbit Propagation
- [ ] 1.1 Implement TLE parser (`cgr_load_tle`): parse two-line element sets, validate checksums and field ranges, extract orbital parameters
- [ ] 1.2 Implement SGP4/SDP4 propagator (`cgr_propagate`): compute satellite position (ECI km) and velocity (ECI km/s) at arbitrary times from TLE data
- [ ] 1.3 Implement line-of-sight computation: given satellite ECI position and ground station geodetic coordinates, compute elevation angle and determine visibility
- [ ] 1.4 Implement contact prediction (`cgr_predict_contacts`): iterate over time steps, find AOS/LOS for each ground station, compute max elevation and Doppler per pass, enforce minimum elevation filter and 24-hour horizon
- [ ] 1.5 Implement `cgr_next_contact_for_station` and `cgr_next_contact_any` convenience functions
- [ ] 1.6 Implement `cgr_get_memory_usage` and pool-based allocation from `POOL_CGR_STATE`
- [ ] 1.7 Write unit tests for TLE parsing (valid/invalid TLEs, checksum failures, field range violations)
- [ ] 1.8 Write unit tests for SGP4 propagation against known reference positions (e.g., ISS TLE with published ephemeris)
- [ ] 1.9 Write property test: CGR prediction invariants (Property 27) — generate random valid TLE + catalog, verify start < end, elevation ≥ min, Doppler within ±10 kHz, horizon ≥ 24h, all direct contacts

## Task 2: TLE Manager — Persistence and Updates
- [ ] 2.1 Implement `tle_init`: load TLE from NVM if available, initialize staleness tracking
- [ ] 2.2 Implement `tle_update`: validate TLE format and epoch, reject if invalid or older, persist to NVM, trigger CGR re-prediction
- [ ] 2.3 Implement `tle_get_age`, `tle_is_stale`, `tle_get_margin_factor`: staleness computation and margin widening
- [ ] 2.4 Implement `tle_persist` and `tle_reload`: NVM round-trip with CRC protection (`nvm_tle_sector_t`)
- [ ] 2.5 Write unit tests for TLE validation (valid TLEs, bad checksums, old epochs, format errors)
- [ ] 2.6 Write property test: TLE update validation (Property 28) — generate random TLEs, verify acceptance/rejection based on validity and epoch
- [ ] 2.7 Write property test: TLE NVM round-trip (Property 29) — generate random valid TLE, persist, reload, assert equality
- [ ] 2.8 Write property test: TLE staleness tracking (Property 30) — generate random epochs/times/thresholds, verify age, stale flag, margin factor

## Task 3: Ground Station Catalog Manager
- [ ] 3.1 Implement `catalog_init`: load catalog from NVM if available
- [ ] 3.2 Implement `catalog_add_station`: validate fields (lat ±90, lon ±180, alt ≥ 0, elev 0–90), add/update entry, trigger CGR re-prediction for affected station
- [ ] 3.3 Implement `catalog_get_station`, `catalog_find_by_callsign`, `catalog_count`
- [ ] 3.4 Implement `catalog_persist` and `catalog_reload`: NVM round-trip with CRC protection (`nvm_catalog_sector_t`)
- [ ] 3.5 Write unit tests for catalog operations (add, update, find, capacity limit at 32, field validation)
- [ ] 3.6 Write property test: Catalog NVM round-trip (Property 31) — generate random valid entries (up to 32), persist, reload, assert equality

## Task 4: Doppler Compensator
- [ ] 4.1 Implement `doppler_init`, `doppler_load_profile`, `doppler_clear_profile`: profile lifecycle management
- [ ] 4.2 Implement `doppler_get_offset`: linear interpolation between profile points for smooth frequency tracking
- [ ] 4.3 Implement `doppler_correct_rx`: NCO frequency shift on RX IQ samples based on current Doppler offset
- [ ] 4.4 Implement `doppler_correct_tx`: inverse NCO shift on TX IQ samples for pre-compensation
- [ ] 4.5 Implement Doppler profile computation from CGR pass geometry (satellite position/velocity relative to ground station → radial velocity → frequency offset at 437 MHz)
- [ ] 4.6 Write unit tests for Doppler interpolation, boundary values (±10 kHz, 0 Hz), profile loading/clearing
- [ ] 4.7 Write property test: Doppler computation bounds (Property 32) — generate random positions/velocities, verify offset within ±10 kHz and consistent with v_radial formula
- [ ] 4.8 Write property test: Doppler correction round-trip (Property 33) — generate random IQ samples + offsets, apply then correct, verify restoration within precision

## Task 5: Flight Transceiver CLA Adaptation
- [ ] 5.1 Adapt CLA from Phase 2 IQ bridge path to direct Flight_Transceiver interface (STM32U585 DMA → DAC/ADC or SPI → transceiver IC)
- [ ] 5.2 Implement `cla_activate_link_doppler`: activate link with Doppler profile, integrate Doppler compensator into IQ streaming pipeline
- [ ] 5.3 Update DSP modulation from GFSK/G3RUH (Phase 2) to GMSK/BPSK for flight link
- [ ] 5.4 Implement flight transceiver initialization, power control, and UHF 437 MHz configuration
- [ ] 5.5 Implement transceiver health check and 3-retry reinitialization logic on failure
- [ ] 5.6 Write unit tests for CLA activation/deactivation, Doppler integration, transceiver failure handling
- [ ] 5.7 Write property test: End-to-end radio path round-trip (Property 14) — generate random bundles, full stack BPv7 → LTP → AX.25 → IQ mod → IQ demod → AX.25 → LTP → BPv7, assert equality
- [ ] 5.8 Write property test: AX.25 callsign framing (Property 15) — generate random bundles, verify frames carry valid callsigns
- [ ] 5.9 Write property test: LTP segmentation/reassembly round-trip (Property 16) — generate random large bundles, segment, reassemble, assert equality

## Task 6: Autonomous Node Controller
- [ ] 6.1 Implement `node_init`: initialize all subsystems (BPA, store, CGR, TLE, catalog, Doppler, power, time, radiation monitor, watchdog, pool allocator, TrustZone)
- [ ] 6.2 Implement `node_run_cycle`: check CGR for contacts, execute bundle transfers if contact active, run cleanup, validate SRAM integrity, kick watchdog — must complete within 1 second
- [ ] 6.3 Implement `node_cold_boot`: reload all state from NVM (bundles, TLE, catalog, statistics), re-compute CGR predictions — must complete within 5 seconds
- [ ] 6.4 Implement `node_main_loop`: autonomous loop (predict → sleep → wake → communicate → sleep), never returns
- [ ] 6.5 Implement `node_generate_telemetry`: package `telemetry_phase3_t` as a DTN bundle
- [ ] 6.6 Implement administrative bundle dispatch (`bpa_dispatch_admin`): route TLE update, catalog update, time sync, telemetry request, key update bundles to appropriate handlers
- [ ] 6.7 Write unit tests for single cycle execution, contact scheduling, telemetry generation, admin bundle dispatch
- [ ] 6.8 Write property test: Sleep decision correctness (Property 21) — generate random system states, verify sleep/wake decision and RTC alarm
- [ ] 6.9 Write property test: No transmission after window end (Property 17) — generate random windows and time sequences, verify no transmission after end
- [ ] 6.10 Write property test: Missed contact retains bundles (Property 18) — generate random failed contacts, verify bundles retained and counter incremented

## Task 7: Radiation Monitor
- [ ] 7.1 Implement `rad_init`, `rad_register_region`: register critical SRAM regions with CRC and redundant copy
- [ ] 7.2 Implement `rad_update_region`: recompute CRC and refresh redundant copy after legitimate writes
- [ ] 7.3 Implement `rad_validate_all`: check CRC on all registered regions, recover from redundant copy on primary corruption, recover from primary on redundant corruption, log unrecoverable dual-corruption
- [ ] 7.4 Implement `rad_get_seu_count` and `rad_validate_nvm_read`
- [ ] 7.5 Write unit tests for region registration, CRC computation, single-bit flip detection, recovery scenarios
- [ ] 7.6 Write property test: SRAM radiation protection (Property 34) — generate random data, flip random bits in primary, verify detection and recovery from redundant, SEU counter increment
- [ ] 7.7 Write property test: NVM read CRC validation (Property 35) — generate random NVM data, corrupt, verify CRC failure detected

## Task 8: Time Manager
- [ ] 8.1 Implement `time_init`, `time_now`: RTC initialization and UTC time retrieval
- [ ] 8.2 Implement `time_sync`: update RTC only if correction exceeds threshold (default 1 second)
- [ ] 8.3 Implement `time_since_last_sync`, `time_is_stale`: staleness tracking (default 7 days)
- [ ] 8.4 Write unit tests for time sync with corrections above/below threshold, staleness detection
- [ ] 8.5 Write property test: Time synchronization threshold (Property 36) — generate random times and thresholds, verify RTC update decision and stale flag

## Task 9: Watchdog Manager
- [ ] 9.1 Implement `wdt_init`: configure IWDG with configurable timeout (default 30 seconds)
- [ ] 9.2 Implement `wdt_kick`: refresh watchdog counter
- [ ] 9.3 Implement `wdt_was_reset_cause`: check RCC reset flags for IWDG reset
- [ ] 9.4 Write unit tests for initialization, reset cause detection

## Task 10: Power Manager Adaptation for Autonomous Operation
- [ ] 10.1 Adapt power manager to receive wake times from CGR-predicted contact plan (replacing Phase 2 UART commands from companion host)
- [ ] 10.2 Implement sleep decision logic: enter Stop 2 iff no active contact, no pending work, no contact within 60 seconds
- [ ] 10.3 Implement RTC alarm setting from next predicted contact window start time before entering Stop 2
- [ ] 10.4 Write unit tests for sleep decision logic, RTC alarm configuration
- [ ] 10.5 Write property test: Power state transition logging (Property 22) — generate random transitions, verify logging correctness

## Task 11: Phase 2 Carry-Forward Components — Adaptation and Property Tests
- [ ] 11.1 Adapt BPA for standalone operation (remove UART command dependency, add admin bundle dispatch)
- [ ] 11.2 Adapt NVM Bundle Store SRAM index for radiation protection (register with radiation monitor)
- [ ] 11.3 Add `POOL_CGR_STATE` to pool allocator
- [ ] 11.4 Write property test: Bundle serialization round-trip (Property 1)
- [ ] 11.5 Write property test: Bundle creation and validation (Property 2)
- [ ] 11.6 Write property test: Store/retrieve round-trip (Property 3)
- [ ] 11.7 Write property test: Priority ordering (Property 4)
- [ ] 11.8 Write property test: Eviction ordering (Property 5)
- [ ] 11.9 Write property test: Capacity bound (Property 6)
- [ ] 11.10 Write property test: Store reload with CRC (Property 7)
- [ ] 11.11 Write property test: Lifetime enforcement (Property 8)
- [ ] 11.12 Write property test: Ping echo correctness (Property 9)
- [ ] 11.13 Write property test: Local vs remote routing (Property 10)
- [ ] 11.14 Write property test: ACK/no-ACK behavior (Property 11)
- [ ] 11.15 Write property test: Bundle retention without contact (Property 12)
- [ ] 11.16 Write property test: No relay (Property 13)
- [ ] 11.17 Write property test: BPSec integrity round-trip (Property 19)
- [ ] 11.18 Write property test: No encryption (Property 20)
- [ ] 11.19 Write property test: Pool exhaustion safety (Property 23)
- [ ] 11.20 Write property test: Rate limiting (Property 24)
- [ ] 11.21 Write property test: Bundle size limit (Property 25)
- [ ] 11.22 Write property test: Statistics monotonicity (Property 26)
- [ ] 11.23 Write property test: Reset recovery completeness (Property 37)

## Task 12: Integration Testing
- [ ] 12.1 End-to-end store-and-forward test: ground station → flight transceiver → STM32U585 store → retrieve → flight transceiver → destination ground station
- [ ] 12.2 End-to-end ping test through full RF path with RTT measurement
- [ ] 12.3 Autonomous pass sequence test: 4–6 CGR-predicted passes with Doppler compensation, verify all bundles delivered
- [ ] 12.4 Power cycle recovery test: populate NVM, power cycle, verify state restored within 5 seconds
- [ ] 12.5 Watchdog reset recovery test: simulate hang, verify watchdog reset, verify state recovery
- [ ] 12.6 Transceiver failure test: mock unresponsive transceiver, verify 3 retries, contact missed, bundles retained
- [ ] 12.7 TLE update flow test: send TLE bundle during pass, verify acceptance, persistence, CGR re-prediction
- [ ] 12.8 Catalog update flow test: send catalog bundle, verify station added, CGR re-prediction
- [ ] 12.9 Time sync flow test: send time sync bundle, verify RTC update based on threshold
- [ ] 12.10 Radiation simulation test: inject bit flips into protected SRAM, verify detection, recovery, SEU counting
- [ ] 12.11 SRAM budget validation test: run all subsystems concurrently, verify total ≤ 786 KB
- [ ] 12.12 Doppler tracking test: simulated pass with realistic Doppler profile, verify demodulator lock
- [ ] 12.13 Stale TLE operation test: operate with TLE > 14 days old, verify warning and widened margins
