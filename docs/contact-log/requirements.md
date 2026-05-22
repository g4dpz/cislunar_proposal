# Requirements Document

## Introduction

This document specifies the requirements for a versioned contact-plan and run-evidence logging system for the RADIANT project. The Contact Log captures both planned (expected) and actual (observed) contact behavior for each DTN session, enabling cross-phase comparison of DTN performance across terrestrial nodes (Phase 1, 9600 baud VHF/UHF), QO-100 GEO tests (Phase 1.5, S-band), CubeSat EM (Phase 2), LEO (Phase 3, UHF 437 MHz), and Cislunar (Phase 4, S-band 2.2 GHz).

Each phase has different link characteristics (delay, bandwidth, availability windows). The Contact Log provides a structured, versioned record that makes results comparable across phases and reproducible over time. The system integrates with the existing Go codebase, HDTN REST API telemetry, and YAML/JSON contact plan files.

## Glossary

- **Contact_Log**: The top-level subsystem responsible for creating, storing, versioning, and querying contact log entries
- **Log_Entry**: A single versioned record capturing the planned contact parameters and observed run evidence for one DTN session
- **Contact_Plan_Snapshot**: An immutable copy of the contact plan parameters that were active at the time a session was initiated
- **Run_Evidence**: The observed telemetry and delivery outcomes collected during and after a DTN session
- **Session**: A single DTN communication attempt between two nodes during a contact window, from link establishment through teardown
- **Phase_Metadata**: Phase-specific parameters (link type, frequency band, expected delay, modulation) that contextualize a log entry
- **Log_Version**: A monotonically increasing integer identifying the schema version of a log entry, enabling forward-compatible evolution
- **Log_Store**: The persistent storage backend for log entries, supporting append, query, and export operations
- **Contact_Plan_Manager**: The existing subsystem that maintains scheduled communication windows between nodes
- **Telemetry_Collector**: The existing HDTN REST API client that collects bundle protocol, LTP, and health telemetry
- **DTN_EID**: A DTN Endpoint Identifier using the "dtn" or "ipn" URI scheme that uniquely addresses a node or application
- **Bundle_ID**: A unique identifier for a bundle consisting of source Endpoint_ID, creation timestamp, and sequence number
- **Priority_Class**: One of four priority levels (critical, expedited, normal, bulk) assigned to bundles
- **Custody_Transfer**: A DTN mechanism where an intermediate or destination node accepts responsibility for bundle delivery

## Requirements

### Requirement 1: Log Entry Creation

**User Story:** As a DTN operator, I want a structured log entry to be automatically created for each DTN session, so that I have a complete record of planned and observed contact behavior without manual intervention.

#### Acceptance Criteria

1. WHEN a Contact_Window becomes active and the Node_Controller initiates a session, THE Contact_Log SHALL create a new Log_Entry containing the Contact_Plan_Snapshot and Phase_Metadata
2. WHEN a session completes (link teardown or Contact_Window end time reached), THE Contact_Log SHALL finalize the Log_Entry by appending the Run_Evidence collected during the session
3. IF a session is interrupted before the Contact_Window end time (link failure, TNC disconnection, or process crash), THEN THE Contact_Log SHALL mark the Log_Entry as incomplete and record the interruption reason and the timestamp of the interruption
4. THE Contact_Log SHALL assign a unique, monotonically increasing entry identifier to each Log_Entry at creation time
5. WHEN a Log_Entry is created, THE Contact_Log SHALL record the log schema version (Log_Version) in the entry so that entries remain interpretable as the schema evolves

### Requirement 2: Contact Plan Snapshot Capture

**User Story:** As a DTN researcher, I want each log entry to contain an immutable snapshot of the contact plan parameters that were active at session start, so that I can see exactly what was planned regardless of later plan changes.

#### Acceptance Criteria

1. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the following Contact_Plan_Snapshot fields: contact window start time (Unix epoch seconds), contact window end time (Unix epoch seconds), predicted duration in seconds, remote node identifier, and data rate in bits per second
2. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the link type (VHF, UHF TNC, UHF IQ, S-band IQ, X-band IQ) and frequency band from the active Contact_Window
3. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the modem type, modulation scheme, and framing assumptions configured for the session
4. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the local node DTN_EID, remote node DTN_EID, and the role of each node (sender, receiver, or bidirectional)
5. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the configured storage limit in bytes for the local node
6. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the Priority_Class distribution of bundles queued for the session (count per priority level)
7. WHEN a Log_Entry is created, THE Contact_Log SHALL capture the retransmission policy (LTP retransmission timeout, maximum retries) and custody transfer setting (enabled or disabled)

### Requirement 3: Run Evidence Collection

**User Story:** As a DTN researcher, I want each log entry to contain the actual observed outcomes of the session, so that I can compare planned versus actual performance.

#### Acceptance Criteria

1. WHEN a session completes, THE Contact_Log SHALL record the actual link establishment time, actual link teardown time, and actual session duration in seconds
2. WHEN a session completes, THE Contact_Log SHALL record the total bundles sent, total bundles received, total bytes sent, total bytes received, and the effective throughput in bits per second
3. WHEN a session completes, THE Contact_Log SHALL record the Bundle_ID and delivery status (delivered, failed, pending, expired) for each bundle transmitted or received during the session
4. WHEN a session completes, THE Contact_Log SHALL record the route selected for each bundle (next-hop node identifier and contact used)
5. WHEN a session completes, THE Contact_Log SHALL record LTP session statistics: segments sent, segments received, retransmissions, and sessions completed versus sessions failed
6. WHEN a session completes, THE Contact_Log SHALL record delivery timing evidence: time from bundle creation to first transmission attempt, time from first transmission to acknowledgment, and end-to-end latency for each delivered bundle
7. IF custody transfer is enabled, THEN THE Contact_Log SHALL record custody acceptance timestamps and custody signal outcomes for each bundle

### Requirement 4: Phase Metadata

**User Story:** As a DTN researcher comparing results across mission phases, I want each log entry to include phase-specific context, so that I can normalize and compare performance metrics across different link environments.

#### Acceptance Criteria

1. WHEN a Log_Entry is created, THE Contact_Log SHALL record the mission phase identifier (terrestrial, qo-100-geo, cubesat-em, leo-cubesat, cislunar)
2. WHEN a Log_Entry is created, THE Contact_Log SHALL record the expected one-way light time in seconds for the link (0 for terrestrial, approximately 0.12 for GEO, variable for LEO and cislunar)
3. WHEN a Log_Entry is created, THE Contact_Log SHALL record the frequency band (VHF 144 MHz, UHF 437 MHz, S-band 2.4 GHz, S-band 2.2 GHz, X-band) and nominal bitrate in bits per second
4. WHEN a Log_Entry is created for a LEO or cislunar phase, THE Contact_Log SHALL record orbital parameters (semi-major axis, eccentricity, inclination) and predicted maximum elevation angle
5. WHEN a Log_Entry is created for a GEO phase, THE Contact_Log SHALL record the transponder identifier and uplink/downlink frequency pair

### Requirement 5: Log Versioning and Schema Evolution

**User Story:** As a long-running project spanning multiple phases, I want the log format to be versioned so that older entries remain readable as the schema evolves over time.

#### Acceptance Criteria

1. THE Contact_Log SHALL include a schema version field (positive integer) in every Log_Entry
2. WHEN the log schema changes (fields added, renamed, or deprecated), THE Contact_Log SHALL increment the schema version number
3. THE Contact_Log SHALL read and interpret Log_Entries written with any prior schema version without error
4. WHEN reading a Log_Entry with an older schema version, THE Contact_Log SHALL apply default values for fields that were added in later schema versions
5. THE Contact_Log SHALL reject any Log_Entry whose schema version is higher than the version supported by the running software, returning a descriptive error identifying the unsupported version

### Requirement 6: Log Serialization and Storage

**User Story:** As a DTN operator, I want log entries to be stored in a machine-readable and human-readable format, so that I can use automated tools for analysis and also review entries manually.

#### Acceptance Criteria

