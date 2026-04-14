package store

import (
	"fmt"
	"sort"
	"sync"

	"terrestrial-dtn/pkg/bpa"
)

// StoreCapacity represents the current capacity status of the bundle store
type StoreCapacity struct {
	TotalBytes  int64
	UsedBytes   int64
	BundleCount int
}

// BundleStore wraps ION-DTN bundle storage operations
type BundleStore struct {
	bundles      map[string]*bpa.Bundle // keyed by bundle ID hash
	maxBytes     int64
	usedBytes    int64
	mu           sync.RWMutex
}

// NewBundleStore creates a new bundle store
func NewBundleStore(maxBytes int64) *BundleStore {
	return &BundleStore{
		bundles:   make(map[string]*bpa.Bundle),
		maxBytes:  maxBytes,
		usedBytes: 0,
	}
}

// Store persists a bundle to the store
// Requirement 17.1: Handle store full with eviction
func (bs *BundleStore) Store(bundle *bpa.Bundle) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bundleSize := int64(bundle.Size())
	bundleHash := bundle.ID.Hash()

	// Check if bundle already exists
	if _, exists := bs.bundles[bundleHash]; exists {
		return fmt.Errorf("bundle %s already exists in store", bundle.ID.String())
	}

	// Check capacity
	if bs.usedBytes+bundleSize > bs.maxBytes {
		return &StoreFullError{
			Required:  bundleSize,
			Available: bs.maxBytes - bs.usedBytes,
			Total:     bs.maxBytes,
		}
	}

	// Store the bundle
	bs.bundles[bundleHash] = bundle
	bs.usedBytes += bundleSize

	return nil
}

// StoreFullError indicates the store is at capacity
type StoreFullError struct {
	Required  int64
	Available int64
	Total     int64
}

func (e *StoreFullError) Error() string {
	return fmt.Sprintf("store full: need %d bytes, have %d available (total %d)",
		e.Required, e.Available, e.Total)
}

// IsStoreFull checks if an error is a StoreFullError
func IsStoreFull(err error) bool {
	_, ok := err.(*StoreFullError)
	return ok
}

// Retrieve fetches a bundle by its ID
func (bs *BundleStore) Retrieve(bundleID bpa.BundleID) (*bpa.Bundle, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	bundleHash := bundleID.Hash()
	bundle, exists := bs.bundles[bundleHash]
	if !exists {
		return nil, fmt.Errorf("bundle %s not found", bundleID.String())
	}

	return bundle, nil
}

// Delete removes a bundle from the store
func (bs *BundleStore) Delete(bundleID bpa.BundleID) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bundleHash := bundleID.Hash()
	bundle, exists := bs.bundles[bundleHash]
	if !exists {
		return fmt.Errorf("bundle %s not found", bundleID.String())
	}

	bundleSize := int64(bundle.Size())
	delete(bs.bundles, bundleHash)
	bs.usedBytes -= bundleSize

	return nil
}

// ListByPriority returns all bundles sorted by priority (highest first)
func (bs *BundleStore) ListByPriority() []*bpa.Bundle {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	bundles := make([]*bpa.Bundle, 0, len(bs.bundles))
	for _, bundle := range bs.bundles {
		bundles = append(bundles, bundle)
	}

	// Sort by priority: critical > expedited > normal > bulk
	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].Priority > bundles[j].Priority
	})

	return bundles
}

// ListByDestination returns all bundles destined for a specific endpoint
func (bs *BundleStore) ListByDestination(destination bpa.EndpointID) []*bpa.Bundle {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	bundles := make([]*bpa.Bundle, 0)
	for _, bundle := range bs.bundles {
		if bundle.Destination.Scheme == destination.Scheme &&
			bundle.Destination.SSP == destination.SSP {
			bundles = append(bundles, bundle)
		}
	}

	// Sort by priority
	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].Priority > bundles[j].Priority
	})

	return bundles
}

// Capacity returns the current capacity status
func (bs *BundleStore) Capacity() StoreCapacity {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	return StoreCapacity{
		TotalBytes:  bs.maxBytes,
		UsedBytes:   bs.usedBytes,
		BundleCount: len(bs.bundles),
	}
}

// EvictLowestPriority evicts the lowest priority bundle
// Returns the evicted bundle or nil if store is empty
func (bs *BundleStore) EvictLowestPriority() (*bpa.Bundle, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if len(bs.bundles) == 0 {
		return nil, fmt.Errorf("store is empty, nothing to evict")
	}

	// Find lowest priority bundle
	var lowestPriority *bpa.Bundle
	var lowestHash string

	for hash, bundle := range bs.bundles {
		if lowestPriority == nil || bundle.Priority < lowestPriority.Priority {
			lowestPriority = bundle
			lowestHash = hash
		}
	}

	// Evict it
	bundleSize := int64(lowestPriority.Size())
	delete(bs.bundles, lowestHash)
	bs.usedBytes -= bundleSize

	return lowestPriority, nil
}

