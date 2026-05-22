# Implementation Plan: Multi-Node Contact Graph

## Overview

This plan implements the multi-node contact graph system bottom-up: starting with topology types and validation, then the contact graph data structure, pairwise contact generation, time-dependent routing, HDTN export, and finally integration with the existing ContactPlanManager. Each layer builds on the previous, with property tests validating correctness at each stage.

## Tasks

- [ ] 1. Implement topology types and validation
  - [ ] 1.1 Create topology types and NodeParams interface in `pkg/contact/topology.go`
    - Define `NodeType` enum (GroundStation, LEOSatellite, GEORelay, CislunarPayload)
    - Define `LinkClass` enum (Terrestrial, GEORelay, LEOPass, CislunarLink)
    - Define `NodeDefinition` struct with ID, Type, HDTNNodeID, Params fields
    - Define `NodeParams` interface with `Validate() error` and `GetNodeType() NodeType`
    - Define `GroundStationParams`, `LEOSatelliteParams`, `GEORelayParams`, `CislunarPayloadParams` structs
    - Implement `Validate()` for each params type with field-level error messages
    - Define `TLEParameters` struct for LEO satellites (alternative to OrbitalParameters)
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

  - [ ] 1.2 Implement TopologyDefinition with node management in `pkg/contact/topology.go`
    - Implement `TopologyDefinition` struct with `sync.RWMutex` and `map[NodeID]*NodeDefinition`
    - Implement `NewTopologyDefinition()` constructor
    - Implement `AddNode()` with validation and duplicate-ID rejection
    - Implement `RemoveNode()` that removes by ID
    - Implement `GetNode()` for retrieval by ID
    - Implement `ListNodes()` with optional NodeType filter
    - Implement `GetNodePairs()` returning all valid directed communication pairs with LinkClass
    - Define `NodePair` struct (Source, Dest, LinkClass)
    - _Requirements: 1.1, 1.6, 1.7_

  - [ ]* 1.3 Write property tests for topology in `pkg/contact/topology_property_test.go`
    - **Property 1: Node Definition Round-Trip**
    - **Property 2: Node Validation Rejects Invalid Definitions**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**

- [ ] 2. Implement contact graph data structure
  - [ ] 2.1 Create DirectedContact type and ContactGraph structure in `pkg/contact/graph.go`
    - Define `DirectedContact` struct with ID, Source, Dest, StartTime, EndTime, DataRate, OWLT, Confidence, LinkClass, ResidualVolume
    - Implement `TransmissionTime(bundleBytes int64) float64` method
    - Implement `RemainingTime(fromTime int64) int64` method
    - Implement `CanTransmit(bundleBytes int64, departureTime int64) bool` method
    - Define `ContactGraph` struct with `sync.RWMutex`, outgoing/incoming adjacency maps, allContacts slice, storageUsed/storageMax maps
    - Implement `NewContactGraph()` constructor
    - _Requirements: 3.1, 3.2_

  - [ ] 2.2 Implement ContactGraph mutation and query methods in `pkg/contact/graph.go`
    - Implement `AddContacts()` with sorted insertion and overlap detection (same source+dest, overlapping time)
    - Implement `RemoveContactsForNode()` removing all contacts where Source or Dest matches
    - Implement `GetOutgoingAfter()` using binary search for O(log n) lookup
    - Implement `GetIncomingInRange()` returning contacts where Dest matches within time range
    - Implement `SetStorageCapacity()`, `GetAvailableStorage()`, `ReserveStorage()`, `ReleaseStorage()`
    - _Requirements: 3.3, 3.4, 3.5, 3.6, 5.3, 5.4_

  - [ ]* 2.3 Write property tests for contact graph in `pkg/contact/graph_property_test.go`
    - **Property 8: Contact Graph Maintains Sorted Order**
    - **Property 9: Contact Field Preservation**
    - **Property 10: Overlap Rejection**
    - **Property 11: Outgoing Query Correctness**
    - **Property 12: Incoming Query Correctness**
    - **Property 18: Storage Accounting Round-Trip**
    - **Property 21: Node Removal Cleans All Contacts**
    - **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 5.3, 5.4, 8.3**

