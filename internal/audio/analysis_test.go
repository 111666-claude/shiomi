package audio

import (
	"math"
	"testing"
)

const statsTolerance = 1e-6

func checkStats(t *testing.T, got WindowStats, min float64, max float64, rms float64, tol float64) {
	t.Helper()
	if math.IsNaN(got.Min) || math.IsNaN(got.Max) || math.IsNaN(got.RMS) ||
		math.IsInf(got.Min, 0) || math.IsInf(got.Max, 0) || math.IsInf(got.RMS, 0) {
		t.Fatalf("ComputeWindowStats() = %+v; want finite stats", got)
	}
	if math.Abs(got.Min-min) > tol {
		t.Errorf("ComputeWindowStats().Min = %f; want %f", got.Min, min)
	}
	if math.Abs(got.Max-max) > tol {
		t.Errorf("ComputeWindowStats().Max = %f; want %f", got.Max, max)
	}
	if math.Abs(got.RMS-rms) > tol {
		t.Errorf("ComputeWindowStats().RMS = %f; want %f", got.RMS, rms)
	}
}

func TestComputeWindowStatsSilence(t *testing.T) {
	// digital silence is 0.5 in the normalized [0, 1] domain
	samples := []float32{0.5, 0.5, 0.5, 0.5}
	checkStats(t, ComputeWindowStats(samples, 1), 0, 0, 0, statsTolerance)
}

func TestComputeWindowStatsEmpty(t *testing.T) {
	checkStats(t, ComputeWindowStats([]float32{}, 1), 0, 0, 0, 0)
	checkStats(t, ComputeWindowStats(nil, 1), 0, 0, 0, 0)
}

func TestComputeWindowStatsConstant(t *testing.T) {
	samples := []float32{1, 1, 1}
	checkStats(t, ComputeWindowStats(samples, 1), 1, 1, 1, statsTolerance)

	samples = []float32{0, 0, 0}
	checkStats(t, ComputeWindowStats(samples, 1), -1, -1, 1, statsTolerance)
}

func TestComputeWindowStatsClipping(t *testing.T) {
	// alternating full-scale samples: peaks at both rails, RMS of 1
	samples := []float32{1, 0, 1, 0}
	checkStats(t, ComputeWindowStats(samples, 1), -1, 1, 1, statsTolerance)
}

func TestComputeWindowStatsImpulse(t *testing.T) {
	samples := []float32{0.5, 0.5, 1, 0.5}
	got := ComputeWindowStats(samples, 1)
	checkStats(t, got, 0, 1, 0.5, statsTolerance)
}

func TestComputeWindowStatsNonFinite(t *testing.T) {
	samples := []float32{0.5, float32(math.NaN()), 0.5, float32(math.Inf(1)), 0.5, float32(math.Inf(-1))}
	checkStats(t, ComputeWindowStats(samples, 1), 0, 0, 0, statsTolerance)

	// a window of only non-finite samples behaves like an empty window
	samples = []float32{float32(math.NaN()), float32(math.Inf(1))}
	checkStats(t, ComputeWindowStats(samples, 1), 0, 0, 0, 0)
}

func TestComputeWindowStatsChannels(t *testing.T) {
	// stereo interleaved [L, R]: L is a constant full-scale signal, R is
	// silence. The RMS of L must not leak into (or be diluted by) R.
	samples := []float32{1, 0.5, 1, 0.5, 1, 0.5, 1, 0.5}
	checkStats(t, ComputeWindowStats(samples, 2), 0, 1, 1, statsTolerance)

	// swapped: energy on R only
	samples = []float32{0.5, 0, 0.5, 0, 0.5, 0, 0.5, 0}
	checkStats(t, ComputeWindowStats(samples, 2), -1, 0, 1, statsTolerance)

	// zero channels falls back to mono
	checkStats(t, ComputeWindowStats(samples, 0), -1, 0, math.Sqrt(0.5), statsTolerance)
}
