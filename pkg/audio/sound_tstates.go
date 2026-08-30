package audio

import "github.com/morozov/river-raid-ebiten/pkg/domain"

// Per-interrupt T-state delay sequences for the dispatcher sounds, computed from
// the routines' instruction timings. Each phase is the interval between two
// OUTs: a delay loop, plus the OUT and loop bookkeeping around it.

// Sound flags bits ($6BB0). The engine routines mask their period out of this
// byte, so a flag whose bit falls inside the mask shifts the engine's pitch
// while that sound plays.
const (
	soundBitFire      = 0x01
	soundSpeedNormal  = 0x02
	soundSpeedFast    = 0x04
	soundSpeedSlow    = 0x06
	soundBitBonusLife = 0x10
)

// Period masks, one per engine routine: $6C5D, $6CB8, $6CD6.
const (
	normalPeriodMask = 0x0F
	fastPeriodMask   = 0x07
	slowPeriodMask   = 0x17
)

// Engine routine timings.
const (
	engineCycles  = 8   // LD C,$08
	engineEntry   = 43  // routine entry to the first OUT
	engineOnBase  = 17  // ON  = engineOnBase + delayLoopStep×period
	engineOffBase = 31  // OFF = engineOffBase + delayLoopStep×period, normal only
	engineFastOff = 98  // $6CB8 holds OFF for a fixed LD D,$04
	engineSlowOff = 226 // $6CD6 holds OFF for a fixed LD D,$0C
	delayLoopStep = 16
)

// appendEngineDelays appends one interrupt's worth of engine delays for the
// given speed and sound flags.
//
// The sequence ends on an ON phase. The routine's final OFF phase is
// indistinguishable from the wait until the next interrupt, since both hold the
// speaker at rest, so the mixer accounts for it as part of the frame remainder.
func appendEngineDelays(dst []int, speed domain.Speed, flags int) []int {
	var period, off int

	switch speed {
	case domain.SpeedFast:
		period = flags & fastPeriodMask
		off = engineFastOff
	case domain.SpeedSlow:
		period = flags & slowPeriodMask
		off = engineSlowOff
	default:
		period = flags & normalPeriodMask
		off = engineOffBase + delayLoopStep*period
	}

	on := engineOnBase + delayLoopStep*period

	dst = append(dst, engineEntry, on)
	for range engineCycles - 1 {
		dst = append(dst, off, on)
	}

	return dst
}

// Fire routine timings ($8A02). Eight pulses per interrupt, over seven
// interrupts: the missile spawns at Y=126, moves 6 px per animate pass, and the
// flag clears at Y=112, which is 3 × 50 / (2 × 12) ≈ 7.
const (
	fireFrames = 7
	fireCycles = 8
	fireEntry  = 25  // routine entry to the first OUT
	fireOn     = 532 // LD D,$20 delay loop
	fireOff    = 50  // no delay loop: two LD D,$20 and an $FD prefix as filler
)

// Explosion routine timings ($6C7B). Four cycles per interrupt over 23
// interrupts, the OFF phase shortening as the counter runs down.
const (
	explosionTicks   = 24 // counter starts at $18; the last call ends the sound
	explosionCycles  = 4
	explosionEntry   = 121
	explosionOffBase = 44 // OFF = explosionOffBase + delayLoopStep × counter
)

// explosionOnTStates is the ON phase of each interrupt. The routine derives it
// from (DE), which the caller never sets, so it carries whatever the interrupted
// code left behind — this is one capture of that, giving the sound its noise.
// The eight values it can take are 17 + 16 × (($10..$48 in steps of 8)).
var explosionOnTStates = [explosionTicks - 1]int{
	657, 1169, 401, 913, 657, 1041, 785, 785,
	785, 1041, 1169, 657, 657, 1041, 529, 657,
	401, 1169, 273, 1169, 913, 529, 657,
}

// delayLoopIterations is what a DEC/JR NZ loop counter of 0 actually runs: DEC
// wraps it to 255 and the loop carries on down to zero.
const delayLoopIterations = 256

// delayLoop returns the T-state cost of a DEC/JR NZ delay loop entered with the
// counter set to n.
func delayLoop(n int) int {
	if n == 0 {
		n = delayLoopIterations
	}

	return delayLoopStep*n - 5
}

