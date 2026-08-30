package audio

import (
	"bytes"
	"math"
	"slices"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// sum returns the total T-states in a delay sequence.
func sum(delays []int) int {
	total := 0
	for _, d := range delays {
		total += d
	}

	return total
}

// frameLevels generates one frame and returns its samples as fractions of full
// scale.
func frameLevels(m *mixer) []float32 {
	m.generateFrame()

	out := make([]float32, frameSize)
	for i := range out {
		lo := uint16(m.frameBuf[i*bytesPerSample])
		hi := uint16(m.frameBuf[i*bytesPerSample+1])
		//nolint:gosec // deliberate uint16→int16 reinterpret, decoding LE PCM
		out[i] = float32(int16(lo|hi<<8)) / math.MaxInt16
	}

	return out
}

// lastActive returns the index of the last sample above threshold, or -1.
func lastActive(samples []float32, threshold float32) int {
	for i := len(samples) - 1; i >= 0; i-- {
		if samples[i] > threshold {
			return i
		}
	}

	return -1
}

// TestRenderFrame_HoldsLevelForWholeFrame verifies that a timeline holding the
// speaker ON for the whole frame renders as a solid level.
func TestRenderFrame_HoldsLevelForWholeFrame(t *testing.T) {
	t.Parallel()

	dst := make([]float32, frameSize)
	// Toggle ON immediately, then hold past the end of the frame.
	got := renderFrame(dst, []int{0, tStatesPerFrame * 2}, 0)

	for i, v := range dst {
		if v != 1 {
			t.Fatalf("sample %d = %v, want 1", i, v)
		}
	}

	if got != 1 {
		t.Errorf("exit level = %d, want 1", got)
	}
}

// TestRenderFrame_HalfFrameBurst verifies the T-state to sample mapping: a hold
// covering the first half of the frame fills the first half of the samples.
func TestRenderFrame_HalfFrameBurst(t *testing.T) {
	t.Parallel()

	dst := make([]float32, frameSize)
	half := tStatesPerFrame / 2
	renderFrame(dst, []int{0, half}, 0)

	if dst[0] != 1 {
		t.Errorf("first sample = %v, want 1", dst[0])
	}

	if dst[frameSize-1] != 0 {
		t.Errorf("last sample = %v, want 0", dst[frameSize-1])
	}

	// The transition lands within a sample of the midpoint.
	mid := lastActive(dst, 0)
	if mid < frameSize/2-1 || mid > frameSize/2 {
		t.Errorf("burst ends at sample %d, want %d±1", mid, frameSize/2)
	}
}

// TestMixer_EngineLevelMatchesDutyCycle pins the burst level. The normal
// engine's carrier is 31 kHz, far above the output Nyquist rate, so the speaker
// cannot follow it and the burst reaches the mean level the port holds — its
// duty cycle, 49/(49+63). Folding the carrier into the output instead lifts the
// burst to 0.62 of full swing and puts a 12.8 kHz tone on it.
func TestMixer_EngineLevelMatchesDutyCycle(t *testing.T) {
	t.Parallel()

	const wantDuty = 49.0 / (49.0 + 63.0)

	out := frameLevels(newMixer())

	peak := float32(0)
	for _, v := range out {
		if v > peak {
			peak = v
		}
	}

	if math.Abs(float64(peak)-wantDuty) > 0.02 {
		t.Errorf("engine burst peaks at %.4f of full swing, want ~%.4f", peak, wantDuty)
	}

	// Across the flat top the level must hold rather than oscillate. It droops
	// slightly, as the DC blocker discharges through the burst, so the bound
	// only has to separate that from the half-swing ripple an aliased carrier
	// produces.
	var plateau []float32

	for _, v := range out {
		if v > peak*0.9 {
			plateau = append(plateau, v)
		}
	}

	if len(plateau) < 4 {
		t.Fatalf("burst plateau is %d samples, too short to check", len(plateau))
	}

	lo, hi := plateau[0], plateau[0]
	for _, v := range plateau {
		lo = min(lo, v)
		hi = max(hi, v)
	}

	if hi-lo > 0.06 {
		t.Errorf("burst plateau ripples by %.4f, want a held level", hi-lo)
	}
}

// TestRenderFrame_ResolvesEngineVariants verifies that variants separated only
// by sub-sample timing stay distinct. At the gameplay periods the normal and
// fast engine routines differ in duty cycle and burst length, but every one of
// their delays falls in the same one-to-two sample range.
func TestRenderFrame_ResolvesEngineVariants(t *testing.T) {
	t.Parallel()

	normalOut := make([]float32, frameSize)
	fastOut := make([]float32, frameSize)
	renderFrame(normalOut, appendEngineDelays(nil, domain.SpeedNormal, soundSpeedNormal), 0)
	renderFrame(fastOut, appendEngineDelays(nil, domain.SpeedFast, soundSpeedFast), 0)

	if lastActive(normalOut, 0) >= lastActive(fastOut, 0) {
		t.Errorf("normal burst ends at %d, fast at %d; fast is the longer routine",
			lastActive(normalOut, 0), lastActive(fastOut, 0))
	}

	same := true

	for i := range normalOut {
		if normalOut[i] != fastOut[i] {
			same = false

			break
		}
	}

	if same {
		t.Error("normal and fast engine render identically")
	}
}

// TestRenderFrame_CarriesLevelAcrossFrames verifies that the speaker level at
// the end of one frame is the level at the start of the next.
func TestRenderFrame_CarriesLevelAcrossFrames(t *testing.T) {
	t.Parallel()

	dst := make([]float32, frameSize)

	// One toggle to ON, no toggle back: the level must leave the frame ON.
	if got := renderFrame(dst, []int{100}, 0); got != 1 {
		t.Fatalf("exit level = %d, want 1", got)
	}

	// Entering ON with no toggles at all holds ON for the whole frame.
	renderFrame(dst, nil, 1)

	for i, v := range dst {
		if v != 1 {
			t.Fatalf("sample %d = %v, want 1 (level carried in)", i, v)
		}
	}
}

// TestRenderFrame_TruncatesOverrun verifies that a timeline longer than one
// frame is cut at the frame boundary rather than spilling.
func TestRenderFrame_TruncatesOverrun(t *testing.T) {
	t.Parallel()

	dst := make([]float32, frameSize)
	// ON for the whole frame, then a toggle that falls beyond the frame.
	renderFrame(dst, []int{0, tStatesPerFrame + 1000, 500}, 0)

	if dst[frameSize-1] != 1 {
		t.Errorf("last sample = %v, want 1 (post-frame toggle must not apply)", dst[frameSize-1])
	}
}

// TestAppendEngineDelays_MatchesReferenceSequences checks the routine timing
// model against the disassembly's documented delay sequences. Those were
// captured with the period nibble forced to 10, so feeding the same flags back
// in must reproduce them exactly.
func TestAppendEngineDelays_MatchesReferenceSequences(t *testing.T) {
	t.Parallel()

	const (
		refNormalFlags = 0x0A // period nibble 10, normal speed bits
		refOtherFlags  = 0x0E // period nibble 10, fast/slow speed bits
	)

	cases := []struct {
		name  string
		speed domain.Speed
		flags int
		on    int
		off   int
	}{
		{"normal", domain.SpeedNormal, refNormalFlags, 177, 191},
		{"fast", domain.SpeedFast, refOtherFlags, 113, 98},
		{"slow", domain.SpeedSlow, refOtherFlags, 113, 226},
	}

	for _, c := range cases {
		want := []int{engineEntry, c.on}
		for range engineCycles - 1 {
			want = append(want, c.off, c.on)
		}

		got := appendEngineDelays(nil, c.speed, c.flags)
		if !slices.Equal(got, want) {
			t.Errorf("%s engine delays = %v, want %v", c.name, got, want)
		}
	}
}

// TestAppendEngineDelays_GameplayPeriods verifies the periods the game actually
// produces. The flags byte holds only the speed bits, and each routine masks its
// period out of it: $0F, $07 and $17 respectively.
func TestAppendEngineDelays_GameplayPeriods(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		speed         domain.Speed
		flags         int
		on, off, span int
	}{
		{"normal", domain.SpeedNormal, soundSpeedNormal, 49, 63, 876},
		{"fast", domain.SpeedFast, soundSpeedFast, 81, 98, 1377},
		{"slow", domain.SpeedSlow, soundSpeedSlow, 113, 226, 2529},
	}

	for _, c := range cases {
		got := appendEngineDelays(nil, c.speed, c.flags)

		if got[1] != c.on {
			t.Errorf("%s ON = %d, want %d", c.name, got[1], c.on)
		}

		if got[2] != c.off {
			t.Errorf("%s OFF = %d, want %d", c.name, got[2], c.off)
		}

		if sum(got) != c.span {
			t.Errorf("%s burst = %d T-states, want %d", c.name, sum(got), c.span)
		}

		if sum(got) >= tStatesPerFrame {
			t.Errorf("%s burst does not fit in a frame", c.name)
		}
	}
}

