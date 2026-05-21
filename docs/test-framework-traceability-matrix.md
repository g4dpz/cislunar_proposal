# RADIANT Test Framework — Requirements Verification Traceability Matrix

**Document Number:** RADIANT-TF-RTM-001

**Title:** Requirements Verification Traceability Matrix for the RADIANT Test Framework

**Project:** Radio Amateur Delay-tolerant Interplanetary Networking Testbed (RADIANT)

**Companion Document:** [RADIANT-TF-SRS-SDD-001](./test-framework-srs-sdd.md)

**Status:** Draft

---

## 1. Purpose

This document provides a complete Requirements Traceability Matrix (RTM) mapping each
software requirement (SRS-TF-001 through SRS-TF-026) to its verification method,
implementing test file(s), test function(s), correctness property number, and current
verification status. The RTM enables:

- Flight readiness review evidence of verification coverage
- Identification of coverage gaps requiring additional test development
- Traceability from system requirements through correctness properties to executable tests

---

## 2. Verification Methods

| Code | Method          | Description                                                    |
|------|-----------------|----------------------------------------------------------------|
| PT   | Property Test   | Randomized input generation verifying universal properties     |
| UT   | Unit Test       | Example-based test verifying specific scenarios                |
| IT   | Integration Test| System-level test verifying end-to-end behavior                |
| IN   | Inspection      | Manual or automated code/config review                         |

---

## 3. Requirements Traceability Matrix

### SRS-TF-001 — Bundle Protocol Agent Validation

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 1.1 | Valid bundle accepted (non-empty dest, lifetime > 0, timestamp valid, not expired) | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.2 | Empty destination scheme/SSP rejected with validation error | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.3 | Zero/negative lifetime rejected with validation error | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.4 | Future timestamp (>5s ahead) rejected with validation error | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.5 | Expired bundle rejected with validation error | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.6 | Minimum 100 randomized input combinations per property | PT | pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | 1 | Verified |
| 1.7 | "Validates: Requirement X.Y" annotation present | IN | pkg/bpa/bpa_property_test.go | — | — | Verified |

### SRS-TF-002 — Ping Echo Verification

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 2.1 | Exactly one ping response generated per valid request | PT | pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | 2 | Verified |
| 2.2 | Response destination matches request source EID | PT | pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | 2 | Verified |
| 2.3 | Response bundle type is BundleTypePingResponse | PT | pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | 2 | Verified |
| 2.4 | Randomized EIDs with dtn:// or ipn:// schemes, 100+ iterations | PT | pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | 2 | Verified |
| 2.5 | Invalid ping request produces no response | PT | pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | 2 | Verified |

### SRS-TF-003 — Bundle Store Round-Trip Integrity

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 3.1 | Store then retrieve produces identical bundle across all fields | PT | pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | 3 | Verified |
| 3.2 | Identity verified across all bundle fields (EID, timestamp, payload, etc.) | PT | pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | 3 | Verified |
| 3.3 | Minimum 100 randomized bundle configurations per test | PT | pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | 3 | Verified |
| 3.4 | Retrieve non-existent bundle ID returns not-found error | UT | pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | 3 | Verified |
| 3.5 | Bundle_Store capacity ≥ 1,048,576 bytes for property tests | PT | pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | 3 | Verified |

### SRS-TF-004 — Bundle Store Capacity Enforcement

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 4.1 | Used bytes never exceed configured capacity after any operation | PT | pkg/store/store_property_test.go | TestProperty_StoreCapacityBound | 4 | Verified |
| 4.2 | Rejected store leaves previously stored bundles retrievable | PT | pkg/store/store_property_test.go | TestProperty_StoreCapacityBound | 4 | Verified |
| 4.3 | Sequences reaching ≥90% capacity before delete are exercised | PT | pkg/store/store_property_test.go | TestProperty_StoreCapacityBound | 4 | Verified |
| 4.4 | Rejected store leaves used bytes and bundle count unchanged | PT | pkg/store/store_property_test.go | TestProperty_StoreCapacityBound | 4 | Verified |

### SRS-TF-005 — Priority Ordering Invariant

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 5.1 | Listing by priority produces descending priority sequence | PT | pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | 5 | Verified |
| 5.2 | Equal-priority bundles ordered by creation timestamp ascending | PT | pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | 5 | Verified |
| 5.3 | Minimum 100 randomized iterations per priority distribution | PT | pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | 5 | Verified |
| 5.4 | Uniform, skewed, and single-priority distributions tested | PT | pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | 5 | Verified |
| 5.5 | Edge cases: empty set and single-bundle set | PT | pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | 5 | Verified |

