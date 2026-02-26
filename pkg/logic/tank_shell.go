package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Tank shell constants.
const (
	shellTrajectorySteps = 8
	shellHorizMultiplier = 2
	shellVertStep        = 1
	shellExplosionFrames = 6
	shellSpeedMask       = 3
)

// FireTankShell launches a shell from the given tank position.
// Speed is pseudo-random, derived from the tick counter.
func FireTankShell(ts *state.TankShell, x, y, tick int, orient domain.Orientation) {
	if ts.IsFlying || ts.IsExploding {
		return
	}

	ts.X = x
	ts.Y = y
	ts.Speed = (tick & shellSpeedMask) + 1
	ts.Orientation = orient
	ts.TrajectoryStep = 0
	ts.IsFlying = true
}

// updateTankShell advances the shell along its trajectory or animates the explosion.
func updateTankShell(ts *state.TankShell, tick int) {
	if ts.IsExploding {
		// Animate on odd ticks only.
		if tick&1 != 0 {
			ts.ExplosionFrame++

			if ts.ExplosionFrame >= shellExplosionFrames || ts.Y >= domain.ViewportHeight {
				clearTankShell(ts)
			}
		}

		return
	}

	if !ts.IsFlying {
		return
	}

	// Horizontal movement.
	dx := ts.Speed * shellHorizMultiplier
	if ts.Orientation == domain.OrientationLeft {
		ts.X -= dx
	} else {
		ts.X += dx
	}

	// Vertical movement (downward).
	ts.Y += shellVertStep
	ts.TrajectoryStep++

	if ts.TrajectoryStep >= shellTrajectorySteps {
		explodeTankShell(ts)

		return
	}

	// Off-screen removal.
	if ts.X < 0 || ts.X >= platform.ScreenWidth || ts.Y >= domain.ViewportHeight {
		clearTankShell(ts)
	}
}

// explodeTankShell transitions the shell from flying to exploding.
func explodeTankShell(ts *state.TankShell) {
	ts.IsFlying = false
	ts.IsExploding = true
	ts.ExplosionFrame = 0
}

// clearTankShell resets all shell state.
func clearTankShell(ts *state.TankShell) {
	*ts = state.TankShell{}
}
