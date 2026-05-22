# Requirements Document

## Introduction

This document specifies the Software Requirements for the RADIANT Test Framework — a requirements-based verification system for the Radio Amateur Delay-tolerant Interplanetary Networking Testbed. The framework validates correctness properties of the RADIANT DTN software stack (BPv7 → LTP → KISS → G3RUH) using property-based testing methodology, modeled after NASA Glenn's HDTN Test Framework (TM-20240014467 / LEW-20818-1).

The Test Framework provides automated, repeatable verification that the RADIANT software meets its functional and performance requirements across all mission phases: terrestrial (Phase 1), QO-100 GEO (Phase 1.5), CubeSat EM (Phase 2), LEO flight (Phase 3), and Cislunar (Phase 4). Each property test traces to one or more system requirements, enabling requirements-based verification suitable for flight experiment proposals and regulatory submissions.

**Methodological Basis**: NASA Glenn Research Center, "HDTN Test Framework," NASA/TM-20240014467 (LEW-20818-1).

## Glossary

- **Test_Framework**: The complete verification system comprising property-based tests, integration tests, CI pipeline, and traceability infrastructure
- **Property_Test**: A test that verifies a correctness property holds for all inputs within a defined domain, using randomized input generation (gopter or rapid libraries)
- **Integration_Test**: A test that verifies system-level behavior using representative examples rather than randomized inputs
- **Smoke_Test**: A lightweight integration test that verifies basic system health (file existence, config parsing, permissions)
- **BPA**: Bundle Protocol Agent — the component responsible for BPv7 bundle creation, validation, and processing
- **Bundle_Store**: The persistent storage component for BPv7 bundles awaiting transmission
- **Contact_Plan_Manager**: The component managing scheduled contact windows and CGR predictions
- **Telemetry_Collector**: The component that retrieves and parses HDTN REST API telemetry data
- **Node_Controller**: The orchestration component managing bundle transmission during contact windows
- **Rate_Limiter**: The security component enforcing bundle acceptance rate limits
- **Link_Budget_Calculator**: The component computing RF link margins for various orbital scenarios
- **CLA**: Convergence Layer Adapter — the interface between BPv7/LTP and physical radio links
- **CGR**: Contact Graph Routing — the algorithm computing optimal routes through scheduled contacts
- **Requirement_Traceability**: The mapping between each property test and the system requirement(s) it validates
- **CI_Pipeline**: The GitHub Actions continuous integration workflow executing all tests on every commit
- **PBT_Library**: Property-based testing library (gopter for gopter-style generators, rapid for pgregory.net/rapid)
- **HDTN**: High-rate Delay Tolerant Networking — NASA Glenn's DTN implementation used as the protocol engine
- **HDTN_REST_API**: The HTTP REST interface exposed by HDTN for telemetry and management
- **HDTN_Process**: A running instance of the HDTN binary (hdtn-one-process)
- **EARS**: Easy Approach to Requirements Syntax — the structured pattern language used for requirement statements
- **ConvergenceLayerAdapter**: The Go interface defining CLA operations (Open, Close, SendBundle, RecvBundle)

## Requirements

### Requirement 1: Bundle Protocol Agent Validation

**User Story:** As a mission assurance engineer, I want the Test Framework to verify that the BPA correctly validates all incoming bundles, so that only well-formed bundles enter the DTN network.

#### Acceptance Criteria

1. WHEN a bundle with a valid destination (non-empty scheme and non-empty SSP), a lifetime greater than 0 seconds, a creation timestamp not exceeding the current time by more than 5 seconds, and a remaining lifetime greater than 0 seconds at the time of receipt is submitted, THE Test_Framework SHALL verify that the BPA accepts the bundle by confirming it is stored or forwarded without returning an error status
2. WHEN a bundle with an empty destination scheme or empty destination SSP is submitted, THE Test_Framework SHALL verify that the BPA rejects the bundle by returning a validation error indicating the destination field that failed and does not store or forward the bundle
3. WHEN a bundle with a lifetime of zero seconds or a negative lifetime value is submitted, THE Test_Framework SHALL verify that the BPA rejects the bundle by returning a validation error indicating invalid lifetime and does not store or forward the bundle
4. WHEN a bundle with a creation timestamp exceeding the current time by more than 5 seconds is submitted, THE Test_Framework SHALL verify that the BPA rejects the bundle by returning a validation error indicating a future timestamp and does not store or forward the bundle
5. WHEN a bundle whose creation timestamp plus lifetime results in an expiration time at or before the current time is submitted, THE Test_Framework SHALL verify that the BPA rejects the bundle by returning a validation error indicating the bundle has expired and does not store or forward the bundle
6. THE Test_Framework SHALL execute a minimum of 100 randomized input combinations per BPA validation property test, covering destination, lifetime, and timestamp fields with values spanning valid ranges and boundary conditions
7. THE Test_Framework SHALL annotate each BPA validation property test with a "Validates: Requirement X.Y" comment tracing to the system-level requirement

