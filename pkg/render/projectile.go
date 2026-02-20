package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// DrawPlayerMissile renders the player's missile and its trail.
func DrawPlayerMissile(screen draw.Image, m *logic.PlayerMissile) {
	if !m.Active {
		return
	}

	sm := assets.SpritePlayerMissile
	drawSprite(screen, sm, m.X, m.Y, colorMissile, false)

	// Trail behind the missile.
	st := assets.SpritePlayerMissileTrail
	drawSprite(screen, st, m.X, m.Y+sm.Height(), colorMissile, false)
}

// DrawTankShell renders the tank shell or its explosion.
func DrawTankShell(screen draw.Image, ts *logic.TankShell) {
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

// DrawHeliMissile renders the helicopter missile as a small projectile.
func DrawHeliMissile(screen draw.Image, hm *logic.HeliMissile) {
	if !hm.Active {
		return
	}

	ink := palette[platform.ColorYellow]
	screen.Set(hm.X, hm.Y, ink)
	screen.Set(hm.X+1, hm.Y, ink)
}