// TestAppendEngineDelays_FireShiftsPeriod pins the quirk that the period is
// masked out of the whole flags byte: firing sets bit 0, which falls inside the
// normal engine's $0F mask and raises its period from 2 to 3.
func TestAppendEngineDelays_FireShiftsPeriod(t *testing.T) {
	t.Parallel()

	quiet := appendEngineDelays(nil, domain.SpeedNormal, soundSpeedNormal)
	firing := appendEngineDelays(nil, domain.SpeedNormal, soundSpeedNormal|soundBitFire)

	if wantOn := engineOnBase + delayLoopStep*3; firing[1] != wantOn {
		t.Errorf("ON while firing = %d, want %d", firing[1], wantOn)
	}

	if firing[1] <= quiet[1] {
		t.Errorf("ON while firing = %d, not longer than %d", firing[1], quiet[1])
	}
}

// TestMixer_SoundFlagsTracksActiveSounds verifies that the byte the engine reads
// carries the bits of the sounds playing alongside it.
func TestMixer_SoundFlagsTracksActiveSounds(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.speed = domain.SpeedSlow

	if got := m.soundFlags(); got != soundSpeedSlow {
		t.Errorf("idle flags = %#02x, want %#02x", got, soundSpeedSlow)
	}

	m.fire.trigger()
	m.bonusLife.trigger()

	want := soundSpeedSlow | soundBitFire | soundBitBonusLife
	if got := m.soundFlags(); got != want {
		t.Errorf("flags = %#02x, want %#02x", got, want)
	}
}

