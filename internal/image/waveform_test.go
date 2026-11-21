package image

import (
	"image/color"
	"image/color/palette"
	"math"
	"testing"
)

func TestNewWaveformImage(t *testing.T) {
	_, err := NewWaveformImage(0, 64, "#000000", "#00ff00", 1)
	if err == nil {
		t.Errorf("NewWaveformImage(0, 64, \"#000000\", \"#00ff00\", 1) = _, nil; want error")
	}

	_, err = NewWaveformImage(128, 0, "#000000", "#00ff00", 1)
	if err == nil {
		t.Errorf("NewWaveformImage(128, 0, \"#000000\", \"#00ff00\", 1) = _, nil; want error")
	}

	wimg, _ := NewWaveformImage(128, 64, "#000000", "#00ff00", 1)
	if wimg == nil {
		t.Errorf("NewWaveformImage(128, 64, \"#000000\", \"#00ff00\", 1) = nil, _; want WaveformImage")
		t.FailNow()
	}

	got := wimg.context.Width()
	if got != 128 {
		t.Errorf("NewWaveformImage(128, 64, \"#000000\", \"#00ff00\", 1) = WaveformImage{context: {Width: %d}}, _; want WaveformImage{context: {Width: 128}}", got)
	}

	got = wimg.context.Height()
	if got != 64 {
		t.Errorf("NewWaveformImage(128, 64, \"#000000\", \"#00ff00\", 1) = WaveformImage{context: {Height: %d}}, _; want WaveformImage{context: {Height: 64}}", got)
	}
}

func TestDrawBackground(t *testing.T) {
	wimg, err := NewWaveformImage(128, 64, "#000000", "#00ff00", 1)
	if wimg == nil {
		t.Fatal(err)
	}

	wimg.DrawBackground()
	img := wimg.Image()
	bgColor := color.RGBA{0x00, 0x00, 0x00, 0xff}

	for x := uint(0); x < wimg.Width; x++ {
		for y := uint(0); y < wimg.Height; y++ {
			got := img.At(int(x), int(y))
			if got != bgColor {
				r, g, b, a := got.RGBA()
				t.Errorf("DrawBackground() draws (0x%02x, 0x%02x, 0x%02x, 0x%02x) at (%d, %d); want (0x00, 0x00, 0x00, 0xff)", r, g, b, a, x, y)
			}
		}
	}
}

func TestDrawPolyLine(t *testing.T) {
	// wimg, err := NewWaveformImage(128, 64, "#000000", "#00ff00", 1)
	// if wimg == nil {
	// 	t.Fatal(err)
	// }
	//
	// left, top := 0.0, 0.0
	// right, bottom := float64(wimg.Width)-1, float64(wimg.Height)-1
	//
	// img := wimg.Image()
	// bgColor := color.RGBA{0x00, 0x00, 0x00, 0x00}
	//
	// for x := uint(0); x < wimg.Width; x++ {
	// 	for y := uint(0); y < wimg.Height; y++ {
	// 		got := img.At(int(x), int(y))
	// 		xf, yf := float64(x), float64(y)
	//
	// 		if (xf == left || xf == right || yf == top || yf == bottom) && got == bgColor {
	// 			r, g, b, a := got.RGBA()
	// 			t.Errorf("DrawPolyLine() draws (0x%02x, 0x%02x, 0x%02x, 0x%02x) at (%d, %d); want (0x00, 0xff, 0x00, 0xff)", r, g, b, a, x, y)
	// 		} else if xf != left && xf != right && yf != top && yf != bottom && got != bgColor {
	// 			r, g, b, a := got.RGBA()
	// 			t.Errorf("DrawPolyLine() draws (0x%02x, 0x%02x, 0x%02x, 0x%02x) at (%d, %d); want (0x00, 0x00, 0x00, 0x00)", r, g, b, a, x, y)
	// 		}
	// 	}
	// }
}

