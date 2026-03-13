package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// heliMissileSpriteWidth is the width of the helicopter missile in pixels (one full tile).
const heliMissileSpriteWidth = 8

// drawPlayerMissile renders the player's missile.
func drawPlayerMissile(screen draw.Image, m *state.PlayerMissile) {
	if !m.Active {
		return
	}

	sm := assets.SpritePlayerMissile
	drawSprite(screen, sm, m.X, m.Y, staticColorFn(colorMissile), false)
}

// drawTankShell renders the tank shell or its explosion.
func drawTankShell(screen draw.Image, ts *state.TankShell) {
	const shellExplosionFrames = 6 // total frames in tank shell explosion

	if ts.IsExploding {
		if ts.ExplosionFrame < shellExplosionFrames {
			s := assets.SpriteShellExplosions[ts.ExplosionFrame]
			// Color cycles each frame: base is green (4), frame offsets through palette.
			colorIdx := platform.Color((int(platform.ColorGreen) + ts.ExplosionFrame) % len(palette))
			drawSprite(screen, s, ts.X, ts.Y, staticColorFn(colorIdx), false)
		}

		return
	}

	if !ts.IsFlying {
		return
	}

	// Shell is a small dot — draw a 2x2 pixel block.
	ink := palette[platform.ColorWhite]
	screen.Set(ts.X, ts.Y, ink)
	screen.Set(ts.X+1, ts.Y, ink)
	screen.Set(ts.X, ts.Y+1, ink)
	screen.Set(ts.X+1, ts.Y+1, ink)
}

// drawHeliMissile renders the helicopter missile as an 8×1 horizontal dash.
func drawHeliMissile(screen draw.Image, hm *state.HeliMissile) {
	if !hm.Active {
		return
	}

	// Determine the leftmost pixel: right-facing uses hm.X directly;
	// left-facing aligns the dash so its right end is at hm.X.
	startX := hm.X
	if hm.Orientation == domain.OrientationLeft {
		startX = hm.X - heliMissileSpriteWidth + 1
	}

	ink := palette[platform.ColorGreen]

	for dx := range heliMissileSpriteWidth {
		screen.Set(startX+dx, hm.Y, ink)
	}
}
