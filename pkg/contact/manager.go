package contact

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"terrestrial-dtn/pkg/bpa"
)

// ContactPlanManager manages scheduled communication windows
type ContactPlanManager struct {
	plan           *ContactPlan
	orbitalParams  map[NodeID]*OrbitalParameters
	mu             sync.RWMutex
}

// NewContactPlanManager creates a new contact plan manager
func NewContactPlanManager() *ContactPlanManager {
	return &ContactPlanManager{
		orbitalParams: make(map[NodeID]*OrbitalParameters),
	}
}

// LoadPlan loads a contact plan
func (cpm *ContactPlanManager) LoadPlan(plan *ContactPlan) error {
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("invalid contact plan: %w", err)
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.plan = plan
	return nil
}

// GetNextContact returns the next contact window with a specific node
func (cpm *ContactPlanManager) GetNextContact(nodeID NodeID, currentTime int64) (*ContactWindow, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.plan == nil {
		return nil, fmt.Errorf("no contact plan loaded")
	}

	// Find the earliest future contact with the specified node
	var nextContact *ContactWindow
	for i := range cpm.plan.Contacts {
		contact := &cpm.plan.Contacts[i]
		if contact.RemoteNode == nodeID && contact.StartTime >= currentTime {
			if nextContact == nil || contact.StartTime < nextContact.StartTime {
				nextContact = contact
			}
		}
	}

	if nextContact == nil {
		return nil, fmt.Errorf("no future contact with node %s", nodeID)
	}

	return nextContact, nil
}

// GetNextContactByEndpoint returns the next contact window for a destination endpoint
// Returns nil if no contact exists (Requirement 17.5)
func (cpm *ContactPlanManager) GetNextContactByEndpoint(destination bpa.EndpointID, currentTime int64) *ContactWindow {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.plan == nil {
		return nil
	}

	// Extract node ID from endpoint
	destNodeID := NodeID(destination.SSP)

	// Find the earliest future contact with the destination node
	var nextContact *ContactWindow
	for i := range cpm.plan.Contacts {
		contact := &cpm.plan.Contacts[i]
		if contact.RemoteNode == destNodeID && contact.StartTime >= currentTime {
			if nextContact == nil || contact.StartTime < nextContact.StartTime {
				nextContact = contact
			}
		}
	}

	return nextContact
}

// GetActiveContacts returns all currently active contact windows
func (cpm *ContactPlanManager) GetActiveContacts(currentTime int64) []ContactWindow {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.plan == nil {
		return nil
	}

	activeContacts := make([]ContactWindow, 0)
	for _, contact := range cpm.plan.Contacts {
		if contact.IsActive(currentTime) {
			activeContacts = append(activeContacts, contact)
		}
	}

	return activeContacts
}

// UpdatePlan updates a specific contact window in the plan
func (cpm *ContactPlanManager) UpdatePlan(contact ContactWindow) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if cpm.plan == nil {
		return fmt.Errorf("no contact plan loaded")
	}

	// Find and update the contact
	found := false
	for i := range cpm.plan.Contacts {
		if cpm.plan.Contacts[i].ContactID == contact.ContactID {
			cpm.plan.Contacts[i] = contact
			found = true
			break
		}
	}

	if !found {
		// Add new contact
		cpm.plan.Contacts = append(cpm.plan.Contacts, contact)
	}

	// Re-validate the plan
	return cpm.plan.Validate()
}

// FindDirectContact looks up the next direct contact window with a destination
func (cpm *ContactPlanManager) FindDirectContact(destination bpa.EndpointID, currentTime int64) (*ContactWindow, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.plan == nil {
		return nil, fmt.Errorf("no contact plan loaded")
	}

	// Extract node ID from endpoint
	// Assuming endpoint SSP contains the node ID
	destNodeID := NodeID(destination.SSP)

	// Find the earliest future contact with the destination node
	var nextContact *ContactWindow
	for i := range cpm.plan.Contacts {
		contact := &cpm.plan.Contacts[i]
		if contact.RemoteNode == destNodeID && contact.StartTime >= currentTime {
			if nextContact == nil || contact.StartTime < nextContact.StartTime {
				nextContact = contact
			}
		}
	}

	if nextContact == nil {
		return nil, fmt.Errorf("no direct contact with destination %s", destination.String())
	}

	return nextContact, nil
}

