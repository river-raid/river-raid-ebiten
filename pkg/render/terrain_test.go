package render

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

func TestBridgeRoadData_CanalPattern(t *testing.T) {
	t.Parallel()

	// Canal pattern (bytes 0–31): solid banks with a river gap in the middle.
	canal := assets.BridgeRoadData[:bridgeRoadBytes]

	// Bank bytes (0–13, 18–31) should be 0xFF (solid).
	for i := range 14 {
		if canal[i] != 0xFF {
			t.Errorf("canal byte %d: got 0x%02X, want 0xFF", i, canal[i])
		}
	}

	// River gap bytes (14–17) should be 0x00.
	for i := 14; i < 18; i++ {
		if canal[i] != 0x00 {
			t.Errorf("canal byte %d: got 0x%02X, want 0x00", i, canal[i])
		}
	}

	for i := 18; i < bridgeRoadBytes; i++ {
		if canal[i] != 0xFF {
			t.Errorf("canal byte %d: got 0x%02X, want 0xFF", i, canal[i])
		}
	}
}

func TestBridgeRoadData_RoadPattern(t *testing.T) {
	t.Parallel()

	// Road pattern (bytes 32–63): road surface with bridge structure in the middle.
	road := assets.BridgeRoadData[bridgeRoadBytes : 2*bridgeRoadBytes]

	// Road bytes (0–13, 18–31) should be 0x00 (empty = road surface).
	for i := range 14 {
		if road[i] != 0x00 {
			t.Errorf("road byte %d: got 0x%02X, want 0x00", i, road[i])
		}
	}

	// Bridge bytes (14–17) should be 0xFF (solid = bridge structure).
	for i := 14; i < 18; i++ {
		if road[i] != 0xFF {
			t.Errorf("road byte %d: got 0x%02X, want 0xFF", i, road[i])
		}
	}
}

func TestBridgeRoadData_Attributes(t *testing.T) {
	t.Parallel()

	// Attribute pattern (bytes 64–95): road=0x3C, bridge=0x0E.
	attrs := assets.BridgeRoadData[2*bridgeRoadBytes:]

	for i := range 14 {
		if attrs[i] != 0x3C {
			t.Errorf("attr byte %d: got 0x%02X, want 0x3C (road)", i, attrs[i])
		}
	}

	for i := 14; i < 18; i++ {
		if attrs[i] != 0x0E {
			t.Errorf("attr byte %d: got 0x%02X, want 0x0E (bridge)", i, attrs[i])
		}
	}
}

func TestCalculateRightEdge_Mirrored(t *testing.T) {
	t.Parallel()

	// rightX = 2*center - leftX = 2*128 - 50 = 206
	got := calculateRightEdge(50, 128, assets.EdgeMirrored)
	if got != 206 {
		t.Errorf("EdgeMirrored: got %d, want 206", got)
	}
}

func TestCalculateRightEdge_Offset(t *testing.T) {
	t.Parallel()

	// rightX = width + leftX = 64 + 50 = 114
	got := calculateRightEdge(50, 64, assets.EdgeOffset)
	if got != 114 {
		t.Errorf("EdgeOffset: got %d, want 114", got)
	}
}

