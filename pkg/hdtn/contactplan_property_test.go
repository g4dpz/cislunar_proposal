package hdtn

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Feature: test-framework-srs-sdd, Property 10: Contact Plan Validation Correctness
// **Validates: SRS-TF-010 (Requirements 10.1, 10.2)**
//
// For any collection of contact entries, validation SHALL accept the collection if and only if:
// every contact has RateBitsPerSec > 0, every contact has StartTime < EndTime, and the total
// number of entries is ≤ 1000. When validation fails, the error SHALL identify the index of
// the first invalid entry.
func TestProperty7_ContactPlanValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a collection of contacts (0 to 1050 entries to test boundary)
		count := rapid.IntRange(0, 1050).Draw(t, "count")
		contacts := make([]Contact, count)

		for i := 0; i < count; i++ {
			// Generate contacts that may or may not be valid
			rate := rapid.Int64Range(-10, 100000).Draw(t, fmt.Sprintf("rate_%d", i))
			start := rapid.Int64Range(0, 1000000).Draw(t, fmt.Sprintf("start_%d", i))
			duration := rapid.Int64Range(-100, 100000).Draw(t, fmt.Sprintf("duration_%d", i))
			end := start + duration

			contacts[i] = Contact{
				Source:         rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("source_%d", i)),
				Dest:           rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("dest_%d", i)),
				StartTime:      start,
				EndTime:        end,
				RateBitsPerSec: rate,
			}
		}

		err := ValidateContacts(contacts)

		// Determine expected validity
		allValid := true
		firstInvalidIdx := -1
		if count > 1000 {
			allValid = false
			// The "too many entries" error doesn't identify a specific index
		} else {
			for i, c := range contacts {
				if c.RateBitsPerSec <= 0 || c.StartTime >= c.EndTime {
					allValid = false
					firstInvalidIdx = i
					break
				}
			}
		}

		if allValid {
			if err != nil {
				t.Fatalf("expected validation to pass for valid contacts, got error: %v", err)
			}
		} else {
			if err == nil {
				t.Fatalf("expected validation to fail for invalid contacts (count=%d, firstInvalid=%d)", count, firstInvalidIdx)
			}
			// If the error is about count > 1000, it won't have an index
			if count <= 1000 && firstInvalidIdx >= 0 {
				// Error should identify the index of the first invalid entry
				expectedIdx := fmt.Sprintf("contact[%d]", firstInvalidIdx)
				if !strings.Contains(err.Error(), expectedIdx) {
					t.Fatalf("error should identify index %d, got: %v", firstInvalidIdx, err)
				}
			}
		}
	})
}

// Feature: test-framework-srs-sdd, Property 11: Active Contacts Filtering Correctness
// **Validates: SRS-TF-011 (Requirements 11.1, 11.2)**
//
// For any set of contacts and any query time T, GetActiveContacts(T) SHALL return exactly
// those contacts where StartTime ≤ T < EndTime, and no others.
func TestProperty8_ActiveContactsFiltering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a set of contacts (1 to 50)
		count := rapid.IntRange(1, 50).Draw(t, "count")
		contacts := make([]Contact, count)

		for i := 0; i < count; i++ {
			start := rapid.Int64Range(0, 1000000).Draw(t, fmt.Sprintf("start_%d", i))
			end := start + rapid.Int64Range(1, 100000).Draw(t, fmt.Sprintf("duration_%d", i))
			contacts[i] = Contact{
				Source:         rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("source_%d", i)),
				Dest:           rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("dest_%d", i)),
				StartTime:      start,
				EndTime:        end,
				RateBitsPerSec: rapid.Int64Range(1, 1000000).Draw(t, fmt.Sprintf("rate_%d", i)),
			}
		}

		// Generate a query time
		queryTime := rapid.Int64Range(-100, 1200000).Draw(t, "queryTime")

		// Set up the contact plan manager with these contacts
		cpm := NewContactPlanManager("http://unused")
		cpm.mu.Lock()
		cpm.contacts = contacts
		cpm.mu.Unlock()

		// Get active contacts
		active := cpm.GetActiveContacts(queryTime)

		// Compute expected active contacts
		var expected []Contact
		for _, c := range contacts {
			if c.StartTime <= queryTime && queryTime < c.EndTime {
				expected = append(expected, c)
			}
		}

		// Assert lengths match
		if len(active) != len(expected) {
			t.Fatalf("expected %d active contacts at time %d, got %d", len(expected), queryTime, len(active))
		}

		// Assert each expected contact is in the active set
		for _, exp := range expected {
			found := false
			for _, act := range active {
				if act.Source == exp.Source && act.Dest == exp.Dest &&
					act.StartTime == exp.StartTime && act.EndTime == exp.EndTime &&
					act.RateBitsPerSec == exp.RateBitsPerSec {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected contact (source=%d, dest=%d, start=%d, end=%d) to be active at time %d",
					exp.Source, exp.Dest, exp.StartTime, exp.EndTime, queryTime)
			}
		}
	})
}

