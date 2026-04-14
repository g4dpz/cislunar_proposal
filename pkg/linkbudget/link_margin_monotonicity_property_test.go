package linkbudget

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 26: Link Margin Monotonically Decreasing with Distance
// For any two distances d1 < d2 with identical transmit parameters, the computed
// link margin at d1 SHALL be strictly greater than the link margin at d2.
//
// **Validates: Requirement 18.3**
func TestProperty_LinkMarginMonotonicallyDecreasingWithDistance(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for distance pairs (d1 < d2)
	// Generate two distances where d2 > d1
	genDistance1 := gen.IntRange(1000, 500000)
	genDistance2 := gen.IntRange(1000, 500000)

	// Generator for frequency (100 MHz to 10 GHz)
	genFrequency := gen.Float64Range(100e6, 10e9)

	// Generator for TX power (0-50 dBm)
	genTxPower := gen.Float64Range(0, 50)

	// Generator for antenna gains (-10 to 50 dBi)
	genAntennaGain := gen.Float64Range(-10, 50)

	// Generator for system loss (0-10 dB)
	genSystemLoss := gen.Float64Range(0, 10)

	// Generator for data rate (100 bps to 100 kbps)
	genDataRate := gen.Float64Range(100, 100000)

	// Generator for required Eb/N0 (0-15 dB)
	genRequiredEbN0 := gen.Float64Range(0, 15)

	properties.Property("link margin decreases with increasing distance", prop.ForAll(
		func(d1 int, d2 int, frequency float64, txPower float64,
			txGain float64, rxGain float64, systemLoss float64,
			dataRate float64, requiredEbN0 float64) bool {

			// Ensure d1 < d2
			if d1 >= d2 {
				d1, d2 = d2-1000, d1+1000
			}
			if d1 < 1000 {
				d1 = 1000
			}

			// Create identical link budget parameters except for distance
			params1 := LinkBudgetParams{
				TxPowerDBm:       txPower,
				TxAntennaGainDBi: txGain,
				RxAntennaGainDBi: rxGain,
				FrequencyHz:      frequency,
				DistanceM:        float64(d1),
				SystemLossDB:     systemLoss,
				DataRateBps:      dataRate,
				RequiredEbN0DB:   requiredEbN0,
				NoiseDensityDBmHz: -174.0,
			}

			params2 := params1
			params2.DistanceM = float64(d2)

			// Compute link budgets
			result1, err1 := ComputeLinkBudget(params1)
			result2, err2 := ComputeLinkBudget(params2)

			if err1 != nil || err2 != nil {
				t.Logf("ComputeLinkBudget failed: err1=%v, err2=%v", err1, err2)
				return false
			}

			// Verify link margin at d1 > link margin at d2
			if result1.LinkMargin <= result2.LinkMargin {
				t.Logf("Link margin did not decrease with distance: d1=%d m margin=%.2f dB, d2=%d m margin=%.2f dB",
					d1, result1.LinkMargin, d2, result2.LinkMargin)
				return false
			}

			return true
		},
		genDistance1,
		genDistance2,
		genFrequency,
		genTxPower,
		genAntennaGain,
		genAntennaGain,
		genSystemLoss,
		genDataRate,
		genRequiredEbN0,
	))

	properties.TestingRun(t)
}

// Property 26 (variant): FSPL increases with distance
// For any two distances d1 < d2 with the same frequency, FSPL at d1 SHALL be
// less than FSPL at d2
//
// **Validates: Requirement 18.3**
func TestProperty_FSPLIncreasesWithDistance(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genDistance1 := gen.IntRange(1000, 500000)
	genDistance2 := gen.IntRange(1000, 500000)
	genFrequency := gen.Float64Range(100e6, 10e9)

	properties.Property("FSPL increases with distance", prop.ForAll(
		func(d1 int, d2 int, frequency float64) bool {
			// Ensure d1 < d2
			if d1 >= d2 {
				d1, d2 = d2-1000, d1+1000
			}
			if d1 < 1000 {
				d1 = 1000
			}

			// Create minimal params to compute FSPL
			params1 := LinkBudgetParams{
				TxPowerDBm:       0,
				TxAntennaGainDBi: 0,
				RxAntennaGainDBi: 0,
				FrequencyHz:      frequency,
				DistanceM:        float64(d1),
				SystemLossDB:     0,
				DataRateBps:      1000,
				RequiredEbN0DB:   0,
				NoiseDensityDBmHz: -174.0,
			}

			params2 := params1
			params2.DistanceM = float64(d2)

			result1, err1 := ComputeLinkBudget(params1)
			result2, err2 := ComputeLinkBudget(params2)

			if err1 != nil || err2 != nil {
				return false
			}

			// FSPL should increase with distance
			if result1.FSPL >= result2.FSPL {
				t.Logf("FSPL did not increase with distance: d1=%d m FSPL=%.2f dB, d2=%d m FSPL=%.2f dB",
					d1, result1.FSPL, d2, result2.FSPL)
				return false
			}

			return true
		},
		genDistance1,
		genDistance2,
		genFrequency,
	))

	properties.TestingRun(t)
}

