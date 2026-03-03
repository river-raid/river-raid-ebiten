package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestBoxOverlap(t *testing.T) {
	t.Parallel()

	if !boxOverlap(0, 0, 10, 10, 5, 5, 10, 10) {
		t.Error("expected overlap")
	}

	if boxOverlap(0, 0, 10, 10, 20, 20, 10, 10) {
		t.Error("expected no overlap")
	}

	if boxOverlap(0, 0, 10, 10, 10, 0, 10, 10) {
		t.Error("expected no overlap at exact edge")
	}
}

func TestCheckCollisions_PlaneVsTerrain(t *testing.T) {
	t.Parallel()

	// Terrain edges that leave no room for the plane.
	leftX := func(_ int) int { return 130 }
	rightX := func(_ int) int { return 200 }

	var m state.PlayerMissile
	var hm state.HeliMissile
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false)

	if !result.PlayerDied {
		t.Error("expected player death from terrain collision")
	}
}

func TestCheckCollisions_PlaneVsFuelDepot(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	var m state.PlayerMissile
	var hm state.HeliMissile
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118, Y: domain.PlaneY - 10, Type: domain.ObjectFuel})

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false)

	if !result.Refueling {
		t.Error("expected refueling from fuel depot collision")
	}

	if result.PlayerDied {
		t.Error("fuel depot should not kill player")
	}
}

func TestCheckCollisions_MissileVsObject(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	m := state.PlayerMissile{X: 100, Y: 50, Active: true}
	var hm state.HeliMissile
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectShip})

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false)

	if result.PointsScored != PointsShip {
		t.Errorf("points = %d, want %d", result.PointsScored, PointsShip)
	}

	if len(result.DestroyObjects) != 1 || result.DestroyObjects[0] != 0 {
		t.Errorf("DestroyObjects = %v, want [0]", result.DestroyObjects)
	}

	if m.Active {
		t.Error("missile should be deactivated after hit")
	}
}

func TestCheckCollisions_HeliMissileVsPlayer(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	var m state.PlayerMissile
	hm := state.HeliMissile{X: 121, Y: domain.PlaneY + 2, Active: true}
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false)

	if !result.PlayerDied {
		t.Error("expected player death from helicopter missile")
	}
}

func TestCheckCollisions_MissileVsBridge(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	m := state.PlayerMissile{X: 128, Y: 50, Active: true}
	var hm state.HeliMissile
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, true, 60, false)

	if !result.BridgeHit {
		t.Error("expected bridge hit")
	}

	if result.PointsScored != PointsBridge {
		t.Errorf("points = %d, want %d", result.PointsScored, PointsBridge)
	}
}

// TestCheckCollisions_MissileVsBridge_SpawnsExplosions checks that hitting a bridge
// produces exactly 6 explosion fragments in the correct 2×3 grid positions.
func TestCheckCollisions_MissileVsBridge_SpawnsExplosions(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	const bridgeY = 60
	m := state.PlayerMissile{X: 128, Y: 50, Active: true}
	var hm state.HeliMissile
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, true, bridgeY, false)

	if len(result.ExplosionFragments) != 6 {
		t.Fatalf("ExplosionFragments count = %d, want 6", len(result.ExplosionFragments))
	}

	// Verify the 2×3 grid: X must be $70 or $80, Y must be bridgeY-{4,12,20}.
	wantXs := map[int]bool{bridgeFragX0: true, bridgeFragX1: true}
	wantYs := map[int]bool{
		bridgeY - bridgeFragRow0: true,
		bridgeY - bridgeFragRow1: true,
		bridgeY - bridgeFragRow2: true,
	}
	for _, frag := range result.ExplosionFragments {
		if !wantXs[frag.X] {
			t.Errorf("unexpected fragment X=%d", frag.X)
		}
		if !wantYs[frag.Y] {
			t.Errorf("unexpected fragment Y=%d", frag.Y)
		}
		if frag.Frame != 1 {
			t.Errorf("fragment Frame = %d, want 1", frag.Frame)
		}
	}
}

// TestApplyBridgeDestroyedTanks_InGap checks that a road tank in the river gap is
// destroyed, awards 250 pts, and spawns 1 fragment.
func TestApplyBridgeDestroyedTanks_InGap(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	// X=102: X+10=112=0x70 → exactly at the left edge of the gap.
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: 0x70 - tankGapProbe, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	})

	result := applyBridgeDestroyedTanks(vp, 8)

	if len(result.RemoveIndices) != 1 || result.RemoveIndices[0] != 0 {
		t.Errorf("RemoveIndices = %v, want [0]", result.RemoveIndices)
	}
	if result.PointsScored != PointsTank {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsTank)
	}
	if len(result.ExplosionFragments) != 1 {
		t.Fatalf("ExplosionFragments count = %d, want 1", len(result.ExplosionFragments))
	}
}

// TestApplyBridgeDestroyedTanks_OnBank_LateLevel checks that a road tank outside the
// gap on a late level (bridge > 7) is converted to a bank-tank, not removed.
func TestApplyBridgeDestroyedTanks_OnBank_LateLevel(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	// X=0x20: X+10=0x2A < 0x70 → on the left bank.
	obj := &state.ViewportObject{
		X: 0x20, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	vp.Objects = append(vp.Objects, obj)

	result := applyBridgeDestroyedTanks(vp, bridgeEarlyLevel+1) // bridge 8 → late level

	if len(result.RemoveIndices) != 0 {
		t.Errorf("RemoveIndices = %v, want empty (tank should become bank-tank)", result.RemoveIndices)
	}
	if obj.TankLocation != domain.TankLocationBank {
		t.Errorf("TankLocation = %v, want TankLocationBank", obj.TankLocation)
	}
}

// TestApplyBridgeDestroyedTanks_OnBank_EarlyLevel checks that a road tank outside the
// gap on an early level (bridge <= 7) is removed.
func TestApplyBridgeDestroyedTanks_OnBank_EarlyLevel(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: 0x20, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	})

	result := applyBridgeDestroyedTanks(vp, bridgeEarlyLevel) // bridge 7 → early level

	if len(result.RemoveIndices) != 1 {
		t.Errorf("RemoveIndices = %v, want [0]", result.RemoveIndices)
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0 (no points for bank removal)", result.PointsScored)
	}
}

// TestMoveTank_Road_FrozenWhenBridgeDestroyed checks that a road tank does not move
// when bridgeDestroyed is true.
func TestMoveTank_Road_FrozenWhenBridgeDestroyed(t *testing.T) {
	t.Parallel()

	obj := &state.ViewportObject{
		X: 128, Orientation: domain.OrientationRight,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	ts := &state.TankShell{}

	moveTank(obj, 0, ts, true) // even tick but bridge destroyed → frozen

	if obj.X != 128 {
		t.Errorf("frozen road tank moved: got X=%d, want 128", obj.X)
	}
}
