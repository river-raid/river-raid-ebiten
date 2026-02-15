package render

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
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

func TestCalculateIslandRightEdge_Mirrored(t *testing.T) {
	t.Parallel()

	// rightX = 2*128 - 140 = 116
	got := calculateIslandRightEdge(140, assets.EdgeMirrored)
	if got != 116 {
		t.Errorf("island EdgeMirrored: got %d, want 116", got)
	}
}

func TestCalculateIslandRightEdge_Offset(t *testing.T) {
	t.Parallel()

	// rightX = 60 + 140 = 200
	got := calculateIslandRightEdge(140, assets.EdgeOffset)
	if got != 200 {
		t.Errorf("island EdgeOffset: got %d, want 200", got)
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