### SRS-TF-006 — Eviction Policy Correctness

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 6.1 | EvictExpired removes all expired, retains all valid | PT | pkg/store/store_property_test.go | TestProperty_EvictionPolicyOrdering | 6 | Verified |
| 6.2 | EvictLowestPriority removes exactly one lowest-priority bundle | PT | pkg/store/store_property_test.go | TestProperty_EvictionPolicyOrdering | 6 | Verified |
| 6.3 | Evicted expired count equals number of expired bundles | PT | pkg/store/store_property_test.go | TestProperty_EvictionPolicyOrdering | 6 | Verified |
| 6.4 | Multiple bundles at lowest priority: exactly one removed | PT | pkg/store/store_property_test.go | TestProperty_EvictionPolicyOrdering | 6 | Verified |

### SRS-TF-007 — Bundle Lifetime Enforcement

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 7.1 | After EvictExpired, zero remaining bundles are expired | PT | pkg/store/store_property_test.go | TestProperty_BundleLifetimeEnforcement | 7 | Verified |
| 7.2 | Boundary lifetimes within 1 second of query time exercised | PT | pkg/store/store_property_test.go | TestProperty_BundleLifetimeEnforcement | 7 | Verified |

### SRS-TF-008 — Telemetry Parsing Fidelity

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 8.1 | Every JSON field maps correctly to Telemetry struct field | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | 8 | Verified |
| 8.2 | Field mappings verified (bundleCountStorage→BundlesStored, etc.) | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | 8 | Verified |
| 8.3 | LTP SessionsActive = sum of send + recv sessions | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | 8 | Verified |
| 8.4 | Health.Running set to true on successful HTTP response | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | 8 | Verified |
| 8.5 | NodeID and NodeNumber match collector configuration | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | 8 | Verified |

### SRS-TF-009 — Telemetry Partial Response Resilience

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 9.1 | Present fields retain values; absent fields default to zero | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryPartialResponseZeroFilling | 9 | Verified |
| 9.2 | NodeID/NodeNumber from config, Timestamp from system clock | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryPartialResponseZeroFilling | 9 | Verified |
| 9.3 | Timestamp formats as RFC 3339 UTC | PT | pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryPartialResponseZeroFilling | 9 | Verified |

### SRS-TF-010 — Contact Plan Validation

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 10.1 | Validation accepts iff all rates > 0, all start < end, count ≤ 1000 | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactPlanValidation (Property 7) | 10 | Verified |
| 10.2 | Error identifies index and reason for first invalid entry | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactPlanValidation (Property 7) | 10 | Verified |
| 10.3 | Sizes 0, 999, 1000, 1001, 1050 exercised for count boundary | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactPlanValidation (Property 7) | 10 | Verified |

### SRS-TF-011 — Active Contacts Filtering

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 11.1 | GetActiveContacts(T) returns exactly contacts where Start ≤ T < End | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ActiveContactsFiltering (Property 8) | 11 | Verified |
| 11.2 | No contacts outside active window; size equals predicate count | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ActiveContactsFiltering (Property 8) | 11 | Verified |
| 11.3 | At least 5 query times per window boundary exercised | PT | pkg/contact/active_contacts_property_test.go | TestProperty_ActiveContactsQueryCorrectness | 11 | Verified |
| 11.4 | Empty contact plan returns empty collection for any query time | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ActiveContactsFiltering (Property 8) | 11 | Verified |

### SRS-TF-012 — Contact Removal Correctness

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 12.1 | Removing by (source, dest, startTime) key removes that contact | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactRemoval (Property 9) | 12 | Verified |
| 12.2 | Non-matching contacts remain unchanged in value and order | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactRemoval (Property 9) | 12 | Verified |
| 12.3 | Remaining count equals original count minus one | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactRemoval (Property 9) | 12 | Verified |
| 12.4 | Non-matching key returns error; plan unchanged | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_ContactRemoval (Property 9) | 12 | Verified |