// Low fuel routine timings ($6CF4). Three cycles per interrupt, each decrementing
// the period at $5F65 modulo 128.
//
// The period therefore runs 127, 126 … 1, 0 and back to 127 — all 128 values,
// zero included, where the delay loop takes its longest path. Three periods are
// consumed per interrupt and 3 is coprime with 128, so the frame pattern only
// repeats after 128 interrupts.
const (
	lowFuelCycles     = 3
	lowFuelFrames     = 128
	lowFuelEntry      = 66
	lowFuelOnBase     = 22 // ON  = lowFuelOnBase + delayLoop(period)
	lowFuelOffBase    = 77 // OFF = lowFuelOffBase + delayLoop(period)
	lowFuelPeriodMask = 0x7F
)

// Bonus life timings ($6C31 into the ROM BEEPER at $03B5). The counter runs to
// $40 over 63 interrupts; pitch = ($40 - counter) >> 3 forms H with L = $FF, and
// BEEPER's half-period is 4 × HL + 115.
const (
	bonusLifeTicks = 64
	// bonusLifeHalves is the number of speaker half-periods per interrupt: the
	// ROM routine is called with DE = 1, so it emits one full cycle plus the
	// edge that starts it.
	bonusLifeHalves  = 3
	bonusLifeEntry   = 305 // first interrupt, including the one-time setup
	bonusLifeReentry = 71  // every later interrupt
	bonusPitchScale  = 4
	bonusPitchBase   = 115
	bonusPitchLow    = 0xFF
	bonusPitchShift  = 3
	bonusPitchByte   = 8
)

// buildFrames assembles one interrupt's delays: the routine's entry cost, then
// alternating ON and OFF phases.
//
// The trailing OFF phase is left out. It holds the speaker at rest, which is
// indistinguishable from the wait until the next interrupt, so the mixer
// accounts for it as frame remainder. Leaving it in would toggle the speaker ON
// for the rest of the frame instead.
func buildFrames(entry, cycles int, phase func(cycle int) (on, off int)) []int {
	f := make([]int, 0, 2*cycles)
	f = append(f, entry)

	for c := range cycles {
		on, off := phase(c)
		f = append(f, on)

		if c < cycles-1 {
			f = append(f, off)
		}
	}

	return f
}

// fireFrameDelays returns the fire burst, one entry per interrupt.
func fireFrameDelays() [][]int {
	frames := make([][]int, fireFrames)
	for i := range frames {
		frames[i] = buildFrames(fireEntry, fireCycles, func(int) (int, int) {
			return fireOn, fireOff
		})
	}

	return frames
}

// explosionFrameDelays returns the explosion, one entry per interrupt.
func explosionFrameDelays() [][]int {
	frames := make([][]int, len(explosionOnTStates))
	for i := range frames {
		counter := explosionTicks - 1 - i
		on := explosionOnTStates[i]
		off := explosionOffBase + delayLoopStep*counter
		frames[i] = buildFrames(explosionEntry, explosionCycles, func(int) (int, int) {
			return on, off
		})
	}

	return frames
}

// lowFuelFrameDelays returns one full warble cycle, one entry per interrupt.
// $5F65 holds 0 before the warning first sounds, so the first period is 127.
func lowFuelFrameDelays() [][]int {
	frames := make([][]int, lowFuelFrames)
	period := 0

	for i := range frames {
		frames[i] = buildFrames(lowFuelEntry, lowFuelCycles, func(int) (int, int) {
			period = (period - 1) & lowFuelPeriodMask
			loop := delayLoop(period)

			return lowFuelOnBase + loop, lowFuelOffBase + loop
		})
	}

	return frames
}

// bonusLifeFrameDelays returns the jingle, one entry per interrupt.
func bonusLifeFrameDelays() [][]int {
	frames := make([][]int, 0, bonusLifeTicks-1)

	for counter := 1; counter < bonusLifeTicks; counter++ {
		hl := (bonusLifeTicks-counter)>>bonusPitchShift<<bonusPitchByte | bonusPitchLow
		half := bonusPitchScale*hl + bonusPitchBase

		entry := bonusLifeReentry
		if counter == 1 {
			entry = bonusLifeEntry
		}

		f := make([]int, 0, bonusLifeHalves+1)
		f = append(f, entry)

		for range bonusLifeHalves {
			f = append(f, half)
		}

		frames = append(frames, f)
	}

	return frames
}