// TestDispatchSound_AppendDelaysAdvances verifies that each call yields the next
// frame's delays and the sound goes inactive after the last one.
func TestDispatchSound_AppendDelaysAdvances(t *testing.T) {
	t.Parallel()

	s := newDispatchSound([][]int{{10}, {20}}, false)
	s.trigger()

	if got := s.appendDelays(nil); len(got) != 1 || got[0] != 10 {
		t.Fatalf("frame 0 delays = %v, want [10]", got)
	}

	if got := s.appendDelays(nil); len(got) != 1 || got[0] != 20 {
		t.Fatalf("frame 1 delays = %v, want [20]", got)
	}

	if s.active() {
		t.Error("sound still active after the last frame")
	}

	if got := s.appendDelays(nil); got != nil {
		t.Errorf("inactive sound appended %v, want nothing", got)
	}
}

// TestDispatchSound_Loops verifies that a looping sound wraps to frame 0.
func TestDispatchSound_Loops(t *testing.T) {
	t.Parallel()

	s := newDispatchSound([][]int{{10}, {20}}, true)
	s.trigger()

	s.appendDelays(nil)
	s.appendDelays(nil)

	if s.frameIdx != 0 {
		t.Errorf("frameIdx after wrap = %d, want 0", s.frameIdx)
	}
}

// TestDCBlock_RemovesOffset verifies that a constant input decays to zero.
func TestDCBlock_RemovesOffset(t *testing.T) {
	t.Parallel()

	d := newDCBlock()

	var out float64
	for range sampleRate {
		out = d.step(1)
	}

	if math.Abs(out) > 0.01 {
		t.Errorf("output after 1 s of DC = %v, want ~0", out)
	}
}

// TestDCBlock_PassesTransients verifies that a level change still produces
// output — the filter must remove the offset, not the signal.
func TestDCBlock_PassesTransients(t *testing.T) {
	t.Parallel()

	d := newDCBlock()

	if got := d.step(1); got < 0.9 {
		t.Errorf("step response = %v, want ~1", got)
	}
}

