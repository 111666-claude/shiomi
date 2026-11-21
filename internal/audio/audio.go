package audio

import (
	"errors"
	"io"
	"math"

	"github.com/madelynnblue/go-dsp/wav"
)

type Audio struct {
	w               *wav.Wav
	freq            uint
	windowSize      uint
	subsamplingRate uint
}

type AudioData struct {
	Samples []float32
	Error   error
}

func NewAudio(f io.Reader, freq uint, windowSize uint, subsamplingRate uint) (*Audio, error) {
	if freq <= 0 {
		return nil, errors.New("freq must be > 0")
	}

	if windowSize <= 0 {
		return nil, errors.New("windowSize must be > 0")
	}

	if subsamplingRate <= 0 {
		return nil, errors.New("subsamplingRate must be > 0")
	}

	w, err := wav.New(f)
	if err != nil {
		return nil, err
	}

	audio := &Audio{w: w, freq: freq, windowSize: windowSize, subsamplingRate: subsamplingRate}
	if audio.RoundedNumSamplesPerWindow() == 0 {
		return nil, errors.New("windowSize is too small for freq and the sample rate")
	}

	return audio, nil
}

func (audio *Audio) NumSamplesPerCycle() float64 {
	return float64(audio.w.Header.SampleRate) / float64(audio.freq)
}

func (audio *Audio) NumSamplesPerWindow() float64 {
	return audio.NumSamplesPerCycle() * float64(audio.windowSize)
}

func (audio *Audio) RoundedNumSamplesPerWindow() uint {
	return uint(math.Round(audio.NumSamplesPerWindow()))
}

func (audio *Audio) NumWindow() uint {
	return uint(audio.w.Samples) / audio.RoundedNumSamplesPerWindow()
}

func (audio *Audio) NumChannels() uint {
	return uint(audio.w.Header.NumChannels)
}

// NumSampleValuesPerWindow is the number of interleaved sample values
// (all channels) covered by one window.
func (audio *Audio) NumSampleValuesPerWindow() uint {
	return audio.RoundedNumSamplesPerWindow() * audio.NumChannels()
}

// NumWindowWithTail counts the windows needed to cover every sample,
// including a trailing window smaller than NumSampleValuesPerWindow.
func (audio *Audio) NumWindowWithTail() uint {
	n := audio.NumSampleValuesPerWindow()
	if n == 0 {
		return 0
	}
	total := uint(audio.w.Samples)
	return (total + n - 1) / n
}

func (audio *Audio) ReadCycles() <-chan *AudioData {
	nSamplesPerWindow := int(audio.RoundedNumSamplesPerWindow())
	nWindow := audio.NumWindow()

	ch := make(chan *AudioData)
	go func() {
		defer close(ch)

		for i := uint(0); i < nWindow; i++ {
			samples, err := audio.w.ReadFloats(nSamplesPerWindow)
			if i%audio.subsamplingRate != 0 {
				continue
			}

			ch <- &AudioData{Samples: samples, Error: err}
		}
	}()

	return ch
}

// ReadWindows is like ReadCycles but reads whole windows accounting for
// channel interleaving, and emits a final partial window when the sample
// count is not a multiple of the window size. Frame i covers exactly the
// sample range [i*NumSampleValuesPerWindow, (i+1)*NumSampleValuesPerWindow)
// clipped to the end of the stream.
func (audio *Audio) ReadWindows() <-chan *AudioData {
	nSamplesPerWindow := int(audio.NumSampleValuesPerWindow())
	nWindow := audio.NumWindowWithTail()
	total := audio.w.Samples

	ch := make(chan *AudioData)
	go func() {
		defer close(ch)

		consumed := 0
		for i := uint(0); i < nWindow; i++ {
			n := nSamplesPerWindow
			if remaining := total - consumed; remaining < n {
				n = remaining
			}
			samples, err := audio.w.ReadFloats(n)
			consumed += n
			if i%audio.subsamplingRate != 0 {
				continue
			}

			ch <- &AudioData{Samples: samples, Error: err}
		}
	}()

	return ch
}
