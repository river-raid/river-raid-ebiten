package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

const (
	bladesHz              = 12                    // blade alternation frequency
	bladesAlterationEvery = domain.Tps / bladesHz // ticks between blade state changes

	fuelBlinkHz    = 3                        // fuel blink toggle rate (→ 1.5 Hz visual cycle)
	fuelBlinkEvery = domain.Tps / fuelBlinkHz // ticks between blink state changes

	tankCaterpillarCycleSize = 4
)

// tankCaterpillarFrames maps X-position-based frame index to caterpillar sprite frame.
var tankCaterpillarFrames = [tankCaterpillarCycleSize]int{0, 1, 0, 2}

// roadTankColorFn returns the road or bridge color for each pixel column.
// A tank straddling the road–bridge boundary is rendered in both colors.
// In the original game, the road tank doesn't have its own colors: it uses the bank color while on the road
// and the river color while on the bridge – this allows to avoid attribute clashing during vertical scrolling.
var roadTankColorFn ColorFn = func(x, _ domain.Px) platform.Color { //nolint:gochecknoglobals // constant lookup, package-level by design
	if x >= bridgeStartX && x < bridgeEndX {
		return colorRiver
	}

	return colorBank
}

// fighterColorFn returns a ColorFn that picks the fighter's apparent XOR color
// based on the terrain beneath each pixel.
func fighterColorFn(tb *TerrainBuffer, scrollY domain.Px) ColorFn {
	return func(x, y domain.Px) platform.Color {
		edge := tb.EdgeAt(scrollY + y)
		onBank := x < edge.LeftX || x >= edge.RightX ||
			(edge.HasIsland && x >= edge.IslandLeftX && x < edge.IslandRightX)
		if onBank {
			return colorRiver
		}

		return colorBank
	}
}

// drawViewportSlots renders all active objects in the viewport.
// tick is the global game tick (s.Tick) used for real-time animations.
func drawViewportSlots(screen draw.Image, vp *state.Viewport, tick int, mode domain.GameplayMode, tb *TerrainBuffer, scrollY domain.Px) {
	for i := range vp.Objects {
		obj := vp.Objects[i]

		x := obj.X.ToPx()
		y := obj.Y.ToPx()

		// Handle rocks separately (they use different sprite selection logic).
		if obj.IsRock {
			drawRock(screen, x, y, obj.RockVariant)
		} else {
			drawObject(screen, x, y, obj.Type, obj.Orientation, tick, mode, obj.TankLocation, tb, scrollY)
		}
	}
}

// drawRock renders a rock.
func drawRock(screen draw.Image, x, y domain.Px, variant int) {
	s := assets.SpriteRocks[variant]
	drawSprite(screen, s, x, y, staticColorFn(colorRock), false)
}

// drawObject renders an interactive object.
func drawObject(screen draw.Image, x, y domain.Px, typ domain.ObjectType, orientation domain.Orientation, tick int, mode domain.GameplayMode, tankLocation domain.TankLocation, tb *TerrainBuffer, scrollY domain.Px) {
	s := assets.SpriteObjects[typ]
	mirror := orientation == domain.OrientationRight
	animate := mode != domain.GameplayScrollIn

	var colorFn ColorFn

	switch typ {
	case domain.ObjectHelicopterReg, domain.ObjectHelicopterAdv:
		colorFn = staticColorFn(colorHelicopter)
	case domain.ObjectShip:
		colorFn = staticColorFn(colorShip)
	case domain.ObjectBalloon:
		colorFn = staticColorFn(colorBalloon)
	case domain.ObjectFuel:
		mirror = false
		if animate && (tick/fuelBlinkEvery)%2 != 0 {
			colorFn = staticColorFn(colorFuelBlinking)
		} else {
			colorFn = staticColorFn(colorFuel)
		}
	case domain.ObjectFighter:
		colorFn = fighterColorFn(tb, scrollY)
	case domain.ObjectTank:
		if tankLocation == domain.TankLocationRoad {
			colorFn = roadTankColorFn
		} else {
			colorFn = staticColorFn(colorTankOnBank)
		}
	}

	drawSprite(screen, s, x, y, colorFn, mirror)

	// Helicopter blades overlay.
	if typ == domain.ObjectHelicopterReg || typ == domain.ObjectHelicopterAdv {
		frameIdx := 0
		if animate && (tick/bladesAlterationEvery)%2 != 0 {
			frameIdx = 1
		}
		blades := assets.SpriteBladesFrames[frameIdx]
		drawSprite(screen, blades, x, y, colorFn, mirror)
	}

	// Tank caterpillar overlay.
	if typ == domain.ObjectTank {
		frameIdx := int(x) % tankCaterpillarCycleSize
		catSprite := assets.SpriteTankCaterpillarFrames[tankCaterpillarFrames[frameIdx]]
		catY := y + domain.Px(s.Height) - domain.Px(catSprite.Height)
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
		drawSprite(screen, assets.SpriteExplosions[idx], f.X.ToPx(), f.Y.ToPx(), staticColorFn(colorExplosion), false)
	}
}