// EvictExpired removes all expired bundles at the given time
// Returns the number of bundles evicted
func (bs *BundleStore) EvictExpired(currentTime int64) int {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	evicted := 0
	for hash, bundle := range bs.bundles {
		if bundle.IsExpired(currentTime) {
			bundleSize := int64(bundle.Size())
			delete(bs.bundles, hash)
			bs.usedBytes -= bundleSize
			evicted++
		}
	}

	return evicted
}

// EvictToFreeSpace evicts bundles to free the required space
// Evicts expired bundles first, then lowest priority bundles
// Returns the number of bytes freed
func (bs *BundleStore) EvictToFreeSpace(requiredBytes int64, currentTime int64) (int64, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	freedBytes := int64(0)

	// First, evict expired bundles
	for hash, bundle := range bs.bundles {
		if bundle.IsExpired(currentTime) {
			bundleSize := int64(bundle.Size())
			delete(bs.bundles, hash)
			bs.usedBytes -= bundleSize
			freedBytes += bundleSize

			if freedBytes >= requiredBytes {
				return freedBytes, nil
			}
		}
	}

	// If still need more space, evict by priority (lowest first)
	// Build a sorted list of bundles by priority
	type bundleEntry struct {
		hash   string
		bundle *bpa.Bundle
	}

	entries := make([]bundleEntry, 0, len(bs.bundles))
	for hash, bundle := range bs.bundles {
		entries = append(entries, bundleEntry{hash, bundle})
	}

	// Sort by priority (lowest first), but preserve critical bundles until last
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].bundle.Priority < entries[j].bundle.Priority
	})

	// Evict bundles in priority order
	for _, entry := range entries {
		bundleSize := int64(entry.bundle.Size())
		delete(bs.bundles, entry.hash)
		bs.usedBytes -= bundleSize
		freedBytes += bundleSize

		if freedBytes >= requiredBytes {
			return freedBytes, nil
		}
	}

	// Could not free enough space
	return freedBytes, fmt.Errorf("could not free %d bytes, only freed %d", requiredBytes, freedBytes)
}

// Flush persists the store state to non-volatile memory
// In a real implementation, this would sync with ION-DTN's persistent storage
func (bs *BundleStore) Flush() error {
	// This would call into ION-DTN to persist the bundle store
	// For now, this is a no-op since we're using in-memory storage
	return nil
}

// Count returns the number of bundles in the store
func (bs *BundleStore) Count() int {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return len(bs.bundles)
}

// RecoverFromPowerLoss reloads the bundle store from NVM after power cycle
// Requirement 17.3, 17.4: Reload store from NVM and validate integrity
func (bs *BundleStore) RecoverFromPowerLoss(nvmBundles []*bpa.Bundle) (int, int, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	recovered := 0
	corrupted := 0

	for _, bundle := range nvmBundles {
		// Validate bundle integrity (CRC check would be done by ION-DTN)
		// For now, we just check basic validity
		if bundle == nil {
			corrupted++
			continue
		}

		bundleSize := int64(bundle.Size())
		bundleHash := bundle.ID.Hash()

		// Check if we have capacity
		if bs.usedBytes+bundleSize > bs.maxBytes {
			// Skip bundles that don't fit
			continue
		}

		// Restore the bundle
		bs.bundles[bundleHash] = bundle
		bs.usedBytes += bundleSize
		recovered++
	}

	return recovered, corrupted, nil
}

// ValidateIntegrity validates the integrity of all bundles in the store
// Requirement 17.4: Validate store integrity via CRC
func (bs *BundleStore) ValidateIntegrity() error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	// In a real implementation, this would validate CRC for each bundle
	// For now, we just check basic consistency
	totalSize := int64(0)
	for _, bundle := range bs.bundles {
		totalSize += int64(bundle.Size())
	}

	if totalSize != bs.usedBytes {
		return fmt.Errorf("store integrity check failed: computed size %d != stored size %d",
			totalSize, bs.usedBytes)
	}

	return nil
}

// HandleCorruption handles bundle corruption by discarding corrupted bundles
// Requirement 17.2: Handle CRC validation failures
func (bs *BundleStore) HandleCorruption(bundleID bpa.BundleID, linkMetrics interface{}) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bundleHash := bundleID.Hash()
	bundle, exists := bs.bundles[bundleHash]
	if !exists {
		return fmt.Errorf("bundle %s not found", bundleID.String())
	}

	// Log corruption event with link metrics
	// In a real implementation, would log to telemetry system
	bundleSize := int64(bundle.Size())
	delete(bs.bundles, bundleHash)
	bs.usedBytes -= bundleSize

	return nil
}
