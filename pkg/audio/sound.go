package audio

import (
	"embed"
	"io"
	"log"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

//go:embed assets/audio
var audioFS embed.FS

// mixerBufferSize is the mixer player's read-ahead, in whole interrupt frames.
const mixerBufferSize = 3 * time.Second / interruptRate

// gameFrameEvery is the number of ticks between game frames.
//
// The refueling and tank-full beeps are emitted once per iteration of the
// original's main loop, not from the 50 Hz interrupt. That loop is free-running:
// it waits on no interrupt, so its rate is whatever the frame's work costs, and
// it averages ~100 ms during play. Five ticks is the closest this port's tick
// rate comes.
const gameFrameEvery = 5

// dispatchSound holds one dispatcher sound as a list of per-frame T-state delay
// sequences — one entry per interrupt the sound runs for.
type dispatchSound struct {
	frames   [][]int
	frameIdx int // cursor, kept while inactive
	on       bool
	loops    bool // wrap the cursor when exhausted (low fuel only)
}

func newDispatchSound(frames [][]int, loops bool) dispatchSound {
	return dispatchSound{frames: frames, loops: loops}
}

func (s *dispatchSound) active() bool {
	return s.on && len(s.frames) > 0
}

// trigger starts a one-shot from its first frame.
func (s *dispatchSound) trigger() {
	s.on = true
	s.frameIdx = 0
}

// setOn starts or stops a sound without disturbing its cursor, for the sounds
// whose counter is persistent state on the original: the low fuel period at
// $5F65 and the bonus life counter at $6C30. Neither is reset by the event that
// starts the sound, only by the sound running to completion.
func (s *dispatchSound) setOn(on bool) {
	s.on = on
}

// appendDelays appends this sound's delays for the current frame to dst and
// advances to the next frame.
func (s *dispatchSound) appendDelays(dst []int) []int {
	if !s.active() {
		return dst
	}

	dst = append(dst, s.frames[s.frameIdx]...)
	s.frameIdx++

	if s.frameIdx >= len(s.frames) {
		s.frameIdx = 0

		if !s.loops {
			s.on = false
		}
	}

	return dst
}

// mixer is an io.Reader implementing the ZX Spectrum single-channel sequential
// mixer. Each frame it concatenates the active dispatcher sounds' T-state delay
// sequences in dispatcher order — exactly as the interrupt handler runs the
// routines back to back — and integrates the resulting 1-bit speaker timeline
// into one 882-sample (3528-byte stereo) frame.
type mixer struct {
	// delays is the frame being built: the active sounds' timelines, end to end.
	delays []int

	// Dispatcher sounds in dispatcher order.
	fire      dispatchSound // 7-frame one-shot
	bonusLife dispatchSound // 63-frame one-shot
	explosion dispatchSound // 23-frame one-shot
	lowFuel   dispatchSound // 128-frame cycle, loops

	// sub is the oversampled integration of delays. It carries the previous
	// frame's filter tail in front of the current frame's subsamples, so
	// decimation needs no look-ahead.
	sub    [filterTaps - 1 + subFrameSize]float32
	kernel [filterTaps]float64
	dc     dcBlock

	// mu protects all fields accessed by both the game goroutine (setState) and
	// the audio goroutine (Read/generateFrame).
	mu sync.Mutex

	// Frame buffer state: generated one frame at a time, drained by Read.
	frameBuf    [frameBytes]byte
	frameOffset int

	// level is the speaker state at the frame boundary, carried across frames.
	level int

	// Current speed selects which engine routine plays.
	speed domain.Speed

	// Output silence without advancing any sound's cursor. Set until setState
	// reports a running game, so a player started before that reads silence.
	suppressed bool
}

// Read implements io.Reader. Called by the Ebiten audio goroutine.
func (m *mixer) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	n := 0

	for n < len(p) {
		if m.frameOffset >= frameBytes {
			m.generateFrame()
			m.frameOffset = 0
		}

		toCopy := min(len(p)-n, frameBytes-m.frameOffset)
		copy(p[n:], m.frameBuf[m.frameOffset:m.frameOffset+toCopy])
		n += toCopy
		m.frameOffset += toCopy
	}

	return n, nil
}