// Property 26 (variant): Link margin decreases by approximately 6 dB per doubling of distance
// For any distance d, the link margin at 2*d should be approximately 6 dB less
// (due to 20*log10(2) ≈ 6 dB increase in FSPL)
//
// **Validates: Requirement 18.3**
func TestProperty_LinkMarginDecreasesBy6dBPerDoubling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genDistance := gen.IntRange(10000, 500000)
	genFrequency := gen.Float64Range(100e6, 10e9)
	genTxPower := gen.Float64Range(0, 50)
	genAntennaGain := gen.Float64Range(-10, 50)
	genSystemLoss := gen.Float64Range(0, 10)
	genDataRate := gen.Float64Range(100, 100000)
	genRequiredEbN0 := gen.Float64Range(0, 15)

	properties.Property("link margin decreases by ~6 dB per doubling of distance", prop.ForAll(
		func(distance int, frequency float64, txPower float64,
			txGain float64, rxGain float64, systemLoss float64,
			dataRate float64, requiredEbN0 float64) bool {

			d1 := float64(distance)
			d2 := d1 * 2.0

			params1 := LinkBudgetParams{
				TxPowerDBm:       txPower,
				TxAntennaGainDBi: txGain,
				RxAntennaGainDBi: rxGain,
				FrequencyHz:      frequency,
				DistanceM:        d1,
				SystemLossDB:     systemLoss,
				DataRateBps:      dataRate,
				RequiredEbN0DB:   requiredEbN0,
				NoiseDensityDBmHz: -174.0,
			}

			params2 := params1
			params2.DistanceM = d2

			result1, err1 := ComputeLinkBudget(params1)
			result2, err2 := ComputeLinkBudget(params2)

			if err1 != nil || err2 != nil {
				return false
			}

			// Margin difference should be approximately 6 dB (20*log10(2))
			marginDiff := result1.LinkMargin - result2.LinkMargin
			expected := 6.0206 // 20*log10(2)

			// Allow 0.1 dB tolerance for floating point
			if marginDiff < expected-0.1 || marginDiff > expected+0.1 {
				t.Logf("Margin difference not ~6 dB: got %.2f dB (d1=%.0f m, d2=%.0f m)",
					marginDiff, d1, d2)
				return false
			}

			return true
		},
		genDistance,
		genFrequency,
		genTxPower,
		genAntennaGain,
		genAntennaGain,
		genSystemLoss,
		genDataRate,
		genRequiredEbN0,
	))

	properties.TestingRun(t)
}

// Property 26 (variant): Link margin is strictly monotonic across multiple distances
// For any sequence of increasing distances, link margins SHALL be strictly decreasing
//
// **Validates: Requirement 18.3**
func TestProperty_LinkMarginStrictlyMonotonicAcrossSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Generator for number of distances to test
	genNumDistances := gen.IntRange(2, 10)
	genBaseDistance := gen.IntRange(10000, 100000)
	genFrequency := gen.Float64Range(100e6, 10e9)

	properties.Property("link margin strictly decreases across distance sequence", prop.ForAll(
		func(numDistances int, baseDistance int, frequency float64) bool {
			// Generate strictly increasing distance sequence
			distances := make([]int, numDistances)
			distances[0] = baseDistance
			for i := 1; i < numDistances; i++ {
				distances[i] = distances[i-1] + 10000 // Increment by 10 km
			}
			baseParams := LinkBudgetParams{
				TxPowerDBm:       30.0,
				TxAntennaGainDBi: 0.0,
				RxAntennaGainDBi: 10.0,
				FrequencyHz:      frequency,
				DistanceM:        0, // will be set per distance
				SystemLossDB:     2.0,
				DataRateBps:      9600.0,
				RequiredEbN0DB:   10.0,
				NoiseDensityDBmHz: -174.0,
			}

			var prevMargin float64
			for i, distance := range distances {
				params := baseParams
				params.DistanceM = float64(distance)

				result, err := ComputeLinkBudget(params)
				if err != nil {
					return false
				}

				if i > 0 {
					// Margin should strictly decrease
					if result.LinkMargin >= prevMargin {
						t.Logf("Margin did not decrease: d[%d]=%.0f m margin=%.2f dB, d[%d]=%.0f m margin=%.2f dB",
							i-1, float64(distances[i-1]), prevMargin, i, float64(distance), result.LinkMargin)
						return false
					}
				}

				prevMargin = result.LinkMargin
			}

			return true
		},
		genNumDistances,
		genBaseDistance,
		genFrequency,
	))

	properties.TestingRun(t)
}
