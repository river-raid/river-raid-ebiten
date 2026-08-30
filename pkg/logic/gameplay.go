package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Physical plane speed.
const planePxSec = 24 // px/sec

// Derived plane movement step (sp/tick).
const planeMovementStep = planePxSec * domain.SubpixelScale / domain.Tps

// Scroll-in constants.
const (
	// scrollInStep is the scroll-in speed in sp/tick, matching SpeedFast.
	scrollInStep = int(domain.SpeedFast) * domain.SubpixelScale / domain.Tps
	// scrollInFrames is the number of ticks needed to populate the viewport and scroll
	// one terrain profile past it.
	// = (TotalViewportHeight + NumLinesPerTerrainProfile) * SubpixelScale / scrollInStep
	// = 160 * 8 / 8 = 160 ticks
	scrollInFrames    = (domain.TotalViewportHeight + domain.NumLinesPerTerrainProfile) * domain.SubpixelScale / scrollInStep
	scrollInScrolling = state.ScrollInScrolling
	scrollInWaiting   = state.ScrollInWaiting
)

// UpdateGameplay updates the gameplay state based on current mode.
func UpdateGameplay(s *state.GameState, terrain TerrainRenderer) {
	switch s.GameplayMode {
	case domain.GameplayScrollIn:
		updateScrollIn(s, terrain)
	case domain.GameplayNormal, domain.GameplayRefuel:
		step(s, s.InputInterface, terrain)
	case domain.GameplayOverview:
		moveEnemies(s.Viewport, s.Tick, s.TankShell, s.HeliMissile, s.GameplayMode, s.BridgeDestroyed)
		prevPx := s.ScrollY.ToPx()
		s.ScrollY -= domain.SP(int(domain.SpeedNormal) * domain.SubpixelScale / domain.Tps)
		if crossed := prevPx - s.ScrollY.ToPx(); crossed > 0 {
			advanceAndRender(s, crossed, terrain)
		}
	case domain.GameplayDying:
		updateDying(s, terrain)
	}
}

// updateScrollIn handles the scroll-in sequence logic.
func updateScrollIn(s *state.GameState, terrain TerrainRenderer) {
	prevPx := s.ScrollY.ToPx()
	s.ScrollY -= domain.SP(scrollInStep)
	crossed := prevPx - s.ScrollY.ToPx()
	if crossed > 0 {
		advanceAndRender(s, crossed, terrain)
	}
	s.ScrollInCount++

	if s.ScrollInCount >= scrollInFrames {
		s.BridgeDestroyed = false
		// Decrement lives, switch mode to Normal so the plane becomes visible, and
		// enter the waiting sub-state — matching the original's tight input loop.
		s.Players[s.CurrentPlayer].Lives--
		s.GameplayMode = domain.GameplayNormal
		s.ScrollInState = scrollInWaiting
	}
}

// resumeIfRequested polls for the input that ends the current halt: any control
// to take over a new life, or the unpause key.
func resumeIfRequested(s *state.GameState, in input.Interface) {
	switch {
	case s.ScrollInState == scrollInWaiting:
		if in.IsLeftPressed() || in.IsRightPressed() || in.IsUpPressed() || in.IsDownPressed() || in.IsFirePressed() {
			s.ScrollInState = scrollInScrolling
		}
	case s.Paused:
		if input.IsUnpausePressed() {
			s.Paused = false
		}
	}
}