// generateFrame fills m.frameBuf with the next interrupt frame.
// Must be called with m.mu held.
func (m *mixer) generateFrame() {
	m.delays = m.delays[:0]

	// Read the flags before the appends consume a frame: the dispatcher reads
	// $6BB0 fresh at each check, and no sound routine clears its own bit, so
	// the engine sees the flags as they stand for this whole interrupt.
	flags := m.soundFlags()

	// Dispatcher order ($6BED): fire, bonus life and explosion each return to
	// the dispatcher (CALL NZ), so their bursts stack up within the frame. Low
	// fuel and the engine are tail-calls (JP), so at most one of them runs, and
	// it runs last.
	if !m.suppressed {
		m.delays = m.fire.appendDelays(m.delays)
		m.delays = m.bonusLife.appendDelays(m.delays)
		m.delays = m.explosion.appendDelays(m.delays)

		if m.lowFuel.active() {
			m.delays = m.lowFuel.appendDelays(m.delays)
		} else {
			m.delays = appendEngineDelays(m.delays, m.speed, flags)
		}
	}

	copy(m.sub[:filterTaps-1], m.sub[subFrameSize:])
	m.level = renderFrame(m.sub[filterTaps-1:], m.delays, m.level)

	for i := range frameSize {
		acc := 0.0
		for j, k := range m.kernel {
			acc += k * float64(m.sub[i*oversample+j])
		}

		writeSample(m.frameBuf[i*bytesPerSample:], scaleSample(m.dc.step(acc)))
	}
}

// soundFlags reconstructs the sound flags byte ($6BB0) as the engine routines
// see it. Low fuel and exploding never appear: low fuel tail-calls past the
// engine, and no engine mask covers the exploding bit.
func (m *mixer) soundFlags() int {
	var f int

	switch m.speed {
	case domain.SpeedFast:
		f = soundSpeedFast
	case domain.SpeedSlow:
		f = soundSpeedSlow
	default:
		f = soundSpeedNormal
	}

	if m.fire.active() {
		f |= soundBitFire
	}

	if m.bonusLife.active() {
		f |= soundBitBonusLife
	}

	return f
}

// resetPositions deactivates all one-shot and looping sounds.
// Called when leaving gameplay so sounds don't bleed into the next session.
func (m *mixer) resetPositions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cursors are left alone. The low fuel and bonus life counters are
	// persistent state on the original, and explosion and fire rewind when they
	// are next triggered.
	m.fire.on = false
	m.lowFuel.on = false
	m.bonusLife.on = false
	m.explosion.on = false
	m.suppressed = true
	m.level = 0
	clear(m.sub[:])
	m.dc.reset()
}

// setState updates the mixer's active sound state.
// Called each game tick (game goroutine); protected by mu.
func (m *mixer) setState(s *state.GameState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	suppressed := s.GameplayMode == domain.GameplayScrollIn || s.Paused
	m.suppressed = suppressed

	if suppressed {
		return
	}

	m.speed = s.Speed

	// Low fuel: starts and stops, but keeps its place in the warble.
	m.lowFuel.setOn(s.Sounds.FuelState == state.FuelStateLow)

	if s.Sounds.Firing {
		m.fire.trigger()
		s.Sounds.Firing = false
	}

	if s.Sounds.Exploding {
		m.explosion.trigger()
		s.Sounds.Exploding = false
	}

	// Bonus life: $9119 only sets the flag. The counter at $6C30 is reset by
	// $6C52 when the jingle finishes, so awarding a life mid-jingle does not
	// restart it, and an interrupted jingle resumes.
	if s.Sounds.BonusLife {
		m.bonusLife.setOn(true)
		s.Sounds.BonusLife = false
	}
}

// newMixer builds the mixer with each dispatcher sound's per-interrupt delays
// computed from its routine.
func newMixer() *mixer {
	return &mixer{
		fire:        newDispatchSound(fireFrameDelays(), false),
		bonusLife:   newDispatchSound(bonusLifeFrameDelays(), false),
		explosion:   newDispatchSound(explosionFrameDelays(), false),
		lowFuel:     newDispatchSound(lowFuelFrameDelays(), true),
		dc:          newDCBlock(),
		kernel:      newLowPass(),
		frameOffset: frameBytes, // generate on the first Read
		speed:       domain.SpeedNormal,
		suppressed:  true,
	}
}

// SoundSystem manages all audio playback for the game.
//
// The dispatcher sounds — fire, bonus life, explosion, low fuel and engine —
// share a single-channel sequential mixer, as they share one speaker and one
// interrupt on the original. The BEEPER one-shots (refuel, fuel-full, shell
// whistle, heli missile launch) run outside the dispatcher and use their own
// players.
type SoundSystem struct {
	mx     *mixer
	player *audio.Player

	// Out-of-mixer one-shot players.
	refuel            *audio.Player
	fuelFull          *audio.Player
	shellWhistle      *audio.Player
	heliMissileLaunch *audio.Player

	// Previous-frame state for edge detection of out-of-mixer sounds.
	prevShellFlying bool
	prevHeliActive  bool
}