func TestConvertToPaletted(t *testing.T) {
	wimg, err := NewWaveformImage(128, 64, "#ff00ff", "#00ff00", 1)
	if wimg == nil {
		t.Fatal(err)
	}

	wimg.DrawBackground()
	pimg := wimg.ConvertToPaletted(palette.Plan9)
	bgColor := color.RGBA{0xff, 0x00, 0xff, 0xff}

	for x := uint(0); x < wimg.Width; x++ {
		for y := uint(0); y < wimg.Height; y++ {
			got := pimg.At(int(x), int(y))
			if got != bgColor {
				r, g, b, a := got.RGBA()
				t.Errorf("ConvertToPaletted(palette.Plan9) draws (0x%02x, 0x%02x, 0x%02x, 0x%02x) at (%d, %d); want (0xff, 0x00, 0xff, 0xff)", r, g, b, a, x, y)
			}
		}
	}
}

func TestDrawPeakRMS(t *testing.T) {
	wimg, err := NewWaveformImage(64, 48, "#000000", "#00ff00", 2)
	if err != nil {
		t.Fatal(err)
	}

	wimg.DrawPeakRMS(-1, 1, 0.5)
	img := wimg.Image()
	bg := color.RGBA{0x00, 0x00, 0x00, 0xff}
	fg := color.RGBA{0x00, 0xff, 0x00, 0xff}

	// the peak bar spans the full height at the horizontal center
	for _, y := range []int{2, 24, 45} {
		if got := img.At(32, y); got != fg {
			t.Errorf("DrawPeakRMS() draws %v at (32, %d); want %v", got, y, fg)
		}
	}

	// the RMS indicators are horizontal ticks at +-0.5 (y = 12 and y = 36)
	for _, p := range [][2]int{{28, 12}, {36, 12}, {28, 36}, {36, 36}} {
		if got := img.At(p[0], p[1]); got != fg {
			t.Errorf("DrawPeakRMS() draws %v at (%d, %d); want %v", got, p[0], p[1], fg)
		}
	}

	// corners and off-center pixels stay background
	for _, p := range [][2]int{{0, 0}, {63, 0}, {0, 47}, {63, 47}, {8, 12}, {56, 36}} {
		if got := img.At(p[0], p[1]); got != bg {
			t.Errorf("DrawPeakRMS() draws %v at (%d, %d); want %v", got, p[0], p[1], bg)
		}
	}
}

func TestDrawPeakRMSSilence(t *testing.T) {
	wimg, err := NewWaveformImage(64, 48, "#000000", "#00ff00", 2)
	if err != nil {
		t.Fatal(err)
	}

	wimg.DrawPeakRMS(0, 0, 0)
	img := wimg.Image()
	bg := color.RGBA{0x00, 0x00, 0x00, 0xff}
	fg := color.RGBA{0x00, 0xff, 0x00, 0xff}

	// silence collapses to a mark at the vertical center
	if got := img.At(32, 24); got != fg {
		t.Errorf("DrawPeakRMS() draws %v at (32, 24); want %v", got, fg)
	}
	for _, p := range [][2]int{{0, 0}, {63, 47}, {32, 0}, {32, 47}} {
		if got := img.At(p[0], p[1]); got != bg {
			t.Errorf("DrawPeakRMS() draws %v at (%d, %d); want %v", got, p[0], p[1], bg)
		}
	}
}

func TestDrawPeakRMSNonFinite(t *testing.T) {
	wimg, err := NewWaveformImage(64, 48, "#000000", "#00ff00", 2)
	if err != nil {
		t.Fatal(err)
	}

	// must not panic and must keep the frame drawable
	wimg.DrawPeakRMS(math.NaN(), math.Inf(1), math.NaN())
	img := wimg.Image()
	bg := color.RGBA{0x00, 0x00, 0x00, 0xff}
	if got := img.At(0, 0); got != bg {
		t.Errorf("DrawPeakRMS() draws %v at (0, 0); want %v", got, bg)
	}
}
