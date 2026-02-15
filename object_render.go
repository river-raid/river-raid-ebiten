package main

import (
	"image/color"
	"image/draw"
)

// fuelBlinkInterval is the tick mask for fuel depot blinking (every 4 ticks).
const fuelBlinkInterval = 4

// tankCaterpillarFrames maps X-position-based frame index to caterpillar sprite.
var tankCaterpillarFrames = [4]SpriteID{ //nolint:gochecknoglobals // constant table
	SpriteTankCaterpillar0,
	SpriteTankCaterpillar1,
	SpriteTankCaterpillar2,
	SpriteTankCaterpillar1,
}

// objectSpriteInfo returns the sprite ID and ink color for a given object type.
func objectSpriteInfo(objType ObjectType, tick int) (SpriteID, color.RGBA) {
	switch objType {
	case ObjectHelicopterReg:
		return SpriteHelicopterReg, Palette[ColorHelicopter]
	case ObjectShip:
		return SpriteShip, Palette[ColorShip]
	case ObjectHelicopterAdv:
		return SpriteHelicopterAdv, Palette[ColorHelicopter]
	case ObjectTank:
		return SpriteTankBody, Palette[ColorRiver] // XOR on green bank → blue
	case ObjectFighter:
		return SpriteFighter, Palette[ColorRiver] // XOR on green bank → blue
	case ObjectBalloon:
		return SpriteBalloon, Palette[ColorBalloon]
	case ObjectFuel:
		ink := Palette[ColorFuel]
		if tick&fuelBlinkInterval != 0 {
			ink = Palette[ColorWhite]
		}

		return SpriteFuelDepot, ink
	default:
		return SpritePlayerLevel, Palette[ColorWhite] // fallback
	}
}

// drawViewportSlots renders all active objects in the viewport.
func drawViewportSlots(screen draw.Image, vp *Viewport) {
	for i := range vp.Slots {
		slot := &vp.Slots[i]
		spriteID, ink := objectSpriteInfo(slot.Type, vp.Tick)
		mirror := slot.Orientation == OrientationLeft

		s := SpriteCatalog[spriteID]
		drawSprite(screen, s, slot.X, slot.Y, ink, mirror)

		// Helicopter rotor overlay.
		if slot.Type == ObjectHelicopterReg || slot.Type == ObjectHelicopterAdv {
			rotorID := SpriteRotorRight
			if mirror {
				rotorID = SpriteRotorLeft
			}

			rotor := SpriteCatalog[rotorID]
			drawSprite(screen, rotor, slot.X, slot.Y-rotor.Height(), ink, false)
		}

		// Tank caterpillar overlay.
		if slot.Type == ObjectTank {
			frameIdx := (slot.X / 2) & 0x03 //nolint:mnd // 4-frame palindromic cycle from 2px shifts
			catSprite := SpriteCatalog[tankCaterpillarFrames[frameIdx]]
			catY := slot.Y + s.Height()
			drawSprite(screen, catSprite, slot.X, catY, ink, mirror)
		}
	}
}