- [ ] 3. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 4. Implement pairwise contact generation
  - [ ] 4.1 Create PairwiseContactGenerator in `pkg/contact/generator.go`
    - Define `PairwiseContactGenerator` struct bound to a `*TopologyDefinition`
    - Implement `NewPairwiseContactGenerator(topology *TopologyDefinition)`
    - Implement `GenerateAllContacts(fromTime, toTime time.Time) ([]DirectedContact, error)` iterating all node pairs
    - Implement `GenerateForNode(nodeID NodeID, fromTime, toTime time.Time) ([]DirectedContact, error)` for incremental updates
    - Implement `GenerateForPair(source, dest NodeID, fromTime, toTime time.Time) ([]DirectedContact, error)`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

  - [ ] 4.2 Implement link-class-specific contact generation logic in `pkg/contact/generator.go`
    - Implement terrestrial contact generation: continuous contact spanning full time range, OWLT < 0.001s, confidence 1.0
    - Implement GEO relay contact generation: continuous if both in footprint, OWLT = 0.250s, GEO data rate
    - Implement LEO pass contact generation: delegate to existing `PredictLEOPasses` / `PredictPasses`, assign confidence from epoch propagation
    - Implement cislunar contact generation: delegate to existing `PredictCislunarPasses` / `PredictPasses`, compute cislunar OWLT
    - Generate bidirectional contacts (A→B and B→A) for each valid pair
    - Assign confidence values based on propagation time from orbital epoch using `computeConfidence`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.7_

  - [ ]* 4.3 Write property tests for contact generation in `pkg/contact/generator_property_test.go`
    - **Property 4: Terrestrial Contacts Are Continuous**
    - **Property 5: GEO Relay Contacts Have Correct Properties**
    - **Property 6: Bidirectional Contact Generation**
    - **Property 7: Confidence Monotonically Decreases With Propagation Time**
    - **Validates: Requirements 2.1, 2.2, 2.5, 2.7**

- [ ] 5. Implement time-dependent routing
  - [ ] 5.1 Create routing types and ContactGraphRouter in `pkg/contact/router.go`
    - Define `BundlePriority` enum (Bulk=0, Normal=1, Expedited=2)
    - Define `RouteRequest` struct (Source, Destination, RequestTime, BundleSize, Priority)
    - Define `Route` struct (Hops []RouteHop, DeliveryTime, TotalLatency, Confidence)
    - Define `RouteHop` struct (Contact, ArrivalTime, DepartureTime, StoreDelay)
    - Define `NoRouteError` struct with Source, Destination, Reason, Details fields
    - Implement `NoRouteError.Error()`, `IsNoRoute()`, `IsStorageExhausted()` helpers
    - Define `ContactGraphRouter` struct with graph and topology references
    - Implement `NewContactGraphRouter(graph *ContactGraph, topology *TopologyDefinition)`
    - _Requirements: 4.1, 4.5, 4.7_

  - [ ] 5.2 Implement FindRoute with time-dependent Dijkstra in `pkg/contact/router.go`
    - Implement min-heap priority queue keyed on arrival time
    - Implement `FindRoute(req *RouteRequest) (*Route, error)` using modified Dijkstra
    - Account for transmission time (bundleSize * 8 / dataRate)
    - Account for one-way light time (OWLT) on each contact
    - Account for store-and-forward delay (time between arrival and next contact departure)
    - Check residual volume capacity on each contact via `CanTransmit`
    - Check storage capacity at intermediate nodes via `GetAvailableStorage`
    - Return `NoRouteError` with appropriate reason when no path exists
    - Implement path reconstruction from predecessor map
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 5.1, 5.2, 5.5_

  - [ ] 5.3 Implement FindRouteWithExclusions and priority preemption in `pkg/contact/router.go`
    - Implement `FindRouteWithExclusions(req *RouteRequest, excludeContacts map[uint64]bool) (*Route, error)`
    - Implement `PreemptAndReroute(highPriority *RouteRequest, existingRoutes map[NodeID]*Route) (*Route, []NodeID, error)`
    - Priority preemption: expedited > normal > bulk; equal or higher priority cannot be preempted
    - Return list of preempted bundle sources for rerouting
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [ ]* 5.4 Write property tests for routing in `pkg/contact/router_property_test.go`
    - **Property 13: Route Optimality (Earliest Arrival)** — brute-force verification on small graphs (≤6 nodes, ≤20 contacts)
    - **Property 14: Delivery Time Correctness**
    - **Property 15: No-Route for Disconnected Graphs**
    - **Property 16: Capacity Constraint Respected**
    - **Property 17: Storage Constraint Respected**
    - **Property 19: Priority Preemption Correctness**
    - **Property 22: Temporal Feasibility of All Routes**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 5.1, 5.2, 5.5, 6.1, 6.3, 6.4, 9.5**