### Requirement 2: Ping Echo Verification

**User Story:** As a network operator, I want the Test Framework to verify that ping echo responses are correctly generated, so that I can confirm end-to-end reachability.

#### Acceptance Criteria

1. WHEN a valid ping request bundle is received by the BPA, THE Test_Framework SHALL verify that exactly one ping response bundle is generated and no additional response bundles are produced for the same request
2. WHEN a ping response is generated, THE Test_Framework SHALL verify that the response destination matches the source endpoint identifier from the original ping request bundle's primary block
3. WHEN a ping response is generated, THE Test_Framework SHALL verify that the response bundle type is set to BundleTypePingResponse
4. THE Test_Framework SHALL generate randomized source and destination endpoint identifiers using valid dtn:// or ipn:// scheme formats with SSP strings between 1 and 255 characters for ping echo property tests, running a minimum of 100 iterations per property
5. IF a ping request bundle fails BPA validation, THEN THE Test_Framework SHALL verify that no ping response bundle is generated

### Requirement 3: Bundle Store Round-Trip Integrity

**User Story:** As a software assurance engineer, I want the Test Framework to verify that bundles stored and retrieved from the Bundle Store are identical, so that I can confirm no data corruption occurs during persistence.

#### Acceptance Criteria

1. THE Test_Framework SHALL verify that for any valid BPv7 bundle with payload sizes between 1 and 255 bytes, priorities between 0 and 3, and lifetimes between 1 and 3600 seconds, storing the bundle and then retrieving it by bundle ID produces a bundle identical to the original across all compared fields
2. THE Test_Framework SHALL verify identity by comparing all bundle fields between the stored and retrieved bundle: source EID scheme, source EID SSP, creation timestamp, sequence number, destination scheme, destination SSP, payload content (byte-for-byte equality), payload length, priority, lifetime, creation time, and bundle type
3. THE Test_Framework SHALL execute a minimum of 100 randomized bundle configurations per store round-trip property test, where each test iteration randomizes payload size, payload content, priority, and lifetime within the specified ranges
4. IF the Test_Framework attempts to retrieve a bundle ID that was not previously stored, THEN the Bundle_Store SHALL return an error indicating the bundle was not found without modifying any existing stored bundles
5. THE Test_Framework SHALL allocate a Bundle_Store with a capacity of at least 1,048,576 bytes for each property test execution to ensure store-full errors do not interfere with round-trip verification

### Requirement 4: Bundle Store Capacity Enforcement

**User Story:** As a systems engineer, I want the Test Framework to verify that the Bundle Store never exceeds its configured capacity, so that I can confirm memory safety on resource-constrained flight hardware.

#### Acceptance Criteria

1. THE Test_Framework SHALL execute a property-based test with a minimum of 100 successful randomized sequences of interleaved store and delete operations, using payload sizes between 1 and 255 bytes against a Bundle_Store configured with a fixed capacity, and verify that used bytes never exceed the configured total bytes after any operation in the sequence
2. WHEN a store operation would exceed capacity, THE Test_Framework SHALL verify that the Bundle_Store rejects the operation and that all previously stored bundles remain retrievable with unchanged payload length, priority, and bundle ID
3. THE Test_Framework SHALL generate randomized operation sequences that include sequences where cumulative store payload reaches at least 90% of configured capacity before any delete occurs, ensuring capacity boundary conditions are exercised
4. IF a store operation is rejected due to insufficient capacity, THEN THE Test_Framework SHALL verify that the Bundle_Store used bytes and bundle count remain unchanged from their values immediately before the rejected operation

### Requirement 5: Priority Ordering Invariant

**User Story:** As a mission planner, I want the Test Framework to verify that bundles are always retrievable in priority order, so that critical telemetry is transmitted before routine data during limited contact windows.

#### Acceptance Criteria

