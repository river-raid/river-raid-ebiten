package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Physical tank shell speeds (px/sec).
const (
	shellHorizSpeedPxSec = 24 // px/sec per speed unit (multiplier 1–4)
	shellVertSpeedPxSec  = 12 // px/sec downward
)

// Derived tank shell constants (subpixels / ticks).
const (
	shellTrajectoryRange domain.SP = 8 * domain.SubpixelScale
	shellHorizUnitSP     domain.SP = shellHorizSpeedPxSec * domain.SubpixelScale / domain.Tps // per speed unit (multiplier 1–4)
	shellVertSP          domain.SP = shellVertSpeedPxSec * domain.SubpixelScale / domain.Tps
	shellExplosionFrames           = domain.Tps / 2 // 0.5 s duration
	shellSpeedMask                 = 3              // pseudo-random multiplier mask (1–4)
)

// FireTankShell launches a shell from the given tank position.
// Speed is pseudo-random, derived from the tick counter.
func FireTankShell(ts *state.TankShell, x, y domain.SP, tick int, orient domain.Orientation) {
	if ts.IsFlying || ts.IsExploding {
		return
	}

	ts.X = x
	ts.Y = y
	ts.Speed = domain.SP((tick&shellSpeedMask)+1) * shellHorizUnitSP
	ts.Orientation = orient
	ts.TrajectoryStep = 0
	ts.IsFlying = true
}

// updateTankShell advances the shell along its trajectory or animates the explosion.
func updateTankShell(ts *state.TankShell, tick int) {
	if ts.IsExploding {
		// Advance explosion at 24 Hz (every 2 ticks at 48 TPS).
		if tick&1 != 0 {
			ts.ExplosionFrame++

			if ts.ExplosionFrame >= shellExplosionFrames || ts.Y >= domain.TotalViewportHeightSP {
				clearTankShell(ts)
			}
		}

		return
	}

	if !ts.IsFlying {
		return
	}

	// Horizontal movement (Speed is sp/tick directly).
	if ts.Orientation == domain.OrientationLeft {
		ts.X -= ts.Speed
	} else {
		ts.X += ts.Speed
	}

	// Vertical movement (downward).
	ts.Y += shellVertSP
	ts.TrajectoryStep += shellVertSP

	if ts.TrajectoryStep >= shellTrajectoryRange {
		explodeTankShell(ts)

		return
	}

	// Off-screen removal.
	if ts.X < 0 || ts.X >= platform.ScreenWidth*domain.SubpixelScale || ts.Y >= domain.TotalViewportHeightSP {
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