1. THE Log_Store SHALL serialize each Log_Entry as a JSON object with consistent field ordering and two-space indentation for human readability
2. THE Log_Store SHALL persist each Log_Entry atomically to the local filesystem, preventing partial writes if the process is interrupted
3. THE Log_Store SHALL organize log entries in a directory structure partitioned by mission phase and date (phase/YYYY-MM-DD/)
4. THE Log_Store SHALL support appending new entries without modifying existing entries (append-only storage)
5. FOR ALL valid Log_Entry objects, serializing a Log_Entry to JSON and then parsing the JSON back SHALL produce a Log_Entry equivalent to the original (round-trip property)
6. THE Log_Store SHALL enforce a configurable maximum log retention period in days, deleting entries older than the retention period during cleanup cycles

### Requirement 7: Log Querying and Export

**User Story:** As a DTN researcher, I want to query and export log entries by phase, time range, node pair, and outcome, so that I can perform cross-phase comparison and trend analysis.

#### Acceptance Criteria

1. WHEN queried with a mission phase filter, THE Log_Store SHALL return all Log_Entries matching the specified phase
2. WHEN queried with a time range (start timestamp, end timestamp), THE Log_Store SHALL return all Log_Entries whose session start time falls within the specified range
3. WHEN queried with a node pair filter (local DTN_EID, remote DTN_EID), THE Log_Store SHALL return all Log_Entries involving the specified node pair in either direction
4. WHEN queried with a delivery outcome filter (all-delivered, partial-delivery, all-failed), THE Log_Store SHALL return Log_Entries matching the specified outcome category
5. THE Log_Store SHALL support exporting query results as a single JSON array document or as newline-delimited JSON (one entry per line) for streaming processing
6. THE Log_Store SHALL complete a query over up to 10,000 stored entries within 5 seconds

### Requirement 8: Integration with Contact Plan Manager

**User Story:** As a DTN operator, I want the contact log to automatically capture contact plan state from the existing Contact_Plan_Manager, so that logging requires no additional manual configuration.

#### Acceptance Criteria

1. WHEN a session begins, THE Contact_Log SHALL retrieve the active Contact_Window from the Contact_Plan_Manager and include it in the Contact_Plan_Snapshot
2. WHEN the Contact_Plan_Manager loads or updates a contact plan, THE Contact_Log SHALL record the plan identifier, plan generation timestamp, and plan valid-from/valid-to range in subsequent Log_Entries
3. THE Contact_Log SHALL reference the Contact_Plan_Manager's contact plan version so that log entries can be correlated with specific plan revisions

### Requirement 9: Integration with Telemetry Collector

**User Story:** As a DTN operator, I want the contact log to automatically collect run evidence from the existing HDTN telemetry system, so that observed metrics are captured without duplicate instrumentation.

#### Acceptance Criteria

1. WHEN a session begins, THE Contact_Log SHALL collect a telemetry baseline snapshot from the Telemetry_Collector (bundle counts, byte counts, LTP session counts)
2. WHEN a session completes, THE Contact_Log SHALL collect a telemetry final snapshot from the Telemetry_Collector and compute deltas (bundles sent during session, bytes transferred during session, retransmissions during session)
3. IF the Telemetry_Collector is unavailable during a session, THEN THE Contact_Log SHALL mark the Run_Evidence telemetry fields as unavailable and record the collection failure reason

### Requirement 10: Cross-Phase Comparison Support

**User Story:** As a DTN researcher, I want to compare performance metrics across mission phases using normalized fields, so that I can evaluate how DTN performs under different link conditions.

#### Acceptance Criteria

1. THE Contact_Log SHALL compute and record a normalized goodput metric: (useful bytes delivered) divided by (actual session duration in seconds) for each completed session
2. THE Contact_Log SHALL compute and record a plan adherence metric: (actual session duration) divided by (planned session duration) for each completed session
3. THE Contact_Log SHALL compute and record a delivery success ratio: (bundles delivered) divided by (bundles attempted) for each completed session
4. WHEN queried for cross-phase comparison, THE Log_Store SHALL return entries from multiple phases with consistent field names and units, enabling direct metric comparison without field mapping

