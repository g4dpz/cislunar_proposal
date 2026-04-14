package ion

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ContactPlanManager manages ION-DTN contact plans
type ContactPlanManager struct {
	lifecycle *NodeLifecycle
	plan      *ContactPlan
}

// ContactPlan represents a DTN contact plan
type ContactPlan struct {
	PlanID    string          `json:"plan_id" yaml:"plan_id"`
	ValidFrom int64           `json:"valid_from" yaml:"valid_from"`
	ValidTo   int64           `json:"valid_to" yaml:"valid_to"`
	Contacts  []Contact       `json:"contacts" yaml:"contacts"`
	Ranges    []Range         `json:"ranges" yaml:"ranges"`
}

// Contact represents a scheduled communication window
type Contact struct {
	ID         string  `json:"id" yaml:"id"`
	StartTime  int64   `json:"start_time" yaml:"start_time"`
	EndTime    int64   `json:"end_time" yaml:"end_time"`
	FromNode   int     `json:"from_node" yaml:"from_node"`
	ToNode     int     `json:"to_node" yaml:"to_node"`
	DataRate   int64   `json:"data_rate" yaml:"data_rate"` // bits per second
	Confidence float64 `json:"confidence" yaml:"confidence"`
}

// Range represents a distance range between nodes
type Range struct {
	StartTime int64 `json:"start_time" yaml:"start_time"`
	FromNode  int   `json:"from_node" yaml:"from_node"`
	ToNode    int   `json:"to_node" yaml:"to_node"`
	Distance  int64 `json:"distance" yaml:"distance"` // kilometers
}

// NewContactPlanManager creates a new contact plan manager
func NewContactPlanManager(lifecycle *NodeLifecycle) *ContactPlanManager {
	return &ContactPlanManager{
		lifecycle: lifecycle,
	}
}

// LoadFromYAML loads a contact plan from a YAML file
func (cpm *ContactPlanManager) LoadFromYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	var plan ContactPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := cpm.validatePlan(&plan); err != nil {
		return fmt.Errorf("invalid contact plan: %w", err)
	}

	cpm.plan = &plan
	return nil
}

// LoadFromJSON loads a contact plan from a JSON file
func (cpm *ContactPlanManager) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var plan ContactPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := cpm.validatePlan(&plan); err != nil {
		return fmt.Errorf("invalid contact plan: %w", err)
	}

	cpm.plan = &plan
	return nil
}

// validatePlan validates a contact plan
func (cpm *ContactPlanManager) validatePlan(plan *ContactPlan) error {
	if plan.ValidFrom >= plan.ValidTo {
		return fmt.Errorf("valid_from must be less than valid_to")
	}

	// Validate all contacts fall within valid range
	for i, contact := range plan.Contacts {
		if contact.StartTime < plan.ValidFrom || contact.EndTime > plan.ValidTo {
			return fmt.Errorf("contact %d falls outside valid time range", i)
		}
		if contact.StartTime >= contact.EndTime {
			return fmt.Errorf("contact %d: start_time must be less than end_time", i)
		}
		if contact.DataRate <= 0 {
			return fmt.Errorf("contact %d: data_rate must be positive", i)
		}
	}

	return nil
}

// Apply applies the contact plan to ION-DTN using ionadmin
func (cpm *ContactPlanManager) Apply() error {
	if cpm.plan == nil {
		return fmt.Errorf("no contact plan loaded")
	}

	if !cpm.lifecycle.IsRunning() {
		return fmt.Errorf("ION-DTN is not running")
	}

	// Generate ionadmin commands
	commands := cpm.generateIONAdminCommands()

	// Execute commands via ionadmin
	if err := cpm.executeIONAdminCommands(commands); err != nil {
		return fmt.Errorf("failed to apply contact plan: %w", err)
	}

	return nil
}

// generateIONAdminCommands generates ionadmin commands from the contact plan
func (cpm *ContactPlanManager) generateIONAdminCommands() []string {
	var commands []string

	// Add contacts
	for _, contact := range cpm.plan.Contacts {
		// Format: a contact +<start> +<duration> <from> <to> <rate> [confidence]
		duration := contact.EndTime - contact.StartTime
		cmd := fmt.Sprintf("a contact +%d +%d %d %d %d",
			contact.StartTime, duration, contact.FromNode, contact.ToNode, contact.DataRate)
		commands = append(commands, cmd)
	}

	// Add ranges
	for _, rng := range cpm.plan.Ranges {
		// Format: a range +<time> <from> <to> <distance>
		cmd := fmt.Sprintf("a range +%d %d %d %d",
			rng.StartTime, rng.FromNode, rng.ToNode, rng.Distance)
		commands = append(commands, cmd)
	}

	return commands
}

// executeIONAdminCommands executes a list of ionadmin commands
func (cpm *ContactPlanManager) executeIONAdminCommands(commands []string) error {
	if len(commands) == 0 {
		return nil
	}

	// Build command input
	input := strings.Join(commands, "\n") + "\nq\n"

	cmdPath := cpm.lifecycle.ionBinPath + "/ionadmin"
	cmd := exec.Command(cmdPath, ".")
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = cpm.lifecycle.buildEnv()
	cmd.Dir = cpm.lifecycle.config.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ionadmin failed: %w, output: %s", err, string(output))
	}

	return nil
}