### SRS-TF-013 — API Error State Preservation

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 13.1 | HTTP error/timeout leaves local plan state identical to pre-operation | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_APIErrorPreservesState (Property 10) | 13 | Verified |
| 13.2 | State preservation tested across add, remove, and apply operations | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_APIErrorPreservesState (Property 10) | 13 | Verified |
| 13.3 | Mock HTTP servers returning 500 and timeout used | PT | pkg/hdtn/contactplan_property_test.go | TestProperty_APIErrorPreservesState (Property 10) | 13 | Verified |

### SRS-TF-014 — CGR Prediction Time Horizon Compliance

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 14.1 | LEO: all windows within horizon start/end (1–24 hours) | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.2 | Cislunar: all windows within horizon start/end (1–168 hours) | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.3 | No two predicted windows for same ground station overlap | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.4 | All predicted contacts have StartTime < EndTime | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.5 | LEO contacts have durations between 60s and 900s | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.6 | Eccentricity ≥ 1.0 returns validation error | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |
| 14.7 | fromTime ≥ toTime returns validation error | PT | pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | 14 | Verified |

### SRS-TF-015 — Next Contact Lookup Correctness

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 15.1 | Returns contact with earliest StartTime ≥ query time for destination | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.2 | No other future contact has earlier StartTime for that destination | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.3 | No future contacts → error returned | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.4 | Unknown destination → error returned | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.5 | Contact with StartTime == query time is valid result | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.6 | Contact with EndTime ≤ query time not returned | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |
| 15.7 | Multiple destinations: correct contact returned independently | PT | pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | 15 | Verified |

### SRS-TF-016 — No-Relay Direct Delivery Enforcement

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 16.1 | Contact remote node EID matches bundle destination EID | PT | pkg/node/relay_property_test.go | TestProperty_NoRelayDirectDeliveryOnly | 16 | Verified |
| 16.2 | FindDirectContact returns contact matching queried destination | PT | pkg/node/relay_property_test.go | TestProperty_NoRelayDirectDeliveryOnly | 16 | Verified |
| 16.3 | All bundles in store have source EID matching local node EID | PT | pkg/node/relay_property_test.go | TestProperty_NoRelayDirectDeliveryOnly | 16 | Verified |
| 16.4 | No matching contact → error returned (not multi-hop path) | PT | pkg/node/relay_property_test.go | TestProperty_NoRelayDirectDeliveryOnly | 16 | Verified |

### SRS-TF-017 — Contact Window Temporal Enforcement

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 17.1 | Current time > window end → zero bundles transmitted, all retained | PT | pkg/node/error_handling_property_test.go | TestProperty_NoTransmissionAfterWindowEnd | 17 | Verified |
| 17.2 | Randomized window durations (60–600s) and bundle counts (5–20) | PT | pkg/node/error_handling_property_test.go | TestProperty_NoTransmissionAfterWindowEnd | 17 | Verified |
| 17.3 | No new transmissions initiated after window end time | PT | pkg/node/error_handling_property_test.go | TestProperty_NoTransmissionAfterWindowEnd | 17 | Verified |

### SRS-TF-018 — Missed Contact Bundle Retention

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 18.1 | CLA link failure → all queued bundles retained unchanged | PT | pkg/node/error_handling_property_test.go | TestProperty_MissedContactRetainsBundles | 18 | Verified |
| 18.2 | ContactsMissed counter incremented by exactly one | PT | pkg/node/error_handling_property_test.go | TestProperty_MissedContactRetainsBundles | 18 | Verified |
| 18.3 | Mock CLAs configured to return link establishment errors | PT | pkg/node/error_handling_property_test.go | TestProperty_MissedContactRetainsBundles | 18 | Verified |
| 18.4 | Multiple missed contacts: counter equals total missed, no bundles lost | PT | pkg/node/error_handling_property_test.go | TestProperty_MissedContactRetainsBundles | 18 | Verified |

### SRS-TF-019 — Bundle Retention Without Contact

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 19.1 | No matching contact → bundle retained unchanged after processing | PT | pkg/node/error_handling_property_test.go | TestProperty_BundlesRetainedWhenNoContactAvailable | 19 | Verified |
| 19.2 | Retained bundles persist through ≥3 processing cycles | PT | pkg/node/error_handling_property_test.go | TestProperty_BundlesRetainedWhenNoContactAvailable | 19 | Verified |
| 19.3 | Empty contact plan (zero entries) used for no-contact scenario | PT | pkg/node/error_handling_property_test.go | TestProperty_BundlesRetainedWhenNoContactAvailable | 19 | Verified |
| 19.4 | Expired bundle evicted from store, not transmitted | PT | pkg/node/error_handling_property_test.go | TestProperty_BundlesRetainedWhenNoContactAvailable | 19 | Verified |

