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

		s := assets.SpriteObjects[slot.Type]
		ink := objectColors[slot.Type]
		mirror := false

		if slot.Type == domain.ObjectFuel {
			if vp.Tick&fuelBlinkInterval != 0 {
				ink = colorFuelBlinking
			}
		} else if slot.Orientation == domain.OrientationRight {
			mirror = true
		}

		drawSprite(screen, s, slot.X, slot.Y, ink, mirror)

		// Helicopter rotor overlay.
		if slot.Type == domain.ObjectHelicopterReg || slot.Type == domain.ObjectHelicopterAdv {
			rotor := assets.SpriteRotorFrames[slot.Orientation]
			drawSprite(screen, rotor, slot.X, slot.Y, ink, mirror)
		}

		// Tank caterpillar overlay.
		if slot.Type == domain.ObjectTank {
			frameIdx := (slot.X / logic.EnemyMoveStep) % tankCaterpillarCycleSize
			catSprite := assets.SpriteTankCaterpillarFrames[tankCaterpillarFrames[frameIdx]]
			catY := slot.Y + s.Height()
			drawSprite(screen, catSprite, slot.X, catY, ink, mirror)
		}
	}
}
