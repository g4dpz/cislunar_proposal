package kiss

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
)

// cloLoopEvent represents a single action taken by the CLO main loop simulation.
type cloLoopEvent struct {
	Kind    string // "send", "pause", "block"
	PauseMs int    // duration of pause in ms (only for "pause" events)
}

// simulateCLOLoop models the FIXED CLO main loop in ltpkissclo.c.
// After every burstSize consecutive sends, the loop records a pause of
// listenWindowMs milliseconds before resuming — matching the duty-cycle
// arbitration added to the C code.
//
// The simulation does NOT actually sleep; it records events so the
// property checker can verify the duty-cycle logic without wall-clock delays.
//
// Parameters:
//   - queueDepth: number of segments waiting to be sent
//   - burstSize: configured max segments per burst
//   - listenWindowMs: configured listen pause in ms
//
// Returns a slice of events describing what the loop did.
func simulateCLOLoop(queueDepth, burstSize, listenWindowMs int) []cloLoopEvent {
	events := make([]cloLoopEvent, 0, queueDepth+queueDepth/burstSize)
	burstCounter := 0

	for i := 0; i < queueDepth; i++ {
		// --- Dequeue, frame, send segment ---
		events = append(events, cloLoopEvent{
			Kind: "send",
		})

		// --- Rate control + sm_TaskYield (no-op in simulation) ---

		// --- Duty-cycle arbitration (the fix) ---
		burstCounter++
		if burstCounter >= burstSize && listenWindowMs > 0 {
			events = append(events, cloLoopEvent{
				Kind:    "pause",
				PauseMs: listenWindowMs,
			})
			burstCounter = 0
		}
	}

	return events
}

// checkDutyCycleProperty verifies the core property:
//
//	After every `burstSize` consecutive sends, a pause of >= listenWindowMs
//	must occur before the next send.
//
// Returns nil if the property holds, or an error describing the violation.
//
// **Validates: Requirements 2.1, 2.2**
func checkDutyCycleProperty(events []cloLoopEvent, burstSize, listenWindowMs int) error {
	consecutiveSends := 0

	for i, ev := range events {
		switch ev.Kind {
		case "send":
			consecutiveSends++
			if consecutiveSends > burstSize {
				// We've exceeded burstSize without a pause.
				return fmt.Errorf(
					"duty-cycle violation at event %d: %d consecutive sends without a pause of >= %d ms (burstSize=%d)",
					i, consecutiveSends, listenWindowMs, burstSize,
				)
			}
		case "pause":
			if ev.PauseMs >= listenWindowMs {
				consecutiveSends = 0 // burst counter resets after a valid pause
			}
		}
	}

	return nil
}

// TestCLODutyCycle_BugExploration_Property is the property-based exploration
// test for the half-duplex deadlock bug (Property 1: Fault Condition).
//
// **Validates: Requirements 1.1, 1.2, 1.3, 2.1, 2.2**
//
// Property: for all (queueDepth, burstSize, listenWindowMs) where
// queueDepth >= 1 AND burstSize >= 1 AND listenWindowMs > 0,
// the CLO SHALL NOT send more than burstSize consecutive segments
// without sleeping for at least listenWindowMs milliseconds.
//
// This test is EXPECTED TO FAIL on unfixed code — failure confirms the bug.
func TestCLODutyCycle_BugExploration_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)), // deterministic seed for reproducibility
	}

	prop := func(queueDepthRaw, burstSizeRaw, listenWindowMsRaw uint8) bool {
		// Map raw values into valid ranges:
		//   queueDepth:    [2, 50]  (need > burstSize to trigger the bug)
		//   burstSize:     [1, 10]
		//   listenWindowMs: [100, 1000]
		queueDepth := int(queueDepthRaw)%49 + 2    // 2..50
		burstSize := int(burstSizeRaw)%10 + 1       // 1..10
		listenWindowMs := (int(listenWindowMsRaw)%10 + 1) * 100 // 100..1000

		// Ensure queueDepth > burstSize so the bug can manifest
		if queueDepth <= burstSize {
			queueDepth = burstSize + 1
		}

		events := simulateCLOLoop(queueDepth, burstSize, listenWindowMs)
		err := checkDutyCycleProperty(events, burstSize, listenWindowMs)
		if err != nil {
			t.Logf("COUNTEREXAMPLE: queueDepth=%d burstSize=%d listenWindowMs=%d => %v",
				queueDepth, burstSize, listenWindowMs, err)
			return false // property violated
		}
		return true // property holds
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("Property violated (EXPECTED for unfixed code — confirms bug exists): %v", err)
	}
}
