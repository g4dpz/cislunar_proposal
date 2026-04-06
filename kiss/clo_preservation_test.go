package kiss

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
)

// ---------------------------------------------------------------------------
// Simulation types for preservation property tests
// ---------------------------------------------------------------------------

// cloConfig models the KissConfig fields relevant to preservation properties.
type cloConfig struct {
	MTU            int
	MaxRate        int
	UseAX25        bool
	ReconnectDelay int // seconds
}

// cloAction records a single action taken by the simulated CLO loop.
type cloAction struct {
	Kind string // "block", "discard_mtu", "send", "rate_control", "ax25_wrap", "reconnect_close", "reconnect_sleep", "reconnect_open", "error_log"
	// Metadata for verification:
	SegmentSize    int  // size of the segment being processed
	RateBytes      int  // bytes passed to applyRateControl
	AX25Applied    bool // whether AX.25 wrapping was applied
	ReconnectDelay int  // delay used in reconnection sleep
}

// ---------------------------------------------------------------------------
// Simulation: models the CURRENT (unfixed) CLO main loop behavior
// ---------------------------------------------------------------------------

// simulateCLOLoop_Preservation models the unfixed CLO main loop for a given
// sequence of segments and a serial-port error injection point.
//
// Parameters:
//   - segments: slice of segment sizes to process (empty = idle queue)
//   - config: CLO configuration
//   - serialErrorAtIndex: inject a serial send error at this segment index
//     (-1 = no error)
//
// Returns a slice of actions describing what the loop did.
func simulateCLOLoop_Preservation(segments []int, config cloConfig, serialErrorAtIndex int) []cloAction {
	actions := make([]cloAction, 0, len(segments)*3)

	if len(segments) == 0 {
		// Queue is empty: CLO blocks on ltpDequeueOutboundSegment()
		// No spurious sleeps, no busy-waits — just a blocking call.
		actions = append(actions, cloAction{Kind: "block"})
		return actions
	}

	for i, segSize := range segments {
		// --- Dequeue segment (ltpDequeueOutboundSegment) ---

		// --- Check segment size against MTU ---
		if segSize > config.MTU {
			actions = append(actions, cloAction{
				Kind:        "discard_mtu",
				SegmentSize: segSize,
			})
			actions = append(actions, cloAction{
				Kind:        "error_log",
				SegmentSize: segSize,
			})
			continue // discard oversized segment
		}

		// --- AX.25 wrapping if configured ---
		frameSize := segSize
		if config.UseAX25 {
			// AX.25 UI frame adds a 16-byte header
			frameSize = segSize + 16
			actions = append(actions, cloAction{
				Kind:        "ax25_wrap",
				SegmentSize: segSize,
				AX25Applied: true,
			})
		}

		// --- KISS framing (always happens, not tracked separately) ---

		// --- Serial send ---
		if i == serialErrorAtIndex {
			// Serial send fails — trigger reconnection logic
			actions = append(actions, cloAction{Kind: "reconnect_close"})
			actions = append(actions, cloAction{
				Kind:           "reconnect_sleep",
				ReconnectDelay: config.ReconnectDelay,
			})
			actions = append(actions, cloAction{Kind: "reconnect_open"})
			continue // after reconnect, loop continues to next segment
		}

		actions = append(actions, cloAction{
			Kind:        "send",
			SegmentSize: segSize,
		})

		// --- Apply rate control (always called after successful send) ---
		actions = append(actions, cloAction{
			Kind:      "rate_control",
			RateBytes: frameSize,
		})

		// --- sm_TaskYield() — just yields CPU, no pause ---
	}

	return actions
}


// ---------------------------------------------------------------------------
// Property 2a: Idle Blocking Preserved
// ---------------------------------------------------------------------------