// NewSoundSystem creates a SoundSystem, computing the dispatcher sounds' delays
// and loading the BEEPER one-shot WAVs.
func NewSoundSystem(ctx *audio.Context) *SoundSystem {
	mx := newMixer()

	p, err := ctx.NewPlayer(mx)
	if err != nil {
		log.Printf("audio: new mixer player: %v", err)
	}

	if p != nil {
		// The mixer synthesizes each frame from the game state at the moment it
		// is read, so the player's read-ahead is latency between an event and
		// the sound of it. The default is half a second.
		p.SetBufferSize(mixerBufferSize)
	}

	return &SoundSystem{
		mx:                mx,
		player:            p,
		refuel:            newPlayerFromBytes(ctx, "refuel.wav"),
		fuelFull:          newPlayerFromBytes(ctx, "fuel-full.wav"),
		shellWhistle:      newPlayerFromBytes(ctx, "shell-whistle.wav"),
		heliMissileLaunch: newPlayerFromBytes(ctx, "heli-missile-launch.wav"),
	}
}

// Update drives audio playback from the current game state.
// Call once per gameplay Update tick. Starts the mixer player on the first call.
func (ss *SoundSystem) Update(gs *state.GameState) {
	// Play fills the player's whole read-ahead synchronously, from whatever
	// state the mixer holds at that moment.
	ss.mx.setState(gs)

	if ss.player != nil {
		ss.player.Play()
	}

	ss.updateRefuel(gs)
	ss.updateFuelFull(gs)
	ss.updateShellWhistle(gs)
	ss.updateHeliMissileLaunch(gs)
}

// StopAll pauses the mixer and all out-of-mixer players, and resets all sound
// positions. Called on any transition away from gameplay.
func (ss *SoundSystem) StopAll() {
	if ss.player != nil {
		ss.player.Pause()
	}

	ss.mx.resetPositions()

	pauseAndRewind(ss.refuel)
	pauseAndRewind(ss.fuelFull)
	pauseAndRewind(ss.shellWhistle)
	pauseAndRewind(ss.heliMissileLaunch)
	ss.prevShellFlying = false
	ss.prevHeliActive = false
}

// updateRefuel plays the refueling beep once per game frame while fuel is being
// added. Suppressed when the tank is full — no fuel is going in.
func (ss *SoundSystem) updateRefuel(gs *state.GameState) {
	if gs.GameplayMode == domain.GameplayRefuel &&
		gs.Sounds.FuelState != state.FuelStateFull &&
		gs.Tick%gameFrameEvery == 0 {
		rewindAndPlay(ss.refuel)
	}
}

// updateFuelFull beeps once per game frame while the tank cannot accept fuel.
func (ss *SoundSystem) updateFuelFull(gs *state.GameState) {
	if gs.Sounds.FuelState == state.FuelStateFull && gs.Tick%gameFrameEvery == 0 {
		rewindAndPlay(ss.fuelFull)
	}
}

func (ss *SoundSystem) updateShellWhistle(gs *state.GameState) {
	flying := gs.TankShell != nil && gs.TankShell.IsFlying

	if flying && !ss.prevShellFlying {
		rewindAndPlay(ss.shellWhistle)
	}

	ss.prevShellFlying = flying
}

func (ss *SoundSystem) updateHeliMissileLaunch(gs *state.GameState) {
	active := gs.HeliMissile != nil && gs.HeliMissile.Active

	if active && !ss.prevHeliActive {
		rewindAndPlay(ss.heliMissileLaunch)
	}

	ss.prevHeliActive = active
}

// --- helpers -----------------------------------------------------------------

// newPlayerFromBytes creates a one-shot audio.Player for out-of-mixer sounds.
func newPlayerFromBytes(ctx *audio.Context, name string) *audio.Player {
	f, err := audioFS.Open("assets/audio/" + name)
	if err != nil {
		log.Printf("audio: open %s: %v", name, err)
		return nil
	}

	stream, err := wav.DecodeWithoutResampling(f)
	if err != nil {
		log.Printf("audio: decode %s: %v", name, err)
		return nil
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("audio: read %s: %v", name, err)
		return nil
	}

	return ctx.NewPlayerFromBytes(data)
}

func rewindAndPlay(p *audio.Player) {
	if p == nil {
		return
	}

	if err := p.Rewind(); err != nil {
		log.Printf("audio: rewind: %v", err)
	}

	p.Play()
}

func pauseAndRewind(p *audio.Player) {
	if p == nil {
		return
	}

	p.Pause()

	if err := p.Rewind(); err != nil {
		log.Printf("audio: rewind: %v", err)
	}
}

// NewContext returns the shared audio context, creating it if needed.
func NewContext() *audio.Context {
	if ctx := audio.CurrentContext(); ctx != nil {
		return ctx
	}

	return audio.NewContext(sampleRate)
}
