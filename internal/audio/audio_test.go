package audio

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"testing"
)

var testTone = "../../test/data/sine-44.1kHz-16bit-1kHz-3s.wav"
var testImpulse = "../../test/data/impulse-44.1kHz-16bit-3s.wav"
var testSilence = "../../test/data/silence-44.1kHz-16bit-3s.wav"

// newTestWav builds an in-memory 16-bit PCM WAV from the given samples.
func newTestWav(samples []int16, sampleRate uint32, numChannels uint16) *bytes.Buffer {
	data := new(bytes.Buffer)
	for _, s := range samples {
		binary.Write(data, binary.LittleEndian, s)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+data.Len()))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, numChannels)
	binary.Write(buf, binary.LittleEndian, sampleRate)
	binary.Write(buf, binary.LittleEndian, sampleRate*uint32(numChannels)*2)
	binary.Write(buf, binary.LittleEndian, numChannels*2)
	binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(data.Len()))
	buf.Write(data.Bytes())
	return buf
}

// makeRamp returns n samples with distinct values 1..n.
func makeRamp(n int) []int16 {
	samples := make([]int16, n)
	for i := range samples {
		samples[i] = int16(i + 1)
	}
	return samples
}

func normalized(v int16) float32 {
	return (float32(v) - math.MinInt16) / (math.MaxInt16 - math.MinInt16)
}

func TestNewAudio(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = NewAudio(f, 0, 3, 10)
	if err == nil {
		t.Errorf("NewAudio(f, 0, 3, 10) = _, nil; want error")
	}

	_, err = NewAudio(f, 1000, 0, 10)
	if err == nil {
		t.Errorf("NewAudio(f, 1000, 0, 10) = _, nil; want error")
	}

	_, err = NewAudio(f, 1000, 3, 0)
	if err == nil {
		t.Errorf("NewAudio(f, 1000, 3, 0) = _, nil; want error")
	}

	a, _ := NewAudio(f, 1000, 3, 10)
	if a == nil {
		t.Errorf("NewAudio(f, 1000, 3, 10) = nil, _; want Audio")
	}
}

func TestNumSamplesPerCycle(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 10)
	if err != nil {
		t.Fatal(err)
	}

	got := a.NumSamplesPerCycle()
	if got != 44.1 {
		t.Errorf("NumSamplesPerCycle() = %f; want 44.1", got)
	}
}

func TestNumSamplesPerWindow(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 10)
	if err != nil {
		t.Fatal(err)
	}

	got := a.NumSamplesPerWindow()
	if got != 132.3 {
		t.Errorf("NumSamplesPerWindow() = %f; want 132.3", got)
	}
}

func TestRoundedNumSamplesPerWindow(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 10)
	if err != nil {
		t.Fatal(err)
	}

	got := a.RoundedNumSamplesPerWindow()
	if got != 132 {
		t.Errorf("RoundedNumSamplesPerWindow() = %d; want 132", got)
	}
}

func TestNumWindow(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 10)
	if err != nil {
		t.Fatal(err)
	}

	got := a.NumWindow()
	if got != 1003 {
		t.Errorf("NumWindow() = %d; want 1003", got)
	}
}