// Feature: test-framework-srs-sdd, Property 12: Contact Removal Correctness
// **Validates: SRS-TF-012 (Requirements 12.1, 12.2, 12.3, 12.4)**
//
// For any contact plan containing at least one contact, removing a contact by its
// (source, dest, startTime) key SHALL result in a local plan that no longer contains
// that contact, and all other contacts remain unchanged.
func TestProperty9_ContactRemovalFromLocalState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a plan with at least one contact (1 to 50)
		count := rapid.IntRange(1, 50).Draw(t, "count")
		contacts := make([]Contact, count)

		for i := 0; i < count; i++ {
			start := rapid.Int64Range(0, 1000000).Draw(t, fmt.Sprintf("start_%d", i))
			end := start + rapid.Int64Range(1, 100000).Draw(t, fmt.Sprintf("duration_%d", i))
			contacts[i] = Contact{
				Source:         rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("source_%d", i)),
				Dest:           rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("dest_%d", i)),
				StartTime:      start,
				EndTime:        end,
				RateBitsPerSec: rapid.Int64Range(1, 1000000).Draw(t, fmt.Sprintf("rate_%d", i)),
			}
		}

		// Pick a random contact to remove
		removeIdx := rapid.IntRange(0, count-1).Draw(t, "removeIdx")
		toRemove := contacts[removeIdx]

		// Set up a mock HTTP server that always returns success for DELETE
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Set up the contact plan manager
		cpm := NewContactPlanManager(server.URL)
		cpm.mu.Lock()
		cpm.contacts = make([]Contact, len(contacts))
		copy(cpm.contacts, contacts)
		cpm.mu.Unlock()

		// Remove the contact
		err := cpm.RemoveContact(toRemove.Source, toRemove.Dest, toRemove.StartTime)
		if err != nil {
			t.Fatalf("unexpected error removing contact: %v", err)
		}

		// Get remaining contacts
		remaining, _ := cpm.ListContacts()

		// Assert the removed contact is gone
		for _, c := range remaining {
			if c.Source == toRemove.Source && c.Dest == toRemove.Dest && c.StartTime == toRemove.StartTime {
				t.Fatalf("removed contact (source=%d, dest=%d, start=%d) still present in plan",
					toRemove.Source, toRemove.Dest, toRemove.StartTime)
			}
		}

		// Assert all other contacts remain
		// Count how many contacts with the same key existed in the original
		sameKeyCount := 0
		for _, c := range contacts {
			if c.Source == toRemove.Source && c.Dest == toRemove.Dest && c.StartTime == toRemove.StartTime {
				sameKeyCount++
			}
		}

		// Expected remaining count: original count minus the number of contacts with the same key
		// that were removed (RemoveContact removes the first match)
		expectedRemaining := count - 1
		if len(remaining) != expectedRemaining {
			t.Fatalf("expected %d remaining contacts, got %d", expectedRemaining, len(remaining))
		}
	})
}

// Feature: test-framework-srs-sdd, Property 13: API Error State Preservation
// **Validates: SRS-TF-013 (Requirements 13.1, 13.2)**
//
// For any contact plan manager with existing local state, if an API operation (add, remove, apply)
// fails due to an API error, the local plan state SHALL be identical to the state before the
// operation was attempted.
func TestProperty10_APIErrorPreservesLocalState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate initial contacts (1 to 20)
		count := rapid.IntRange(1, 20).Draw(t, "count")
		contacts := make([]Contact, count)

		for i := 0; i < count; i++ {
			start := rapid.Int64Range(0, 1000000).Draw(t, fmt.Sprintf("start_%d", i))
			end := start + rapid.Int64Range(1, 100000).Draw(t, fmt.Sprintf("duration_%d", i))
			contacts[i] = Contact{
				Source:         rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("source_%d", i)),
				Dest:           rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("dest_%d", i)),
				StartTime:      start,
				EndTime:        end,
				RateBitsPerSec: rapid.Int64Range(1, 1000000).Draw(t, fmt.Sprintf("rate_%d", i)),
			}
		}

		// Set up a mock HTTP server that always returns 500 (failure)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		// Choose which operation to test
		operation := rapid.IntRange(0, 2).Draw(t, "operation")

		// Set up the contact plan manager with initial state
		cpm := NewContactPlanManager(server.URL)
		cpm.mu.Lock()
		cpm.contacts = make([]Contact, len(contacts))
		copy(cpm.contacts, contacts)
		cpm.mu.Unlock()

		// Snapshot the state before the operation
		beforeState, _ := cpm.ListContacts()

		switch operation {
		case 0:
			// Test AddContact failure
			newContact := Contact{
				Source:         rapid.IntRange(1, 100).Draw(t, "new_source"),
				Dest:           rapid.IntRange(1, 100).Draw(t, "new_dest"),
				StartTime:      rapid.Int64Range(0, 1000000).Draw(t, "new_start"),
				EndTime:        rapid.Int64Range(1000001, 2000000).Draw(t, "new_end"),
				RateBitsPerSec: rapid.Int64Range(1, 1000000).Draw(t, "new_rate"),
			}
			_ = cpm.AddContact(newContact)

		case 1:
			// Test RemoveContact failure
			if count > 0 {
				removeIdx := rapid.IntRange(0, count-1).Draw(t, "removeIdx")
				toRemove := contacts[removeIdx]
				_ = cpm.RemoveContact(toRemove.Source, toRemove.Dest, toRemove.StartTime)
			}

		case 2:
			// Test Apply failure
			_ = cpm.Apply()
		}

		// Assert local state is unchanged
		afterState, _ := cpm.ListContacts()

		if len(afterState) != len(beforeState) {
			t.Fatalf("local state changed after API error: had %d contacts, now have %d",
				len(beforeState), len(afterState))
		}

		for i := range beforeState {
			if beforeState[i] != afterState[i] {
				t.Fatalf("contact[%d] changed after API error: before=%+v, after=%+v",
					i, beforeState[i], afterState[i])
			}
		}
	})
}