1. FOR ALL generated sets of 1 to 200 bundles with arbitrary priority values (0-3, where 0=bulk, 1=normal, 2=expedited, 3=critical), THE Test_Framework SHALL verify that listing bundles by priority produces a sequence where each bundle's priority is greater than or equal to the next bundle's priority
2. FOR ALL generated sets containing bundles with equal priority, THE Test_Framework SHALL verify that bundles within the same priority level are ordered by creation timestamp ascending (oldest first)
3. THE Test_Framework SHALL execute a minimum of 100 randomized test iterations per priority distribution type
4. THE Test_Framework SHALL generate randomized priority distributions including: uniform (equal count per priority level), skewed (at least 70% of bundles assigned to a single priority level), and single-priority (all bundles assigned the same priority value)
5. THE Test_Framework SHALL include edge cases of an empty bundle set (which produces an empty sequence) and a single-bundle set (which trivially satisfies ordering)

### Requirement 6: Eviction Policy Correctness

**User Story:** As a flight software engineer, I want the Test Framework to verify that the eviction policy removes expired bundles first and lowest-priority bundles second, so that critical data is preserved under storage pressure.

#### Acceptance Criteria

1. FOR ALL sets of 1 to 100 bundles containing both expired and valid bundles (where a bundle is expired if creation_timestamp plus lifetime is less than or equal to current_time), THE Test_Framework SHALL verify that EvictExpired removes all expired bundles and retains all valid bundles
2. FOR ALL sets of 2 to 100 bundles with mixed integer priorities in the range 0 to 255, THE Test_Framework SHALL verify that EvictLowestPriority removes exactly one bundle whose priority is less than or equal to all remaining bundles
3. THE Test_Framework SHALL verify that the count of evicted expired bundles equals the number of expired bundles in the original set
4. IF multiple bundles share the lowest priority value, THEN THE Test_Framework SHALL verify that EvictLowestPriority removes exactly one of those bundles and retains the others

### Requirement 7: Bundle Lifetime Enforcement

**User Story:** As a network architect, I want the Test Framework to verify that no expired bundles persist after a cleanup cycle, so that stale data does not consume storage or bandwidth.

#### Acceptance Criteria

1. FOR ALL sets of 1 to 100 bundles with arbitrary lifetimes (1-2000 seconds) and arbitrary current times (1000-3000 seconds), THE Test_Framework SHALL verify that after EvictExpired, zero remaining bundles have creation timestamp plus lifetime less than or equal to the current time
2. THE Test_Framework SHALL generate bundles with lifetimes that place the expiration time within 1 second of the query time (both above and below) to exercise the boundary between expired and valid

### Requirement 8: Telemetry Parsing Fidelity

**User Story:** As a ground station operator, I want the Test Framework to verify that HDTN telemetry is parsed without data loss, so that I can trust the displayed statistics for mission operations.

#### Acceptance Criteria

1. FOR ALL valid HDTN REST API JSON responses with arbitrary non-negative integer field values (0 to 2^63-1), THE Test_Framework SHALL verify that every field maps correctly to the corresponding Telemetry structure field
2. THE Test_Framework SHALL verify the following field mappings: bundleCountStorage to BundlesStored, bundleCountEgress to BundlesSent, bundleCountIngress to BundlesReceived, bundleByteCountEgress to BytesSent, bundleByteCountIngress to BytesReceived, usedSpaceBytes to StorageUsedBytes, totalSpaceBytes to StorageQuotaBytes
3. THE Test_Framework SHALL verify that LTP SessionsActive equals the sum of numActiveSendSessions and numActiveRecvSessions
4. WHEN a successful HTTP response is received from the HDTN_REST_API, THE Test_Framework SHALL verify that Health.Running is set to true
5. THE Test_Framework SHALL verify that NodeID equals the node_id string from the Telemetry_Collector configuration and NodeNumber equals the node_number integer from the Telemetry_Collector configuration

### Requirement 9: Telemetry Partial Response Resilience

**User Story:** As a fault tolerance engineer, I want the Test Framework to verify that partial telemetry responses are handled gracefully with zero-filling, so that missing fields do not cause crashes or incorrect displays.

#### Acceptance Criteria