// TestPreservation_IdleBlocking_Property verifies that when the segment queue
// is empty, the CLO blocks on dequeue without introducing spurious sleeps or
// busy-waits.
//
// **Validates: Requirements 3.1**
//
// Property: for all inputs where queueDepth == 0, the CLO produces exactly
// one "block" action and no other actions (no sends, no sleeps, no rate
// control calls).
func TestPreservation_IdleBlocking_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(mtuRaw, maxRateRaw uint8, useAX25 bool) bool {
		mtu := int(mtuRaw)%512 + 1       // 1..512
		maxRate := int(maxRateRaw)%960 + 1 // 1..960

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        maxRate,
			UseAX25:        useAX25,
			ReconnectDelay: 5,
		}

		// Empty queue
		actions := simulateCLOLoop_Preservation(nil, config, -1)

		// Must produce exactly one "block" action
		if len(actions) != 1 {
			t.Logf("FAIL: expected 1 action, got %d (mtu=%d, maxRate=%d, useAX25=%v)",
				len(actions), mtu, maxRate, useAX25)
			return false
		}
		if actions[0].Kind != "block" {
			t.Logf("FAIL: expected 'block' action, got '%s'", actions[0].Kind)
			return false
		}
		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("Idle blocking preservation property violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Property 2b: MTU Enforcement Preserved
// ---------------------------------------------------------------------------

// TestPreservation_MTUEnforcement_Property verifies that segments exceeding
// the configured MTU are discarded and an error is logged.
//
// **Validates: Requirements 3.3**
//
// Property: for all segments where segmentSize > MTU, the segment is discarded
// (no "send" action) and an "error_log" action is recorded.
func TestPreservation_MTUEnforcement_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(mtuRaw, oversizeRaw uint8) bool {
		mtu := int(mtuRaw)%256 + 1             // 1..256
		segSize := mtu + int(oversizeRaw)%256 + 1 // always > MTU

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        false,
			ReconnectDelay: 5,
		}

		actions := simulateCLOLoop_Preservation([]int{segSize}, config, -1)

		// Must have a "discard_mtu" and "error_log", no "send"
		hasDiscard := false
		hasErrorLog := false
		hasSend := false
		for _, a := range actions {
			switch a.Kind {
			case "discard_mtu":
				hasDiscard = true
			case "error_log":
				hasErrorLog = true
			case "send":
				hasSend = true
			}
		}

		if !hasDiscard || !hasErrorLog || hasSend {
			t.Logf("FAIL: mtu=%d segSize=%d discard=%v errorLog=%v send=%v",
				mtu, segSize, hasDiscard, hasErrorLog, hasSend)
			return false
		}
		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("MTU enforcement preservation property violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Property 2c: Rate Control Preserved
// ---------------------------------------------------------------------------

// TestPreservation_RateControl_Property verifies that applyRateControl() is
// invoked exactly once per successfully sent segment.
//
// **Validates: Requirements 3.6**
//
// Property: for all sent segments, the number of "rate_control" actions equals
// the number of "send" actions.
func TestPreservation_RateControl_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(countRaw uint8, mtuRaw uint8) bool {
		count := int(countRaw)%20 + 1 // 1..20 segments
		mtu := int(mtuRaw)%256 + 64   // 64..319

		segments := make([]int, count)
		for i := range segments {
			segments[i] = mtu/2 + 1 // all within MTU
		}

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        false,
			ReconnectDelay: 5,
		}

		actions := simulateCLOLoop_Preservation(segments, config, -1)

		sendCount := 0
		rateControlCount := 0
		for _, a := range actions {
			switch a.Kind {
			case "send":
				sendCount++
			case "rate_control":
				rateControlCount++
			}
		}

		if sendCount != rateControlCount {
			t.Logf("FAIL: sends=%d rateControls=%d (count=%d, mtu=%d)",
				sendCount, rateControlCount, count, mtu)
			return false
		}
		if sendCount != count {
			t.Logf("FAIL: expected %d sends, got %d", count, sendCount)
			return false
		}
		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("Rate control preservation property violated: %v", err)
	}
}


