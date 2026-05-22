# Requirements Document

## Introduction

This feature extends the RADIANT project's existing single-hop contact plan system to support multi-node contact graphs with time-dependent routing and automated contact plan distribution. The system models a distributed ground station network where bundles can be routed through multiple intermediate nodes (ground stations, LEO satellites, GEO relays, cislunar payloads) using store-and-forward semantics. A Contact Graph Routing (CGR) algorithm computes optimal multi-hop paths that minimize end-to-end delivery time while respecting storage constraints, link capacities, and time-varying contact windows.

The system also provides mechanisms for propagating contact plans to all nodes in the network: ground stations receive plans via a REST API from a central planning service, while space nodes (LEO satellites, cislunar payloads) receive plan updates as administrative DTN bundles transmitted over-the-air during contact windows.

## Glossary

- **Contact_Graph**: A directed graph where nodes represent DTN endpoints and edges represent time-bounded communication opportunities (contacts) between node pairs
- **Contact_Graph_Router**: The component that computes time-dependent shortest paths through the Contact_Graph using modified Dijkstra's algorithm
- **Network_Topology**: The definition of all nodes, their types, locations, orbital parameters, and pairwise link capabilities
- **Node**: A DTN endpoint capable of sending, receiving, or relaying bundles; classified as Ground_Station, LEO_Satellite, GEO_Relay, or Cislunar_Payload
- **Ground_Station**: A terrestrial node with a fixed geographic location and radio equipment for space and/or terrestrial links
- **LEO_Satellite**: A low Earth orbit spacecraft (altitude < 1600 km) with limited onboard storage and periodic ground contact
- **GEO_Relay**: A geostationary orbit relay (e.g., QO-100) providing continuous coverage within its footprint
- **Cislunar_Payload**: A spacecraft beyond LEO (lunar orbit, Lagrange points) with long-duration contact windows and high latency
- **Contact**: A time-bounded, directed communication opportunity between two nodes with defined start time, end time, data rate, and one-way light time
- **Route**: An ordered sequence of Contacts from source to destination, with store-and-forward delays at intermediate nodes
- **Store_And_Forward_Delay**: The time a bundle waits at an intermediate node between arrival on one Contact and departure on the next
- **Pairwise_Contact_Generator**: The component that computes Contact windows between each pair of nodes based on their types and orbital parameters
- **Link_Class**: A categorization of communication links: Terrestrial (ground-to-ground wired/RF), GEO_Relay (via QO-100), LEO_Pass (ground-to-LEO), or Cislunar_Link (ground/relay-to-cislunar)
- **HDTN_Contact_Plan**: The JSON format used by NASA HDTN for contact plan ingestion, containing source, dest, startTime, endTime, rateBitsPerSec, and owlt fields
- **Bundle_Volume**: The size of a bundle in bytes, used to determine transmission time over a Contact
- **Storage_Capacity**: The available buffer space at a relay node for storing bundles awaiting forwarding
- **Delivery_Time**: The total elapsed time from bundle creation at the source to delivery at the final destination, including transmission, propagation, and store-and-forward delays
- **ContactPlanManager**: The existing component that manages scheduled communication windows and provides contact lookup
- **Planning_Service**: The central server that generates the full contact graph and distributes per-node views to ground stations via REST API
- **Plan_Update_Bundle**: An administrative DTN bundle carrying a contact plan update payload, transmitted OTA to space nodes during contact windows
- **Bootstrap_Plan**: A minimal seed contact plan loaded onto a spacecraft before launch, providing enough routing information to receive initial plan update bundles
- **Plan_Version**: A monotonically increasing integer identifying a specific revision of a node's contact plan, used for conflict resolution (latest version wins)

## Requirements

### Requirement 1: Network Topology Definition

**User Story:** As a network operator, I want to define a multi-node network topology with heterogeneous node types, so that the system can model all communication paths across the RADIANT network.

#### Acceptance Criteria

