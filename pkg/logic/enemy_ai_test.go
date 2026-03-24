package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
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

func (m *mockTerrainBuffer) GetEdges(x, y domain.Px, spriteHeight int) (leftX, rightX domain.Px) {
	m.queriedPositions = append(m.queriedPositions, struct{ x, y int }{int(x), int(y)})

	// Initialize with widest boundaries
	leftX = 0
	rightX = 255

	// Check all scanlines the sprite overlaps
	for dy := range spriteHeight {
		scanlineY := int(y) + dy
		if edges, ok := m.edgesByY[scanlineY]; ok {
			// Use most restrictive boundaries
			if domain.Px(edges.left) > leftX {
				leftX = domain.Px(edges.left)
			}
			if domain.Px(edges.right) < rightX {
				rightX = domain.Px(edges.right)
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

func (m *mockTerrainBuffer) RenderFragment(_ assets.TerrainFragment, _ int, _ bool) {}
func (m *mockTerrainBuffer) Clear()                                                 {}

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

	scrollY := domain.SP(100)

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

	scrollY := domain.SP(100)

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

	// Tank on the left bank (spawn X=32px < leftEdge=64px; stored as SP).
	obj := &state.ViewportObject{
		X:            32 * domain.SubpixelScale,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationBank,
	}

	initializeObjectBoundaries(obj, mock, 0)

	// Left bank: MinX=0, MaxX < leftEdge*SubpixelScale.
	if obj.MinX != 0 {
		t.Errorf("left bank tank MinX: got %d, want 0", obj.MinX)
	}
	if obj.MaxX >= 64*domain.SubpixelScale {
		t.Errorf("left bank tank MaxX: got %d, want < %d (river edge in sp)", obj.MaxX, 64*domain.SubpixelScale)
	}
}

func TestInitializeObjectBoundaries_BankTankRightBank(t *testing.T) {
	t.Parallel()

	mock := newMockTerrainBuffer()
	mock.setEdges(0, 64, 192) // river: left bank edge at 64, right bank edge at 192

	// Tank on the right bank (spawn X=220px > rightEdge=192px; stored as SP).
	obj := &state.ViewportObject{
		X:            220 * domain.SubpixelScale,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationBank,
	}

	initializeObjectBoundaries(obj, mock, 0)

	// Right bank: MinX > rightEdge*SubpixelScale.
	if obj.MinX <= 192*domain.SubpixelScale {
		t.Errorf("right bank tank MinX: got %d, want > %d (river edge in sp)", obj.MinX, 192*domain.SubpixelScale)
	}
}

func TestMoveTank_Road_NeverFiresShell(t *testing.T) {
	t.Parallel()

	// Cross centre from both sides — no shell should ever be fired.
	cases := []struct {
		desc string
		ori  domain.Orientation
		x    domain.SP
	}{
		{"crossing centre from left", domain.OrientationRight, 126 * domain.SubpixelScale},
		{"crossing centre from right", domain.OrientationLeft, 130 * domain.SubpixelScale},
		{"not crossing centre", domain.OrientationRight, 100 * domain.SubpixelScale},
	}

	for _, tc := range cases {
		obj := &state.ViewportObject{X: tc.x, Orientation: tc.ori, TankLocation: domain.TankLocationRoad, Activated: true}
		ts := &state.TankShell{}

		moveTank(obj, 0, ts, false)

		if ts.IsFlying {
			t.Errorf("road tank fired shell (%s)", tc.desc)
		}
	}
}

func TestMoveTank_Bank_StopsAndFiresAtMinBoundary(t *testing.T) {
	t.Parallel()

	// X = MinX + speedStandard so one step lands exactly on MinX.
	const minX = 50
	obj := &state.ViewportObject{
		X: minX + speedStandard, MinX: minX, MaxX: 200,
		Orientation:  domain.OrientationLeft,
		TankLocation: domain.TankLocationBank,
		Activated:    true,
	}
	ts := &state.TankShell{}

	moveTank(obj, 0, ts, false) // terrainAhead was true; advances to MinX

	if obj.X != minX {
		t.Errorf("bank tank X: got %d, want %d", obj.X, minX)
	}
	if ts.IsFlying {
		t.Errorf("shell fired prematurely while terrain was still ahead")
	}

	// Next call: now at MinX, terrainAhead is false → fires.
	moveTank(obj, 4, ts, false)

	if obj.X != minX {
		t.Errorf("bank tank moved past MinX: got X=%d, want %d", obj.X, minX)
	}
	if !ts.IsFlying {
		t.Errorf("shell not flying after bank tank reached MinX")
	}
	// Orientation must NOT reverse.
	if obj.Orientation != domain.OrientationLeft {
		t.Errorf("bank tank changed orientation after reaching MinX: got %v, want Left", obj.Orientation)
	}
}

func TestMoveTank_Bank_StopsAndFiresAtMaxBoundary(t *testing.T) {
	t.Parallel()

	// X = MaxX - speedStandard so one step lands exactly on MaxX.
	const maxX = 200
	obj := &state.ViewportObject{
		X: maxX - speedStandard, MinX: 50, MaxX: maxX,
		Orientation:  domain.OrientationRight,
		TankLocation: domain.TankLocationBank,
		Activated:    true,
	}
	ts := &state.TankShell{}

	moveTank(obj, 0, ts, false) // terrainAhead was true; advances to MaxX

	if obj.X != maxX {
		t.Errorf("bank tank X: got %d, want %d", obj.X, maxX)
	}
	if ts.IsFlying {
		t.Errorf("shell fired prematurely while terrain was still ahead")
	}

	// Next call: now at MaxX, terrainAhead is false → fires.
	moveTank(obj, 4, ts, false)

	if obj.X != maxX {
		t.Errorf("bank tank moved past MaxX: got X=%d, want %d", obj.X, maxX)
	}
	if !ts.IsFlying {
		t.Errorf("shell not flying after bank tank reached MaxX")
	}
	// Orientation must NOT reverse.
	if obj.Orientation != domain.OrientationRight {
		t.Errorf("bank tank changed orientation after reaching MaxX: got %v, want Right", obj.Orientation)
	}
}

func TestMoveTank_Bank_AtEdgeDoesNotFireWhileShellFlying(t *testing.T) {
	t.Parallel()

	// Tank already at MinX boundary.
	obj := &state.ViewportObject{
		X: 50, MinX: 50, MaxX: 200,
		Orientation:  domain.OrientationLeft,
		TankLocation: domain.TankLocationBank,
		Activated:    true,
	}
	ts := &state.TankShell{IsFlying: true} // shell already in flight

	moveTank(obj, 0, ts, false)

	// X must not change.
	if obj.X != 50 {
		t.Errorf("bank tank at edge moved: got X=%d, want 50", obj.X)
	}
	// FireTankShell is a no-op while IsFlying, so shell remains flying.
	if !ts.IsFlying {
		t.Errorf("shell cleared unexpectedly")
	}
}

func TestMoveTank_Bank_AtEdgeFiresAgainWhenShellGone(t *testing.T) {
	t.Parallel()

	// Tank already at MinX boundary, no shell active.
	obj := &state.ViewportObject{
		X: 50, MinX: 50, MaxX: 200,
		Orientation:  domain.OrientationLeft,
		TankLocation: domain.TankLocationBank,
		Activated:    true,
	}
	ts := &state.TankShell{}

	moveTank(obj, 0, ts, false)

	if !ts.IsFlying {
		t.Errorf("bank tank at edge did not fire when shell was gone")
	}
	if obj.X != 50 {
		t.Errorf("bank tank at edge moved after firing: got X=%d, want 50", obj.X)
	}
}

func TestMoveTank_Road_MovesEachTick(t *testing.T) {
	t.Parallel()

	obj := &state.ViewportObject{
		X:            126 * domain.SubpixelScale,
		Orientation:  domain.OrientationRight,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	ts := &state.TankShell{}

	startX := obj.X
	moveTank(obj, 0, ts, false)

	if obj.X != startX+speedStandard {
		t.Errorf("road tank did not move: got X=%d, want %d", obj.X, startX+speedStandard)
	}
	if ts.IsFlying {
		t.Errorf("road tank fired shell unexpectedly")
	}
}

func TestMoveFighter_WrapsLeft(t *testing.T) {
	t.Parallel()

	// X just above zero; one step takes it to or below fighterWrapLeftX=0.
	obj := state.ViewportObject{X: speedFighter - 1, Orientation: domain.OrientationLeft, Activated: true}
	moveFighter(&obj)

	if obj.X != fighterResetLeftX {
		t.Errorf("fighter wrap left: got X=%d, want %d", obj.X, fighterResetLeftX)
	}
}

func TestMoveFighter_WrapsRight(t *testing.T) {
	t.Parallel()

	// X just below fighterWrapRightX; one step crosses it.
	obj := state.ViewportObject{X: fighterWrapRightX - speedFighter + 1, Orientation: domain.OrientationRight, Activated: true}
	moveFighter(&obj)

	if obj.X != fighterResetRightX {
		t.Errorf("fighter wrap right: got X=%d, want %d", obj.X, fighterResetRightX)
	}
}

func TestMoveShipOrHelicopter_MovesEachTick(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{
		X:           100 * domain.SubpixelScale,
		MinX:        0,
		MaxX:        200 * domain.SubpixelScale,
		Orientation: domain.OrientationRight,
		Activated:   true,
	}
	startX := obj.X
	moveShipOrHelicopter(&obj)

	if obj.X != startX+speedStandard {
		t.Errorf("ship/helicopter did not move: got X=%d, want %d", obj.X, startX+speedStandard)
	}
}

func TestMoveShipOrHelicopter_ReversesAtBoundary(t *testing.T) {
	t.Parallel()

	// Start at MaxX; next step should reverse direction.
	const maxX = 200 * domain.SubpixelScale
	obj := state.ViewportObject{
		X: maxX, MinX: 0, MaxX: maxX,
		Orientation: domain.OrientationRight,
		Activated:   true,
	}
	moveShipOrHelicopter(&obj)

	if obj.Orientation != domain.OrientationLeft {
		t.Errorf("expected orientation to reverse at MaxX, got %v", obj.Orientation)
	}
}

func TestMoveBalloon_MovesEachTick(t *testing.T) {
	t.Parallel()

	obj := state.ViewportObject{
		X:           100 * domain.SubpixelScale,
		MinX:        0,
		MaxX:        200 * domain.SubpixelScale,
		Orientation: domain.OrientationRight,
		Activated:   true,
	}
	startX := obj.X
	moveBalloon(&obj)

	if obj.X != startX+speedBalloon {
		t.Errorf("balloon did not move: got X=%d, want %d", obj.X, startX+speedBalloon)
	}
}

func TestMoveEnemies_AdvancedHelicopterFiresMissileInNormalMode(t *testing.T) {
	t.Parallel()

	vp := &state.Viewport{
		Objects: []*state.ViewportObject{
			{
				Type:        domain.ObjectHelicopterAdv,
				X:           100 * domain.SubpixelScale,
				Y:           40 * domain.SubpixelScale,
				Orientation: domain.OrientationLeft,
				Activated:   true,
				MaxX:        200 * domain.SubpixelScale,
				MinX:        0,
			},
		},
	}
	ts := &state.TankShell{}
	hm := &state.HeliMissile{}

	moveEnemies(vp, 0, ts, hm, domain.GameplayNormal, false)

	if !hm.Active {
		t.Error("advanced helicopter did not fire missile during normal gameplay")
	}
}

func TestMoveEnemies_AdvancedHelicopterDoesNotFireDuringScrollIn(t *testing.T) {
	t.Parallel()

	vp := &state.Viewport{
		Objects: []*state.ViewportObject{
			{
				Type:        domain.ObjectHelicopterAdv,
				X:           100 * domain.SubpixelScale,
				Y:           40 * domain.SubpixelScale,
				Orientation: domain.OrientationLeft,
				Activated:   true,
				MaxX:        200 * domain.SubpixelScale,
				MinX:        0,
			},
		},
	}
	ts := &state.TankShell{}
	hm := &state.HeliMissile{}

	moveEnemies(vp, 0, ts, hm, domain.GameplayScrollIn, false)

	if hm.Active {
		t.Error("advanced helicopter fired missile during scroll-in")
	}
}

func TestMoveEnemies_RegularHelicopterNeverFiresMissile(t *testing.T) {
	t.Parallel()

	vp := &state.Viewport{
		Objects: []*state.ViewportObject{
			{
				Type:        domain.ObjectHelicopterReg,
				X:           100 * domain.SubpixelScale,
				Y:           40 * domain.SubpixelScale,
				Orientation: domain.OrientationLeft,
				Activated:   true,
				MaxX:        200 * domain.SubpixelScale,
				MinX:        0,
			},
		},
	}
	ts := &state.TankShell{}
	hm := &state.HeliMissile{}

	moveEnemies(vp, 0, ts, hm, domain.GameplayNormal, false)

	if hm.Active {
		t.Error("regular helicopter fired a missile")
	}
}
