package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func TestViewportCanvas_Clipping(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, domain.TotalViewportHeight+16))
	vc := newViewportCanvas(img)
	ink := palette[platform.ColorWhite]

	t.Run("AboveViewport", func(t *testing.T) {
		vc.Set(0, -1, ink)
		assertColor(t, img, 0, 0, color.RGBA{}, "y=-1 not drawn")
	})

	t.Run("InBlankZone", func(t *testing.T) {
		vc.Set(0, 0, ink) // game y=0 is in the blank zone
		assertColor(t, img, 0, 0, color.RGBA{}, "y=0 (blank zone) not drawn")
	})

	t.Run("TopEdge", func(t *testing.T) {
		vc.Set(0, domain.ViewportBlankZone, ink) // game y=8 → screen y=0
		assertColor(t, img, 0, 0, ink, "y=ViewportBlankZone drawn at screen y=0")
	})

	t.Run("BottomEdge", func(t *testing.T) {
		vc.Set(0, domain.TotalViewportHeight-1, ink) // game y=143 → screen y=VisibleViewportHeight-1
		assertColor(t, img, 0, domain.VisibleViewportHeight-1, ink, "y=TotalViewportHeight-1 drawn at screen y=VisibleViewportHeight-1")
	})

	t.Run("BelowViewport", func(t *testing.T) {
		vc.Set(0, domain.TotalViewportHeight, ink) // game y=144 → clipped
		assertColor(t, img, 0, domain.VisibleViewportHeight, color.RGBA{}, "y=TotalViewportHeight not drawn")
	})
}
