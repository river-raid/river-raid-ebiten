package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func TestViewportCanvas_Clipping(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, domain.ViewportHeight+16))
	vc := newViewportCanvas(img)
	ink := palette[platform.ColorWhite]

	t.Run("AboveViewport", func(t *testing.T) {
		vc.Set(0, -1, ink)
		assertColor(t, img, 0, 0, color.RGBA{}, "y=-1 not drawn")
	})

	t.Run("TopEdge", func(t *testing.T) {
		vc.Set(0, 0, ink)
		assertColor(t, img, 0, 0, ink, "y=0 drawn")
	})

	t.Run("BottomEdge", func(t *testing.T) {
		vc.Set(0, domain.ViewportHeight-1, ink)
		assertColor(t, img, 0, domain.ViewportHeight-1, ink, "y=ViewportHeight-1 drawn")
	})

	t.Run("BelowViewport", func(t *testing.T) {
		vc.Set(0, domain.ViewportHeight, ink)
		assertColor(t, img, 0, domain.ViewportHeight, color.RGBA{}, "y=ViewportHeight not drawn")
	})
}