func TestNumWindowWithTail(t *testing.T) {
	// 104 samples, 24 samples per window -> 4 full windows + tail of 8
	a, err := NewAudio(newTestWav(makeRamp(104), 8000, 1), 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	if got := a.NumWindow(); got != 4 {
		t.Errorf("NumWindow() = %d; want 4", got)
	}
	if got := a.NumWindowWithTail(); got != 5 {
		t.Errorf("NumWindowWithTail() = %d; want 5", got)
	}
	if got := a.NumSampleValuesPerWindow(); got != 24 {
		t.Errorf("NumSampleValuesPerWindow() = %d; want 24", got)
	}
}

func TestNumWindowWithTailEmpty(t *testing.T) {
	a, err := NewAudio(newTestWav([]int16{}, 8000, 1), 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	if got := a.NumWindowWithTail(); got != 0 {
		t.Errorf("NumWindowWithTail() = %d; want 0", got)
	}

	count := 0
	for range a.ReadWindows() {
		count++
	}
	if count != 0 {
		t.Errorf("ReadWindows() yielded %d frames; want 0", count)
	}
}

func TestNewAudioDegenerateWindow(t *testing.T) {
	// 8000 Hz / 20000 Hz * 1 cycle rounds to a zero-sample window
	_, err := NewAudio(newTestWav(makeRamp(16), 8000, 1), 20000, 1, 1)
	if err == nil {
		t.Errorf("NewAudio(f, 20000, 1, 1) = _, nil; want error")
	}
}

func TestReadWindowsCorrespondence(t *testing.T) {
	samples := makeRamp(104)
	a, err := NewAudio(newTestWav(samples, 8000, 1), 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	// frames must cover the sample ranges [0,24) [24,48) [48,72) [72,96) [96,104)
	wantLens := []int{24, 24, 24, 24, 8}
	offset := 0
	i := 0
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		if i >= len(wantLens) {
			t.Fatalf("ReadWindows() yielded more than %d frames", len(wantLens))
		}
		if len(data.Samples) != wantLens[i] {
			t.Errorf("frame %d has %d samples; want %d", i, len(data.Samples), wantLens[i])
		}
		for j, s := range data.Samples {
			if want := normalized(samples[offset+j]); s != want {
				t.Errorf("frame %d sample %d = %f; want %f", i, j, s, want)
			}
		}
		offset += len(data.Samples)
		i++
	}
	if i != len(wantLens) {
		t.Errorf("ReadWindows() yielded %d frames; want %d", i, len(wantLens))
	}
	if offset != len(samples) {
		t.Errorf("ReadWindows() covered %d samples; want %d", offset, len(samples))
	}
}

func TestReadWindowsSubsampling(t *testing.T) {
	samples := makeRamp(104)
	a, err := NewAudio(newTestWav(samples, 8000, 1), 1000, 3, 2)
	if err != nil {
		t.Fatal(err)
	}

	// windows 0, 2, 4 -> sample ranges [0,24) [48,72) [96,104)
	wantLens := []int{24, 24, 8}
	wantOffsets := []int{0, 48, 96}
	i := 0
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		if i >= len(wantLens) {
			t.Fatalf("ReadWindows() yielded more than %d frames", len(wantLens))
		}
		if len(data.Samples) != wantLens[i] {
			t.Errorf("frame %d has %d samples; want %d", i, len(data.Samples), wantLens[i])
		}
		for j, s := range data.Samples {
			if want := normalized(samples[wantOffsets[i]+j]); s != want {
				t.Errorf("frame %d sample %d = %f; want %f", i, j, s, want)
			}
		}
		i++
	}
	if i != len(wantLens) {
		t.Errorf("ReadWindows() yielded %d frames; want %d", i, len(wantLens))
	}
}

func TestReadWindowsStereo(t *testing.T) {
	// 96 interleaved values (48 stereo frames); a window spans 24 samples
	// per channel = 48 interleaved values
	samples := makeRamp(96)
	a, err := NewAudio(newTestWav(samples, 8000, 2), 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.NumChannels(); got != 2 {
		t.Errorf("NumChannels() = %d; want 2", got)
	}

	wantLens := []int{48, 48}
	offset := 0
	i := 0
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		if i >= len(wantLens) {
			t.Fatalf("ReadWindows() yielded more than %d frames", len(wantLens))
		}
		if len(data.Samples) != wantLens[i] {
			t.Errorf("frame %d has %d samples; want %d", i, len(data.Samples), wantLens[i])
		}
		for j, s := range data.Samples {
			if want := normalized(samples[offset+j]); s != want {
				t.Errorf("frame %d sample %d = %f; want %f", i, j, s, want)
			}
		}
		offset += len(data.Samples)
		i++
	}
	if i != len(wantLens) {
		t.Errorf("ReadWindows() yielded %d frames; want %d", i, len(wantLens))
	}
}

func TestReadWindowsSampleRates(t *testing.T) {
	// the same signal at 44.1kHz: 132 samples per window, 132296 samples
	// -> 1002 full windows + tail of 32
	f, err := os.Open(testSilence)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.NumWindowWithTail(); got != 1003 {
		t.Errorf("NumWindowWithTail() = %d; want 1003", got)
	}

	var lastLen, count int
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		lastLen = len(data.Samples)
		count++
	}
	if count != 1003 {
		t.Errorf("ReadWindows() yielded %d frames; want 1003", count)
	}
	if lastLen != 32 {
		t.Errorf("last window has %d samples; want 32", lastLen)
	}
}

func TestReadWindowsImpulseStats(t *testing.T) {
	f, err := os.Open(testImpulse)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	stats := make([]WindowStats, 0, a.NumWindowWithTail())
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		stats = append(stats, ComputeWindowStats(data.Samples, a.NumChannels()))
	}

	// +full-scale impulse at sample 0 -> window 0
	checkStats(t, stats[0], 0, 1, 1/math.Sqrt(132), 1e-3)

	// -full-scale impulse at sample 66150 -> window 501
	checkStats(t, stats[501], -1, 0, 1/math.Sqrt(132), 1e-3)

	// a window without impulses is silent
	checkStats(t, stats[100], 0, 0, 0, 1e-4)
}

func TestReadWindowsSilenceStats(t *testing.T) {
	f, err := os.Open(testSilence)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		checkStats(t, ComputeWindowStats(data.Samples, a.NumChannels()), 0, 0, 0, 1e-4)
		count++
	}
	if count != 1003 {
		t.Errorf("ReadWindows() yielded %d frames; want 1003", count)
	}
}

func TestReadWindowsSineStats(t *testing.T) {
	f, err := os.Open(testTone)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	a, err := NewAudio(f, 1000, 3, 1)
	if err != nil {
		t.Fatal(err)
	}

	stats := make([]WindowStats, 0, a.NumWindowWithTail())
	for data := range a.ReadWindows() {
		if data.Error != nil {
			t.Fatal(data.Error)
		}
		stats = append(stats, ComputeWindowStats(data.Samples, a.NumChannels()))
	}

	// 1kHz sine with amplitude ~0.69: peaks near +-0.69, RMS near 0.49
	checkStats(t, stats[500], -0.6919, 0.6919, 0.4892, 1e-2)
}