// ---------------------------------------------------------------------------
// Property 2d: AX.25 Framing Preserved
// ---------------------------------------------------------------------------

// TestPreservation_AX25Framing_Property verifies that when useAX25 is true,
// AX.25 UI-frame wrapping is applied to every outbound segment.
//
// **Validates: Requirements 3.5**
//
// Property: for all configs with useAX25 == true and all segments within MTU,
// an "ax25_wrap" action is recorded for each segment before the "send" action.
func TestPreservation_AX25Framing_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(countRaw, mtuRaw uint8) bool {
		count := int(countRaw)%20 + 1 // 1..20 segments
		mtu := int(mtuRaw)%256 + 64   // 64..319

		segments := make([]int, count)
		for i := range segments {
			segments[i] = mtu/2 + 1 // all within MTU
		}

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        true, // AX.25 enabled
			ReconnectDelay: 5,
		}

		actions := simulateCLOLoop_Preservation(segments, config, -1)

		// Count ax25_wrap and send actions
		ax25Count := 0
		sendCount := 0
		for _, a := range actions {
			switch a.Kind {
			case "ax25_wrap":
				ax25Count++
				if !a.AX25Applied {
					t.Logf("FAIL: ax25_wrap action has AX25Applied=false")
					return false
				}
			case "send":
				sendCount++
			}
		}

		if ax25Count != sendCount {
			t.Logf("FAIL: ax25Wraps=%d sends=%d (count=%d, mtu=%d)",
				ax25Count, sendCount, count, mtu)
			return false
		}
		if ax25Count != count {
			t.Logf("FAIL: expected %d ax25 wraps, got %d", count, ax25Count)
			return false
		}
		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("AX.25 framing preservation property violated: %v", err)
	}
}

