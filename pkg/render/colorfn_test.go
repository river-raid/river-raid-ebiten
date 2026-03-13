package render

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func TestStaticColorFn(t *testing.T) {
	t.Parallel()

	fn := staticColorFn(platform.ColorCyan)
	for _, tc := range []struct{ x, y int }{{0, 0}, {100, 50}, {255, 135}} {
		if got := fn(tc.x, tc.y); got != platform.ColorCyan {
			t.Errorf("staticColorFn(%d,%d) = %d, want ColorCyan", tc.x, tc.y, got)
		}
	}
}

func TestFighterColorFn_OnBank(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	// River from x=50 to x=200 at buffer row 10.
	tb.edges[10] = TerrainEdges{LeftX: 50, RightX: 200}
	fn := fighterColorFn(tb, 0)

	// Left bank (x=10 < LeftX=50) → blue.
	if got := fn(10, 10); got != platform.ColorBlue {
		t.Errorf("left bank: got %d, want ColorBlue", got)
	}

	// Right bank (x=210 >= RightX=200) → blue.
	if got := fn(210, 10); got != platform.ColorBlue {
		t.Errorf("right bank: got %d, want ColorBlue", got)
	}
}

func TestFighterColorFn_OnRiver(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	tb.edges[10] = TerrainEdges{LeftX: 50, RightX: 200}
	fn := fighterColorFn(tb, 0)

	// River pixel (LeftX ≤ x < RightX) → green.
	if got := fn(100, 10); got != platform.ColorGreen {
		t.Errorf("river: got %d, want ColorGreen", got)
	}
}

func TestFighterColorFn_OnIsland(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	tb.edges[20] = TerrainEdges{
		LeftX:        30,
		RightX:       220,
		HasIsland:    true,
		IslandLeftX:  100,
		IslandRightX: 150,
	}
	fn := fighterColorFn(tb, 0)

	// Island pixel (IslandLeftX ≤ x < IslandRightX) → blue.
	if got := fn(120, 20); got != platform.ColorBlue {
		t.Errorf("island: got %d, want ColorBlue", got)
	}

	// River shoulder pixel (between left bank and island) → green.
	if got := fn(60, 20); got != platform.ColorGreen {
		t.Errorf("river shoulder: got %d, want ColorGreen", got)
	}
}

func TestFighterColorFn_ScrollYOffset(t *testing.T) {
	t.Parallel()

	tb := NewTerrainBuffer()
	// Edge stored at buffer row 18 (scrollY=8, viewport y=10 → bufY=18).
	tb.edges[18] = TerrainEdges{LeftX: 40, RightX: 180}
	fn := fighterColorFn(tb, 8)

	// At viewport game-y=10, bufY = 8+10 = 18 → bank pixel at x=10 → blue.
	if got := fn(10, 10); got != platform.ColorBlue {
		t.Errorf("with scrollY offset: got %d, want ColorBlue", got)
	}
}

func TestRoadTankColorFn_Road(t *testing.T) {
	t.Parallel()

	got := roadTankColorFn(0, 0)
	if got != colorRoad {
		t.Errorf("road column: got %d, want colorRoad (%d)", got, colorRoad)
	}
}

func TestRoadTankColorFn_Bridge(t *testing.T) {
	t.Parallel()

	got := roadTankColorFn(bridgeStartX, 0)
	if got != colorBridge {
		t.Errorf("bridge column: got %d, want colorBridge (%d)", got, colorBridge)
	}
}

func TestRoadTankColorFn_BridgeEdge(t *testing.T) {
	t.Parallel()

	// bridgeEndX is the first column past the bridge band → road color.
	got := roadTankColorFn(bridgeEndX, 0)
	if got != colorRoad {
		t.Errorf("past bridge end: got %d, want colorRoad (%d)", got, colorRoad)
	}
}