1. FOR ALL HDTN REST API JSON responses with randomly omitted fields (each field independently present or absent), THE Test_Framework SHALL verify that present fields retain their correct values and absent fields default to zero (0 for integers, 0.0 for floats, false for booleans)
2. THE Test_Framework SHALL verify that NodeID is populated from the Telemetry_Collector's configured node_id, NodeNumber is populated from the Telemetry_Collector's configured node_number, and Timestamp is populated from the system clock at collection time, regardless of which API fields are present in the response
3. THE Test_Framework SHALL verify that Timestamp formats as RFC 3339 with UTC timezone designator

### Requirement 10: Contact Plan Validation

**User Story:** As a mission planning engineer, I want the Test Framework to verify that contact plan validation correctly accepts valid plans and rejects invalid plans, so that only physically realizable contact schedules are loaded.

#### Acceptance Criteria

1. FOR ALL collections of contact entries, THE Test_Framework SHALL verify that validation accepts the collection if and only if every contact has RateBitsPerSec greater than zero, every contact has StartTime strictly less than EndTime, and the total number of entries is less than or equal to 1000
2. WHEN validation fails, THE Test_Framework SHALL verify that the error message identifies the zero-based index of the first invalid entry and the reason for rejection (zero rate, invalid time range, or count exceeded)
3. THE Test_Framework SHALL generate contact collections at sizes 0, 999, 1000, 1001, and 1050 to exercise the count limit boundary

### Requirement 11: Active Contacts Filtering

**User Story:** As a contact scheduling engineer, I want the Test Framework to verify that active contact filtering returns exactly the correct set of contacts for any query time, so that the node transmits only during valid windows.

#### Acceptance Criteria

1. FOR ALL sets of contacts and any query time T, THE Test_Framework SHALL verify that GetActiveContacts(T) returns exactly those contacts where StartTime is less than or equal to T and T is strictly less than EndTime
2. THE Test_Framework SHALL verify that no contacts outside the active window are included in the result, and that the returned collection size equals the number of contacts satisfying the active window predicate
3. THE Test_Framework SHALL generate at least 5 query times per contact window: one before StartTime, one equal to StartTime, one between StartTime and EndTime, one equal to EndTime minus 1 millisecond, and one equal to or after EndTime
4. WHEN the contact plan is empty, THE Test_Framework SHALL verify that GetActiveContacts returns an empty collection for any query time

### Requirement 12: Contact Removal Correctness

**User Story:** As a mission operations engineer, I want the Test Framework to verify that removing a contact from the plan removes only that contact and preserves all others, so that plan modifications are safe.

#### Acceptance Criteria

1. FOR ALL contact plans containing at least one contact, THE Test_Framework SHALL verify that removing a contact by its (source, dest, startTime) key results in a plan that no longer contains that contact
2. THE Test_Framework SHALL verify that all contacts not matching the removal key remain unchanged in value and order after the removal operation
3. THE Test_Framework SHALL verify that the remaining contact count equals the original count minus one
4. IF a removal is attempted with a (source, dest, startTime) key that does not match any contact in the plan, THEN THE Test_Framework SHALL verify that the operation returns an error and the plan remains unchanged with the original contact count preserved

### Requirement 13: API Error State Preservation

**User Story:** As a reliability engineer, I want the Test Framework to verify that API failures do not corrupt local contact plan state, so that transient network errors do not cause data loss.

#### Acceptance Criteria

1. FOR ALL contact plan managers with existing local state, IF an API operation (add, remove, or apply) fails due to an HTTP error (status codes 400-599) or a request timeout exceeding 5 seconds, THEN THE Test_Framework SHALL verify that the local plan state is identical to the state before the operation was attempted
2. THE Test_Framework SHALL test state preservation across add, remove, and apply operations independently, verifying each operation type with at least one failure scenario
3. THE Test_Framework SHALL use mock HTTP servers returning 500 status codes to simulate server errors and 408 or connection timeout to simulate network failures

### Requirement 14: CGR Prediction Time Horizon Compliance

**User Story:** As an orbital mechanics engineer, I want the Test Framework to verify that all CGR-predicted contact windows fall within the requested time horizon, so that the contact plan does not schedule transmissions outside valid prediction bounds.

#### Acceptance Criteria

