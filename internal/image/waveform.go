package image

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

type WaveformImage struct {
	Width     uint
	Height    uint
	BgColor   string
	FgColor   string
	LineWidth float64
	context   *gg.Context
}

type Point = gg.Point

func NewWaveformImage(width uint, height uint, bgColor string, fgColor string, lineWidth float64) (*WaveformImage, error) {
	if width <= 0 || height <= 0 {
		return nil, errors.New("width and height must be > 0")
	}

	context := gg.NewContext(int(width), int(height))
	return &WaveformImage{Width: width, Height: height, BgColor: bgColor, FgColor: fgColor, LineWidth: lineWidth, context: context}, nil
}

func (wimg *WaveformImage) DrawBackground() {
	wimg.context.SetHexColor(wimg.BgColor)
	wimg.context.DrawRectangle(0, 0, float64(wimg.Width), float64(wimg.Height))
	wimg.context.Fill()
}

func (wimg *WaveformImage) DrawPolyLine(points []Point) {
	wimg.context.SetHexColor(wimg.FgColor)
	wimg.context.SetLineWidth(wimg.LineWidth)

	for i, p1 := range points[:len(points)-1] {
		p2 := points[i+1]
		wimg.context.DrawLine(p1.X, p1.Y, p2.X, p2.Y)
	}

	wimg.context.Stroke()
}

func (wimg *WaveformImage) DrawWaveform(data []float32) {
	widthf := float64(wimg.Width)
	heightf := float64(wimg.Height)
	xScale := widthf / float64(len(data)-1)

	wimg.DrawBackground()

	var points = make([]gg.Point, len(data))
	for i, d := range data {
		x := math.Round(float64(i) * xScale)
		y := heightf - math.Round(float64(d)*heightf)
		points[i] = gg.Point{X: x, Y: y}
	}

	wimg.DrawPolyLine(points)
}

// DrawPeakRMS draws one peak/RMS frame: a vertical bar spanning the window's
// negative-to-positive peak range and a pair of horizontal RMS indicators,
// centered in the image. Values are in the bipolar domain [-1, 1] (as
// produced by audio.ComputeWindowStats); non-finite values are treated as 0
// and out-of-range values are clamped to the image bounds.
func (wimg *WaveformImage) DrawPeakRMS(min float64, max float64, rms float64) {
	widthf := float64(wimg.Width)
	heightf := float64(wimg.Height)
	centerX := math.Round(widthf / 2)
	centerY := heightf / 2

	sanitize := func(v float64) float64 {
		if math.IsNaN(v) {
			return 0
		}
		return math.Max(-1, math.Min(1, v))
	}
	yOf := func(v float64) float64 {
		return math.Round(centerY - sanitize(v)*centerY)
	}

	wimg.DrawBackground()

	wimg.context.SetHexColor(wimg.FgColor)
	wimg.context.SetLineWidth(wimg.LineWidth)

	// peak range bar
	wimg.context.DrawLine(centerX, yOf(max), centerX, yOf(min))
	wimg.context.Stroke()

	// RMS indicators
	rms = math.Abs(sanitize(rms))
	tickHalf := math.Max(wimg.LineWidth*4, widthf/16)
	wimg.context.DrawLine(centerX-tickHalf, yOf(rms), centerX+tickHalf, yOf(rms))
	wimg.context.DrawLine(centerX-tickHalf, yOf(-rms), centerX+tickHalf, yOf(-rms))
	wimg.context.Stroke()
}

func (wimg *WaveformImage) Clear() {
	wimg.context.Clear()
}

func (wimg *WaveformImage) Image() image.Image {
	return wimg.context.Image()
}

func (wimg *WaveformImage) ConvertToPaletted(colorPalette color.Palette) *image.Paletted {
	img := wimg.Image()
	bounds := img.Bounds()
	pimg := image.NewPaletted(bounds, colorPalette)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pimg.Set(x, y, img.At(x, y))
		}
	}

	return pimg
}
