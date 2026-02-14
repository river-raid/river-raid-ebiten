package main

import (
	"image/color"
	"image/draw"
)

// Sprite holds a 1-bit-per-pixel bitmap and its pixel dimensions.
// The bitmap is stored as ceil(Width/8) bytes per row, MSB first.
type Sprite struct {
	data          []byte
	Width, Height int
	bytesPerRow   int
}

// SpriteID identifies a sprite in the catalog.
type SpriteID int

// Sprite catalog entries.
const (
	SpritePlayerLevel SpriteID = iota
	SpritePlayerBanked
	SpritePlayerMissile
	SpriteMissileTrail
	SpriteExplosionSmall
	SpriteExplosionMedium
	SpriteExplosionLarge
	SpriteRock0
	SpriteRock1
	SpriteRock2
	SpriteRock3
	SpriteHelicopterReg
	SpriteShip
	SpriteHelicopterAdv
	SpriteTankBody
	SpriteFighter
	SpriteTankCaterpillar0
	SpriteTankCaterpillar1
	SpriteTankCaterpillar2
	SpriteBalloon
	SpriteFuelDepot
	SpriteRotorLeft
	SpriteRotorRight
	SpriteShellExplosion0
	SpriteShellExplosion1
	SpriteShellExplosion2
	SpriteShellExplosion3
	SpriteShellExplosion4
	SpriteShellExplosion5
	spriteCount
)

// SpriteCatalog contains all game sprites, indexed by SpriteID.
//
//nolint:mnd // sprite dimensions are inherent to the extracted data
var SpriteCatalog = [spriteCount]Sprite{
	SpritePlayerLevel:      newSprite(spritePlayerLevel[:], 8, 8),
	SpritePlayerBanked:     newSprite(spritePlayerBanked[:], 8, 8),
	SpritePlayerMissile:    newSprite(spritePlayerMissile[:], 2, 8),
	SpriteMissileTrail:     newSprite(spriteMissileTrail[:], 2, 8),
	SpriteExplosionSmall:   newSprite(spriteExplosionSmall[:], 10, 8),
	SpriteExplosionMedium:  newSprite(spriteExplosionMedium[:], 10, 8),
	SpriteExplosionLarge:   newSprite(spriteExplosionLarge[:], 10, 8),
	SpriteRock0:            newSprite(spriteRock0[:], 18, 16),
	SpriteRock1:            newSprite(spriteRock1[:], 18, 16),
	SpriteRock2:            newSprite(spriteRock2[:], 18, 16),
	SpriteRock3:            newSprite(spriteRock3[:], 18, 16),
	SpriteHelicopterReg:    newSprite(spriteHelicopterReg[:], 10, 8),
	SpriteShip:             newSprite(spriteShip[:], 18, 8),
	SpriteHelicopterAdv:    newSprite(spriteHelicopterAdv[:], 10, 8),
	SpriteTankBody:         newSprite(spriteTankBody[:], 10, 8),
	SpriteFighter:          newSprite(spriteFighter[:], 10, 8),
	SpriteTankCaterpillar0: newSprite(spriteTankCaterpillar0[:], 10, 3),
	SpriteTankCaterpillar1: newSprite(spriteTankCaterpillar1[:], 10, 3),
	SpriteTankCaterpillar2: newSprite(spriteTankCaterpillar2[:], 10, 3),
	SpriteBalloon:          newSprite(spriteBalloon[:], 10, 16),
	SpriteFuelDepot:        newSprite(spriteFuelDepot[:], 16, 25),
	SpriteRotorLeft:        newSprite(spriteRotorLeft[:], 10, 2),
	SpriteRotorRight:       newSprite(spriteRotorRight[:], 10, 2),
	SpriteShellExplosion0:  newSprite(spriteShellExplosion0[:], 10, 16),
	SpriteShellExplosion1:  newSprite(spriteShellExplosion1[:], 10, 16),
	SpriteShellExplosion2:  newSprite(spriteShellExplosion2[:], 10, 16),
	SpriteShellExplosion3:  newSprite(spriteShellExplosion3[:], 10, 16),
	SpriteShellExplosion4:  newSprite(spriteShellExplosion4[:], 10, 16),
	SpriteShellExplosion5:  newSprite(spriteShellExplosion5[:], 10, 16),
}

// newSprite creates a Sprite from raw 1bpp bitmap data.
// w and h are the visual width and height in pixels.
func newSprite(data []byte, w, h int) Sprite {
	bpr := (w + 7) / 8 //nolint:mnd // ceiling division by 8 bits per byte

	return Sprite{
		data:        data,
		Width:       w,
		Height:      h,
		bytesPerRow: bpr,
	}
}

// drawSprite draws a sprite at pixel position (x, y) onto screen.
// Set bits are drawn in ink color; unset bits are left unchanged (transparent).
// If mirror is true, the sprite is flipped horizontally.
func drawSprite(screen draw.Image, s Sprite, x, y int, ink color.RGBA, mirror bool) {
	for row := range s.Height {
		for col := range s.Width {
			byteIdx := row*s.bytesPerRow + col/8 //nolint:mnd // 8 bits per byte
			bitIdx := 7 - col%8                  //nolint:mnd // MSB first, 8 bits per byte

			if s.data[byteIdx]&(1<<bitIdx) != 0 {
				px := col
				if mirror {
					px = s.Width - 1 - col
				}

				screen.Set(x+px, y+row, ink)
			}
		}
	}
}