1. FOR ALL valid LEO orbital parameters and time horizons (1 to 24 hours), THE Test_Framework SHALL verify that all predicted contact windows have StartTime greater than or equal to the horizon start and EndTime less than or equal to the horizon end
2. FOR ALL valid cislunar orbital parameters and time horizons (1 to 168 hours), THE Test_Framework SHALL verify that all predicted contact windows have StartTime greater than or equal to the horizon start and EndTime less than or equal to the horizon end
3. THE Test_Framework SHALL verify that no two predicted windows for the same ground station overlap in time
4. THE Test_Framework SHALL verify that all predicted contacts have StartTime strictly less than EndTime
5. THE Test_Framework SHALL verify that LEO predicted contacts have durations between 60 seconds and 900 seconds (physically reasonable for LEO passes)
6. WHEN orbital parameters with eccentricity greater than or equal to 1.0 are provided, THE Test_Framework SHALL verify that the prediction function returns a validation error
7. WHEN a time horizon with fromTime greater than or equal to toTime is provided, THE Test_Framework SHALL verify that the prediction function returns a validation error

### Requirement 15: Next Contact Lookup Correctness

**User Story:** As a routing engineer, I want the Test Framework to verify that next contact lookup returns the earliest future contact for a given destination, so that CGR selects optimal transmission opportunities.

#### Acceptance Criteria

1. FOR ALL contact plans, destination nodes, and query times, THE Test_Framework SHALL verify that GetNextContact returns the contact with the earliest StartTime that is greater than or equal to the query time and matches the destination node
2. WHEN multiple future contacts exist for the same destination, THE Test_Framework SHALL verify that the contact with the earliest StartTime is returned and that no other future contact for that destination has a StartTime earlier than the returned contact
3. IF no future contacts exist for the specified destination at the given query time, THEN THE Test_Framework SHALL verify that an error is returned indicating no future contact is available
4. IF the destination node does not exist in the contact plan, THEN THE Test_Framework SHALL verify that an error is returned indicating the destination is unknown
5. THE Test_Framework SHALL verify the boundary condition: a contact with StartTime exactly equal to the query time is included as a valid result
6. THE Test_Framework SHALL verify that a contact with EndTime less than or equal to the query time is not returned as a next contact
7. THE Test_Framework SHALL verify that when multiple destination nodes exist, the correct contact for each specific destination is returned independently

### Requirement 16: No-Relay Direct Delivery Enforcement

**User Story:** As a network security engineer, I want the Test Framework to verify that nodes only transmit bundles to their final destination (no relay forwarding), so that the single-hop architecture constraint is enforced.

#### Acceptance Criteria

1. FOR ALL bundles transmitted during any contact window, THE Test_Framework SHALL verify that the contact's remote node EID matches the bundle's destination EID exactly
2. WHEN FindDirectContact is called with a destination EID, THE Test_Framework SHALL verify that the returned contact's remote node EID matches the queried destination EID, confirming single-hop routing
3. FOR ALL bundles in a node's Bundle_Store, THE Test_Framework SHALL verify that every bundle's source EID matches the local node's EID (no relay bundles accepted from other originators)
4. IF FindDirectContact is called with a destination EID that has no matching contact in the current contact plan, THEN THE Test_Framework SHALL verify that an error is returned rather than a multi-hop path through an intermediate node

### Requirement 17: Contact Window Temporal Enforcement

**User Story:** As a flight software engineer, I want the Test Framework to verify that no transmission occurs after a contact window ends, so that the node does not transmit into void or interfere with other scheduled contacts.

#### Acceptance Criteria

1. FOR ALL contact windows where the current time exceeds the window end time, THE Test_Framework SHALL verify that zero bundles are transmitted and all bundles queued for that contact's destination remain in the Bundle_Store with their count unchanged
2. THE Test_Framework SHALL generate randomized contact window durations between 60 and 600 seconds and bundle counts between 5 and 20 bundles per test iteration to exercise the temporal boundary
3. WHEN the current time is within 1 second before the contact window end time and a bundle transmission is in progress, THE Test_Framework SHALL verify that no new bundle transmissions are initiated after the window end time

### Requirement 18: Missed Contact Bundle Retention

**User Story:** As a reliability engineer, I want the Test Framework to verify that bundles are retained when a contact is missed, so that data is not lost due to transient link failures.

#### Acceptance Criteria

1. WHEN the CLA fails to establish a link during a scheduled contact window, THE Test_Framework SHALL verify that all bundles queued for that contact's destination remain in the Bundle_Store with their count and content unchanged
2. WHEN a contact is missed due to CLA link establishment failure, THE Test_Framework SHALL verify that the ContactsMissed counter is incremented by exactly one
3. THE Test_Framework SHALL use mock CLAs configured to return link establishment errors to simulate missed contacts
4. IF multiple contacts are missed in sequence, THEN THE Test_Framework SHALL verify that the ContactsMissed counter equals the total number of missed contacts and that no bundles are lost from the Bundle_Store across all missed windows

