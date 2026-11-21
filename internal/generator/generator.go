package generator

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/gif"
	"io"
	"os"

	"github.com/akiomik/shiomi/internal/audio"
	simage "github.com/akiomik/shiomi/internal/image"
)

type Mode string

const (
	// ModeWaveform draws the instantaneous waveform (default).
	ModeWaveform Mode = "waveform"
	// ModePeakRMS draws the peak range and RMS energy of each window.
	ModePeakRMS Mode = "peak-rms"
)

type Config struct {
	Mode            Mode
	Frequency       uint
	WindowSize      uint
	SubsamplingRate uint
	LineWidth       float64
	Width           uint
	Height          uint
	BgColor         string
	FgColor         string
	Delay           uint
}

func GenerateWaveformGIF(inputFile io.Reader, outputFile io.Writer, config *Config) {
	a, err := audio.NewAudio(inputFile, config.Frequency, config.WindowSize, config.SubsamplingRate)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	colorPalette := palette.Plan9
	wimg, err := simage.NewWaveformImage(config.Width, config.Height, config.BgColor, config.FgColor, config.LineWidth)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	imgGen := make(chan *image.Paletted)
	go func() {
		defer close(imgGen)

		draw := func(samples []float32) {
			wimg.DrawWaveform(samples)
		}
		var cycles <-chan *audio.AudioData
		if config.Mode == ModePeakRMS {
			cycles = a.ReadWindows()
			draw = func(samples []float32) {
				stats := audio.ComputeWindowStats(samples, a.NumChannels())
				wimg.DrawPeakRMS(stats.Min, stats.Max, stats.RMS)
			}
		} else {
			cycles = a.ReadCycles()
		}

		for audioData := range cycles {
			samples := audioData.Samples
			if audioData.Error != nil {
				fmt.Println(audioData.Error.Error())
				os.Exit(1)
			}

			draw(samples)
			pimg := wimg.ConvertToPaletted(colorPalette)
			wimg.Clear()

			imgGen <- pimg
		}
	}()

	animeGIF, err := simage.GenerateAnimationGIF(imgGen, config.Width, config.Height, colorPalette, config.Delay)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	gif.EncodeAll(outputFile, animeGIF)
}
