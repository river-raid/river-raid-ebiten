package main

import (
	"image"
	"image/color"
	"testing"
)

func TestNewSprite_Dimensions(t *testing.T) {
	tests := []struct {
		name                  string
		data                  []byte
		w, h                  int
		wantW, wantH, wantBPR int
	}{
		{
			name:    "8px wide",
			data:    spritePlayerLevel[:],
			w:       8,
			h:       8,
			wantW:   8,
			wantH:   8,
			wantBPR: 1,
		},
		{
			name:    "10px wide",
			data:    spriteHelicopterReg[:],
			w:       10,
			h:       8,
			wantW:   10,
			wantH:   8,
			wantBPR: 2,
		},
		{
			name:    "18px wide",
			data:    spriteShip[:],
			w:       18,
			h:       8,
			wantW:   18,
			wantH:   8,
			wantBPR: 3,
		},
		{
			name:    "2px wide",
			data:    spritePlayerMissile[:],
			w:       2,
			h:       8,
			wantW:   2,
			wantH:   8,
			wantBPR: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSprite(tt.data, tt.w, tt.h)
			if s.Width != tt.wantW {
				t.Errorf("Width = %d, want %d", s.Width, tt.wantW)
			}
			if s.Height != tt.wantH {
				t.Errorf("Height = %d, want %d", s.Height, tt.wantH)
			}
			if s.bytesPerRow != tt.wantBPR {
				t.Errorf("bytesPerRow = %d, want %d", s.bytesPerRow, tt.wantBPR)
			}
		})
	}
}

func TestDrawSprite_InkPixels(t *testing.T) {
	// spritePlayerLevel row 0 is 0x10 = 00010000 — only bit 4 set (pixel x=3).
	s := newSprite(spritePlayerLevel[:], 8, 8)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ink := Palette[ColorYellow]

	drawSprite(img, s, 0, 0, ink, false)

	// Pixel (3, 0) should be ink (bit 4 of 0x10 is set).
	assertColor(t, img, 3, 0, ink, "ink at (3,0)")

	// Pixel (0, 0) should be untouched (transparent / zero).
	assertColor(t, img, 0, 0, color.RGBA{}, "transparent at (0,0)")
}

func TestDrawSprite_Position(t *testing.T) {
	// Draw player sprite at offset (4, 2).
	s := newSprite(spritePlayerLevel[:], 8, 8)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ink := Palette[ColorYellow]

	drawSprite(img, s, 4, 2, ink, false)

	// Row 0 of sprite (0x10): bit 4 set → pixel x=3. At offset (4,2), pixel (7, 2).
	assertColor(t, img, 7, 2, ink, "ink at offset")

	// Origin should be untouched.
	assertColor(t, img, 0, 0, color.RGBA{}, "origin untouched")
}

func TestDrawSprite_Mirror(t *testing.T) {
	// spritePlayerLevel row 0 is 0x10 = 00010000 — bit 4 set → pixel x=3.
	// Mirrored: x = width-1-3 = 4.
	s := newSprite(spritePlayerLevel[:], 8, 8)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ink := Palette[ColorYellow]

	drawSprite(img, s, 0, 0, ink, true)

	// Mirrored pixel should be at x=4.
	assertColor(t, img, 4, 0, ink, "mirrored ink at (4,0)")

	// Original position x=3 should be transparent.
	assertColor(t, img, 3, 0, color.RGBA{}, "original position transparent")
}

func TestDrawSprite_WideSprite(t *testing.T) {
	// spriteHelicopterReg is 10px wide (2 bytes/row).
	// Row 0: 0xf0 0x00 = 11110000 00000000
	// Bits 0-3 set (pixels x=0..3), rest clear.
	s := newSprite(spriteHelicopterReg[:], 10, 8)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ink := Palette[ColorHelicopter]

	drawSprite(img, s, 0, 0, ink, false)

	for x := range 4 {
		assertColor(t, img, x, 0, ink, "ink in rotor row")
	}

	// Pixel 4 should be transparent.
	assertColor(t, img, 4, 0, color.RGBA{}, "transparent after rotor")
}

func TestDrawSprite_Transparent(t *testing.T) {
	// Unset bits should not overwrite existing pixels.
	s := newSprite(spritePlayerLevel[:], 8, 8)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	bg := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// Fill background.
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, bg)
		}
	}

	ink := Palette[ColorYellow]
	drawSprite(img, s, 0, 0, ink, false)

	// Row 0, pixel 0: bit not set → background should remain.
	assertColor(t, img, 0, 0, bg, "background preserved")

	// Row 0, pixel 3: bit set → ink.
	assertColor(t, img, 3, 0, ink, "ink drawn")
}

func TestDrawSprite_MissileWidth(t *testing.T) {
	// spritePlayerMissile is 2px wide. Row 0 = 0xc0 = 11000000.
	// Only pixels 0 and 1 should be drawn (visual width is 2).
	s := newSprite(spritePlayerMissile[:], 2, 8)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	ink := Palette[ColorMissile]

	drawSprite(img, s, 0, 0, ink, false)

	assertColor(t, img, 0, 0, ink, "missile pixel 0")
	assertColor(t, img, 1, 0, ink, "missile pixel 1")

	// Pixel 2 should be transparent (outside visual width).
	assertColor(t, img, 2, 0, color.RGBA{}, "outside missile width")
}
