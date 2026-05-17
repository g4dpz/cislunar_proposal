package hdtn

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadFromFile_ValidJSON(t *testing.T) {
	contacts := contactPlanFile{
		Contacts: []Contact{
			{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
			{Source: 2, Dest: 1, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		},
	}

	data, err := json.Marshal(contacts)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	cpm := NewContactPlanManager("http://unused")
	if err := cpm.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("expected no error loading valid JSON, got: %v", err)
	}

	loaded, _ := cpm.ListContacts()
	if len(loaded) != 2 {
		t.Fatalf("expected 2 contacts, got %d", len(loaded))
	}
}

func TestLoadFromFile_ValidYAML(t *testing.T) {
	contacts := contactPlanFile{
		Contacts: []Contact{
			{Source: 1, Dest: 2, StartTime: 0, EndTime: 200, RateBitsPerSec: 4800},
			{Source: 2, Dest: 1, StartTime: 10, EndTime: 200, RateBitsPerSec: 4800},
			{Source: 1, Dest: 3, StartTime: 50, EndTime: 150, RateBitsPerSec: 1200},
		},
	}

	data, err := yaml.Marshal(contacts)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile := filepath.Join(t.TempDir(), "plan.yaml")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	cpm := NewContactPlanManager("http://unused")
	if err := cpm.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("expected no error loading valid YAML, got: %v", err)
	}

	loaded, _ := cpm.ListContacts()
	if len(loaded) != 3 {
		t.Fatalf("expected 3 contacts, got %d", len(loaded))
	}
}

func TestValidateContacts_RejectsRateZero(t *testing.T) {
	contacts := []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 2, Dest: 1, StartTime: 0, EndTime: 100, RateBitsPerSec: 0},
	}

	err := ValidateContacts(contacts)
	if err == nil {
		t.Fatal("expected validation error for rate = 0")
	}
	if !contains(err.Error(), "contact[1]") {
		t.Fatalf("expected error to identify contact[1], got: %v", err)
	}
	if !contains(err.Error(), "rate") {
		t.Fatalf("expected error to mention rate, got: %v", err)
	}
}

func TestValidateContacts_RejectsNegativeRate(t *testing.T) {
	contacts := []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: -500},
	}

	err := ValidateContacts(contacts)
	if err == nil {
		t.Fatal("expected validation error for negative rate")
	}
	if !contains(err.Error(), "contact[0]") {
		t.Fatalf("expected error to identify contact[0], got: %v", err)
	}
}

func TestValidateContacts_RejectsStartEqualEnd(t *testing.T) {
	contacts := []Contact{
		{Source: 1, Dest: 2, StartTime: 100, EndTime: 100, RateBitsPerSec: 9600},
	}

	err := ValidateContacts(contacts)
	if err == nil {
		t.Fatal("expected validation error for start == end")
	}
	if !contains(err.Error(), "start time") {
		t.Fatalf("expected error to mention start time, got: %v", err)
	}
}

func TestValidateContacts_RejectsStartAfterEnd(t *testing.T) {
	contacts := []Contact{
		{Source: 1, Dest: 2, StartTime: 200, EndTime: 100, RateBitsPerSec: 9600},
	}

	err := ValidateContacts(contacts)
	if err == nil {
		t.Fatal("expected validation error for start > end")
	}
}

func TestValidateContacts_RejectsMoreThan1000(t *testing.T) {
	contacts := make([]Contact, 1001)
	for i := range contacts {
		contacts[i] = Contact{
			Source:         1,
			Dest:           2,
			StartTime:      int64(i),
			EndTime:        int64(i + 100),
			RateBitsPerSec: 9600,
		}
	}

	err := ValidateContacts(contacts)
	if err == nil {
		t.Fatal("expected validation error for > 1000 contacts")
	}
	if !contains(err.Error(), "1000") {
		t.Fatalf("expected error to mention 1000, got: %v", err)
	}
}

func TestValidateContacts_AcceptsExactly1000(t *testing.T) {
	contacts := make([]Contact, 1000)
	for i := range contacts {
		contacts[i] = Contact{
			Source:         1,
			Dest:           2,
			StartTime:      int64(i),
			EndTime:        int64(i + 100),
			RateBitsPerSec: 9600,
		}
	}

	err := ValidateContacts(contacts)
	if err != nil {
		t.Fatalf("expected no error for exactly 1000 valid contacts, got: %v", err)
	}
}

func TestApply_StopsOnFirstFailure(t *testing.T) {
	var callCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)
	cpm.mu.Lock()
	cpm.contacts = []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 2, Dest: 1, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 1, Dest: 3, StartTime: 0, EndTime: 100, RateBitsPerSec: 4800},
	}
	cpm.mu.Unlock()

	err := cpm.Apply()
	if err == nil {
		t.Fatal("expected error from Apply")
	}

	// Should have stopped after the second call (which failed)
	if atomic.LoadInt32(&callCount) != 2 {
		t.Fatalf("expected 2 API calls (stop on first failure), got %d", atomic.LoadInt32(&callCount))
	}

	if !contains(err.Error(), "contact[1]") {
		t.Fatalf("expected error to identify contact[1], got: %v", err)
	}
}