### Requirement 19: Bundle Retention Without Contact

**User Story:** As a store-and-forward engineer, I want the Test Framework to verify that bundles are retained when no contact is available for their destination, so that data persists until a future contact opportunity.

#### Acceptance Criteria

1. FOR ALL bundles whose destination EID has no matching contact window in the current contact plan, THE Test_Framework SHALL verify that the Bundle_Store retains the bundle with its content unchanged after processing
2. THE Test_Framework SHALL verify that retained bundles persist through at least 3 subsequent processing cycles until their lifetime expires, confirming no premature deletion
3. THE Test_Framework SHALL use an empty contact plan (zero contact entries) to exercise the no-contact-available scenario
4. IF a bundle's lifetime expires while retained without a contact, THEN THE Test_Framework SHALL verify that the bundle is evicted from the Bundle_Store and not transmitted

### Requirement 20: Rate Limiting Enforcement

**User Story:** As a security engineer, I want the Test Framework to verify that the rate limiter correctly enforces bundle acceptance limits, so that denial-of-service conditions are mitigated.

#### Acceptance Criteria

1. FOR ALL configured rate limits between 1 and 100 bundles per second, THE Test_Framework SHALL verify that the number of accepted bundles within any 1-second window does not exceed the configured maximum rate
2. FOR ALL submission sequences where the number of submissions exceeds the configured rate limit within a 1-second window, THE Test_Framework SHALL verify that at least one bundle is rejected with a rate-limit-exceeded error
3. THE Test_Framework SHALL verify the accounting invariant: accepted count plus rejected count equals total submission attempts for every test iteration
4. FOR ALL configured rate limits between 1 and 100 bundles per second, THE Test_Framework SHALL verify that a submission sequence with a count equal to or less than the configured rate limit within a 1-second window results in all bundles being accepted
5. WHEN 1 second has elapsed since the last rate window started, THE Test_Framework SHALL verify that the rate limiter resets and accepts new bundles up to the configured maximum rate
6. THE Test_Framework SHALL verify the monotonicity property: the value returned by GetCurrentRate never exceeds the configured maximum rate at any point during a submission sequence

### Requirement 21: Link Budget Monotonicity

**User Story:** As an RF engineer, I want the Test Framework to verify that link margin decreases monotonically with distance, so that I can trust the link budget calculator for mission planning.

#### Acceptance Criteria

1. FOR ALL sequences of 2 to 10 increasing distances in the range 1,000 m to 500,000,000 m with identical transmit parameters, THE Test_Framework SHALL verify that the computed link margin strictly decreases with each distance increment, using a minimum of 100 randomized test iterations
2. THE Test_Framework SHALL verify that the LEO UHF link budget (2W TX, omni TX antenna, 12 dBi Yagi RX antenna, 437 MHz, 9600 bps, 10 dB required Eb/N0) closes with a positive margin exceeding 20 dB at 500 km distance
3. THE Test_Framework SHALL verify that the cislunar S-band link budget (5W TX, 10 dBi patch TX antenna, 35 dBi dish RX antenna, 2.2 GHz, 500 bps, 2 dB required Eb/N0) closes with a positive margin between 5.0 and 7.0 dB at 384,000 km (lunar distance)
4. WHEN zero or negative distance is provided, THE Test_Framework SHALL verify that the Link_Budget_Calculator returns a validation error indicating the distance must be positive
5. WHEN zero frequency or zero data rate is provided, THE Test_Framework SHALL verify that the Link_Budget_Calculator returns a validation error indicating the invalid parameter
6. FOR ALL distances d where d is in the range 10,000 m to 500,000 m, THE Test_Framework SHALL verify that doubling the distance reduces the link margin by 6.02 dB (±0.1 dB tolerance), confirming correct free-space path loss computation

### Requirement 22: S-Band CLA Bundle Serialization Round-Trip

**User Story:** As a communications engineer, I want the Test Framework to verify that bundle serialization and deserialization through the S-band CLA preserves all bundle fields, so that no data is lost during radio transmission framing.

#### Acceptance Criteria