// TestScaleSample_Clamps verifies that out-of-range levels saturate rather than
// wrapping.
func TestScaleSample_Clamps(t *testing.T) {
	t.Parallel()

	if got := scaleSample(100); got != math.MaxInt16 {
		t.Errorf("scaleSample(100) = %d, want %d", got, math.MaxInt16)
	}

	if got := scaleSample(-100); got != math.MinInt16 {
		t.Errorf("scaleSample(-100) = %d, want %d", got, math.MinInt16)
	}

	if got := scaleSample(0); got != 0 {
		t.Errorf("scaleSample(0) = %d, want 0", got)
	}
}

// TestMixer_DispatcherOrder verifies that a CALL NZ sound is laid down before
// the engine tail-call, pushing the engine burst later in the frame.
func TestMixer_DispatcherOrder(t *testing.T) {
	t.Parallel()

	engineEnd := lastActive(frameLevels(newMixer()), 0.02)

	withExplosion := newMixer()
	withExplosion.explosion.trigger()
	combinedEnd := lastActive(frameLevels(withExplosion), 0.02)

	if combinedEnd <= engineEnd {
		t.Errorf("explosion+engine ends at sample %d, engine alone at %d; "+
			"the engine must follow the explosion in the frame", combinedEnd, engineEnd)
	}
}

// TestMixer_LowFuelReplacesEngine verifies the JP NZ tail-call: when low fuel is
// active the engine does not play.
func TestMixer_LowFuelReplacesEngine(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.lowFuel.trigger()
	m.generateFrame()

	if got, want := sum(m.delays), sum(lowFuelFrameDelays()[0]); got != want {
		t.Errorf("frame timeline = %d T-states, want %d (low fuel only)", got, want)
	}
}

// TestMixer_SuppressedIsSilentAndHoldsPosition verifies that suppression outputs
// silence without consuming a sound's frames.
func TestMixer_SuppressedIsSilentAndHoldsPosition(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.explosion.trigger()
	m.suppressed = true

	buf := make([]byte, frameBytes)
	if _, err := m.Read(buf); err != nil {
		t.Fatalf("Read: %v", err)
	}

	for i, b := range buf {
		if b != 0 {
			t.Fatalf("byte %d = %d, want 0 (suppressed)", i, b)
		}
	}

	if m.explosion.frameIdx != 0 {
		t.Errorf("explosion advanced to frame %d while suppressed, want 0", m.explosion.frameIdx)
	}
}

// TestMixer_ReadFillsAnyLength verifies that Read satisfies the whole request
// across frame boundaries.
func TestMixer_ReadFillsAnyLength(t *testing.T) {
	t.Parallel()

	m := newMixer()

	buf := make([]byte, frameBytes*3/2)

	n, err := m.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if n != len(buf) {
		t.Errorf("Read returned %d bytes, want %d", n, len(buf))
	}
}

// TestMixer_OutputIsStereo verifies that both channels carry the same sample.
func TestMixer_OutputIsStereo(t *testing.T) {
	t.Parallel()

	m := newMixer()

	buf := make([]byte, frameBytes)
	if _, err := m.Read(buf); err != nil {
		t.Fatalf("Read: %v", err)
	}

	for i := 0; i+bytesPerSample <= len(buf); i += bytesPerSample {
		left := buf[i : i+bytesPerChan]
		right := buf[i+bytesPerChan : i+bytesPerSample]

		if !bytes.Equal(left, right) {
			t.Fatalf("sample at byte %d differs between channels", i)
		}
	}
}

// TestDispatcherFramesReturnSpeakerToRest is the invariant every routine holds:
// it leaves the speaker OFF, so the rest of the frame is silence. A frame with
// an odd number of delays ends on a toggle to ON and holds full scale until the
// next interrupt, which inverts alternate frames and swamps the sound with a
// 25 Hz square.
func TestDispatcherFramesReturnSpeakerToRest(t *testing.T) {
	t.Parallel()

	sounds := map[string][][]int{
		"fire":      fireFrameDelays(),
		"explosion": explosionFrameDelays(),
		"lowFuel":   lowFuelFrameDelays(),
		"bonusLife": bonusLifeFrameDelays(),
	}

	sounds["engine normal"] = [][]int{appendEngineDelays(nil, domain.SpeedNormal, soundSpeedNormal)}
	sounds["engine fast"] = [][]int{appendEngineDelays(nil, domain.SpeedFast, soundSpeedFast)}
	sounds["engine slow"] = [][]int{appendEngineDelays(nil, domain.SpeedSlow, soundSpeedSlow)}

	for name, frames := range sounds {
		for i, f := range frames {
			if len(f)%2 != 0 {
				t.Errorf("%s frame %d has %d delays; an odd count leaves the speaker ON", name, i, len(f))
			}

			if got := sum(f); got >= tStatesPerFrame {
				t.Errorf("%s frame %d spans %d T-states, over the %d budget", name, i, got, tStatesPerFrame)
			}
		}
	}
}