func TestAddContact_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)

	contact := Contact{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600}
	err := cpm.AddContact(contact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contacts, _ := cpm.ListContacts()
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	if contacts[0] != contact {
		t.Fatalf("contact mismatch: got %+v, want %+v", contacts[0], contact)
	}
}

func TestAddContact_Failure_LocalStateUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)
	// Pre-populate with one contact
	existing := Contact{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600}
	cpm.mu.Lock()
	cpm.contacts = []Contact{existing}
	cpm.mu.Unlock()

	newContact := Contact{Source: 3, Dest: 4, StartTime: 50, EndTime: 200, RateBitsPerSec: 4800}
	err := cpm.AddContact(newContact)
	if err == nil {
		t.Fatal("expected error from AddContact")
	}

	contacts, _ := cpm.ListContacts()
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact (unchanged), got %d", len(contacts))
	}
	if contacts[0] != existing {
		t.Fatalf("existing contact was modified")
	}
}

func TestRemoveContact_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)
	cpm.mu.Lock()
	cpm.contacts = []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 2, Dest: 1, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
	}
	cpm.mu.Unlock()

	err := cpm.RemoveContact(1, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contacts, _ := cpm.ListContacts()
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact remaining, got %d", len(contacts))
	}
	if contacts[0].Source != 2 || contacts[0].Dest != 1 {
		t.Fatalf("wrong contact remaining: %+v", contacts[0])
	}
}

func TestRemoveContact_Failure_LocalStateUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)
	contacts := []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 2, Dest: 1, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
	}
	cpm.mu.Lock()
	cpm.contacts = make([]Contact, len(contacts))
	copy(cpm.contacts, contacts)
	cpm.mu.Unlock()

	err := cpm.RemoveContact(1, 2, 0)
	if err == nil {
		t.Fatal("expected error from RemoveContact")
	}

	remaining, _ := cpm.ListContacts()
	if len(remaining) != 2 {
		t.Fatalf("expected 2 contacts (unchanged), got %d", len(remaining))
	}
}

func TestRemoveContact_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cpm := NewContactPlanManager(server.URL)
	cpm.mu.Lock()
	cpm.contacts = []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
	}
	cpm.mu.Unlock()

	err := cpm.RemoveContact(99, 99, 0)
	if err == nil {
		t.Fatal("expected error for non-existent contact")
	}
	if !contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %v", err)
	}
}

func TestGetActiveContacts_Various(t *testing.T) {
	cpm := NewContactPlanManager("http://unused")
	cpm.mu.Lock()
	cpm.contacts = []Contact{
		{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
		{Source: 2, Dest: 1, StartTime: 50, EndTime: 150, RateBitsPerSec: 9600},
		{Source: 1, Dest: 3, StartTime: 200, EndTime: 300, RateBitsPerSec: 4800},
	}
	cpm.mu.Unlock()

	tests := []struct {
		name     string
		time     int64
		expected int
	}{
		{"before all", -1, 0},
		{"at start of first", 0, 1},
		{"during first only", 25, 1},
		{"during first and second", 50, 2},
		{"during first and second (mid)", 75, 2},
		{"at end of first (exclusive)", 100, 1},
		{"during second only", 120, 1},
		{"between second and third", 175, 0},
		{"during third", 250, 1},
		{"at end of third (exclusive)", 300, 0},
		{"after all", 500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			active := cpm.GetActiveContacts(tt.time)
			if len(active) != tt.expected {
				t.Fatalf("at time %d: expected %d active contacts, got %d", tt.time, tt.expected, len(active))
			}
		})
	}
}

func TestLoadFromFile_InvalidFormat(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "plan.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cpm := NewContactPlanManager("http://unused")
	err := cpm.LoadFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !contains(err.Error(), "unsupported") {
		t.Fatalf("expected 'unsupported' error, got: %v", err)
	}
}

func TestLoadFromFile_ValidationFailure(t *testing.T) {
	contacts := contactPlanFile{
		Contacts: []Contact{
			{Source: 1, Dest: 2, StartTime: 0, EndTime: 100, RateBitsPerSec: 9600},
			{Source: 2, Dest: 1, StartTime: 100, EndTime: 50, RateBitsPerSec: 9600}, // invalid: start > end
		},
	}

	data, err := json.Marshal(contacts)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	cpm := NewContactPlanManager("http://unused")
	err = cpm.LoadFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !contains(err.Error(), "contact[1]") {
		t.Fatalf("expected error to identify contact[1], got: %v", err)
	}
}

// contains checks if s contains substr (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
