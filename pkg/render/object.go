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

// roadTankColorFn returns the road or bridge color for each pixel column.
// A tank straddling the road–bridge boundary is rendered in both colors.
var roadTankColorFn ColorFn = func(x, _ int) platform.Color { //nolint:gochecknoglobals // constant lookup, package-level by design
	if x >= bridgeStartX && x < bridgeEndX {
		return colorBridge
	}

	return colorRoad
}

// fighterColorFn returns a ColorFn that picks the fighter's apparent XOR color
// based on the terrain beneath each pixel.
func fighterColorFn(tb *TerrainBuffer, scrollY int) ColorFn {
	return func(x, y int) platform.Color {
		edge := tb.EdgeAt(scrollY + y)
		onBank := x < edge.LeftX || x >= edge.RightX ||
			(edge.HasIsland && x >= edge.IslandLeftX && x < edge.IslandRightX)
		if onBank {
			return platform.ColorBlue // bank is green; XOR → paper = blue
		}

		return platform.ColorGreen // river is blue; XOR → ink = green
	}
}

// drawViewportSlots renders all active objects in the viewport.
func drawViewportSlots(screen draw.Image, vp *state.Viewport, mode domain.GameplayMode, tb *TerrainBuffer, scrollY int) {
	for i := range vp.Objects {
		obj := vp.Objects[i]

		// Handle rocks separately (they use different sprite selection logic).
		if obj.IsRock {
			drawRock(screen, obj.X, obj.Y, obj.RockVariant)
		} else {
			drawObject(screen, obj.X, obj.Y, obj.Type, obj.Orientation, vp.Tick, mode, obj.TankLocation, tb, scrollY)
		}
	}
}

// drawRock renders a rock.
func drawRock(screen draw.Image, x, y, variant int) {
	s := assets.SpriteRocks[variant]
	drawSprite(screen, s, x, y, staticColorFn(colorRock), false)
}

// drawObject renders an interactive object.
func drawObject(screen draw.Image, x, y int, typ domain.ObjectType, orientation domain.Orientation, tick int, mode domain.GameplayMode, tankLocation domain.TankLocation, tb *TerrainBuffer, scrollY int) {
	s := assets.SpriteObjects[typ]
	mirror := false
	animate := mode != domain.GameplayScrollIn

	var colorFn ColorFn

	switch typ {
	case domain.ObjectFuel:
		if animate && tick&fuelBlinkInterval != 0 {
			colorFn = staticColorFn(colorFuelBlinking)
		} else {
			colorFn = staticColorFn(objectColors[typ])
		}
	case domain.ObjectFighter:
		colorFn = fighterColorFn(tb, scrollY)
		if orientation == domain.OrientationRight {
			mirror = true
		}
	case domain.ObjectTank:
		if tankLocation == domain.TankLocationRoad {
			colorFn = roadTankColorFn
		} else {
			colorFn = staticColorFn(platform.ColorBlue)
		}
		if orientation == domain.OrientationRight {
			mirror = true
		}
	default:
		colorFn = staticColorFn(objectColors[typ])
		if orientation == domain.OrientationRight {
			mirror = true
		}
	}

	drawSprite(screen, s, x, y, colorFn, mirror)

	// Helicopter blades overlay.
	if typ == domain.ObjectHelicopterReg || typ == domain.ObjectHelicopterAdv {
		frameIdx := 0
		if animate && tick&bladesAlterationInterval != 0 {
			frameIdx = 1
		}
		blades := assets.SpriteBladesFrames[frameIdx]
		drawSprite(screen, blades, x, y, colorFn, mirror)
	}

	// Tank caterpillar overlay.
	if typ == domain.ObjectTank {
		frameIdx := (x / logic.EnemyMoveStep) % tankCaterpillarCycleSize
		catSprite := assets.SpriteTankCaterpillarFrames[tankCaterpillarFrames[frameIdx]]
		catY := y + s.Height - catSprite.Height
		drawSprite(screen, catSprite, x, catY, colorFn, mirror)
	}
}

// explosionSpriteIndex maps an animation frame to a SpriteExplosions index.
//
//	Frame 0,4 → Small  (index 0)
//	Frame 1,3 → Medium (index 1)
//	Frame 2   → Large  (index 2)
var explosionSpriteIndex = [domain.NumExplosionSpriteFrames]int{0, 1, 2, 1, 0} //nolint:gochecknoglobals // constant lookup table

// drawExplosionFragments renders all active explosion fragments.
// All fragments share the same animation frame stored in ex.Frame.
func drawExplosionFragments(screen draw.Image, ex state.Explosion) {
	if len(ex.Fragments) == 0 || ex.Frame < 0 || ex.Frame >= len(explosionSpriteIndex) {
		return
	}

	idx := explosionSpriteIndex[ex.Frame]

	for _, f := range ex.Fragments {
		drawSprite(screen, assets.SpriteExplosions[idx], f.X, f.Y, staticColorFn(platform.ColorGreen), false)
	}
}