// step advances gameplay by one frame.
func step(s *state.GameState, in input.Interface, terrain TerrainRenderer) {
	if s.Halted() {
		resumeIfRequested(s, in)

		return
	}

	// step 1: Check pause.
	if input.IsPausePressed() {
		s.Paused = true
		return
	}

	// step 2: Animate explosions.
	s.Explosion = animateExplosion(s.Explosion)

	// step 4: Handle collisions.
	terrainLeftX := func(y int) domain.Px {
		left, _ := terrain.GetEdges(s.PlaneX.ToPx(), s.ScrollY.ToPx()+domain.Px(y), 1)
		return left
	}
	terrainRightX := func(y int) domain.Px {
		_, right := terrain.GetEdges(s.PlaneX.ToPx(), s.ScrollY.ToPx()+domain.Px(y), 1)
		return right
	}
	collision := CheckCollisions(
		s.PlaneX,
		s.Missile,
		s.HeliMissile,
		s.Viewport,
		terrainLeftX,
		terrainRightX,
		s.BridgeSection,
		s.BridgeYPosition,
		s.BridgeDestroyed,
		s.BridgeIndex,
	)
	s.Viewport.RemoveByIndices(collision.DestroyObjects)
	s.Explosion = spawnExplosionFragments(s.Explosion, collision.ExplosionFragments, &s.Sounds)
	if collision.PointsScored > 0 {
		addScore(&s.Players[s.CurrentPlayer], &s.Sounds, collision.PointsScored)
	}
	if collision.BridgeHit {
		s.BridgeDestroyed = true
		s.Viewport.ActivationInterval = domain.ActivationIntervalFast
		s.Players[s.CurrentPlayer].BridgeCounter++
		// Re-render the bridge fragment into the terrain buffer with the destruction
		// gap so the visual change is immediate. The fragment was last rendered without
		// the gap (bridgeDestroyed=false); we overwrite it now at the same buffer Y.
		terrain.RenderFragment(s.BridgeFragment, s.BridgeFragBufY, true)
	}
	if collision.Refueling {
		s.GameplayMode = domain.GameplayRefuel
	} else if s.GameplayMode == domain.GameplayRefuel {
		s.GameplayMode = domain.GameplayNormal
	}

	if collision.PlayerDied {
		triggerDeath(s)
		return
	}

	// step 5: Process viewport objects (AI).
	moveEnemies(s.Viewport, s.Tick, s.TankShell, s.HeliMissile, s.GameplayMode, s.BridgeDestroyed)

	// step 6: Animate player missile.
	updateMissile(s.Missile, s.PlaneX)

	// step 7: Process tank shells.
	updateTankShell(s.TankShell, s.Tick)

	// step 8: Process helicopter missiles.
	updateHeliMissile(s.HeliMissile, terrain, s.ScrollY)

	// step 9: Advance scroll and viewport.
	prevPx := s.ScrollY.ToPx()
	s.ScrollY -= domain.SP(int(s.Speed) * domain.SubpixelScale / domain.Tps)
	if crossed := prevPx - s.ScrollY.ToPx(); crossed > 0 {
		advanceAndRender(s, crossed, terrain)
	}

	// step 10: Handle fuel consumption.
	s.Fuel, s.Sounds.FuelState = UpdateFuel(s.Fuel, s.Tick, s.GameplayMode == domain.GameplayRefuel)
	if s.Sounds.FuelState == state.FuelStateEmpty {
		triggerDeath(s)
		return
	}

	// step 11: Scan in for next frame.
	applyInput(s, in)
}

// applyInput processes player input for movement and firing.
func applyInput(s *state.GameState, in input.Interface) {
	// Reset per-frame flags
	s.Speed = domain.SpeedNormal
	s.PlaneSpriteBank = 0 // Assuming 0 is normal, non-banked. Wait, spec says PlaneSpriteBank.

	if in.IsLeftPressed() {
		s.PlaneX -= planeMovementStep
		s.PlaneSpriteBank = 1 // Banked left
	}
	if in.IsRightPressed() {
		s.PlaneX += planeMovementStep
		s.PlaneSpriteBank = 2 // Banked right
	}

	if in.IsUpPressed() {
		s.Speed = domain.SpeedFast
	}
	if in.IsDownPressed() {
		s.Speed = domain.SpeedSlow
	}

	if in.IsFirePressed() {
		if !s.Missile.Active {
			s.Sounds.Firing = true
		}

		FireMissile(s.Missile, s.PlaneX)
	}
}