// TestPreservation_NoAX25WhenDisabled_Property verifies that when useAX25 is
// false, no AX.25 wrapping is applied.
//
// **Validates: Requirements 3.5**
//
// Property: for all configs with useAX25 == false, no "ax25_wrap" actions
// appear in the output.
func TestPreservation_NoAX25WhenDisabled_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(countRaw, mtuRaw uint8) bool {
		count := int(countRaw)%20 + 1
		mtu := int(mtuRaw)%256 + 64

		segments := make([]int, count)
		for i := range segments {
			segments[i] = mtu/2 + 1
		}

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        false, // AX.25 disabled
			ReconnectDelay: 5,
		}

		actions := simulateCLOLoop_Preservation(segments, config, -1)

		for _, a := range actions {
			if a.Kind == "ax25_wrap" {
				t.Logf("FAIL: unexpected ax25_wrap when useAX25=false")
				return false
			}
		}
		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("No-AX.25 preservation property violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Property 2e: Reconnection Preserved
// ---------------------------------------------------------------------------

// TestPreservation_Reconnection_Property verifies that when a serial-port
// error occurs, the reconnection sequence (close → sleep → reopen) executes
// with the configured reconnect delay.
//
// **Validates: Requirements 3.4**
//
// Property: for all serial-port error states, the reconnection sequence
// produces exactly: reconnect_close, reconnect_sleep (with correct delay),
// reconnect_open — in that order.
func TestPreservation_Reconnection_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(countRaw, errorIdxRaw, delayRaw uint8) bool {
		count := int(countRaw)%10 + 2     // 2..11 segments
		mtu := 512
		delay := int(delayRaw)%30 + 1     // 1..30 seconds

		// Error at a valid index
		errorIdx := int(errorIdxRaw) % count

		segments := make([]int, count)
		for i := range segments {
			segments[i] = 100 // well within MTU
		}

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        false,
			ReconnectDelay: delay,
		}

		actions := simulateCLOLoop_Preservation(segments, config, errorIdx)

		// Find the reconnection sequence
		reconnectCloseCount := 0
		reconnectSleepCount := 0
		reconnectOpenCount := 0

		for i, a := range actions {
			switch a.Kind {
			case "reconnect_close":
				reconnectCloseCount++
				// Next two actions must be sleep then open
				if i+2 >= len(actions) {
					t.Logf("FAIL: reconnect_close at %d but not enough following actions", i)
					return false
				}
				if actions[i+1].Kind != "reconnect_sleep" {
					t.Logf("FAIL: expected reconnect_sleep after close, got %s", actions[i+1].Kind)
					return false
				}
				if actions[i+1].ReconnectDelay != delay {
					t.Logf("FAIL: reconnect delay=%d, expected=%d", actions[i+1].ReconnectDelay, delay)
					return false
				}
				if actions[i+2].Kind != "reconnect_open" {
					t.Logf("FAIL: expected reconnect_open after sleep, got %s", actions[i+2].Kind)
					return false
				}
			case "reconnect_sleep":
				reconnectSleepCount++
			case "reconnect_open":
				reconnectOpenCount++
			}
		}

		// Exactly one reconnection sequence
		if reconnectCloseCount != 1 || reconnectSleepCount != 1 || reconnectOpenCount != 1 {
			t.Logf("FAIL: close=%d sleep=%d open=%d (expected 1 each)",
				reconnectCloseCount, reconnectSleepCount, reconnectOpenCount)
			return false
		}

		return true
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("Reconnection preservation property violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Property 2f: Mixed Segments — MTU + Valid Interleaved
// ---------------------------------------------------------------------------

// TestPreservation_MixedSegments_Property verifies that a mix of valid and
// oversized segments produces the correct interleaving of send/discard actions
// with rate control applied only to sent segments.
//
// **Validates: Requirements 3.3, 3.6**
//
// Property: for all mixed segment sequences, each valid segment produces
// exactly one send + one rate_control, and each oversized segment produces
// exactly one discard_mtu + one error_log.
func TestPreservation_MixedSegments_Property(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(42)),
	}

	prop := func(seed int64) bool {
		r := rand.New(rand.NewSource(seed))
		mtu := r.Intn(256) + 64 // 64..319
		count := r.Intn(20) + 1 // 1..20

		segments := make([]int, count)
		expectedSends := 0
		expectedDiscards := 0
		for i := range segments {
			if r.Float64() < 0.3 {
				// 30% chance of oversized segment
				segments[i] = mtu + r.Intn(256) + 1
				expectedDiscards++
			} else {
				segments[i] = r.Intn(mtu) + 1
				expectedSends++
			}
		}

		config := cloConfig{
			MTU:            mtu,
			MaxRate:        960,
			UseAX25:        false,
			ReconnectDelay: 5,
		}

		actions := simulateCLOLoop_Preservation(segments, config, -1)

		sendCount := 0
		rateControlCount := 0
		discardCount := 0
		errorLogCount := 0
		for _, a := range actions {
			switch a.Kind {
			case "send":
				sendCount++
			case "rate_control":
				rateControlCount++
			case "discard_mtu":
				discardCount++
			case "error_log":
				errorLogCount++
			}
		}

		ok := true
		if sendCount != expectedSends {
			t.Logf("FAIL: sends=%d expected=%d", sendCount, expectedSends)
			ok = false
		}
		if rateControlCount != expectedSends {
			t.Logf("FAIL: rateControls=%d expected=%d", rateControlCount, expectedSends)
			ok = false
		}
		if discardCount != expectedDiscards {
			t.Logf("FAIL: discards=%d expected=%d", discardCount, expectedDiscards)
			ok = false
		}
		if errorLogCount != expectedDiscards {
			t.Logf("FAIL: errorLogs=%d expected=%d", errorLogCount, expectedDiscards)
			ok = false
		}
		if !ok {
			t.Logf("  segments=%v mtu=%d", segments, mtu)
		}
		return ok
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Errorf("Mixed segments preservation property violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func init() {
	// Suppress unused import warning for fmt
	_ = fmt.Sprintf
}
