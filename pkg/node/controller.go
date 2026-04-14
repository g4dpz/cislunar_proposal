package node

import (
	"fmt"
	"sync"
	"time"

	"terrestrial-dtn/pkg/bpa"
	"terrestrial-dtn/pkg/cla"
	"terrestrial-dtn/pkg/contact"
	"terrestrial-dtn/pkg/store"
)

// NodeController orchestrates the DTN node operation
type NodeController struct {
	config     NodeConfig
	bpa        *bpa.BundleProtocolAgent
	store      *store.BundleStore
	planner    *contact.ContactPlanManager
	cla        cla.ConvergenceLayerAdapter
	stats      NodeStatistics
	startTime  time.Time
	mu         sync.RWMutex
}

// NewNodeController creates a new node controller
func NewNodeController(
	config NodeConfig,
	bpaAgent *bpa.BundleProtocolAgent,
	bundleStore *store.BundleStore,
	planManager *contact.ContactPlanManager,
	claAdapter cla.ConvergenceLayerAdapter,
) *NodeController {
	return &NodeController{
		config:    config,
		bpa:       bpaAgent,
		store:     bundleStore,
		planner:   planManager,
		cla:       claAdapter,
		startTime: time.Now(),
	}
}

// Initialize initializes the node controller
func (nc *NodeController) Initialize() error {
	// Validate configuration
	if nc.config.NodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if len(nc.config.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint must be configured")
	}
	if nc.config.MaxStorageBytes <= 0 {
		return fmt.Errorf("max storage bytes must be positive")
	}

	return nil
}

// RunCycle executes one operation cycle
func (nc *NodeController) RunCycle(currentTime int64) error {
	// Step 1: Check for active contact windows
	activeContacts := nc.planner.GetActiveContacts(currentTime)

	// Step 2: For each active contact, deliver queued bundles
	for _, contactWindow := range activeContacts {
		if err := nc.executeContactWindow(contactWindow, currentTime); err != nil {
			// Log error but continue with other contacts
			continue
		}
	}

	// Step 3: Receive any incoming bundles
	if err := nc.receiveIncomingBundles(currentTime); err != nil {
		// Log error but continue
	}

	// Step 4: Expire old bundles
	evicted := nc.store.EvictExpired(currentTime)
	if evicted > 0 {
		nc.mu.Lock()
		nc.stats.TotalBundlesReceived -= int64(evicted) // Adjust stats
		nc.mu.Unlock()
	}

	// Step 5: Persist store state
	if err := nc.store.Flush(); err != nil {
		return fmt.Errorf("failed to flush store: %w", err)
	}

	return nil
}

// executeContactWindow handles bundle delivery during a contact window
// Requirement 9.2: Cease transmission when contact window ends
// Requirement 9.4: Retain bundles when contact missed
func (nc *NodeController) executeContactWindow(contactWindow contact.ContactWindow, currentTime int64) error {
	// Open the link
	if err := nc.cla.Open(contactWindow); err != nil {
		// Requirement 9.4: Mark contact as missed, retain bundles
		nc.mu.Lock()
		nc.stats.ContactsMissed++
		nc.mu.Unlock()
		return fmt.Errorf("failed to open link: %w", err)
	}
	defer nc.cla.Close()

	// Get bundles destined for this contact's remote node
	// Convert NodeID to EndpointID
	destEndpoint := bpa.EndpointID{
		Scheme: "dtn",
		SSP:    string(contactWindow.RemoteNode),
	}
	bundles := nc.store.ListByDestination(destEndpoint)

	// Transmit bundles in priority order
	bundlesSent := 0
	for _, bundle := range bundles {
		// Requirement 9.2: Check if contact window is still active
		if currentTime >= contactWindow.EndTime {
			break
		}

		// Send the bundle
		metrics, err := nc.cla.SendBundle(bundle)
		if err != nil {
			// Link degraded, stop sending
			// Requirement 5.5: Retain bundle for retry
			break
		}

		// Bundle sent successfully, delete from store
		if err := nc.store.Delete(bundle.ID); err != nil {
			// Log error but continue
			continue
		}

		bundlesSent++

		// Update statistics
		nc.mu.Lock()
		nc.stats.TotalBundlesSent++
		nc.stats.TotalBytesSent += int64(bundle.Size())
		nc.mu.Unlock()

		// Update metrics
		_ = metrics // Use metrics for telemetry
	}

	// Update contact statistics
	nc.mu.Lock()
	nc.stats.ContactsCompleted++
	nc.mu.Unlock()

	return nil
}