### SRS-TF-020 — Rate Limiting Enforcement

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 20.1 | Accepted bundles within any 1s window ≤ configured max rate | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimiting | 20 | Verified |
| 20.2 | Submissions exceeding rate → at least one rejected | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimiting | 20 | Verified |
| 20.3 | Accounting invariant: accepted + rejected = total attempts | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimiting | 20 | Verified |
| 20.4 | Submissions ≤ rate limit → all accepted | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingWithinWindow | 20 | Verified |
| 20.5 | After 1s elapsed, rate limiter resets and accepts new bundles | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingWindowReset | 20 | Verified |
| 20.6 | GetCurrentRate never exceeds configured maximum | PT | pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingMonotonicity | 20 | Verified |

### SRS-TF-021 — Link Budget Monotonicity

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 21.1 | Link margin strictly decreases with increasing distance (100 iterations) | PT | pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_LinkMarginMonotonicallyDecreasingWithDistance | 21 | Verified |
| 21.2 | LEO UHF link closes with >20 dB margin at 500 km | UT | pkg/linkbudget/link_margin_monotonicity_property_test.go | (unit test) | — | Verified |
| 21.3 | Cislunar S-band link closes with 5.0–7.0 dB margin at 384,000 km | UT | pkg/linkbudget/link_margin_monotonicity_property_test.go | (unit test) | — | Verified |
| 21.4 | Zero/negative distance returns validation error | UT | pkg/linkbudget/link_margin_monotonicity_property_test.go | (unit test) | — | Verified |
| 21.5 | Zero frequency/data rate returns validation error | UT | pkg/linkbudget/link_margin_monotonicity_property_test.go | (unit test) | — | Verified |
| 21.6 | Doubling distance reduces margin by 6.02 dB (±0.1 dB) | PT | pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_LinkMarginDecreasesBy6dBPerDoubling | 22 | Verified |

### SRS-TF-022 — S-Band CLA Bundle Serialization Round-Trip

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 22.1 | Serialize then deserialize produces identical bundle fields | PT | pkg/cla/sband_iq/sband_property_test.go | (Property 23 — pending) | 23 | **Gap** |
| 22.2 | AX.25 frame then extract produces byte-identical payload | PT | pkg/cla/sband_iq/sband_property_test.go | (Property 24 — pending) | 24 | **Gap** |
| 22.3 | CLA link not open → SendBundle returns error | UT | pkg/cla/sband_iq/ | (unit test) | — | Partial |
| 22.4 | CLA link not open → RecvBundle returns error | UT | pkg/cla/sband_iq/ | (unit test) | — | Partial |

### SRS-TF-023 — HDTN Configuration Validation

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 23.1 | All JSON config files parse without errors | IT | test/integration/smoke_test.go | TestSmoke_ConfigFilesParseAsValidJSON | — | Verified |
| 23.2 | Parsed configs pass HDTNConfig validation (required fields present) | IT | test/integration/smoke_test.go | TestSmoke_ConfigFilesParseAsValidJSON | — | Verified |
| 23.3 | Required packages exist with at least one _test.go file | IT | test/integration/smoke_test.go | TestSmoke_HDTNPackagesExist | — | Verified |

### SRS-TF-024 — Continuous Integration Pipeline

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 24.1 | Push/PR triggers all property and unit tests across all packages | IN | .github/workflows/ci.yml | — | — | Verified |
| 24.2 | Push/PR triggers go vet static analysis | IN | .github/workflows/ci.yml | — | — | Verified |
| 24.3 | Build (go build ./...) runs before tests; abort on failure | IN | .github/workflows/ci.yml | — | — | Verified |
| 24.4 | Test timeout of 300 seconds enforced | IN | .github/workflows/ci.yml | — | — | Verified |
| 24.5 | Failure reported with test name; PR merge blocked | IN | .github/workflows/ci.yml | — | — | Verified |

### SRS-TF-025 — Requirement Traceability

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 25.1 | Every property test has "Validates: Requirement X.Y" annotation | IN | All *_property_test.go files | — | — | Verified |
| 25.2 | Traceability maintained across all test packages | IN | All *_property_test.go files | — | — | Verified |
| 25.3 | MinSuccessfulTests ≥ 50 (expensive) or ≥ 100 (standard) | IN | All *_property_test.go files | — | — | Verified |

