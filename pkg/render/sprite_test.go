package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func TestDrawSprite_InkPixels(t *testing.T) {
	// 8x8 test sprite. Row 0 is 0x10 = 00010000 — only bit 4 set (pixel x=3).
	s := assets.Sprite{
		Data:        []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Width:       8,
		BytesPerRow: 1,
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	drawSprite(img, s, 0, 0, platform.ColorYellow, false)

	// Pixel (3, 0) should be ink (bit 4 of 0x10 is set).
	assertColor(t, img, 3, 0, palette[platform.ColorYellow], "ink at (3,0)")

	// Pixel (0, 0) should be untouched (transparent / zero).
	assertColor(t, img, 0, 0, color.RGBA{}, "transparent at (0,0)")
}

func TestDrawSprite_Position(t *testing.T) {
	// Draw 8x8 test sprite at offset (4, 2).
	s := assets.Sprite{
		Data:        []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Width:       8,
		BytesPerRow: 1,
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	drawSprite(img, s, 4, 2, platform.ColorYellow, false)

	// Row 0 of sprite (0x10): bit 4 set → pixel x=3. At offset (4,2), pixel (7, 2).
	assertColor(t, img, 7, 2, palette[platform.ColorYellow], "ink at offset")

	// Origin should be untouched.
	assertColor(t, img, 0, 0, color.RGBA{}, "origin untouched")
}

func TestDrawSprite_Mirror(t *testing.T) {
	// 8x8 test sprite. Row 0 is 0x10 = 00010000 — bit 4 set → pixel x=3.
	// Mirrored: x = width-1-3 = 4.
	s := assets.Sprite{
		Data:        []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Width:       8,
		BytesPerRow: 1,
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	drawSprite(img, s, 0, 0, platform.ColorYellow, true)

	// Mirrored pixel should be at x=4.
	assertColor(t, img, 4, 0, palette[platform.ColorYellow], "mirrored ink at (4,0)")

	// Original position x=3 should be transparent.
	assertColor(t, img, 3, 0, color.RGBA{}, "original position transparent")
}

func TestDrawSprite_WideSprite(t *testing.T) {
	// 10px wide (2 bytes/row).
	// Row 0: 0xf0 0x00 = 11110000 00000000
	// Bits 0-3 set (pixels x=0..3), rest clear.
	s := assets.Sprite{
		Data:        []byte{0xf0, 0x00},
		Width:       10,
		BytesPerRow: 2,
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	drawSprite(img, s, 0, 0, colorHelicopter, false)

	for x := range 4 {
		assertColor(t, img, x, 0, palette[colorHelicopter], "ink in rotor row")
	}

	// Pixel 4 should be transparent.
	assertColor(t, img, 4, 0, color.RGBA{}, "transparent after rotor")
}

func TestDrawSprite_Transparent(t *testing.T) {
	// Unset bits should not overwrite existing pixels.
	s := assets.Sprite{
		Data:        []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Width:       8,
		BytesPerRow: 1,
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	bg := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// Fill background.
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, bg)
		}
	}

	drawSprite(img, s, 0, 0, platform.ColorYellow, false)

	// Row 0, pixel 0: bit not set → background should remain.
	assertColor(t, img, 0, 0, bg, "background preserved")

	// Row 0, pixel 3: bit set → ink.
	assertColor(t, img, 3, 0, palette[platform.ColorYellow], "ink drawn")
}

func TestDrawSprite_MissileWidth(t *testing.T) {
	// Test sprite is 2px wide. Row 0 = 0xc0 = 11000000.
	// Only pixels 0 and 1 should be drawn (visual width is 2).
	s := assets.Sprite{
		Data:        []byte{0xc0},
		Width:       2,
		BytesPerRow: 1,
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	drawSprite(img, s, 0, 0, colorMissile, false)

	assertColor(t, img, 0, 0, palette[colorMissile], "missile pixel 0")
	assertColor(t, img, 1, 0, palette[colorMissile], "missile pixel 1")

	// Pixel 2 should be transparent (outside visual width).
	assertColor(t, img, 2, 0, color.RGBA{}, "outside missile width")
}