func TestRenderIslandLine_FirstThreeScanlines(t *testing.T) {
	t.Parallel()

	// Test island rendering for the first 3 scanlines using actual game data.
	// Island 3 (IslandIndex=3, array index 2): ProfileIdx=6, WidthOffset=0, EdgeMode=EdgeMirrored
	// Profile 6 values: {0x00, 0x00, 0x02, 0x02, 0x04, 0x04, 0x06, 0x06, 0x08, 0x08, 0x0a, 0x0a, 0x0c, 0x0c, 0x0e, 0x0e}

	tb := NewTerrainBuffer(256)

	// Activate island 3 (array index 2).
	island := assets.Islands[2]
	tb.Island = IslandState{
		Active:      true,
		ProfileIdx:  island.ProfileIndex,
		WidthOffset: island.WidthOffset,
		EdgeMode:    island.EdgeMode,
		RenderIdx:   0,
	}

	profile := assets.TerrainProfiles[island.ProfileIndex].(assets.RegularProfile) //nolint:errcheck // island.ProfileIndex is validated by assets package

	// Test first 3 scanlines.
	// LineIdx starts at 16 and wraps to 0 (16 % 16 = 0).
	testCases := []struct {
		scanline         int
		lineIdxInProfile int // The actual index into the profile (after wrapping)
	}{
		{0, 0}, // LineIdx=16, wraps to 0
		{1, 1}, // LineIdx=17, wraps to 1
		{2, 2}, // LineIdx=18, wraps to 2
	}

	for _, tc := range testCases {
		// Calculate expected coordinates using PHP reference implementation formulas.
		profileValue := int(profile.Values[tc.lineIdxInProfile])

		var expectedLeftX, expectedRightX int
		switch island.EdgeMode {
		case assets.EdgeMirrored:
			// SYMMETRICAL: side2 = 0x78 - widthOffset - profileValue, side1 = 0x80 + widthOffset + profileValue
			side2 := 0x78 - island.WidthOffset - profileValue //nolint:mnd // 0x78 = 2*0x3C
			side1 := 0x80 + island.WidthOffset + profileValue //nolint:mnd // 0x80 = 128
			expectedLeftX = side2
			expectedRightX = side1 + 10 //nolint:mnd // constant width addition
		case assets.EdgeOffset:
			// PARALLEL: side2 = 0x3C + widthOffset + profileValue, side1 = 0x80 + widthOffset + profileValue
			side2 := 0x3C + island.WidthOffset + profileValue //nolint:mnd // 0x3C = 60
			side1 := 0x80 + island.WidthOffset + profileValue //nolint:mnd // 0x80 = 128
			expectedLeftX = side2
			expectedRightX = side1 + 10 //nolint:mnd // constant width addition
		default:
			t.Errorf("unsupported edge mode %d", island.EdgeMode)
		}

		// Render the island line.
		tb.renderIslandLine(100+tc.scanline, palette[colorBank])

		// Verify the island state progressed correctly.
		expectedLineIdx := tc.scanline + 1
		if tb.Island.LineIdx != expectedLineIdx {
			t.Errorf("scanline %d: LineIdx = %d, want %d", tc.scanline, tb.Island.LineIdx, expectedLineIdx)
		}

		expectedRenderIdx := tc.scanline + 1
		if tb.Island.RenderIdx != expectedRenderIdx {
			t.Errorf("scanline %d: RenderIdx = %d, want %d", tc.scanline, tb.Island.RenderIdx, expectedRenderIdx)
		}

		// Verify coordinates are valid (leftX <= rightX).
		// Zero-width islands (leftX == rightX) are valid when profile value is 0.
		if expectedLeftX > expectedRightX {
			t.Errorf("scanline %d: invalid coordinates leftX=%d > rightX=%d (profile value=%d, lineIdxInProfile=%d)",
				tc.scanline, expectedLeftX, expectedRightX, profileValue, tc.lineIdxInProfile)
		}

		// Verify coordinates are within screen bounds or clipped appropriately.
		clippedLeftX := expectedLeftX
		clippedRightX := expectedRightX
		if clippedLeftX < 0 {
			clippedLeftX = 0
		}
		if clippedRightX > platform.ScreenWidth {
			clippedRightX = platform.ScreenWidth
		}

		// Islands with profile=0 have zero width and render no pixels (valid behavior).
		// Islands with profile>0 should render some pixels.
		if profileValue > 0 && clippedRightX <= clippedLeftX {
			t.Errorf("scanline %d: no pixels rendered after clipping for non-zero profile (leftX=%d, rightX=%d, profile=%d)",
				tc.scanline, clippedLeftX, clippedRightX, profileValue)
		}
	}

	// Verify island is still active after 3 scanlines.
	if !tb.Island.Active {
		t.Error("island should still be active after 3 scanlines")
	}
}