1. THE Network_Topology SHALL accept node definitions containing a unique NodeID, node type (Ground_Station, LEO_Satellite, GEO_Relay, or Cislunar_Payload), and type-specific parameters
2. WHEN a Ground_Station node is defined, THE Network_Topology SHALL require latitude, longitude, altitude, minimum elevation angle, and supported Link_Class list
3. WHEN a LEO_Satellite node is defined, THE Network_Topology SHALL require OrbitalParameters or TLE data, onboard storage capacity in bytes, and supported data rates
4. WHEN a GEO_Relay node is defined, THE Network_Topology SHALL require longitude of the geostationary position, footprint boundaries, and supported data rates
5. WHEN a Cislunar_Payload node is defined, THE Network_Topology SHALL require OrbitalParameters, onboard storage capacity in bytes, and supported data rates
6. IF a node definition contains invalid parameters, THEN THE Network_Topology SHALL return a descriptive validation error identifying the invalid field
7. THE Network_Topology SHALL support incremental addition of nodes without requiring regeneration of the entire Contact_Graph

### Requirement 2: Pairwise Contact Generation

**User Story:** As a network operator, I want the system to automatically compute contact windows between all reachable node pairs, so that I do not need to manually specify every communication opportunity.

#### Acceptance Criteria

1. WHEN two Ground_Station nodes share a Terrestrial Link_Class, THE Pairwise_Contact_Generator SHALL produce a continuous Contact with the configured data rate and sub-millisecond one-way light time
2. WHEN two Ground_Station nodes are both within the GEO_Relay footprint, THE Pairwise_Contact_Generator SHALL produce a continuous Contact with 250ms one-way light time and the GEO relay's data rate
3. WHEN a Ground_Station and a LEO_Satellite are defined, THE Pairwise_Contact_Generator SHALL compute Contact windows using the existing PredictLEOPasses algorithm with the station's minimum elevation angle
4. WHEN a Ground_Station and a Cislunar_Payload are defined, THE Pairwise_Contact_Generator SHALL compute Contact windows using the existing PredictCislunarPasses algorithm
5. THE Pairwise_Contact_Generator SHALL produce directed Contacts (source → destination) for each valid communication direction between a node pair
6. WHEN new orbital parameters are provided for a LEO_Satellite, THE Pairwise_Contact_Generator SHALL recompute Contact windows for all Ground_Station pairs involving that satellite
7. THE Pairwise_Contact_Generator SHALL assign a confidence value to each predicted Contact based on the propagation time from the orbital epoch

### Requirement 3: Unified Contact Graph Assembly

**User Story:** As a routing algorithm, I want all contacts across all node pairs assembled into a single time-sorted graph structure, so that I can compute optimal multi-hop routes.

#### Acceptance Criteria

1. THE Contact_Graph SHALL contain all Contacts produced by the Pairwise_Contact_Generator, sorted by start time in ascending order
2. THE Contact_Graph SHALL associate each Contact with its source node, destination node, start time, end time, data rate, and one-way light time
3. WHEN a Contact is added to the Contact_Graph, THE Contact_Graph SHALL reject the Contact if it overlaps in time with another Contact on the same directed link (same source and destination)
4. THE Contact_Graph SHALL support querying all Contacts departing from a given node after a specified time
5. THE Contact_Graph SHALL support querying all Contacts arriving at a given node within a specified time range
6. WHEN the Contact_Graph is updated incrementally, THE Contact_Graph SHALL maintain sorted order without requiring a full rebuild

### Requirement 4: Time-Dependent Route Computation

**User Story:** As a bundle agent, I want to find the route that delivers a bundle to its destination in the shortest total time, so that delay-tolerant networking operates efficiently across multiple hops.

#### Acceptance Criteria

