# Implementation Plan: RADIANT Test Framework SRS/SDD

## Overview

This plan creates the formal Software Requirements Specification (SRS) and Software Design Description (SDD) document for the RADIANT Test Framework, modeled after NASA Glenn's HDTN Test Framework (TM-20240014467 / LEW-20818-1). The deliverables are documentation artifacts (the combined SRS/SDD document and traceability matrix), an audit of existing property test annotations, and gap-filling property tests where the existing ~35 tests don't cover the requirements.

## Tasks

- [x] 1. Create the formal SRS/SDD document structure
  - [x] 1.1 Create `docs/test-framework-srs-sdd.md` with front matter, document control, and section scaffolding
    - Create the file with NASA TM-style front matter (title, document number, authors, revision history)
    - Include sections: 1. Introduction, 2. Scope, 3. Referenced Documents, 4. Glossary, 5. Software Requirements Specification, 6. Software Design Description, 7. Verification Approach, 8. Appendices
    - Populate Section 1 (Introduction) with purpose, system overview, and document conventions
    - Populate Section 2 (Scope) with system boundaries, mission phases, and exclusions
    - Populate Section 3 (Referenced Documents) citing NASA/TM-20240014467, RFC 9171 (BPv7), RFC 5050, CCSDS 734.2-B-1 (LTP)
    - Populate Section 4 (Glossary) from the requirements document glossary terms
    - _Requirements: 1.1–1.7, 2.1–2.5, 3.1–3.5, 4.1–4.4, 5.1–5.5, 6.1–6.4, 7.1–7.2, 8.1–8.5, 9.1–9.3, 10.1–10.3, 11.1–11.4, 12.1–12.4, 13.1–13.3, 14.1–14.7, 15.1–15.7, 16.1–16.4, 17.1–17.3, 18.1–18.4, 19.1–19.4, 20.1–20.6, 21.1–21.6, 22.1–22.4, 23.1–23.3, 24.1–24.5, 25.1–25.3, 26.1–26.4_

  - [x] 1.2 Populate Section 5 (SRS) with all 26 requirements in formal EARS notation
    - Transcribe each requirement from requirements.md into formal SRS format with unique identifiers (SRS-TF-001 through SRS-TF-026)
    - Include acceptance criteria as SHALL-statements with verification method (Property Test, Unit Test, Integration Test, Inspection)
    - Add priority classification (Critical, High, Medium) for each requirement
    - Add parent-child traceability to system-level requirements where applicable
    - _Requirements: 1.1–1.7, 2.1–2.5, 3.1–3.5, 4.1–4.4, 5.1–5.5, 6.1–6.4, 7.1–7.2, 8.1–8.5, 9.1–9.3, 10.1–10.3, 11.1–11.4, 12.1–12.4, 13.1–13.3, 14.1–14.7, 15.1–15.7, 16.1–16.4, 17.1–17.3, 18.1–18.4, 19.1–19.4, 20.1–20.6, 21.1–21.6, 22.1–22.4, 23.1–23.3, 24.1–24.5, 25.1–25.3, 26.1–26.4_

  - [x] 1.3 Populate Section 6 (SDD) with architecture, components, interfaces, and data models
    - Document the layered architecture (CI Layer → Test Specification Layer → Test Infrastructure Layer → SUT)
    - Document each component: Property Test Files, PBT Library Interface (gopter/rapid), Mock Infrastructure, Integration/Smoke Tests, CI Pipeline
    - Include interface definitions (ConvergenceLayerAdapter, PBT library APIs)
    - Include data models (Bundle, Contact, Telemetry, TestParameters)
    - Include Mermaid architecture and sequence diagrams from the design document
    - _Requirements: 24.1–24.5, 25.1–25.3, 26.1–26.4_

  - [x] 1.4 Populate Section 7 (Verification Approach) with correctness properties and test strategy
    - Document all 24 correctness properties with formal property statements
    - Map each property to its implementing test file and function
    - Document the dual testing approach (PBT + unit/example-based)
    - Document CI pipeline verification flow
    - Include test execution parameters (MinSuccessfulTests, timeout, shrinking)
    - _Requirements: 24.1–24.5, 25.1–25.3_

- [x] 2. Create the requirements verification traceability matrix
  - [x] 2.1 Create `docs/test-framework-traceability-matrix.md` with the full RTM
    - Create a table mapping each requirement (SRS-TF-001 through SRS-TF-026) to its verification method, implementing test file(s), test function(s), and verification status (Verified/Partial/Gap)
    - Include columns: Req ID, Requirement Title, Acceptance Criterion, Verification Method, Test File, Test Function, Property #, Status
    - Cross-reference the existing `Validates: Requirement X.Y` annotations from all property test files
    - Identify coverage gaps where requirements have no corresponding property test
    - _Requirements: 25.1, 25.2_

  - [x] 2.2 Add a coverage summary section to the traceability matrix
    - Calculate and document: total requirements, requirements with full PBT coverage, requirements with partial coverage, requirements with gaps
    - List specific gaps identified (e.g., Requirement 22 S-Band CLA serialization has unit tests but no property tests)
    - Include a per-package coverage table showing property count per package
    - _Requirements: 25.1, 25.2, 25.3_