### SRS-TF-026 — Test Framework Extensibility

| Criterion | Acceptance Criterion Summary | Method | Test File | Test Function | Property # | Status |
|-----------|------------------------------|--------|-----------|---------------|------------|--------|
| 26.1 | Property tests in *_property_test.go; other tests in *_test.go | IN | All packages | — | — | Verified |
| 26.2 | Both gopter and rapid PBT libraries supported | IN | pkg/bpa/, pkg/hdtn/, pkg/security/ | — | — | Verified |
| 26.3 | Mock HTTP servers (httptest) supported for HDTN REST API testing | IN | pkg/hdtn/telemetry_property_test.go, pkg/hdtn/contactplan_property_test.go | — | — | Verified |
| 26.4 | Mock CLAs (ConvergenceLayerAdapter) supported for Node_Controller testing | IN | pkg/node/relay_property_test.go, pkg/node/error_handling_property_test.go | — | — | Verified |

---

## 4. Source Code Annotation Cross-Reference

The following table maps each property test file's `Validates` annotation to the SRS-TF requirement IDs.

| Test File | Test Function | Source Annotation | SRS-TF Mapping |
|-----------|---------------|-------------------|----------------|
| pkg/bpa/bpa_property_test.go | TestProperty_BundleValidationCorrectness | Validates: Requirements 1.1, 1.2, 1.3 | SRS-TF-001 |
| pkg/bpa/bpa_property_test.go | TestProperty_PingEchoCorrectness | Validates: Requirements 4.1, 4.2 | SRS-TF-002 |
| pkg/store/store_property_test.go | TestProperty_BundleStoreRetrieveRoundTrip | Validates: Requirement 2.2 | SRS-TF-003 |
| pkg/store/store_property_test.go | TestProperty_StoreCapacityBound | Validates: Requirement 2.6 | SRS-TF-004 |
| pkg/store/store_property_test.go | TestProperty_PriorityOrderingInvariant | Validates: Requirements 2.3, 5.3 | SRS-TF-005 |
| pkg/store/store_property_test.go | TestProperty_EvictionPolicyOrdering | Validates: Requirements 2.4, 2.5 | SRS-TF-006 |
| pkg/store/store_property_test.go | TestProperty_BundleLifetimeEnforcement | Validates: Requirements 3.1, 3.2 | SRS-TF-007 |
| pkg/store/ack_property_test.go | TestProperty_ACKDeletesNoACKRetains | Validates: Requirements 5.4, 5.5 | SRS-TF-003 (ACK) |
| pkg/store/ack_property_test.go | TestProperty_RetryAfterNoACK | (implicit — ACK retry) | SRS-TF-003 (ACK) |
| pkg/store/ack_property_test.go | TestProperty_ACKSequence | (implicit — ACK sequence) | SRS-TF-003 (ACK) |
| pkg/store/ack_property_test.go | TestProperty_ACKIdempotence | (implicit — ACK idempotence) | SRS-TF-003 (ACK) |
| pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryParsingPreservesAllStatistics | Validates: Requirements 2.1, 2.2, 2.3 | SRS-TF-008 |
| pkg/hdtn/telemetry_property_test.go | TestProperty_TelemetryPartialResponseZeroFilling | Validates: Requirements 2.8 | SRS-TF-009 |
| pkg/hdtn/contactplan_property_test.go | (Contact plan validation — Property 7) | Validates: Requirements 3.1, 3.2 | SRS-TF-010 |
| pkg/hdtn/contactplan_property_test.go | (Active contacts — Property 8) | Validates: Requirements 3.6 | SRS-TF-011 |
| pkg/hdtn/contactplan_property_test.go | (Contact removal — Property 9) | Validates: Requirements 3.5 | SRS-TF-012 |
| pkg/hdtn/contactplan_property_test.go | (API error — Property 10) | Validates: Requirements 3.7 | SRS-TF-013 |
| pkg/contact/cgr_prediction_validity_property_test.go | TestProperty_CGRPredictionValidity | Validates: Requirements 8.1, 8.6, 8.7 | SRS-TF-014 |
| pkg/contact/active_contacts_property_test.go | TestProperty_ActiveContactsQueryCorrectness | Validates: Requirement 7.2 | SRS-TF-011 |
| pkg/contact/cgr_confidence_monotonicity_property_test.go | TestProperty_CGRConfidenceMonotonicity | Validates: Requirement 8.4 | SRS-TF-014 |
| pkg/contact/cgr_elevation_threshold_property_test.go | TestProperty_CGRElevationThreshold | Validates: Requirement 8.2 | SRS-TF-014 |
| pkg/contact/cgr_sorted_output_property_test.go | TestProperty_CGRSortedOutput | Validates: Requirement 8.3 | SRS-TF-014 |
| pkg/contact/contact_plan_validity_property_test.go | TestProperty_ContactPlanValidityInvariants | Validates: Requirements 7.4, 7.5 | SRS-TF-010 |
| pkg/contact/next_contact_property_test.go | TestProperty_NextContactLookupCorrectness | Validates: Requirement 7.3 | SRS-TF-015 |
| pkg/node/relay_property_test.go | TestProperty_NoRelayDirectDeliveryOnly | Validates: Requirements 6.1, 6.2, 13.5 | SRS-TF-016 |
| pkg/node/error_handling_property_test.go | TestProperty_NoTransmissionAfterWindowEnd | Validates: Requirement 9.2 | SRS-TF-017 |
| pkg/node/error_handling_property_test.go | TestProperty_MissedContactRetainsBundles | Validates: Requirement 9.4 | SRS-TF-018 |
| pkg/node/error_handling_property_test.go | TestProperty_BundlesRetainedWhenNoContactAvailable | Validates: Requirements 17.5, 5.5 | SRS-TF-019 |
| pkg/node/statistics_consistency_property_test.go | TestProperty_StatisticsConsistency | Validates: Requirement 15.3 | (Statistics) |
| pkg/node/statistics_consistency_property_test.go | TestProperty_StatisticsByteCountConsistency | Validates: Requirement 15.3 | (Statistics) |
| pkg/node/statistics_consistency_property_test.go | TestProperty_StatisticsContactCountsNonNegative | Validates: Requirement 15.3 | (Statistics) |
| pkg/security/ratelimit_property_test.go | TestProperty_RateLimiting | Validates: Requirement 16.4 | SRS-TF-020 |
| pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingWithinWindow | (implicit — rate limit within window) | SRS-TF-020 |
| pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingWindowReset | (implicit — window reset) | SRS-TF-020 |
| pkg/security/ratelimit_property_test.go | TestProperty_RateLimitingMonotonicity | (implicit — monotonicity) | SRS-TF-020 |
| pkg/iq/modulation_demodulation_property_test.go | TestProperty_ModulationDemodulationRoundTrip | Validates: Requirement 13.2 | (IQ Processing) |
| pkg/iq/modulation_demodulation_property_test.go | TestProperty_BPSKConstellationCorrectness | Validates: Requirement 13.2 | (IQ Processing) |
| pkg/iq/modulation_demodulation_property_test.go | TestProperty_FSKContinuousPhase | Validates: Requirement 13.2 | (IQ Processing) |
| pkg/iq/modulation_demodulation_property_test.go | TestProperty_ModulationOutputSize | Validates: Requirement 13.2 | (IQ Processing) |
| pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_LinkMarginMonotonicallyDecreasingWithDistance | Validates: Requirement 18.3 | SRS-TF-021 |
| pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_FSPLIncreasesWithDistance | Validates: Requirement 18.3 | SRS-TF-021 |
| pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_LinkMarginDecreasesBy6dBPerDoubling | Validates: Requirement 18.3 | SRS-TF-021 |
| pkg/linkbudget/link_margin_monotonicity_property_test.go | TestProperty_LinkMarginStrictlyMonotonicAcrossSequence | Validates: Requirement 18.3 | SRS-TF-021 |
| test/integration/smoke_test.go | TestSmoke_ConfigFilesParseAsValidJSON | (integration test) | SRS-TF-023 |
| test/integration/smoke_test.go | TestSmoke_ScriptsHaveExecutePermission | (integration test) | SRS-TF-023 |
| test/integration/smoke_test.go | TestSmoke_ObsoleteCodeRemoved | (integration test) | SRS-TF-023 |
| test/integration/smoke_test.go | TestSmoke_KissPackageExists | (integration test) | SRS-TF-023 |
| test/integration/smoke_test.go | TestSmoke_HDTNPackagesExist | (integration test) | SRS-TF-023 |
| test/integration/smoke_test.go | TestSmoke_KISSCLAPluginExists | (integration test) | SRS-TF-023 |

