package main

import "image/draw"

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
	Orientation    Orientation
	IsFlying       bool
	IsExploding    bool
}

// Fire launches a shell from the given tank position.
// Speed is pseudo-random, derived from the tick counter.
func (ts *TankShell) Fire(x, y, tick int, orient Orientation) {
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

			if ts.ExplosionFrame >= shellExplosionFrames || ts.Y >= ViewportHeight {
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
	if ts.Orientation == OrientationLeft {
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
	if ts.X < 0 || ts.X >= ScreenWidth || ts.Y >= ViewportHeight {
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

// Draw renders the shell or its explosion.
func (ts *TankShell) Draw(screen draw.Image) {
	if ts.IsExploding {
		if ts.ExplosionFrame < shellExplosionFrames {
			s := SpriteCatalog[SpriteShellExplosion0+SpriteID(ts.ExplosionFrame)]
			// Color cycles each frame: base is green (4), frame offsets through palette.
			colorIdx := (int(ColorGreen) + ts.ExplosionFrame) % len(Palette)
			ink := Palette[colorIdx]
			drawSprite(screen, s, ts.X, ts.Y, ink, false)
		}

		return
	}

	if !ts.IsFlying {
		return
	}

	// Shell is a small dot — draw a 2x2 pixel block.
	ink := Palette[ColorWhite]
	screen.Set(ts.X, ts.Y, ink)
	screen.Set(ts.X+1, ts.Y, ink)
	screen.Set(ts.X, ts.Y+1, ink)
	screen.Set(ts.X+1, ts.Y+1, ink)
}