- [x] 3. Checkpoint - Review document structure
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Audit and standardize existing property test annotations
  - [x] 4.1 Standardize annotation format across all property test files
    - Audit all `*_property_test.go` files for annotation consistency
    - Standardize to the format: `// Feature: test-framework-srs-sdd, Property N: {title}` followed by `// **Validates: Requirements X.Y, X.Z**`
    - Update annotations in `pkg/bpa/bpa_property_test.go` — currently uses old format `// Validates: Requirements 1.1, 1.2, 1.3` (missing `**` markers and Feature prefix)
    - Update annotations in `pkg/store/store_property_test.go` — currently uses old format `// Validates: Requirement 2.2` (missing `**` markers and Feature prefix)
    - Ensure requirement numbers in annotations match the SRS-TF requirement IDs from the formal document
    - _Requirements: 25.1, 25.2_

  - [x] 4.2 Update annotations in `pkg/hdtn/telemetry_property_test.go` and `pkg/hdtn/contactplan_property_test.go`
    - Current annotations reference `hdtn-migration` feature — add cross-reference to SRS-TF requirement IDs
    - Verify property numbers align with the SRS/SDD document's property numbering (Properties 8–13)
    - _Requirements: 25.1, 25.2_

  - [x] 4.3 Update annotations in `pkg/node/`, `pkg/security/`, `pkg/linkbudget/`, `pkg/contact/`, and `pkg/iq/` property test files
    - Verify all property tests have consistent annotation format
    - Map existing requirement references (e.g., `Requirement 9.2`, `Requirement 18.3`) to the new SRS-TF numbering scheme
    - Add missing Feature prefix where absent
    - _Requirements: 25.1, 25.2_

- [x] 5. Fill identified property test gaps
  - [x] 5.1 Create `pkg/cla/sband_iq/sband_property_test.go` with Property 23 (Bundle Serialization Round-Trip)
    - Implement property test using gopter: for any valid bundle with payload 1–1500 bytes, serializeBundle then deserializeBundle produces identical bundle type, priority, lifetime, destination, and payload
    - Use generators for payload size (1–1500), priority (0–3), lifetime (1–3600), and random payload content
    - MinSuccessfulTests = 100
    - Annotate: `// Feature: test-framework-srs-sdd, Property 23: Bundle Serialization Round-Trip`
    - `// **Validates: Requirements 22.1**`
    - _Requirements: 22.1_

  - [x] 5.2 Add Property 24 (AX.25 Framing Round-Trip) to `pkg/cla/sband_iq/sband_property_test.go`
    - Implement property test using gopter: for any valid payload with size 1–1500 bytes, createAX25Frame then extractAX25Frame produces byte-identical payload
    - Use generators for payload size (1–1500) and random byte content
    - MinSuccessfulTests = 100
    - Annotate: `// Feature: test-framework-srs-sdd, Property 24: AX.25 Framing Round-Trip`
    - `// **Validates: Requirements 22.2**`
    - _Requirements: 22.2_

  - [ ]* 5.3 Write property test for store ACK handling if not already covered
    - Verify `pkg/store/ack_property_test.go` covers ACK round-trip and idempotency properties
    - If gaps exist, add missing property tests with proper annotations
    - _Requirements: 3.1, 3.2_

- [x] 6. Checkpoint - Verify all property tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Final document integration and cross-references
  - [x] 7.1 Add cross-references between SRS/SDD document and traceability matrix
    - Add hyperlinks from the SRS/SDD verification section to the traceability matrix
    - Add a "Document References" section in the traceability matrix pointing back to the SRS/SDD
    - Ensure property numbers are consistent across all three artifacts (SRS/SDD, traceability matrix, source code annotations)
    - _Requirements: 25.1, 25.2_

  - [x] 7.2 Add Appendix A to SRS/SDD: Test Execution Evidence template
    - Create a template section showing how to generate test execution evidence for flight readiness reviews
    - Include `go test -v` output format, coverage report generation (`go test -cover`), and CI badge status
    - Document the command to regenerate the traceability matrix from source annotations
    - _Requirements: 24.1–24.5, 25.1–25.3_

  - [x] 7.3 Add Appendix B to SRS/SDD: Future Phase Extension Guide
    - Document how to add new property tests for QO-100, CubeSat EM, LEO, and Cislunar phases
    - Include the naming convention, annotation format, and CI integration steps
    - Reference the extensibility architecture from the design document
    - _Requirements: 26.1–26.4_

- [x] 8. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- The primary deliverables are documentation artifacts (`docs/test-framework-srs-sdd.md` and `docs/test-framework-traceability-matrix.md`)
- Property tests in task 5 fill gaps identified during the audit — the existing ~35 property tests are NOT recreated
- Annotation standardization (task 4) updates comment formatting only — no functional test changes
- The SRS/SDD document follows NASA TM formatting conventions per the referenced HDTN Test Framework document

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3", "2.1"] },
    { "id": 2, "tasks": ["1.4", "2.2", "4.1"] },
    { "id": 3, "tasks": ["4.2", "4.3", "5.1"] },
    { "id": 4, "tasks": ["5.2", "5.3"] },
    { "id": 5, "tasks": ["7.1", "7.2", "7.3"] }
  ]
}
```