// TestDispatcherFrameCounts checks each sound against the routine that drives it.
func TestDispatcherFrameCounts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		frames [][]int
		want   int
	}{
		{"fire", fireFrameDelays(), 7},
		{"explosion", explosionFrameDelays(), 23},
		{"lowFuel", lowFuelFrameDelays(), 128},
		{"bonusLife", bonusLifeFrameDelays(), 63},
	}

	for _, c := range cases {
		if got := len(c.frames); got != c.want {
			t.Errorf("%s: %d frames, want %d", c.name, got, c.want)
		}
	}
}

// TestDerivedFramesMatchReferenceSequences checks the computed delays against
// the sequences the disassembly documents.
func TestDerivedFramesMatchReferenceSequences(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		frame []int
		want  []int
	}{
		{"fire", fireFrameDelays()[0], []int{25, 532, 50, 532, 50, 532, 50, 532, 50, 532, 50, 532, 50, 532, 50, 532}},
		{"explosion 0", explosionFrameDelays()[0], []int{121, 657, 412, 657, 412, 657, 412, 657}},
		{"explosion 1", explosionFrameDelays()[1], []int{121, 1169, 396, 1169, 396, 1169, 396, 1169}},
		{"explosion 22", explosionFrameDelays()[22], []int{121, 657, 60, 657, 60, 657, 60, 657}},
		{"bonusLife 0", bonusLifeFrameDelays()[0], []int{305, 8303, 8303, 8303}},
		{"bonusLife 8", bonusLifeFrameDelays()[8], []int{71, 7279, 7279, 7279}},
		{"bonusLife 62", bonusLifeFrameDelays()[62], []int{71, 1135, 1135, 1135}},
	}

	for _, c := range cases {
		if !slices.Equal(c.frame, c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.frame, c.want)
		}
	}
}

// TestLowFuelPhaseDelays checks the period-to-delay mapping against the values
// the disassembly documents. Frame alignment is not comparable: $5F65 holds 0
// before the warning first sounds, so the first period is 127, where the
// documented sequence starts one cycle later at 126.
func TestLowFuelPhaseDelays(t *testing.T) {
	t.Parallel()

	cases := []struct {
		period, on, off int
	}{
		{126, 2033, 2088},
		{125, 2017, 2072},
		{3, 65, 120},
		{2, 49, 104},
		{1, 33, 88},
		// A counter of 0 does not skip the loop: DEC wraps to 255 and it runs
		// 256 times, so period 0 is the longest phase of the warble, not the
		// shortest.
		{0, 4113, 4168},
	}

	for _, c := range cases {
		if got := lowFuelOnBase + delayLoop(c.period); got != c.on {
			t.Errorf("period %d ON = %d, want %d", c.period, got, c.on)
		}

		if got := lowFuelOffBase + delayLoop(c.period); got != c.off {
			t.Errorf("period %d OFF = %d, want %d", c.period, got, c.off)
		}
	}
}

// TestLowFuelCycleVisitsEveryPeriod verifies that the warble runs the counter
// through all 128 values. Three periods are consumed per interrupt and 3 is
// coprime with 128, so the frame pattern only repeats after 128 interrupts —
// sizing the cycle to the 126 non-zero periods drops both 0 and 127 and puts a
// seam in the middle of the sweep.
func TestLowFuelCycleVisitsEveryPeriod(t *testing.T) {
	t.Parallel()

	seen := map[int]int{}

	for _, f := range lowFuelFrameDelays() {
		// Phases alternate ON, OFF, ON, OFF, ON after the entry cost; each ON
		// identifies its period.
		for i := 1; i < len(f); i += 2 {
			seen[f[i]-lowFuelOnBase]++
		}
	}

	for period := range lowFuelPeriodMask + 1 {
		if seen[delayLoop(period)] == 0 {
			t.Errorf("period %d never occurs in the cycle", period)
		}
	}

	if got, want := len(seen), lowFuelPeriodMask+1; got != want {
		t.Errorf("cycle visits %d distinct periods, want %d", got, want)
	}
}

