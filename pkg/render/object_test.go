package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TestDrawExplosionFragments_ConvertsToScreenPixels verifies that
// drawExplosionFragments shifts SP coordinates to screen pixels before drawing.
func TestDrawExplosionFragments_ConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10 // screen pixels
	ex := state.Explosion{
		Fragments: []state.ExplosionFragment{
			{X: xPx * domain.SubpixelScale, Y: yPx * domain.SubpixelScale},
		},
		Frame: 0,
	}
	img := newTestImage()

	drawExplosionFragments(img, ex)

	s := assets.SpriteExplosions[explosionSpriteIndex[0]]
	if !hasAnyInk(img, image.Rect(xPx, yPx, xPx+s.Width, yPx+s.Height)) {
		t.Error("no ink at screen-pixel position")
	}
	if hasAnyInk(img, image.Rect(xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, xPx*domain.SubpixelScale+s.Width, yPx*domain.SubpixelScale+s.Height)) {
		t.Error("unexpected ink at raw SP position")
	}
}

// TestDrawViewportSlots_ConvertsToScreenPixels verifies that drawViewportSlots
// shifts SP coordinates to screen pixels before drawing objects.
func TestDrawViewportSlots_ConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10 // screen pixels
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X:    xPx * domain.SubpixelScale,
		Y:    yPx * domain.SubpixelScale,
		Type: domain.ObjectShip,
	})
	img := newTestImage()

	s := assets.SpriteObjects[domain.ObjectShip]
	drawViewportSlots(img, vp, 0, domain.GameplayNormal, nil, 0)

	if !hasAnyInk(img, image.Rect(xPx, yPx, xPx+s.Width, yPx+s.Height)) {
		t.Error("no ink at screen-pixel position")
	}
	if hasAnyInk(img, image.Rect(xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, xPx*domain.SubpixelScale+s.Width, yPx*domain.SubpixelScale+s.Height)) {
		t.Error("unexpected ink at raw SP position")
	}
}

// TestDrawViewportSlots_RockConvertsToScreenPixels verifies that rocks (which
// take a separate path in drawViewportSlots) are also drawn at screen pixels.
func TestDrawViewportSlots_RockConvertsToScreenPixels(t *testing.T) {
	t.Parallel()

	const xPx, yPx = 8, 10
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X:           xPx * domain.SubpixelScale,
		Y:           yPx * domain.SubpixelScale,
		IsRock:      true,
		RockVariant: 0,
	})
	img := newTestImage()

	s := assets.SpriteRocks[0]
	drawViewportSlots(img, vp, 0, domain.GameplayNormal, nil, 0)

	if !hasAnyInk(img, image.Rect(xPx, yPx, xPx+s.Width, yPx+s.Height)) {
		t.Error("no ink at screen-pixel position for rock")
	}
	if hasAnyInk(img, image.Rect(xPx*domain.SubpixelScale, yPx*domain.SubpixelScale, xPx*domain.SubpixelScale+s.Width, yPx*domain.SubpixelScale+s.Height)) {
		t.Error("unexpected ink at raw SP position for rock")
	}
}

// firstInkColor returns the first non-black pixel color found within r.
func firstInkColor(img *image.RGBA, r image.Rectangle) (color.RGBA, bool) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c != (color.RGBA{}) {
				return c, true
			}
		}
	}
	return color.RGBA{}, false
}

// TestDrawObject_FuelDepotBlink verifies that the fuel depot toggles between its
// normal color and the blinking color at the spec-required rate.
// Spec (§16): toggle every 4 original frames at ~12 fps → 3 Hz toggle → domain.Tps/3 ticks.
func TestDrawObject_FuelDepotBlink(t *testing.T) {
	t.Parallel()

	// wantToggleAt is the first tick at which the depot must switch to blinking color.
	// Derived from spec directly, not from fuelBlinkEvery, so a misconfigured constant fails.
	const wantToggleAt = domain.Tps / 3

	const x, y domain.Px = 8, 10
	s := assets.SpriteObjects[domain.ObjectFuel]
	r := image.Rect(int(x), int(y), int(x)+s.Width, int(y)+s.Height)

	// tick=0: first half of cycle → normal color.
	img0 := newTestImage()
	drawObject(img0, x, y, domain.ObjectFuel, domain.OrientationLeft, 0, domain.GameplayNormal, domain.TankLocationRoad, nil, 0)
	c0, ok0 := firstInkColor(img0, r)
	if !ok0 {
		t.Fatal("tick=0: no ink pixels drawn")
	}
	if c0 != palette[colorFuel] {
		t.Errorf("tick=0: got %v, want colorFuel %v", c0, palette[colorFuel])
	}

	// tick=wantToggleAt: second half of cycle → blinking color.
	img1 := newTestImage()
	drawObject(img1, x, y, domain.ObjectFuel, domain.OrientationLeft, wantToggleAt, domain.GameplayNormal, domain.TankLocationRoad, nil, 0)
	c1, ok1 := firstInkColor(img1, r)
	if !ok1 {
		t.Fatalf("tick=%d: no ink pixels drawn", wantToggleAt)
	}
	if c1 != palette[colorFuelBlinking] {
		t.Errorf("tick=%d: got %v, want colorFuelBlinking %v", wantToggleAt, c1, palette[colorFuelBlinking])
	}

	// tick=wantToggleAt*2: back to normal color.
	img2 := newTestImage()
	drawObject(img2, x, y, domain.ObjectFuel, domain.OrientationLeft, wantToggleAt*2, domain.GameplayNormal, domain.TankLocationRoad, nil, 0)
	c2, ok2 := firstInkColor(img2, r)
	if !ok2 {
		t.Fatalf("tick=%d: no ink pixels drawn", wantToggleAt*2)
	}
	if c2 != palette[colorFuel] {
		t.Errorf("tick=%d: got %v, want colorFuel %v", wantToggleAt*2, c2, palette[colorFuel])
	}
}

// TestDrawObject_FuelDepotNoBlinkDuringScrollIn verifies that the fuel depot
// does not blink during scroll-in (animate=false).
func TestDrawObject_FuelDepotNoBlinkDuringScrollIn(t *testing.T) {
	t.Parallel()

	const x, y domain.Px = 8, 10
	s := assets.SpriteObjects[domain.ObjectFuel]
	r := image.Rect(int(x), int(y), int(x)+s.Width, int(y)+s.Height)

	img := newTestImage()
	drawObject(img, x, y, domain.ObjectFuel, domain.OrientationLeft, fuelBlinkEvery, domain.GameplayScrollIn, domain.TankLocationRoad, nil, 0)
	got, ok := firstInkColor(img, r)
	if !ok {
		t.Fatal("scroll-in: no ink pixels drawn")
	}
	if got != palette[colorFuel] {
		t.Errorf("scroll-in: got %v, want colorFuel %v (no blink during scroll-in)", got, palette[colorFuel])
	}
}

// TestDrawExplosionFragments_EmptyFrame draws nothing when Frame is out of range.
func TestDrawExplosionFragments_EmptyFrame(t *testing.T) {
	t.Parallel()

	ex := state.Explosion{
		Fragments: []state.ExplosionFragment{{X: 8 * domain.SubpixelScale, Y: 10 * domain.SubpixelScale}},
		Frame:     -1,
	}
	img := newTestImage()

	drawExplosionFragments(img, ex)

	if hasAnyInk(img, image.Rect(0, 0, 256, 256)) {
		t.Error("expected no ink for out-of-range frame")
	}
}
