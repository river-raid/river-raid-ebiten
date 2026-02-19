package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// fuelBlinkInterval is the tick mask for fuel depot blinking (every 4 ticks).
const fuelBlinkInterval = 4

const tankCaterpillarCycleSize = 4

// tankCaterpillarFrames maps X-position-based frame index to caterpillar sprite frame.
var tankCaterpillarFrames = [tankCaterpillarCycleSize]int{0, 1, 2, 1}

// DrawViewportSlots renders all active objects in the viewport.
func DrawViewportSlots(screen draw.Image, vp *state.Viewport) {
	for i := range vp.Slots {
		slot := &vp.Slots[i]

		// Handle rocks separately (they use different sprite selection logic).
		if slot.IsRock {
			drawRock(screen, slot.X, slot.Y, slot.RockVariant)
		} else {
			drawObject(screen, slot.X, slot.Y, slot.Type, slot.Orientation, vp.Tick)
		}
	}
}

// drawRock renders a rock.
func drawRock(screen draw.Image, x, y, variant int) {
	s := assets.SpriteRocks[variant]
	drawSprite(screen, s, x, y, colorRock, false)
}

// drawObject renders an interactive object.
func drawObject(screen draw.Image, x, y int, typ domain.ObjectType, orientation domain.Orientation, tick int) {
	s := assets.SpriteObjects[typ]
	ink := objectColors[typ]
	mirror := false

	if typ == domain.ObjectFuel {
		if tick&fuelBlinkInterval != 0 {
			ink = colorFuelBlinking
		}
	} else if orientation == domain.OrientationRight {
		mirror = true
	}

	drawSprite(screen, s, x, y, ink, mirror)

	// Helicopter rotor overlay.
	if typ == domain.ObjectHelicopterReg || typ == domain.ObjectHelicopterAdv {
		rotor := assets.SpriteRotorFrames[orientation]
		drawSprite(screen, rotor, x, y, ink, mirror)
	}

	// Tank caterpillar overlay.
	if typ == domain.ObjectTank {
		frameIdx := (x / logic.EnemyMoveStep) % tankCaterpillarCycleSize
		catSprite := assets.SpriteTankCaterpillarFrames[tankCaterpillarFrames[frameIdx]]
		catY := y + s.Height()
		drawSprite(screen, catSprite, x, catY, ink, mirror)
	}
}