// AddContact adds a single contact to the running ION-DTN node
func (cpm *ContactPlanManager) AddContact(contact Contact) error {
	if !cpm.lifecycle.IsRunning() {
		return fmt.Errorf("ION-DTN is not running")
	}

	// Validate contact
	if contact.StartTime >= contact.EndTime {
		return fmt.Errorf("start_time must be less than end_time")
	}
	if contact.DataRate <= 0 {
		return fmt.Errorf("data_rate must be positive")
	}

	// Generate and execute command
	duration := contact.EndTime - contact.StartTime
	cmd := fmt.Sprintf("a contact +%d +%d %d %d %d",
		contact.StartTime, duration, contact.FromNode, contact.ToNode, contact.DataRate)

	if err := cpm.executeIONAdminCommands([]string{cmd}); err != nil {
		return fmt.Errorf("failed to add contact: %w", err)
	}

	// Add to local plan if loaded
	if cpm.plan != nil {
		cpm.plan.Contacts = append(cpm.plan.Contacts, contact)
	}

	return nil
}

// RemoveContact removes a contact from the running ION-DTN node
func (cpm *ContactPlanManager) RemoveContact(fromNode, toNode int, startTime int64) error {
	if !cpm.lifecycle.IsRunning() {
		return fmt.Errorf("ION-DTN is not running")
	}

	// Format: d contact +<start> <from> <to>
	cmd := fmt.Sprintf("d contact +%d %d %d", startTime, fromNode, toNode)

	if err := cpm.executeIONAdminCommands([]string{cmd}); err != nil {
		return fmt.Errorf("failed to remove contact: %w", err)
	}

	// Remove from local plan if loaded
	if cpm.plan != nil {
		filtered := []Contact{}
		for _, c := range cpm.plan.Contacts {
			if !(c.FromNode == fromNode && c.ToNode == toNode && c.StartTime == startTime) {
				filtered = append(filtered, c)
			}
		}
		cpm.plan.Contacts = filtered
	}

	return nil
}

// UpdateContact updates an existing contact
func (cpm *ContactPlanManager) UpdateContact(contact Contact) error {
	// Remove old contact and add new one
	if err := cpm.RemoveContact(contact.FromNode, contact.ToNode, contact.StartTime); err != nil {
		return fmt.Errorf("failed to remove old contact: %w", err)
	}

	if err := cpm.AddContact(contact); err != nil {
		return fmt.Errorf("failed to add updated contact: %w", err)
	}

	return nil
}

// ListContacts returns all contacts in the current plan
func (cpm *ContactPlanManager) ListContacts() ([]Contact, error) {
	if cpm.plan == nil {
		return nil, fmt.Errorf("no contact plan loaded")
	}

	return cpm.plan.Contacts, nil
}

// GetActiveContacts returns contacts active at the given time
func (cpm *ContactPlanManager) GetActiveContacts(currentTime int64) []Contact {
	if cpm.plan == nil {
		return nil
	}

	var active []Contact
	for _, contact := range cpm.plan.Contacts {
		if contact.StartTime <= currentTime && currentTime < contact.EndTime {
			active = append(active, contact)
		}
	}

	return active
}

// GetNextContact returns the next contact with the specified node
func (cpm *ContactPlanManager) GetNextContact(toNode int, afterTime int64) *Contact {
	if cpm.plan == nil {
		return nil
	}

	var nextContact *Contact
	for i := range cpm.plan.Contacts {
		contact := &cpm.plan.Contacts[i]
		if contact.ToNode == toNode && contact.StartTime >= afterTime {
			if nextContact == nil || contact.StartTime < nextContact.StartTime {
				nextContact = contact
			}
		}
	}

	return nextContact
}

// SaveToYAML saves the current contact plan to a YAML file
func (cpm *ContactPlanManager) SaveToYAML(path string) error {
	if cpm.plan == nil {
		return fmt.Errorf("no contact plan loaded")
	}

	data, err := yaml.Marshal(cpm.plan)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// SaveToJSON saves the current contact plan to a JSON file
func (cpm *ContactPlanManager) SaveToJSON(path string) error {
	if cpm.plan == nil {
		return fmt.Errorf("no contact plan loaded")
	}

	data, err := json.MarshalIndent(cpm.plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// CreateAlwaysOnPlan creates a simple always-on contact plan for terrestrial nodes
func CreateAlwaysOnPlan(nodeA, nodeB int, dataRate int64) *ContactPlan {
	now := time.Now().Unix()
	
	return &ContactPlan{
		PlanID:    fmt.Sprintf("terrestrial-%d-%d", nodeA, nodeB),
		ValidFrom: now,
		ValidTo:   now + (365 * 24 * 3600), // 1 year
		Contacts: []Contact{
			{
				ID:         "always-on-a-to-b",
				StartTime:  now,
				EndTime:    now + (365 * 24 * 3600),
				FromNode:   nodeA,
				ToNode:     nodeB,
				DataRate:   dataRate,
				Confidence: 1.0,
			},
			{
				ID:         "always-on-b-to-a",
				StartTime:  now,
				EndTime:    now + (365 * 24 * 3600),
				FromNode:   nodeB,
				ToNode:     nodeA,
				DataRate:   dataRate,
				Confidence: 1.0,
			},
		},
		Ranges: []Range{
			{
				StartTime: now,
				FromNode:  nodeA,
				ToNode:    nodeB,
				Distance:  1, // 1 km for terrestrial
			},
		},
	}
}