- [ ] 6. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Implement HDTN export
  - [ ] 7.1 Create HDTNExporter in `pkg/contact/hdtn_export.go`
    - Define `HDTNContactEntry` struct with JSON tags (contact, source, dest, startTime, endTime, rateBitsPerSec, owlt)
    - Define `HDTNContactPlan` struct with Contacts slice
    - Define `HDTNExporter` struct with graph and topology references
    - Implement `NewHDTNExporter(graph *ContactGraph, topology *TopologyDefinition)`
    - Implement `ExportForNode(nodeID NodeID) (*HDTNContactPlan, error)` extracting outgoing contacts and mapping to HDTN integer node IDs
    - Implement `ExportForNodeJSON(nodeID NodeID) ([]byte, error)` for JSON serialization
    - Include range entries with one-way light time for each contact
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

  - [ ]* 7.2 Write property tests for HDTN export in `pkg/contact/hdtn_export_property_test.go`
    - **Property 20: HDTN Export Field Correctness**
    - **Validates: Requirements 7.1, 7.2, 7.4**

- [ ] 8. Implement integration with ContactPlanManager
  - [ ] 8.1 Create MultiHopContactPlanManager in `pkg/contact/multihop_manager.go`
    - Define `MultiHopContactPlanManager` struct embedding `*ContactPlanManager` with router, graph, topology fields
    - Implement `NewMultiHopContactPlanManager(topology, graph, router)`
    - Implement `FindMultiHopRoute(req *RouteRequest) (*Route, error)` delegating to router
    - Override `FindDirectContact(destination bpa.EndpointID, currentTime int64) (*ContactWindow, error)`:
      - If direct contact exists, return it (no multi-hop invocation)
      - If no direct contact, fall back to multi-hop routing and return first hop's ContactWindow
    - Ensure concurrent access safety (read lock for route computation, no exclusive locks)
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [ ] 8.2 Implement incremental topology update workflow in `pkg/contact/multihop_manager.go`
    - Implement `AddStation(node *NodeDefinition) error` — adds node, generates contacts for new pairs, merges into graph
    - Implement `UpdateSatelliteTLE(nodeID NodeID, params *OrbitalParameters) error` — invalidates old contacts, recomputes
    - Implement `RemoveNode(nodeID NodeID) error` — removes node and all associated contacts
    - Ensure incremental update completes within 5 seconds for 50-node networks
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ]* 8.3 Write property tests for integration in `pkg/contact/integration_property_test.go`
    - **Property 3: Incremental Node Addition Preserves Existing Contacts**
    - **Property 23: FindDirectContact Fallback Behavior**
    - **Validates: Requirements 1.7, 8.1, 10.3, 10.4**