// PredictContacts computes predicted contact windows using CGR-based orbit propagation
// Automatically selects LEO or cislunar propagation based on orbital parameters
func (cpm *ContactPlanManager) PredictContacts(
	orbitalParams *OrbitalParameters,
	stations []GroundStationLocation,
	fromTime, toTime int64,
) ([]PredictedContact, error) {
	if err := orbitalParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid orbital parameters: %w", err)
	}

	if fromTime >= toTime {
		return nil, fmt.Errorf("fromTime must be less than toTime")
	}

	for _, station := range stations {
		if err := station.Validate(); err != nil {
			return nil, fmt.Errorf("invalid ground station: %w", err)
		}
	}

	// Use unified pass prediction (automatically selects LEO or cislunar)
	fromTimeObj := time.Unix(fromTime, 0)
	toTimeObj := time.Unix(toTime, 0)
	
	// Determine appropriate time step based on orbit type
	timeStep := 30 // Default for LEO
	if orbitalParams.DetermineOrbitType() == OrbitTypeCislunar {
		timeStep = 60 // Slower dynamics for cislunar
	}
	
	predicted, err := PredictPasses(
		orbitalParams,
		stations,
		fromTimeObj,
		toTimeObj,
		timeStep,
	)
	if err != nil {
		return nil, fmt.Errorf("CGR pass prediction failed: %w", err)
	}

	return predicted, nil
}

// UpdateOrbitalParameters updates orbital parameters for a space node
func (cpm *ContactPlanManager) UpdateOrbitalParameters(nodeID NodeID, params *OrbitalParameters) error {
	if err := params.Validate(); err != nil {
		return fmt.Errorf("invalid orbital parameters: %w", err)
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.orbitalParams[nodeID] = params

	// Orbital parameters updated - can trigger re-prediction of contact windows
	// using PredictContacts() or UpdateContactPlanWithPredictions()

	return nil
}

// GetNextPredictedPass gets the next predicted pass of a space node over a ground station
func (cpm *ContactPlanManager) GetNextPredictedPass(
	spaceNodeID, groundStationID NodeID,
	currentTime int64,
) (*PredictedContact, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	if cpm.plan == nil {
		return nil, fmt.Errorf("no contact plan loaded")
	}

	// Find the earliest future predicted contact
	var nextPass *PredictedContact
	for i := range cpm.plan.PredictedContacts {
		pc := &cpm.plan.PredictedContacts[i]
		if pc.Window.RemoteNode == groundStationID && pc.Window.StartTime >= currentTime {
			if nextPass == nil || pc.Window.StartTime < nextPass.Window.StartTime {
				nextPass = pc
			}
		}
	}

	if nextPass == nil {
		return nil, fmt.Errorf("no predicted pass found")
	}

	return nextPass, nil
}

// computeConfidence calculates prediction confidence based on time from epoch
// Confidence decreases as propagation time increases
func computeConfidence(epoch, predictionTime int64) float64 {
	// Time difference in days
	timeDiff := float64(predictionTime-epoch) / 86400.0

	// Confidence decreases exponentially with time
	// After 7 days, confidence is ~0.5
	// After 14 days, confidence is ~0.25
	confidence := math.Exp(-timeDiff / 10.0)

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// SortContactsByStartTime sorts contacts by start time (ascending)
func SortContactsByStartTime(contacts []ContactWindow) {
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].StartTime < contacts[j].StartTime
	})
}

// SortPredictedContactsByStartTime sorts predicted contacts by start time (ascending)
func SortPredictedContactsByStartTime(contacts []PredictedContact) {
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Window.StartTime < contacts[j].Window.StartTime
	})
}