1. FOR ALL valid bundles with payload sizes from 1 to 1500 bytes, THE Test_Framework SHALL verify that serializing then deserializing produces a bundle with identical bundle type, priority, lifetime, destination endpoint, and payload content, using a minimum of 100 randomized test iterations
2. FOR ALL valid payloads with sizes from 1 to 1500 bytes, THE Test_Framework SHALL verify that AX.25 framing (createAX25Frame) then extraction (extractAX25Frame) produces a byte sequence identical to the original payload, using a minimum of 100 randomized test iterations
3. IF the CLA link is not open, THEN THE Test_Framework SHALL verify that SendBundle returns an error indicating the link is not open without transmitting data
4. IF the CLA link is not open, THEN THE Test_Framework SHALL verify that RecvBundle returns an error indicating the link is not open without attempting to receive data

### Requirement 23: HDTN Configuration Validation

**User Story:** As a deployment engineer, I want the Test Framework to verify that all HDTN configuration files parse correctly and pass validation, so that deployment errors are caught before runtime.

#### Acceptance Criteria

1. THE Test_Framework SHALL verify that all JSON files in the configs/ and configs/simulation/ directories parse as valid JSON without errors using Go's encoding/json decoder
2. THE Test_Framework SHALL verify that parsed HDTN configurations pass the HDTNConfig validation function, confirming that required fields (node EID, storage path, at least one induct, at least one outduct, and at least one contact plan entry) are present and valid
3. THE Test_Framework SHALL verify that the following packages exist in the project and contain at least one _test.go file: hdtn, hdtnconfig, kiss, bpa, node, contact, store, security, linkbudget, and sband_iq

### Requirement 24: Continuous Integration Pipeline

**User Story:** As a quality assurance engineer, I want the Test Framework to execute all property tests and integration tests automatically on every commit, so that regressions are detected immediately.

#### Acceptance Criteria

1. WHEN a push to main or a pull request targeting main occurs, THE CI_Pipeline SHALL execute all property-based and unit tests across all packages (pkg/hdtn, pkg/hdtnconfig, kiss, cmd/dtn-node, pkg/cla/sband_iq, pkg/store, pkg/bpa, pkg/node, pkg/contact, pkg/iq, pkg/linkbudget, pkg/security, test/integration)
2. WHEN a push to main or a pull request targeting main occurs, THE CI_Pipeline SHALL execute go vet static analysis on all packages
3. THE CI_Pipeline SHALL execute the full build (go build ./...) before running any tests, and abort the pipeline if the build fails
4. THE CI_Pipeline SHALL enforce a test timeout of 300 seconds per test run to prevent hanging tests from blocking the pipeline
5. IF any test or static analysis step fails, THEN THE CI_Pipeline SHALL report the failure with the failing test name and block the pull request from merging

### Requirement 25: Requirement Traceability

**User Story:** As a mission assurance engineer, I want every property test to trace to one or more system requirements via explicit annotations, so that verification coverage can be assessed for flight readiness reviews.

#### Acceptance Criteria

1. THE Test_Framework SHALL include a "Validates: Requirement X.Y" annotation in the documentation comment of every property test function, where X is the requirement number and Y is the criterion number
2. THE Test_Framework SHALL maintain traceability from property tests to system-level requirements across all test packages: bpa, store, contact, hdtn, hdtnconfig, node, security, linkbudget, and sband_iq
3. THE Test_Framework SHALL use a minimum of 50 successful randomized test iterations (MinSuccessfulTests = 50) for computationally expensive property tests (CGR pass prediction, link margin monotonicity across sequences) and a minimum of 100 iterations (MinSuccessfulTests = 100) for all other property tests

### Requirement 26: Test Framework Extensibility

**User Story:** As a systems engineer, I want the Test Framework architecture to support adding new property tests for future mission phases (QO-100, CubeSat EM, LEO, Cislunar), so that verification coverage grows with the system.

#### Acceptance Criteria

1. THE Test_Framework SHALL organize property tests by package, with each package's property tests in files named with the suffix _property_test.go and other tests in files named with the suffix _test.go, following Go testing conventions
2. THE Test_Framework SHALL support both gopter (github.com/leanovate/gopter) and rapid (pgregory.net/rapid) property-based testing libraries for property test implementation
3. THE Test_Framework SHALL support mock HTTP servers (net/http/httptest) for testing components that interact with HDTN REST APIs without requiring a running HDTN_Process
4. THE Test_Framework SHALL support mock CLAs (implementing the ConvergenceLayerAdapter interface) for testing Node_Controller behavior without hardware dependencies or serial port access
