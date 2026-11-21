package audio

import "math"

// WindowStats describes the energy of a single audio window in the bipolar
// domain: normalized samples in [0, 1] are mapped to [-1, 1].
type WindowStats struct {
	Min float64 // most negative (centered) sample in the window
	Max float64 // most positive (centered) sample in the window
	RMS float64 // largest per-channel root-mean-square energy
}

// ComputeWindowStats calculates peak (min/max) and RMS values for a window of
// normalized samples. Samples are de-interleaved into numChannels channels so
// that energy from one channel never leaks into the measurement of another;
// the reported peaks bound all channels and the RMS is the maximum of the
// per-channel RMS values. Non-finite samples (NaN, ±Inf) are ignored so the
// result is always finite. An empty (or fully non-finite) window yields zero
// stats, which is also the result for digital silence.
func ComputeWindowStats(samples []float32, numChannels uint) WindowStats {
	if numChannels == 0 {
		numChannels = 1
	}
	n := int(numChannels)

	mins := make([]float64, n)
	maxs := make([]float64, n)
	sumsq := make([]float64, n)
	counts := make([]int, n)
	for ch := 0; ch < n; ch++ {
		mins[ch] = math.Inf(1)
		maxs[ch] = math.Inf(-1)
	}

	for i, s := range samples {
		v := float64(s)*2 - 1
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		ch := i % n
		if v < mins[ch] {
			mins[ch] = v
		}
		if v > maxs[ch] {
			maxs[ch] = v
		}
		sumsq[ch] += v * v
		counts[ch]++
	}

	stats := WindowStats{Min: 0, Max: 0, RMS: 0}
	first := true
	for ch := 0; ch < n; ch++ {
		if counts[ch] == 0 {
			continue
		}
		if first || mins[ch] < stats.Min {
			stats.Min = mins[ch]
		}
		if first || maxs[ch] > stats.Max {
			stats.Max = maxs[ch]
		}
		first = false
		if rms := math.Sqrt(sumsq[ch] / float64(counts[ch])); rms > stats.RMS {
			stats.RMS = rms
		}
	}

	return stats
}
