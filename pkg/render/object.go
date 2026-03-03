package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

const (
	bladesAlterationInterval = 2
	fuelBlinkInterval        = 4
	tankCaterpillarCycleSize = 4
)

// tankCaterpillarFrames maps X-position-based frame index to caterpillar sprite frame.
var tankCaterpillarFrames = [tankCaterpillarCycleSize]int{0, 1, 0, 2}

// drawViewportSlots renders all active objects in the viewport.
func drawViewportSlots(screen draw.Image, vp *state.Viewport, mode domain.GameplayMode) {
	for i := range vp.Objects {
		obj := vp.Objects[i]

		// Handle rocks separately (they use different sprite selection logic).
		if obj.IsRock {
			drawRock(screen, obj.X, obj.Y, obj.RockVariant)
		} else {
			drawObject(screen, obj.X, obj.Y, obj.Type, obj.Orientation, vp.Tick, mode)
		}
	}
}

// drawRock renders a rock.
func drawRock(screen draw.Image, x, y, variant int) {
	s := assets.SpriteRocks[variant]
	drawSprite(screen, s, x, y, colorRock, false)
}

// drawObject renders an interactive object.
func drawObject(screen draw.Image, x, y int, typ domain.ObjectType, orientation domain.Orientation, tick int, mode domain.GameplayMode) {
	s := assets.SpriteObjects[typ]
	ink := objectColors[typ]
	mirror := false
	animate := mode != domain.GameplayScrollIn

	if typ == domain.ObjectFuel {
		if animate && tick&fuelBlinkInterval != 0 {
			ink = colorFuelBlinking
		}
	} else if orientation == domain.OrientationRight {
		mirror = true
	}

	drawSprite(screen, s, x, y, ink, mirror)

	// Helicopter blades overlay.
	if typ == domain.ObjectHelicopterReg || typ == domain.ObjectHelicopterAdv {
		frameIdx := 0
		if animate && tick&bladesAlterationInterval != 0 {
			frameIdx = 1
		}
		blades := assets.SpriteBladesFrames[frameIdx]
		drawSprite(screen, blades, x, y, ink, mirror)
	}

	// Tank caterpillar overlay.
	if typ == domain.ObjectTank {
		frameIdx := (x / logic.EnemyMoveStep) % tankCaterpillarCycleSize
		catSprite := assets.SpriteTankCaterpillarFrames[tankCaterpillarFrames[frameIdx]]
		catY := y + s.Height - catSprite.Height
		drawSprite(screen, catSprite, x, catY, ink, mirror)
	}
}

// explosionSpriteIndex maps an animation frame (1–6) to a SpriteExplosions index.
// Frame 6 is the erase frame — no sprite is drawn; returns -1 as sentinel.
//
//	Frame 1,5 → Small  (index 0)
//	Frame 2,4 → Medium (index 1)
//	Frame 3   → Large  (index 2)
//	Frame 6   → Erase  (no draw)
var explosionSpriteIndex = [7]int{-1, 0, 1, 2, 1, 0, -1} //nolint:gochecknoglobals // constant lookup table

// drawExplosionFragments renders all active explosion fragments.
// Each fragment is drawn at its (X, Y) position using the sprite corresponding to its
// animation frame. Frame 6 (erase) is skipped — the fragment is no longer visible.
func drawExplosionFragments(screen draw.Image, fragments []state.ExplodingFragment) {
	for _, f := range fragments {
		if f.Frame < 1 || f.Frame >= len(explosionSpriteIndex) {
			continue
		}

		idx := explosionSpriteIndex[f.Frame]
		if idx < 0 {
			continue // erase frame — nothing to draw
		}

		drawSprite(screen, assets.SpriteExplosions[idx], f.X, f.Y, platform.ColorGreen, false)
	}
}