1. WHEN a route is requested from a source node to a destination node at a given time, THE Contact_Graph_Router SHALL compute the path that minimizes total Delivery_Time
2. THE Contact_Graph_Router SHALL account for transmission time on each Contact as Bundle_Volume divided by the Contact's data rate
3. THE Contact_Graph_Router SHALL account for one-way light time (propagation delay) on each Contact
4. THE Contact_Graph_Router SHALL account for Store_And_Forward_Delay at each intermediate node, computed as the time between bundle arrival and the next available departing Contact
5. WHEN no route exists from source to destination within the Contact_Graph's valid time range, THE Contact_Graph_Router SHALL return an explicit no-route-found indication
6. THE Contact_Graph_Router SHALL not select a Contact whose remaining capacity (data rate × remaining time) is insufficient to transmit the Bundle_Volume
7. THE Contact_Graph_Router SHALL produce a Route containing the ordered sequence of Contacts and the estimated Delivery_Time

### Requirement 5: Storage-Constrained Routing

**User Story:** As a network operator, I want routing to respect storage limits at relay nodes, so that bundles are not routed through nodes that cannot buffer them.

#### Acceptance Criteria

1. WHEN computing a route through an intermediate node, THE Contact_Graph_Router SHALL verify that the node's available Storage_Capacity is sufficient to hold the Bundle_Volume
2. WHEN a relay node's Storage_Capacity is insufficient for the Bundle_Volume, THE Contact_Graph_Router SHALL exclude that node from the route and seek an alternative path
3. WHILE a bundle is stored at a relay node awaiting forwarding, THE Contact_Graph_Router SHALL decrement the node's available Storage_Capacity by the Bundle_Volume
4. WHEN a bundle departs a relay node, THE Contact_Graph_Router SHALL restore the node's available Storage_Capacity by the Bundle_Volume
5. IF all routes to a destination require relay nodes with insufficient Storage_Capacity, THEN THE Contact_Graph_Router SHALL return a no-route-found indication with a storage-exhausted reason

### Requirement 6: Route Selection with Priority

**User Story:** As a bundle agent, I want route selection to consider bundle priority, so that high-priority bundles can preempt capacity reservations on constrained links.

#### Acceptance Criteria

1. WHEN multiple bundles compete for the same Contact capacity, THE Contact_Graph_Router SHALL allocate capacity to higher-priority bundles first
2. THE Contact_Graph_Router SHALL support at least three priority levels: bulk, normal, and expedited
3. WHEN an expedited bundle requires a Contact that has capacity reserved by bulk bundles, THE Contact_Graph_Router SHALL preempt the bulk reservation and reroute the bulk bundle
4. THE Contact_Graph_Router SHALL not preempt bundles of equal or higher priority

### Requirement 7: HDTN Contact Plan Export

**User Story:** As a node operator, I want to export the local node's view of the contact graph in HDTN-compatible JSON format, so that the RADIANT node integrates with NASA HDTN for bundle forwarding.

#### Acceptance Criteria

1. WHEN an HDTN export is requested for a specific node, THE Contact_Graph SHALL produce a JSON document containing all Contacts where that node is the source, formatted with source, dest, startTime, endTime, rateBitsPerSec, and owlt fields
2. THE Contact_Graph SHALL assign integer node identifiers consistent with the existing HDTN configuration's myNodeId convention
3. WHEN the Contact_Graph is updated, THE Contact_Graph SHALL support re-export of the HDTN contact plan without service interruption
4. THE Contact_Graph SHALL include range entries with one-way light time for each Contact in the exported plan

### Requirement 8: Incremental Topology Updates

**User Story:** As a network operator, I want to add new ground stations or update satellite TLEs without rebuilding the entire contact graph, so that the system adapts to a growing network (SatNOGS-style).

#### Acceptance Criteria

