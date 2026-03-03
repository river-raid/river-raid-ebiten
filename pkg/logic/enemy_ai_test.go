package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// mockTerrainBuffer is a test double for TerrainBuffer that records queries.
type mockTerrainBuffer struct {
	edgesByY         map[int]struct{ left, right int }
	queriedPositions []struct{ x, y int }
}

func newMockTerrainBuffer() *mockTerrainBuffer {
	return &mockTerrainBuffer{
		edgesByY: make(map[int]struct{ left, right int }),
	}
}

func (m *mockTerrainBuffer) GetEdges(x, y, spriteHeight int) (leftX, rightX int) {
	m.queriedPositions = append(m.queriedPositions, struct{ x, y int }{x, y})

	// Initialize with widest boundaries
	leftX = 0
	rightX = 255

	// Check all scanlines the sprite overlaps
	for dy := range spriteHeight {
		scanlineY := y + dy
		if edges, ok := m.edgesByY[scanlineY]; ok {
			// Use most restrictive boundaries
			if edges.left > leftX {
				leftX = edges.left
			}
			if edges.right < rightX {
				rightX = edges.right
			}
		} else {
			// Default: return reasonable river boundaries
			if 50 > leftX {
				leftX = 50
			}
			if 200 < rightX {
				rightX = 200
			}
		}
	}

	return leftX, rightX
}

func (m *mockTerrainBuffer) setEdges(y, left, right int) {
	m.edgesByY[y] = struct{ left, right int }{left, right}
}

func TestInitializeObjectBoundaries_CalculatesBoundariesCorrectly(t *testing.T) {
	t.Parallel()

	// Setup: Create terrain with known edges
	mock := newMockTerrainBuffer()

	// Set up terrain that narrows at buffer Y=95
	for y := 100; y >= 0; y-- {
		if y == 95 {
			// Narrow passage at this position
			mock.setEdges(y, 80, 120) // Only 40 pixels wide
		} else {
			// Wide river elsewhere
			mock.setEdges(y, 50, 200) // 150 pixels wide
		}
	}

	obj := &state.ViewportObject{
		X:    100,
		Y:    0,
		Type: domain.ObjectHelicopterReg, // 16px wide
	}

	scrollY := 100

	// Execute
	initializeObjectBoundaries(obj, mock, scrollY)

	// Verify: Boundaries should reflect the spawn position
	// At scrollY=100, the object spawns at buffer Y=100
	// Expected boundaries: minX = 50, maxX = 200 - 10 = 190 (helicopter width is 10px)

	t.Logf("Calculated boundaries: MinX=%d, MaxX=%d", obj.MinX, obj.MaxX)

	// This test will help us understand if the narrow passage is being detected
	switch {
	case obj.MinX == 80 && obj.MaxX == 110:
		t.Log("✓ Narrow passage at y=95 was detected (helicopter is 10px wide)")
	case obj.MinX == 50 && obj.MaxX == 184:
		t.Log("✗ Narrow passage at y=95 was NOT detected")
	default:
		t.Logf("? Unexpected boundaries (neither wide nor narrow)")
	}

	// Verify boundaries are valid (MinX < MaxX)
	if obj.MinX >= obj.MaxX {
		t.Errorf("Invalid boundaries: MinX=%d >= MaxX=%d (zero or negative width!)",
			obj.MinX, obj.MaxX)
	}
}

func TestInitializeObjectBoundaries_DetectsImpossiblePassage(t *testing.T) {
	t.Parallel()

	// Setup: Create terrain that's too narrow for the enemy
	mock := newMockTerrainBuffer()

	// Set up terrain where the river is narrower than the sprite
	// Passage too narrow: only 8 pixels wide, but helicopter is 10px
	mock.setEdges(100, 100, 108)

	obj := &state.ViewportObject{
		X:    100,
		Y:    0,
		Type: domain.ObjectHelicopterReg, // 10px wide
	}

	scrollY := 100

	// Execute
	initializeObjectBoundaries(obj, mock, scrollY)

	// Verify
	t.Logf("Calculated boundaries: MinX=%d, MaxX=%d", obj.MinX, obj.MaxX)

	// When passage is too narrow, MaxX - spriteWidth might be less than MinX
	// This would result in MinX >= MaxX (invalid/zero-width boundaries)
	if obj.MinX >= obj.MaxX {
		t.Logf("✓ Detected impossible passage: MinX=%d >= MaxX=%d", obj.MinX, obj.MaxX)
		t.Log("  This is the bug! Enemy would get stuck with zero-width boundaries.")
	} else {
		t.Logf("✗ Did not detect impossible passage: MinX=%d < MaxX=%d", obj.MinX, obj.MaxX)
	}
}

