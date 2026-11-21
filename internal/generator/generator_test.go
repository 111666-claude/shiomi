package generator

import (
	"bytes"
	"image/gif"
	"os"
	"testing"
)

var testSine = "../../test/data/sine-44.1kHz-16bit-1kHz-3s.wav"
var testImpulse = "../../test/data/impulse-44.1kHz-16bit-3s.wav"
var testSilence = "../../test/data/silence-44.1kHz-16bit-3s.wav"

func generate(t *testing.T, path string, mode Mode, rate uint) *gif.GIF {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	config := &Config{
		Mode:            mode,
		Frequency:       1000,
		WindowSize:      3,
		SubsamplingRate: rate,
		LineWidth:       2.0,
		Width:           64,
		Height:          48,
		BgColor:         "#000000",
		FgColor:         "#ffffff",
		Delay:           7,
	}

	buf := new(bytes.Buffer)
	GenerateWaveformGIF(f, buf, config)

	g, err := gif.DecodeAll(buf)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func checkGIF(t *testing.T, g *gif.GIF, wantFrames int) {
	t.Helper()

	if len(g.Image) != wantFrames {
		t.Errorf("generated GIF has %d frames; want %d", len(g.Image), wantFrames)
	}
	if g.Config.Width != 64 || g.Config.Height != 48 {
		t.Errorf("generated GIF is %dx%d; want 64x48", g.Config.Width, g.Config.Height)
	}
	for i, d := range g.Delay {
		if d != 7 {
			t.Errorf("frame %d has delay %d; want 7", i, d)
		}
	}
	for i, img := range g.Image {
		b := img.Bounds()
		if b.Dx() != 64 || b.Dy() != 48 {
			t.Errorf("frame %d is %dx%d; want 64x48", i, b.Dx(), b.Dy())
		}
	}
}

func TestGenerateWaveformGIFDefaultMode(t *testing.T) {
	// 132496 samples / 132 samples per window = 1003 full windows;
	// subsampling by 10 yields windows 0, 10, ..., 1000 -> 101 frames
	checkGIF(t, generate(t, testSine, ModeWaveform, 10), 101)

	// without subsampling the trailing partial window is dropped
	checkGIF(t, generate(t, testSine, ModeWaveform, 1), 1003)
}

func TestGeneratePeakRMSGIF(t *testing.T) {
	// peak-rms mode includes the trailing partial window:
	// ceil(132496 / 132) = 1004 windows -> frames 0, 10, ..., 1000
	checkGIF(t, generate(t, testSine, ModePeakRMS, 10), 101)
	checkGIF(t, generate(t, testSine, ModePeakRMS, 1), 1004)

	// 132296 samples -> ceil(132296 / 132) = 1003 windows
	checkGIF(t, generate(t, testImpulse, ModePeakRMS, 10), 101)
	checkGIF(t, generate(t, testSilence, ModePeakRMS, 10), 101)
	checkGIF(t, generate(t, testSilence, ModePeakRMS, 1), 1003)
}