1. WHEN a new Ground_Station is added to the Network_Topology, THE Pairwise_Contact_Generator SHALL compute Contacts only for pairs involving the new station and merge them into the existing Contact_Graph
2. WHEN updated TLE data is received for a LEO_Satellite, THE Pairwise_Contact_Generator SHALL invalidate existing predicted Contacts for that satellite and recompute them using the new orbital parameters
3. WHEN a node is removed from the Network_Topology, THE Contact_Graph SHALL remove all Contacts involving that node and invalidate any Routes that traversed it
4. THE Contact_Graph SHALL complete an incremental update for a single new Ground_Station within 5 seconds for a network of up to 50 nodes

### Requirement 9: Multi-Hop Scenario Support

**User Story:** As a mission planner, I want the system to correctly route bundles through the four key relay scenarios (ground relay, LEO relay, multi-station coverage, GEO relay), so that all RADIANT mission phases are supported.

#### Acceptance Criteria

1. WHEN a Ground_Station has an earlier LEO pass than the originating station, THE Contact_Graph_Router SHALL discover the ground-relay route (Origin → Terrestrial → Ground_B → LEO_Pass → Satellite)
2. WHEN a LEO_Satellite has sequential passes over two different Ground_Stations, THE Contact_Graph_Router SHALL discover the satellite-relay route (Ground_A → LEO_Upload → Satellite_Store → LEO_Download → Ground_C)
3. WHEN multiple Ground_Stations have upcoming passes over the same LEO_Satellite, THE Contact_Graph_Router SHALL select the station with the earliest pass for download
4. WHEN a GEO_Relay connects two Ground_Stations and one has a LEO pass, THE Contact_Graph_Router SHALL discover the GEO-relay route (Ground_A → GEO → Ground_B → LEO_Pass → Satellite)
5. FOR ALL computed Routes, THE Contact_Graph_Router SHALL verify that each Contact in the sequence starts after the bundle's arrival time at that node (temporal feasibility)

### Requirement 10: Integration with ContactPlanManager

**User Story:** As a developer, I want the multi-node contact graph to integrate with the existing ContactPlanManager interface, so that existing single-hop code continues to function alongside multi-hop routing.

#### Acceptance Criteria

1. THE Contact_Graph_Router SHALL expose a FindRoute method compatible with the existing FindDirectContact signature, returning the first hop's ContactWindow for backward compatibility
2. THE Contact_Graph_Router SHALL extend the ContactPlanManager to support a FindMultiHopRoute method that returns the complete Route
3. WHEN FindDirectContact is called and a direct Contact exists, THE ContactPlanManager SHALL return the direct Contact without invoking multi-hop routing
4. WHEN FindDirectContact is called and no direct Contact exists, THE ContactPlanManager SHALL fall back to the Contact_Graph_Router to find a multi-hop Route and return the first hop
5. THE Contact_Graph_Router SHALL operate concurrently with the existing ContactPlanManager without requiring exclusive locks during route computation

### Requirement 11: REST API Distribution for Ground Stations

**User Story:** As a ground station operator, I want my node to automatically receive its local contact plan view from a central planning service via REST API, so that my station always has an up-to-date routing table without manual configuration.

#### Acceptance Criteria

1. THE Planning_Service SHALL expose a REST endpoint `GET /api/contact-plan/{nodeID}` that returns the HDTN-format contact plan for the specified node
2. WHEN the Contact_Graph is updated (TLE refresh, station added/removed), THE Planning_Service SHALL increment the Plan_Version and make the updated plan available within 10 seconds
3. THE Planning_Service SHALL include a Plan_Version field and a valid-from/valid-to time range in the API response, enabling clients to detect stale plans
4. WHEN a Ground_Station polls for its plan and the Plan_Version has not changed since the last retrieval, THE Planning_Service SHALL return HTTP 304 Not Modified to conserve bandwidth
5. THE Planning_Service SHALL support a webhook notification endpoint where ground stations can register to receive push notifications when their plan changes
6. THE Planning_Service SHALL reject requests for unknown NodeIDs with HTTP 404 and a descriptive error message

### Requirement 12: OTA Contact Plan Distribution for Space Nodes