- [ ] 9. Implement multi-hop relay scenario unit tests
  - [ ] 9.1 Write unit tests for relay scenarios in `pkg/contact/router_test.go`
    - Test ground-relay route: Origin → Terrestrial → Ground_B → LEO_Pass → Satellite
    - Test satellite-relay route: Ground_A → LEO_Upload → Satellite_Store → LEO_Download → Ground_C
    - Test multi-station coverage: select station with earliest pass for download
    - Test GEO-relay route: Ground_A → GEO → Ground_B → LEO_Pass → Satellite
    - Verify temporal feasibility for all computed routes (each hop starts after previous arrival)
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 10. Implement contact plan distribution
  - [ ] 10.1 Create PlanningService and PlanVersionManager in `pkg/contact/planning_service.go`
    - Define `PlanVersionManager` struct with monotonic version counter and max-known-version tracking
    - Implement `Increment()`, `IsAcceptable(version)`, `GetCurrent()`
    - Define `VersionedPlan` struct (Version, GeneratedAt, ValidFrom, ValidTo, Plan)
    - Define `PlanningService` struct with topology, generator, graph, exporter, version manager
    - Implement `NewPlanningService(topology)` constructor
    - Implement `Regenerate(fromTime, toTime)` — recomputes graph and increments version
    - Implement `GetPlanForNode(nodeID)` — returns VersionedPlan for a specific node
    - _Requirements: 11.1, 11.2, 11.3, 14.1, 14.4_

  - [ ] 10.2 Create REST API for ground station distribution in `pkg/contact/plan_api.go`
    - Define `PlanDistributionAPI` struct with PlanningService and WebhookRegistry
    - Implement `GET /api/contact-plan/{nodeID}` handler returning VersionedPlan JSON
    - Implement HTTP 304 Not Modified when client's If-None-Match matches current version
    - Implement HTTP 404 for unknown NodeIDs
    - Implement `POST /api/webhooks/register` for push notification registration
    - Implement `DELETE /api/webhooks/{webhookID}` for unregistration
    - Define `WebhookRegistry` with HMAC-signed payload delivery and retry logic
    - _Requirements: 11.1, 11.3, 11.4, 11.5, 11.6_

  - [ ] 10.3 Create OTA Plan_Update_Bundle generator and receiver in `pkg/contact/plan_ota.go`
    - Define `AdminServicePath` constant (`admin/contactplan`)
    - Define `PlanUpdatePayload` struct (Version, GeneratedAt, ValidFrom, ValidTo, Contacts)
    - Define `MaxPlanUpdatePayloadBytes` constant (5000)
    - Implement `PlanUpdateBundleGenerator.GeneratePlanUpdateBundle(targetNodeID, plan)` — creates expedited-priority bundle with serialized plan payload ≤ 5KB
    - Implement `PlanUpdateReceiver.HandlePlanUpdate(bundle)` — validates version > current, reloads HDTN plan
    - Reject stale versions (V_new ≤ V_current) with stale-plan-rejected log
    - Reject out-of-range versions (V_new > maxKnownVersion) with version-out-of-range log
    - Auto-queue plan updates targeting ground station with earliest upcoming pass
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 14.2, 14.3, 14.5_

  - [ ] 10.4 Create BootstrapPlanGenerator in `pkg/contact/plan_bootstrap.go`
    - Implement `BootstrapPlanGenerator.GenerateBootstrapPlan(spacecraftID, initialTLE, stations)` 
    - Generate 7-day plan from initial TLE with conservative parameters (15° min elevation, 0.5 confidence)
    - Include contacts with all registered stations having line-of-sight to orbital plane
    - Handle boot-with-no-plan scenario: load bootstrap from NVM
    - Log bootstrap-to-operational transition on first Plan_Update_Bundle receipt
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

  - [ ]* 10.5 Write property tests for distribution in `pkg/contact/distribution_property_test.go`
    - **Property 24: Plan Version Monotonicity**
    - **Property 25: Version Acceptance Correctness**
    - **Property 26: Plan Update Payload Size Constraint**
    - **Property 27: Bootstrap Plan Coverage**
    - **Property 28: API Conditional Response**
    - **Validates: Requirements 11.4, 12.6, 13.1, 13.2, 14.1, 14.2, 14.3**