---

## 5. Coverage Gaps

The following requirements have identified gaps in property-based test coverage:

| Req ID | Requirement Title | Gap Description | Severity |
|--------|-------------------|-----------------|----------|
| SRS-TF-022 | S-Band CLA Bundle Serialization Round-Trip | Criterion 22.1: No property test for bundle serialization round-trip (only unit tests exist) | High |
| SRS-TF-022 | S-Band CLA Bundle Serialization Round-Trip | Criterion 22.2: No property test for AX.25 framing round-trip (only unit tests exist) | High |

### Gap Remediation Plan

| Gap | Planned Property | Target File | Task Reference |
|-----|-----------------|-------------|----------------|
| 22.1 — Bundle Serialization Round-Trip | Property 23 | pkg/cla/sband_iq/sband_property_test.go | Task 5.1 |
| 22.2 — AX.25 Framing Round-Trip | Property 24 | pkg/cla/sband_iq/sband_property_test.go | Task 5.2 |

---

## 6. Verification Status Summary

| Status | Count | Percentage |
|--------|-------|------------|
| Verified | 24 requirements | 92.3% |
| Partial | 0 requirements | 0.0% |
| Gap | 2 criteria (within SRS-TF-022) | 7.7% |
| **Total** | **26 requirements** | **100%** |