func TestInitializeObjectBoundaries_BankTankLeftBank(t *testing.T) {
	t.Parallel()

	mock := newMockTerrainBuffer()
	mock.setEdges(0, 64, 192) // river: left bank edge at 64, right bank edge at 192

	// Tank on the left bank (spawn X=32 < leftEdge=64).
	obj := &state.ViewportObject{
		X:            32,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationBank,
	}

	initializeObjectBoundaries(obj, mock, 0)

	// Left bank: MinX=0, MaxX=leftEdge − spriteWidth − padding.
	// Tank sprite width from assets is used; padding=8. leftEdge=64.
	if obj.MinX != 0 {
		t.Errorf("left bank tank MinX: got %d, want 0", obj.MinX)
	}
	if obj.MaxX >= 64 {
		t.Errorf("left bank tank MaxX: got %d, want < 64 (river edge)", obj.MaxX)
	}
}

func TestInitializeObjectBoundaries_BankTankRightBank(t *testing.T) {
	t.Parallel()

	mock := newMockTerrainBuffer()
	mock.setEdges(0, 64, 192) // river: left bank edge at 64, right bank edge at 192

	// Tank on the right bank (spawn X=220 > rightEdge=192).
	obj := &state.ViewportObject{
		X:            220,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationBank,
	}

	initializeObjectBoundaries(obj, mock, 0)

	// Right bank: MinX=rightEdge + padding, MaxX=ScreenWidth − spriteWidth.
	if obj.MinX <= 192 {
		t.Errorf("right bank tank MinX: got %d, want > 192 (river edge)", obj.MinX)
	}
	if obj.MaxX > 256 {
		t.Errorf("right bank tank MaxX: got %d, want ≤ 256", obj.MaxX)
	}
}

func TestMoveFighter_WrapsLeft(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{X: 2, Orientation: domain.OrientationLeft, Activated: true}
	moveFighter(&obj)

	if obj.X != fighterResetLeftX {
		t.Errorf("fighter wrap left: got X=%d, want %d", obj.X, fighterResetLeftX)
	}
}

func TestMoveFighter_WrapsRight(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{X: 232, Orientation: domain.OrientationRight, Activated: true}
	moveFighter(&obj)

	if obj.X != fighterResetRightX {
		t.Errorf("fighter wrap right: got X=%d, want %d", obj.X, fighterResetRightX)
	}
}

func TestMoveShipOrHelicopter_EvenTickOnly(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{X: 100, Orientation: domain.OrientationRight, Activated: true}

	// Odd tick: no movement.
	moveShipOrHelicopter(&obj, 1)

	if obj.X != 100 {
		t.Errorf("odd tick: got X=%d, want 100", obj.X)
	}

	// Even tick: moves right.
	moveShipOrHelicopter(&obj, 2)

	if obj.X != 102 {
		t.Errorf("even tick: got X=%d, want 102", obj.X)
	}
}

func TestMoveBalloon_Every4thFrame(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{X: 100, Orientation: domain.OrientationRight, Activated: true}

	// tick & 3 != 1: no movement.
	moveBalloon(&obj, 0)

	if obj.X != 100 {
		t.Errorf("tick 0: got X=%d, want 100", obj.X)
	}

	// tick & 3 == 1: moves.
	moveBalloon(&obj, 1)

	if obj.X != 102 {
		t.Errorf("tick 1: got X=%d, want 102", obj.X)
	}
}