**User Story:** As a mission operator, I want to upload contact plan updates to spacecraft as DTN bundles during contact windows, so that space nodes maintain current routing information despite intermittent connectivity.

#### Acceptance Criteria

1. THE system SHALL define a Plan_Update_Bundle type using a reserved BPv7 service demux path (e.g., dtn://nodeID/admin/contactplan) that carries a serialized HDTN contact plan as its payload
2. WHEN a Plan_Update_Bundle is received by a space node's BPA, THE node SHALL validate the plan payload, verify the Plan_Version is greater than the currently loaded version, and reload the local HDTN contact plan
3. IF a Plan_Update_Bundle contains a Plan_Version less than or equal to the currently loaded version, THEN THE node SHALL discard the bundle and log a stale-plan-rejected event
4. THE Plan_Update_Bundle SHALL be assigned expedited priority to ensure plan updates are transmitted before routine data bundles during limited contact windows
5. THE Planning_Service SHALL automatically generate and queue Plan_Update_Bundles for space nodes whenever the Contact_Graph is updated, targeting the ground station with the earliest upcoming pass to that spacecraft
6. THE Plan_Update_Bundle payload size SHALL not exceed 5,000 bytes to ensure transmission within a single LEO pass at 9600 bps (approximately 4 seconds of link time)

### Requirement 13: Bootstrap Contact Plan for Space Nodes

**User Story:** As a spacecraft integrator, I want to pre-load a minimal seed contact plan before launch, so that the spacecraft can receive its first plan update bundle without requiring a pre-existing contact plan.

#### Acceptance Criteria

1. THE system SHALL generate a Bootstrap_Plan from the spacecraft's initial TLE and the registered ground station list, covering at least 7 days of predicted passes
2. THE Bootstrap_Plan SHALL include contacts with all registered ground stations that have line-of-sight to the spacecraft's orbital plane
3. THE Bootstrap_Plan SHALL use conservative parameters: minimum elevation angle of 15 degrees (higher than operational 10 degrees) and confidence value of 0.5 to account for TLE drift before first update
4. WHEN a space node boots with no stored contact plan, THE node SHALL load the Bootstrap_Plan from non-volatile memory and begin accepting bundles on the first predicted contact window
5. WHEN a space node receives its first Plan_Update_Bundle, THE node SHALL replace the Bootstrap_Plan with the received plan and log a bootstrap-to-operational transition event

### Requirement 14: Plan Versioning and Conflict Resolution

**User Story:** As a network operator managing multiple ground stations that may independently attempt to upload plans to a spacecraft, I want deterministic conflict resolution so that the spacecraft always converges to the latest plan.

#### Acceptance Criteria

1. THE Planning_Service SHALL assign a monotonically increasing Plan_Version (positive integer) to each generated contact plan revision
2. WHEN a node receives a plan with Plan_Version V_new and currently holds Plan_Version V_current, THE node SHALL accept the plan if and only if V_new > V_current
3. WHEN multiple ground stations upload Plan_Update_Bundles to the same spacecraft during overlapping passes, THE spacecraft SHALL accept only the bundle with the highest Plan_Version and discard others
4. THE Planning_Service SHALL include a generation timestamp (Unix epoch seconds) in each plan, enabling operators to determine when the plan was computed
5. IF a node receives a Plan_Update_Bundle with a Plan_Version higher than any version the Planning_Service has generated, THEN THE node SHALL reject the bundle and log a version-out-of-range error (protection against corrupted bundles)

### Requirement 15: Contact Graph Management UI — Topology View

**User Story:** As a network operator, I want a web-based interface to view and manage the network topology on a world map, so that I can visually add ground stations, view satellite orbital tracks, and understand the network layout.

#### Acceptance Criteria

1. THE UI SHALL display all Ground_Station nodes as markers on a Leaflet.js world map at their configured latitude and longitude
2. THE UI SHALL display LEO_Satellite orbital ground tracks as polylines on the map, updated from current TLE data
3. THE UI SHALL provide a form to add a new Ground_Station node with fields for NodeID, callsign, latitude, longitude, altitude, minimum elevation angle, and supported link classes
4. THE UI SHALL provide a form to add a new LEO_Satellite node with fields for NodeID, TLE line 1, TLE line 2, storage capacity, and data rate
5. THE UI SHALL provide a form to add a new GEO_Relay node with fields for NodeID, longitude, footprint boundaries, and data rate
6. WHEN a node is added via the UI, THE system SHALL validate the parameters and display field-level error messages for invalid inputs
7. THE UI SHALL provide a delete action for each node with a confirmation dialog before removal
8. THE UI SHALL be implemented using Deno, Oak, Handlebars templates, Bootstrap CSS, and SQLite for persistence, consistent with the existing website architecture

### Requirement 16: Contact Graph Management UI — Contact Timeline

**User Story:** As a network operator, I want to view all contact windows across all node pairs on a timeline, so that I can understand when communication opportunities exist and identify coverage gaps.

#### Acceptance Criteria

1. THE UI SHALL display a Gantt-chart-style timeline showing all Contact windows for a configurable time range (default: next 24 hours)
2. THE UI SHALL color-code contacts by Link_Class: green for Terrestrial, blue for GEO_Relay, orange for LEO_Pass, purple for Cislunar_Link
3. WHEN a user clicks a Contact on the timeline, THE UI SHALL display a detail panel showing source node, destination node, start time, end time, data rate, OWLT, confidence, and residual capacity
4. THE UI SHALL provide time range controls allowing the operator to view the next 1 hour, 6 hours, 24 hours, 7 days, or a custom range
5. THE UI SHALL group contacts by source node, with each node occupying a row on the timeline
6. THE UI SHALL auto-refresh the timeline every 60 seconds to reflect contact graph updates

### Requirement 17: Contact Graph Management UI — Route Planner

**User Story:** As a network operator, I want to compute and visualize multi-hop routes between any two nodes, so that I can verify routing decisions and troubleshoot delivery failures.

#### Acceptance Criteria

1. THE UI SHALL provide a route planning form with source node selector, destination node selector, bundle size input (bytes), priority selector (bulk/normal/expedited), and departure time picker
2. WHEN the operator submits a route request, THE UI SHALL invoke the Contact_Graph_Router and display the computed Route as an ordered list of hops with timing breakdown (departure time, transmission time, OWLT, store-and-forward delay, arrival time)
3. THE UI SHALL display the computed Route as a path overlay on the topology map, highlighting each hop's source and destination nodes
4. WHEN no route exists, THE UI SHALL display the NoRouteError reason (no-path, storage-exhausted, or capacity-exhausted) with a human-readable explanation
5. THE UI SHALL display the total estimated delivery time and minimum confidence for the computed route

### Requirement 18: Contact Graph Management UI — Distribution Status

**User Story:** As a network operator, I want to monitor the plan distribution status of all nodes, so that I can verify all nodes have current routing information and troubleshoot distribution failures.

#### Acceptance Criteria

1. THE UI SHALL display a table of all nodes showing NodeID, node type, current Plan_Version, last update timestamp, and distribution status (current, stale, unreachable, bootstrap)
2. FOR Ground_Station nodes, THE UI SHALL show the last API poll time and whether the node has acknowledged the current plan version
3. FOR space nodes (LEO_Satellite, Cislunar_Payload), THE UI SHALL show the last OTA upload attempt time, target ground station, and delivery confirmation status
4. THE UI SHALL provide a "Regenerate Plan" button that triggers a full Contact_Graph recomputation and increments the Plan_Version
5. THE UI SHALL provide a "Push Update" action per space node that generates and queues a Plan_Update_Bundle for the next available contact window
6. THE UI SHALL provide a "Generate Bootstrap" action for new spacecraft that produces a downloadable Bootstrap_Plan JSON file