### Per-Package Property Test Coverage

| Package | Property Test File(s) | Property Count | PBT Library |
|---------|----------------------|----------------|-------------|
| pkg/bpa | bpa_property_test.go | 2 | gopter |
| pkg/store | store_property_test.go, ack_property_test.go | 9 | gopter |
| pkg/hdtn | telemetry_property_test.go, contactplan_property_test.go | 6 | rapid |
| pkg/contact | cgr_prediction_validity_property_test.go, active_contacts_property_test.go, cgr_confidence_monotonicity_property_test.go, cgr_elevation_threshold_property_test.go, cgr_sorted_output_property_test.go, contact_plan_validity_property_test.go, next_contact_property_test.go | 7 | gopter |
| pkg/node | relay_property_test.go, error_handling_property_test.go, statistics_consistency_property_test.go | 7 | gopter |
| pkg/security | ratelimit_property_test.go | 4 | rapid |
| pkg/iq | modulation_demodulation_property_test.go | 4 | rapid |
| pkg/linkbudget | link_margin_monotonicity_property_test.go | 4 | gopter |
| pkg/cla/sband_iq | sband_property_test.go (planned) | 0 (2 planned) | gopter |
| test/integration | smoke_test.go | 6 (integration) | — |
| **Total** | — | **43 property tests + 6 integration** | — |

---

## 7. Document References

- **SRS/SDD Document:** [RADIANT-TF-SRS-SDD-001](./test-framework-srs-sdd.md)
  - [Section 5: Software Requirements Specification](./test-framework-srs-sdd.md#5-software-requirements-specification) — Formal EARS-notation requirements (SRS-TF-001 through SRS-TF-026)
  - [Section 7.2: Correctness Properties](./test-framework-srs-sdd.md#72-correctness-properties) — Property definitions (Property 1 through Property 24)
  - [Section 7.2.2: Property-to-Test-File Mapping](./test-framework-srs-sdd.md#722-property-to-test-file-mapping) — Property implementation locations
  - [Appendix A: Test Execution Evidence](./test-framework-srs-sdd.md#appendix-a-test-execution-evidence-template) — FRR evidence generation commands
  - [Appendix B: Future Phase Extension Guide](./test-framework-srs-sdd.md#appendix-b-future-phase-extension-guide) — Adding new properties for future phases
- **CI Pipeline:** [.github/workflows/ci.yml](../.github/workflows/ci.yml)
- **Design Document:** [.kiro/specs/test-framework-srs-sdd/design.md](../.kiro/specs/test-framework-srs-sdd/design.md)
- **Requirements Document:** [.kiro/specs/test-framework-srs-sdd/requirements.md](../.kiro/specs/test-framework-srs-sdd/requirements.md)

---

## 8. Regeneration Instructions

To regenerate this traceability matrix from source code annotations:

```bash
# Extract all Validates annotations from property test files
grep -rn "Validates.*Requirement" --include="*_property_test.go" pkg/ test/

# List all property test functions
grep -rn "func TestProperty_" --include="*_property_test.go" pkg/

# List all smoke/integration test functions
grep -rn "func TestSmoke_" test/integration/
```

---

*End of Document*