- [ ] 11. Implement contact graph management UI
  - [ ] 11.1 Create topology management page in `website/views/pages/admin/topology.hbs` and `website/routes/topology.ts`
    - Create Handlebars template with Leaflet.js map (ground station markers, orbital tracks)
    - Create Oak route serving the page with node data from SQLite
    - Add Bootstrap modal forms for adding Ground_Station, LEO_Satellite, GEO_Relay, and Cislunar_Payload nodes
    - Implement server-side validation with field-level error messages
    - Add delete action with confirmation dialog
    - Add SQLite schema migration for topology_nodes table
    - Include satellite.js for client-side SGP4 orbital track rendering
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7, 15.8_

  - [ ] 11.2 Create contact timeline page in `website/views/pages/admin/contacts.hbs` and `website/routes/contacts.ts`
    - Create Handlebars template with SVG-based Gantt chart container
    - Create Oak route serving contacts data as JSON for the time range
    - Implement client-side timeline rendering with color-coded link classes (green/blue/orange/purple)
    - Add click handler showing contact detail panel (source, dest, times, data rate, OWLT, confidence)
    - Add time range controls (1h, 6h, 24h, 7d, custom)
    - Implement 60-second auto-refresh via setInterval
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6_

  - [ ] 11.3 Create route planner page in `website/views/pages/admin/routes.hbs` and `website/routes/route-planner.ts`
    - Create Handlebars template with route request form (source, dest, size, priority, time)
    - Create Oak route that proxies route computation to the Go backend
    - Display computed route as hop table (From, To, Depart, Arrive, TX Time, OWLT, Store Delay)
    - Display route path overlay on Leaflet map
    - Display NoRouteError with human-readable explanation when no route exists
    - Show total delivery time and minimum confidence
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5_

  - [ ] 11.4 Create distribution status page in `website/views/pages/admin/distribution.hbs` and `website/routes/distribution.ts`
    - Create Handlebars template with Bootstrap table showing all nodes' plan status
    - Add status badges (Current=green, Stale=yellow, Unreachable=red, Bootstrap=grey)
    - Show ground station last poll time and space node last OTA attempt
    - Add "Regenerate Plan" button (POST /api/graph/regenerate)
    - Add per-node "Push Update" button for space nodes
    - Add "Generate Bootstrap" button producing downloadable JSON
    - Add SQLite schema migration for plan_versions and distribution_status tables
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6_

  - [ ] 11.5 Add admin navigation links and integrate with existing website auth
    - Add "Topology", "Contacts", "Routes", "Distribution" links to admin nav partial
    - Ensure all new pages require admin role authentication (existing auth middleware)
    - Add admin-statistics integration showing contact graph metrics
    - _Requirements: 15.8_

- [ ] 12. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document (31 properties total)
- Unit tests validate specific relay scenarios and edge cases
- The implementation uses `github.com/leanovate/gopter` for property-based testing (consistent with existing project)
- All new Go code goes in `pkg/contact/` alongside existing types.go, cgr.go, manager.go
- The `MultiHopContactPlanManager` embeds the existing `ContactPlanManager` for backward compatibility
- Distribution tasks (10.x) implement REST API for ground stations and OTA bundle mechanism for space nodes
- UI tasks (11.x) use Deno, Oak, Handlebars, Bootstrap, and SQLite consistent with the existing website

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "2.1"] },
    { "id": 2, "tasks": ["1.3", "2.2"] },
    { "id": 3, "tasks": ["2.3", "4.1"] },
    { "id": 4, "tasks": ["4.2"] },
    { "id": 5, "tasks": ["4.3", "5.1"] },
    { "id": 6, "tasks": ["5.2", "7.1"] },
    { "id": 7, "tasks": ["5.3", "7.2"] },
    { "id": 8, "tasks": ["5.4", "8.1"] },
    { "id": 9, "tasks": ["8.2", "9.1"] },
    { "id": 10, "tasks": ["8.3", "10.1"] },
    { "id": 11, "tasks": ["10.2", "10.3", "10.4"] },
    { "id": 12, "tasks": ["10.5", "11.1"] },
    { "id": 13, "tasks": ["11.2", "11.3", "11.4"] },
    { "id": 14, "tasks": ["11.5"] }
  ]
}
```
