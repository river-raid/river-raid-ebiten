package render

import (
	"image"
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func newTestTerrainBuffer(h int) (*TerrainBuffer, *image.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, platform.ScreenWidth, h))
	return &TerrainBuffer{buffer: img, edges: make([]TerrainEdges, h)}, img
}

func TestRenderBandedLines(t *testing.T) {
	t.Parallel()

	tb, img := newTestTerrainBuffer(1)
	tb.renderBandedLines(0, 1, colorBank, colorRoad)

	for x := range platform.ScreenWidth {
		got := img.RGBAAt(x, 0)
		want := palette[colorBank]
		if x >= bridgeStartX && x < bridgeEndX {
			want = palette[colorRoad]
		}
		if got != want {
			t.Errorf("x=%d: got %v, want %v", x, got, want)
		}
	}
}

func TestCalculateRightEdge_Mirrored(t *testing.T) {
	t.Parallel()

	// rightX = 2*center - leftX = 2*128 - 50 = 206
	got := calculateOtherEdge(128, 50, assets.EdgeMirrored)
	if got != 206 {
		t.Errorf("EdgeMirrored: got %d, want 206", got)
	}
}

func TestCalculateRightEdge_Offset(t *testing.T) {
	t.Parallel()

	// rightX = width + leftX = 64 + 50 = 114
	got := calculateOtherEdge(64, 50, assets.EdgeOffset)
	if got != 114 {
		t.Errorf("EdgeOffset: got %d, want 114", got)
	}
}
