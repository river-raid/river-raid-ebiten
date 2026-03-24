package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// hasAnyInk reports whether any pixel in r is non-black.
func hasAnyInk(img *image.RGBA, r image.Rectangle) bool {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				return true
			}
		}
	}
	return false
}

// newTestImage returns a blank 256×256 RGBA image, large enough to reveal
// both the correct screen-pixel position and the raw-SP position.
func newTestImage() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, 256, 256))
}

// TestDrawHeliMissile_ConvertsToScreenPixels verifies that drawHeliMissile
// shifts SP coordinates to screen pixels before drawing.
func TestDrawHeliMissile_ConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10 // screen pixels
	hm := &state.HeliMissile{
		X:           xPx * domain.SubpixelScale,
		Y:           yPx * domain.SubpixelScale,
		Active:      true,
		Orientation: domain.OrientationRight,
	}
	img := newTestImage()

	drawHeliMissile(img, hm)

	ink := palette[platform.ColorGreen]
	assertColor(t, img, xPx, yPx, ink, "ink at screen position")
	assertColor(t, img, xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, color.RGBA{}, "no ink at raw SP position")
}

// TestDrawTankShell_ConvertsToScreenPixels verifies that drawTankShell
// shifts SP coordinates to screen pixels before drawing.
func TestDrawTankShell_ConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10 // screen pixels
	ts := &state.TankShell{
		X:        xPx * domain.SubpixelScale,
		Y:        yPx * domain.SubpixelScale,
		IsFlying: true,
	}
	img := newTestImage()

	drawTankShell(img, ts)

	ink := palette[platform.ColorWhite]
	assertColor(t, img, xPx, yPx, ink, "2×2 block top-left")
	assertColor(t, img, xPx+1, yPx, ink, "2×2 block top-right")
	assertColor(t, img, xPx, yPx+1, ink, "2×2 block bottom-left")
	assertColor(t, img, xPx+1, yPx+1, ink, "2×2 block bottom-right")
	assertColor(t, img, xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, color.RGBA{}, "no ink at raw SP position")
}

// TestDrawPlayerMissile_ConvertsToScreenPixels verifies that drawPlayerMissile
// shifts SP coordinates to screen pixels before drawing.
func TestDrawPlayerMissile_ConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10 // screen pixels
	m := &state.PlayerMissile{
		X:      xPx * domain.SubpixelScale,
		Y:      yPx * domain.SubpixelScale,
		Active: true,
	}
	img := newTestImage()

	drawPlayerMissile(img, m)

	// Missile sprite is 2×N; check that ink lands in the expected region.
	const missileW, missileH = 2, 8
	if !hasAnyInk(img, image.Rect(xPx, yPx, xPx+missileW, yPx+missileH)) {
		t.Error("no ink at screen-pixel position")
	}
	if hasAnyInk(img, image.Rect(xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, xPx*domain.SubpixelScale+missileW, yPx*domain.SubpixelScale+missileH)) {
		t.Error("unexpected ink at raw SP position")
	}
}
