package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Player movement constant.
const planeMovementStep = 2

// Scroll-in sub-states.
const (
	scrollInFrames    = 38
	scrollInScrolling = 0
	scrollInWaiting   = 1
)

// UpdateGameplay updates the gameplay state based on current mode.
func UpdateGameplay(s *state.GameState, terrain TerrainRenderer) {
	switch s.GameplayMode {
	case domain.GameplayScrollIn:
		updateScrollIn(s, terrain)
	case domain.GameplayNormal, domain.GameplayRefuel:
		in := input.ScanGameplay()
		step(s, in, terrain)
	case domain.GameplayOverview:
	}
}

// updateScrollIn handles the scroll-in sequence logic.
func updateScrollIn(s *state.GameState, terrain TerrainRenderer) {
	switch s.ScrollInState {
	case scrollInScrolling:
		// Advance scroll atomically: updates scroll state, renders terrain, and updates viewport.
		advanceAndRender(s, int(domain.SpeedFast), terrain)
		s.ScrollInCount++

		if s.ScrollInCount >= scrollInFrames {
			s.BridgeDestroyed = false
			s.ScrollInState = scrollInWaiting
		}
	case scrollInWaiting:
		// Wait for any gameplay input (not Enter) to begin.
		in := input.ScanGameplay()
		if in.Left || in.Right || in.Up || in.Down || in.Fire {
			s.GameplayMode = domain.GameplayNormal
		}
	}
}

// step implements the 11-step frame ordering as defined in the architectural specification.
func step(s *state.GameState, in input.Input, terrain TerrainRenderer) {
	if s.Paused {
		if input.IsPausePressed() {
			s.Paused = false
		}
		return
	}

	// step 1: Check pause.
	if input.IsPausePressed() {
		s.Paused = true
		return
	}

	// step 2: Increment frame tick.
	s.Tick++

	// step 5: Process viewport objects (AI).
	moveEnemies(s.Viewport)

	// step 6: Animate player missile.
	updateMissile(s.Missile)

	// step 7: Process tank shells.
	updateTankShell(s.TankShell, int(s.Tick))

	// step 8: Process helicopter missiles.
	updateHeliMissile(s.HeliMissile)

	// step 9: Advance scroll and viewport.
	advanceAndRender(s, int(s.Speed), terrain)

	// step 11: Scan in for next frame.
	applyInput(s, in)
}

// applyInput processes player input for movement and firing.
func applyInput(s *state.GameState, in input.Input) {
	// Reset per-frame flags
	s.Speed = domain.SpeedNormal
	s.PlaneSpriteBank = 0 // Assuming 0 is normal, non-banked. Wait, spec says PlaneSpriteBank.

	if in.Left {
		s.PlaneX -= planeMovementStep
		s.PlaneSpriteBank = 1 // Banked left
	}
	if in.Right {
		s.PlaneX += planeMovementStep
		s.PlaneSpriteBank = 2 // Banked right
	}

	if in.Up {
		s.Speed = domain.SpeedFast
	}
	if in.Down {
		s.Speed = domain.SpeedSlow
	}

	if in.Fire {
		FireMissile(s.Missile, s.PlaneX)
	}
}
