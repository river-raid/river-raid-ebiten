package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// Tank shell constants.
const (
	shellTrajectorySteps = 8
	shellHorizMultiplier = 2
	shellVertStep        = 1
	shellExplosionFrames = 6
)

// TankShell tracks the tank shell projectile state.
type TankShell struct {
	X              int
	Y              int
	Speed          int // 1-4 horizontal pixels per frame
	TrajectoryStep int // 0-7
	ExplosionFrame int
	Orientation    domain.Orientation
	IsFlying       bool
	IsExploding    bool
}

// Fire launches a shell from the given tank position.
// Speed is pseudo-random, derived from the tick counter.
func (ts *TankShell) Fire(x, y, tick int, orient domain.Orientation) {
	if ts.IsFlying || ts.IsExploding {
		return
	}

	ts.X = x
	ts.Y = y
	ts.Speed = (tick & 0x03) + 1 //nolint:mnd // pseudo-random speed 1-4
	ts.Orientation = orient
	ts.TrajectoryStep = 0
	ts.IsFlying = true
}

// Update advances the shell along its trajectory or animates the explosion.
func (ts *TankShell) Update(tick int) {
	if ts.IsExploding {
		// Animate on odd ticks only.
		if tick&1 != 0 {
			ts.ExplosionFrame++

			if ts.ExplosionFrame >= shellExplosionFrames || ts.Y >= domain.ViewportHeight {
				ts.Clear()
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
		ts.Explode()

		return
	}

	// Off-screen removal.
	if ts.X < 0 || ts.X >= platform.ScreenWidth || ts.Y >= domain.ViewportHeight {
		ts.Clear()
	}
}

// Explode transitions the shell from flying to exploding.
func (ts *TankShell) Explode() {
	ts.IsFlying = false
	ts.IsExploding = true
	ts.ExplosionFrame = 0
}

// Clear resets all shell state.
func (ts *TankShell) Clear() {
	*ts = TankShell{}
}