// TestEnginePeriodIsNeverZero guards the same zero-counter hazard for the engine
// routines: each masks its period out of the flags byte, and every reachable
// flags value keeps a speed bit inside the mask, so the loop counter cannot
// reach 0 and the closed-form ON delay stays valid.
func TestEnginePeriodIsNeverZero(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		speed           domain.Speed
		speedBits, mask int
	}{
		{domain.SpeedNormal, soundSpeedNormal, normalPeriodMask},
		{domain.SpeedFast, soundSpeedFast, fastPeriodMask},
		{domain.SpeedSlow, soundSpeedSlow, slowPeriodMask},
	} {
		for _, extra := range []int{0, soundBitFire, soundBitBonusLife, soundBitFire | soundBitBonusLife} {
			if got := (c.speedBits | extra) & c.mask; got == 0 {
				t.Errorf("speed %v with extra flags %#02x gives period 0", c.speed, extra)
			}
		}
	}
}

// TestLowFuelResumesAcrossEpisodes pins the persistence of $5F65. Nothing
// outside $6CF4 writes it, and stopping the warning does not reset it, so a
// later episode continues the sweep rather than restarting it.
func TestLowFuelResumesAcrossEpisodes(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.lowFuel.setOn(true)

	for range 7 {
		m.generateFrame()
	}

	at := m.lowFuel.frameIdx
	if at == 0 {
		t.Fatal("cursor did not advance")
	}

	m.lowFuel.setOn(false)
	m.generateFrame()
	m.lowFuel.setOn(true)

	if m.lowFuel.frameIdx != at {
		t.Errorf("cursor is %d after restarting, want %d", m.lowFuel.frameIdx, at)
	}

	// Leaving gameplay must not reset it either.
	m.resetPositions()

	if m.lowFuel.frameIdx != at {
		t.Errorf("cursor is %d after resetPositions, want %d", m.lowFuel.frameIdx, at)
	}
}

// TestTriggerRewindsOneShot verifies the other half: a one-shot restarts from
// its first frame every time it is triggered.
func TestTriggerRewindsOneShot(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.explosion.trigger()

	for range 5 {
		m.generateFrame()
	}

	m.explosion.trigger()

	if m.explosion.frameIdx != 0 {
		t.Errorf("explosion resumed at frame %d, want 0", m.explosion.frameIdx)
	}
}

// TestBonusLifeResumesRatherThanRewinding pins the counter at $6C30. The award
// site $9119 only sets the flag; $6C52 resets the counter when the jingle
// finishes. So a second award mid-jingle does not restart it, and a jingle cut
// short by leaving gameplay resumes.
func TestBonusLifeResumesRatherThanRewinding(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.bonusLife.setOn(true)

	for range 9 {
		m.generateFrame()
	}

	at := m.bonusLife.frameIdx
	if at == 0 {
		t.Fatal("cursor did not advance")
	}

	// A second award while the jingle plays.
	m.bonusLife.setOn(true)

	if m.bonusLife.frameIdx != at {
		t.Errorf("cursor is %d after a second award, want %d", m.bonusLife.frameIdx, at)
	}

	// Interrupted by leaving gameplay, then awarded again.
	m.resetPositions()
	m.bonusLife.setOn(true)

	if m.bonusLife.frameIdx != at {
		t.Errorf("cursor is %d after an interrupted jingle, want %d", m.bonusLife.frameIdx, at)
	}
}

// TestBonusLifeRewindsOnCompletion verifies the other half of $6C52: running to
// the end returns the counter to zero, so the next award starts the jingle over.
func TestBonusLifeRewindsOnCompletion(t *testing.T) {
	t.Parallel()

	m := newMixer()
	m.bonusLife.setOn(true)

	for range len(bonusLifeFrameDelays()) {
		m.generateFrame()
	}

	if m.bonusLife.active() {
		t.Error("jingle still active after its last frame")
	}

	if m.bonusLife.frameIdx != 0 {
		t.Errorf("cursor is %d after completion, want 0", m.bonusLife.frameIdx)
	}
}