// receiveIncomingBundles processes incoming bundles
func (nc *NodeController) receiveIncomingBundles(currentTime int64) error {
	// Try to receive a bundle
	bundle, metrics, err := nc.cla.RecvBundle()
	if err != nil {
		// No bundle available or error
		return err
	}

	// Update statistics
	nc.mu.Lock()
	nc.stats.TotalBundlesReceived++
	nc.stats.TotalBytesReceived += int64(bundle.Size())
	nc.mu.Unlock()

	// Process the bundle based on type
	switch bundle.BundleType {
	case bpa.BundleTypePingRequest:
		// Handle ping request
		response, err := nc.bpa.HandlePing(bundle)
		if err != nil {
			return fmt.Errorf("failed to handle ping: %w", err)
		}

		// Store the response for delivery
		if err := nc.store.Store(response); err != nil {
			return fmt.Errorf("failed to store ping response: %w", err)
		}

	case bpa.BundleTypeData:
		// Process incoming data bundle
		if err := nc.processIncomingBundle(bundle, currentTime); err != nil {
			return fmt.Errorf("failed to process incoming bundle: %w", err)
		}

	case bpa.BundleTypePingResponse:
		// Ping response received - calculate RTT
		// In a real implementation, would match with original ping request
		_ = metrics // Use metrics for RTT calculation
	}

	return nil
}

// processIncomingBundle validates, stores, and handles an incoming bundle
// Requirement 17.1: Handle store full with eviction
// Requirement 17.5: Retain bundles when no contact available
func (nc *NodeController) processIncomingBundle(bundle *bpa.Bundle, currentTime int64) error {
	// Validate the bundle
	if err := nc.bpa.ValidateBundle(bundle, currentTime); err != nil {
		nc.mu.Lock()
		nc.stats.TotalBundlesReceived-- // Don't count invalid bundles
		nc.mu.Unlock()
		return fmt.Errorf("bundle validation failed: %w", err)
	}

	// Check if destination is local
	if nc.bpa.IsLocalEndpoint(bundle.Destination) {
		// Deliver to local application
		if err := nc.bpa.DeliverBundle(bundle); err != nil {
			return fmt.Errorf("failed to deliver bundle locally: %w", err)
		}
		nc.mu.Lock()
		nc.stats.BundlesForwarded++
		nc.mu.Unlock()
		return nil
	}

	// Destination is remote - store for forwarding
	// Requirement 17.5: Check if direct contact exists
	nextContact := nc.planner.GetNextContactByEndpoint(bundle.Destination, currentTime)
	if nextContact == nil {
		// No contact available - still store the bundle
		// It will be re-evaluated when contact plan is updated
	}

	// Try to store the bundle
	if err := nc.store.Store(bundle); err != nil {
		// Requirement 17.1: Store full - try to evict
		if store.IsStoreFull(err) {
			_, evictErr := nc.store.EvictToFreeSpace(int64(bundle.Size()), currentTime)
			if evictErr != nil {
				nc.mu.Lock()
				nc.stats.TotalBundlesReceived-- // Don't count dropped bundles
				nc.mu.Unlock()
				return fmt.Errorf("store full and eviction failed: %w", evictErr)
			}

			// Try storing again after eviction
			if err := nc.store.Store(bundle); err != nil {
				nc.mu.Lock()
				nc.stats.TotalBundlesReceived-- // Don't count dropped bundles
				nc.mu.Unlock()
				return fmt.Errorf("failed to store bundle after eviction: %w", err)
			}
		} else {
			nc.mu.Lock()
			nc.stats.TotalBundlesReceived-- // Don't count dropped bundles
			nc.mu.Unlock()
			return fmt.Errorf("failed to store bundle: %w", err)
		}
	}

	return nil
}

// Shutdown gracefully shuts down the node
func (nc *NodeController) Shutdown() error {
	// Close CLA
	if err := nc.cla.Close(); err != nil {
		return fmt.Errorf("failed to close CLA: %w", err)
	}

	// Flush store
	if err := nc.store.Flush(); err != nil {
		return fmt.Errorf("failed to flush store: %w", err)
	}

	return nil
}

// HealthCheck returns the current health status
func (nc *NodeController) HealthCheck() NodeHealth {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	capacity := nc.store.Capacity()
	storageUsedPercent := float64(capacity.UsedBytes) / float64(capacity.TotalBytes) * 100.0

	uptime := time.Since(nc.startTime).Seconds()

	health := NodeHealth{
		UptimeSeconds:      int64(uptime),
		StorageUsedPercent: storageUsedPercent,
		BundlesStored:      capacity.BundleCount,
		BundlesForwarded:   int(nc.stats.BundlesForwarded),
		BundlesDropped:     int(nc.stats.TotalBundlesReceived - nc.stats.BundlesForwarded - int64(capacity.BundleCount)),
	}

	return health
}

// GetStatistics returns cumulative statistics
func (nc *NodeController) GetStatistics() NodeStatistics {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	// Return a copy
	return nc.stats
}

// GetBundleStore returns the bundle store for testing purposes
func (nc *NodeController) GetBundleStore() *store.BundleStore {
	return nc.store
}
