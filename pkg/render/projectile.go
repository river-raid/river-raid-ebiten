package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// drawPlayerMissile renders the player's missile.
func drawPlayerMissile(screen draw.Image, m *state.PlayerMissile) {
	if !m.Active {
		return
	}

	sm := assets.SpritePlayerMissile
	drawSprite(screen, sm, m.X, m.Y, colorMissile, false)
}

// drawTankShell renders the tank shell or its explosion.
func drawTankShell(screen draw.Image, ts *state.TankShell) {
	if ts.IsExploding {
		if ts.ExplosionFrame < 6 { //nolint:mnd // shell explosion has 6 frames
			s := assets.SpriteShellExplosions[ts.ExplosionFrame]
			// Color cycles each frame: base is green (4), frame offsets through palette.
			colorIdx := platform.Color((int(platform.ColorGreen) + ts.ExplosionFrame) % len(palette))
			drawSprite(screen, s, ts.X, ts.Y, colorIdx, false)
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

// drawHeliMissile renders the helicopter missile as a small projectile.
func drawHeliMissile(screen draw.Image, hm *state.HeliMissile) {
	if !hm.Active {
		return
	}

	ink := palette[platform.ColorYellow]
	screen.Set(hm.X, hm.Y, ink)
	screen.Set(hm.X+1, hm.Y, ink)
}
