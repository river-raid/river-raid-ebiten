package audio

import (
	"embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

//go:embed assets/audio
var audioFS embed.FS

const sampleRate = 44100

// SoundSystem manages audio playback.
// Continuous sounds (engine, low fuel) loop while their condition holds.
// One-shot sounds (fire, explosion, bonus life, refuel, etc.) play once per trigger.
// The refuel beep is re-triggered every frame while refueling (it finishes before the next frame).
type SoundSystem struct {
	// Continuous (looping) players.
	engineNormal *audio.Player
	engineFast   *audio.Player
	engineSlow   *audio.Player
	lowFuel      *audio.Player

	// One-shot players.
	refuel            *audio.Player
	fire              *audio.Player
	explosion         *audio.Player
	bonusLife         *audio.Player
	fuelFull          *audio.Player
	shellWhistle      *audio.Player
	heliMissileLaunch *audio.Player

	// Currently active engine player (one of the three above, or nil).
	activeEngine *audio.Player

	// Previous-frame state for edge detection.
	prevSpeed       domain.Speed
	prevShellFlying bool
	prevHeliActive  bool
	prevLowFuel     bool
}

// NewSoundSystem creates a SoundSystem, loading all WAV files from the embedded FS.
// Returns nil if the audio context cannot be used (e.g., no audio device).
func NewSoundSystem(ctx *audio.Context) *SoundSystem {
	s := &SoundSystem{}

	s.engineNormal = loadLooping(ctx, "engine-normal.wav")
	s.engineFast = loadLooping(ctx, "engine-fast.wav")
	s.engineSlow = loadLooping(ctx, "engine-slow.wav")
	s.lowFuel = loadLooping(ctx, "low-fuel.wav")

	s.refuel = loadOneShot(ctx, "refuel.wav")
	s.fire = loadOneShot(ctx, "fire.wav")
	s.explosion = loadOneShot(ctx, "explosion.wav")
	s.bonusLife = loadOneShot(ctx, "bonus-life.wav")
	s.fuelFull = loadOneShot(ctx, "fuel-full.wav")
	s.shellWhistle = loadOneShot(ctx, "shell-whistle.wav")
	s.heliMissileLaunch = loadOneShot(ctx, "heli-missile-launch.wav")

	return s
}

// Update drives audio playback from the current game state.
// Call once per gameplay Update tick.
// All sounds are suppressed during scroll-in and while paused.
func (s *SoundSystem) Update(gs *state.GameState) {
	if gs.GameplayMode == domain.GameplayScrollIn || gs.Paused {
		s.StopAll()

		return
	}

	s.updateEngine(gs.Speed)
	s.updateLowFuel(gs.Controls.LowFuel)
	s.updateRefuel(gs)
	s.updateFire(gs)
	s.updateExplosion(gs)
	s.updateBonusLife(gs)
	s.updateFuelFull(gs)
	s.updateShellWhistle(gs)
	s.updateHeliMissileLaunch(gs)
}

// StopAll pauses all active players (e.g., on screen transition away from gameplay).
func (s *SoundSystem) StopAll() {
	pauseAndRewind(s.activeEngine)
	s.activeEngine = nil
	pauseAndRewind(s.lowFuel)
	// One-shot sounds stop naturally; pause them too for a clean transition.
	pauseAndRewind(s.refuel)
	pauseAndRewind(s.fire)
	pauseAndRewind(s.explosion)
	pauseAndRewind(s.bonusLife)
	pauseAndRewind(s.fuelFull)
	pauseAndRewind(s.shellWhistle)
	pauseAndRewind(s.heliMissileLaunch)

	s.prevSpeed = 0
	s.prevShellFlying = false
	s.prevHeliActive = false
	s.prevLowFuel = false
}

// updateEngine switches engine tone immediately when speed changes.
func (s *SoundSystem) updateEngine(speed domain.Speed) {
	if speed == s.prevSpeed && s.activeEngine != nil {
		return
	}

	pauseAndRewind(s.activeEngine)

	s.activeEngine = s.engineForSpeed(speed)
	s.prevSpeed = speed

	if s.activeEngine == nil {
		return
	}

	if !s.activeEngine.IsPlaying() {
		s.activeEngine.Play()
	}
}

// updateLowFuel starts/stops the low fuel warning loop.
func (s *SoundSystem) updateLowFuel(low bool) {
	switch {
	case low && !s.prevLowFuel:
		play(s.lowFuel)
	case !low && s.prevLowFuel:
		pauseAndRewind(s.lowFuel)
	}

	s.prevLowFuel = low
}

// updateRefuel plays the refueling beep once per frame while actively receiving fuel.
// Suppressed when the tank is full (FuelFull is set) — no fuel is being added.
// The beep (~14.6 ms) finishes well before the next frame, so there is no overlap.
func (s *SoundSystem) updateRefuel(gs *state.GameState) {
	if gs.GameplayMode == domain.GameplayRefuel && !gs.Controls.FuelFull {
		rewindAndPlay(s.refuel)
	}
}

// updateFire plays the fire burst on each new missile launch.
func (s *SoundSystem) updateFire(gs *state.GameState) {
	if gs.Controls.FireSound {
		rewindAndPlay(s.fire)
		gs.Controls.FireSound = false
	}
}

// updateExplosion plays the explosion sound on the first frame the flag is set.
func (s *SoundSystem) updateExplosion(gs *state.GameState) {
	if gs.Controls.Exploding {
		rewindAndPlay(s.explosion)
		gs.Controls.Exploding = false
	}
}

// updateBonusLife plays the rising-pitch jingle once per bonus life.
func (s *SoundSystem) updateBonusLife(gs *state.GameState) {
	if gs.Controls.BonusLife {
		rewindAndPlay(s.bonusLife)
		gs.Controls.BonusLife = false
	}
}

// updateFuelFull plays the tank-full beep when the fuel cap is hit.
func (s *SoundSystem) updateFuelFull(gs *state.GameState) {
	if gs.Controls.FuelFull {
		rewindAndPlay(s.fuelFull)
		gs.Controls.FuelFull = false
	}
}

// updateShellWhistle plays the descending whistle when a shell starts flying.
func (s *SoundSystem) updateShellWhistle(gs *state.GameState) {
	flying := gs.TankShell != nil && gs.TankShell.IsFlying

	if flying && !s.prevShellFlying {
		rewindAndPlay(s.shellWhistle)
	}

	s.prevShellFlying = flying
}

// updateHeliMissileLaunch plays a short beep when an advanced helicopter fires.
func (s *SoundSystem) updateHeliMissileLaunch(gs *state.GameState) {
	active := gs.HeliMissile != nil && gs.HeliMissile.Active

	if active && !s.prevHeliActive {
		rewindAndPlay(s.heliMissileLaunch)
	}

	s.prevHeliActive = active
}

// engineForSpeed returns the player for the given speed variant.
func (s *SoundSystem) engineForSpeed(speed domain.Speed) *audio.Player {
	switch speed {
	case domain.SpeedNormal:
		return s.engineNormal
	case domain.SpeedFast:
		return s.engineFast
	case domain.SpeedSlow:
		return s.engineSlow
	}

	return s.engineNormal
}

// --- helpers -----------------------------------------------------------------

func loadLooping(ctx *audio.Context, name string) *audio.Player {
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

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	p, err := ctx.NewPlayer(loop)
	if err != nil {
		log.Printf("audio: new player %s: %v", name, err)
		return nil
	}

	return p
}

func loadOneShot(ctx *audio.Context, name string) *audio.Player {
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

	p, err := ctx.NewPlayer(stream)
	if err != nil {
		log.Printf("audio: new player %s: %v", name, err)
		return nil
	}

	return p
}

func play(p *audio.Player) {
	if p == nil || p.IsPlaying() {
		return
	}

	p.Play()
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

// NewContext returns the shared audio context at the sample rate expected by the WAV files,
// creating it if it does not yet exist.
func NewContext() *audio.Context {
	if ctx := audio.CurrentContext(); ctx != nil {
		return ctx
	}

	return audio.NewContext(sampleRate)
}
