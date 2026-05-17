package hdtn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Contact represents a scheduled communication window.
// Used by ContactPlanManager for runtime contact plan operations.
type Contact struct {
	Source         int   `json:"source" yaml:"source"`
	Dest           int   `json:"dest" yaml:"dest"`
	StartTime      int64 `json:"startTime" yaml:"start_time"`
	EndTime        int64 `json:"endTime" yaml:"end_time"`
	RateBitsPerSec int64 `json:"rateBitsPerSec" yaml:"rate_bps"`
}

// ContactPlanManager manages contacts via HDTN REST API.
type ContactPlanManager struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	contacts   []Contact
}

// NewContactPlanManager creates a contact plan manager.
func NewContactPlanManager(baseURL string) *ContactPlanManager {
	return &ContactPlanManager{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		contacts:   []Contact{},
	}
}

// ValidateContacts validates a slice of contacts.
// Returns an error identifying the index of the first invalid entry and the reason.
func ValidateContacts(contacts []Contact) error {
	if len(contacts) > 1000 {
		return fmt.Errorf("contact plan exceeds maximum of 1000 entries (has %d)", len(contacts))
	}
	for i, c := range contacts {
		if c.RateBitsPerSec <= 0 {
			return fmt.Errorf("contact[%d]: rate must be greater than 0 (got %d)", i, c.RateBitsPerSec)
		}
		if c.StartTime >= c.EndTime {
			return fmt.Errorf("contact[%d]: start time (%d) must be before end time (%d)", i, c.StartTime, c.EndTime)
		}
	}
	return nil
}

// contactPlanFile represents the structure of a contact plan file (JSON or YAML).
type contactPlanFile struct {
	Contacts []Contact `json:"contacts" yaml:"contacts"`
}

// LoadFromFile loads and validates a contact plan from a JSON or YAML file.
// On validation failure, returns an error identifying which entry failed and why,
// without submitting anything.
func (cpm *ContactPlanManager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read contact plan file: %w", err)
	}

	var plan contactPlanFile
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &plan); err != nil {
			return fmt.Errorf("failed to parse JSON contact plan: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &plan); err != nil {
			return fmt.Errorf("failed to parse YAML contact plan: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s (use .json or .yaml)", ext)
	}

	if err := ValidateContacts(plan.Contacts); err != nil {
		return err
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	cpm.contacts = plan.Contacts
	return nil
}

// Apply submits each contact to the HDTN REST API via POST /api/v1/contacts.
// Stops on first failure, returns error identifying the failed contact.
// Previously submitted contacts remain.
func (cpm *ContactPlanManager) Apply() error {
	cpm.mu.Lock()
	contacts := make([]Contact, len(cpm.contacts))
	copy(contacts, cpm.contacts)
	cpm.mu.Unlock()

	for i, c := range contacts {
		if err := cpm.postContact(c); err != nil {
			return fmt.Errorf("failed to apply contact[%d] (source=%d, dest=%d, start=%d): %w",
				i, c.Source, c.Dest, c.StartTime, err)
		}
	}
	return nil
}

// AddContact submits a single contact to the API and updates local state on success.
// On API error, local state is unchanged.
func (cpm *ContactPlanManager) AddContact(contact Contact) error {
	if err := cpm.postContact(contact); err != nil {
		return err
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	cpm.contacts = append(cpm.contacts, contact)
	return nil
}

// RemoveContact identifies a contact by (source, dest, startTime), issues DELETE to API,
// and removes from local state on success. On API error, local state is unchanged.
func (cpm *ContactPlanManager) RemoveContact(source, dest int, startTime int64) error {
	cpm.mu.Lock()
	idx := -1
	for i, c := range cpm.contacts {
		if c.Source == source && c.Dest == dest && c.StartTime == startTime {
			idx = i
			break
		}
	}
	cpm.mu.Unlock()

	if idx == -1 {
		return fmt.Errorf("contact not found (source=%d, dest=%d, startTime=%d)", source, dest, startTime)
	}

	if err := cpm.deleteContact(source, dest, startTime); err != nil {
		return err
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()
	// Re-find the index in case state changed between unlock and lock
	for i, c := range cpm.contacts {
		if c.Source == source && c.Dest == dest && c.StartTime == startTime {
			cpm.contacts = append(cpm.contacts[:i], cpm.contacts[i+1:]...)
			break
		}
	}
	return nil
}

// GetActiveContacts returns contacts where StartTime ≤ currentTime < EndTime.
func (cpm *ContactPlanManager) GetActiveContacts(currentTime int64) []Contact {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	var active []Contact
	for _, c := range cpm.contacts {
		if c.StartTime <= currentTime && currentTime < c.EndTime {
			active = append(active, c)
		}
	}
	return active
}

// ListContacts returns all contacts in local state.
func (cpm *ContactPlanManager) ListContacts() ([]Contact, error) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	result := make([]Contact, len(cpm.contacts))
	copy(result, cpm.contacts)
	return result, nil
}

// postContact submits a single contact to the HDTN REST API.
func (cpm *ContactPlanManager) postContact(contact Contact) error {
	body, err := json.Marshal(contact)
	if err != nil {
		return fmt.Errorf("failed to marshal contact: %w", err)
	}

	url := cpm.baseURL + "/api/v1/contacts"
	resp, err := cpm.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	return nil
}

// deleteContact issues a DELETE request to the HDTN REST API.
func (cpm *ContactPlanManager) deleteContact(source, dest int, startTime int64) error {
	url := fmt.Sprintf("%s/api/v1/contacts?source=%d&dest=%d&startTime=%d",
		cpm.baseURL, source, dest, startTime)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := cpm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	return nil
}
