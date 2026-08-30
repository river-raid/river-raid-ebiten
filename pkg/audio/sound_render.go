package audio

import (
	"encoding/binary"
	"math"
)

// Audio timing constants.
const (
	sampleRate     = 44100
	interruptRate  = 50
	frameSize      = sampleRate / interruptRate // 882 samples per 20 ms interrupt frame
	bytesPerChan   = 2                          // int16
	bytesPerSample = 2 * bytesPerChan           // stereo: L + R
	frameBytes     = frameSize * bytesPerSample // 3528 bytes per frame
)

// Speaker rendering constants.
const (
	// tStatesPerFrame is the ZX Spectrum 48K frame length, and the T-state
	// budget one interrupt has. One output sample spans ~79.2 T-states of it.
	tStatesPerFrame = 69888

	// dcBlockHz is the corner frequency of the speaker's AC coupling model. It
	// sits below the 50 Hz interrupt rate, which carries most of the signal.
	dcBlockHz = 25.0

	// oversample is how many subsamples the speaker timeline is integrated into
	// per output sample. The engine's carrier runs at 31 kHz, well above the
	// output Nyquist rate, so it has to be resolved here and filtered out before
	// decimation. Integrating straight to the output rate folds it back as a
	// 12.8 kHz tone and lifts the engine to 0.62 of full swing, where its duty
	// cycle is 0.4375.
	oversample = 8

	// lowPassHz and filterTaps define the windowed-sinc low-pass applied before
	// decimation. 10 kHz passes everything the game produces as a tone — the
	// fire burst's 6 kHz carrier comes through unattenuated — and puts the
	// engine's carrier deep in the stopband.
	lowPassHz  = 10_000
	filterTaps = 33

	hammingA0 = 0.54
	hammingA1 = 0.46
)

// subFrameSize is the oversampled length of one frame.
const subFrameSize = frameSize * oversample

// newLowPass builds the Hamming-windowed sinc low-pass for the oversampled
// rate, normalized to unity gain at DC.
func newLowPass() [filterTaps]float64 {
	var (
		k   [filterTaps]float64
		sum float64
	)

	fc := float64(lowPassHz) / float64(sampleRate*oversample)
	mid := float64(filterTaps-1) / 2

	for i := range k {
		x := float64(i) - mid

		v := 2 * fc
		if x != 0 {
			v = math.Sin(2*math.Pi*fc*x) / (math.Pi * x)
		}

		v *= hammingA0 - hammingA1*math.Cos(2*math.Pi*float64(i)/float64(filterTaps-1))
		k[i] = v
		sum += v
	}

	for i := range k {
		k[i] /= sum
	}

	return k
}

// renderFrame integrates a 1-bit speaker timeline into len(dst) levels in the
// range 0..1, spanning exactly one frame.
//
// delays holds successive T-state waits; the speaker toggles after each one and
// holds its final level to the end of the frame. level is the speaker state on
// entry, and the state on exit is returned so a hold carries across the frame
// boundary. A timeline longer than one frame is truncated.
//
// Each value is the mean speaker level over the T-states it spans. Called with
// an oversampled dst, so the routines' 49 T-state half-periods — under two
// thirds of an output sample — are resolved rather than folded into the output.
func renderFrame(dst []float32, delays []int, level int) int {
	di := 0 // index into delays
	held := 0
	t := 0 // T-state position, always equal to the sample window start

	for i := range dst {
		start := i * tStatesPerFrame / len(dst)
		end := (i + 1) * tStatesPerFrame / len(dst)
		on := 0

		for t < end {
			if di >= len(delays) {
				if level == 1 {
					on += end - t
				}

				t = end

				break
			}

			step := min(end-t, delays[di]-held)
			if level == 1 {
				on += step
			}

			t += step
			held += step

			if held >= delays[di] {
				di++
				held = 0
				level ^= 1
			}
		}

		dst[i] = float32(on) / float32(end-start)
	}

	return level
}

// dcBlock is a one-pole high-pass filter modeling the speaker's AC coupling: a
// held level decays to rest rather than offsetting the output.
type dcBlock struct {
	coeff   float64
	prevIn  float64
	prevOut float64
}

func newDCBlock() dcBlock {
	return dcBlock{coeff: math.Exp(-2 * math.Pi * dcBlockHz / sampleRate)}
}

func (d *dcBlock) step(x float64) float64 {
	y := x - d.prevIn + d.coeff*d.prevOut
	d.prevIn = x
	d.prevOut = y

	return y
}

func (d *dcBlock) reset() {
	d.prevIn = 0
	d.prevOut = 0
}

// writeSample writes one mono sample as a stereo 16-bit LE pair.
func writeSample(dst []byte, v int16) {
	u := uint16(v) //nolint:gosec // deliberate int16→uint16 reinterpret for PCM encoding
	binary.LittleEndian.PutUint16(dst, u)
	binary.LittleEndian.PutUint16(dst[bytesPerChan:], u)
}

// scaleSample converts a filtered speaker level to a 16-bit sample. The speaker
// port is one bit with no level control, so it maps to full scale, matching the
// BEEPER one-shot WAVs that share the output.
func scaleSample(v float64) int16 {
	s := v * math.MaxInt16

	if s > math.MaxInt16 {
		return math.MaxInt16
	}

	if s < math.MinInt16 {
		return math.MinInt16
	}

	return int16(s)
}
