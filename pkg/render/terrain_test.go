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

func TestTerrainBuffer_EdgeAt(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	// Manually set an edge at buffer Y=5.
	tb.edges[5] = TerrainEdges{LeftX: 20, RightX: 200, HasIsland: false}

	got := tb.EdgeAt(5)
	if got.LeftX != 20 || got.RightX != 200 {
		t.Errorf("EdgeAt(5): got {%d, %d}, want {20, 200}", got.LeftX, got.RightX)
	}
}

func TestTerrainBuffer_EdgeAt_Wraps(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	height := len(tb.edges)
	// Set an edge at the last slot.
	tb.edges[height-1] = TerrainEdges{LeftX: 10, RightX: 100}

	// Query with bufY = -1 (wraps to height-1).
	got := tb.EdgeAt(-1)
	if got.LeftX != 10 || got.RightX != 100 {
		t.Errorf("EdgeAt(-1): got {%d, %d}, want {10, 100}", got.LeftX, got.RightX)
	}

	// Query with bufY = height (wraps to 0).
	tb.edges[0] = TerrainEdges{LeftX: 50, RightX: 150}
	got = tb.EdgeAt(height)
	if got.LeftX != 50 || got.RightX != 150 {
		t.Errorf("EdgeAt(height): got {%d, %d}, want {50, 150}", got.LeftX, got.RightX)
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
