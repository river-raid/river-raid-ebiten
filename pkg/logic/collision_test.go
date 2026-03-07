package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// openTerrain returns terrain functions that leave the full screen width open.
func openTerrain() (leftX, rightX func(int) int) {
	return func(_ int) int { return 0 }, func(_ int) int { return 256 }
}

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

func TestBridgeTarget_Inactive(t *testing.T) {
	t.Parallel()

	bt := bridgeTarget{vp: state.NewViewport(), y: 100, active: false, destroyed: false}
	m := playerMissile{x: 128, y: 90}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("inactive bridge should not register a hit")
	}
}

func TestBridgeTarget_Destroyed(t *testing.T) {
	t.Parallel()

	bt := bridgeTarget{vp: state.NewViewport(), y: 100, active: true, destroyed: true}
	m := playerMissile{x: 128, y: 90}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("destroyed bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileAbove(t *testing.T) {
	t.Parallel()

	// bridgeTop = 100 - 22 = 78; missile bottom = 70 + 8 = 78 → just outside.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 70}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile above bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileBelow(t *testing.T) {
	t.Parallel()

	// missile top = 100 → at bridgeY, just outside.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 100}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile below bridge should not register a hit")
	}
}

func TestBridgeTarget_Hit_PointsAndFragments(t *testing.T) {
	t.Parallel()

	const by = 100
	bt := bridgeTarget{vp: state.NewViewport(), y: by, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 85} // inside [78, 100)

	hit, ok := bt.checkHit(m, &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != -1 {
		t.Errorf("objectIdx = %d, want -1", hit.objectIdx)
	}
	if hit.points != PointsBridge {
		t.Errorf("points = %d, want %d", hit.points, PointsBridge)
	}
	if len(hit.explosionFragments) != 6 {
		t.Fatalf("fragment count = %d, want 6", len(hit.explosionFragments))
	}

	wantXs := map[int]bool{bridgeFragX0: true, bridgeFragX1: true}
	wantYs := map[int]bool{
		by - bridgeFragRow0: true,
		by - bridgeFragRow1: true,
		by - bridgeFragRow2: true,
	}
	for _, f := range hit.explosionFragments {
		if !wantXs[f.X] {
			t.Errorf("unexpected fragment X=%d", f.X)
		}
		if !wantYs[f.Y] {
			t.Errorf("unexpected fragment Y=%d", f.Y)
		}
	}
}

func TestViewportObjectsTarget_EmptyViewport(t *testing.T) {
	t.Parallel()

	vot := viewportObjectsTarget{vp: state.NewViewport()}
	m := playerMissile{x: 100, y: 50}

	if _, ok := vot.checkHit(m, &CollisionResult{}); ok {
		t.Error("empty viewport should not register a hit")
	}
}

func TestViewportObjectsTarget_MissileIgnoresTank(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 100, Y: 50, Type: domain.ObjectTank})

	vot := viewportObjectsTarget{vp: vp}
	m := playerMissile{x: 100, y: 50}

	if _, ok := vot.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile should pass through tank")
	}
}

func TestViewportObjectsTarget_PlaneIgnoresFuel(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 120, Y: domain.PlaneY, Type: domain.ObjectFuel})

	vot := viewportObjectsTarget{vp: vp}
	p := playerPlane{x: 120}

	if _, ok := vot.checkHit(p, &CollisionResult{}); ok {
		t.Error("plane should not hit fuel depot via checkHit (fuel is handled separately)")
	}
}

func TestViewportObjectsTarget_NoSpatialOverlap(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 200, Y: 200, Type: domain.ObjectHelicopterReg})

	vot := viewportObjectsTarget{vp: vp}
	m := playerMissile{x: 0, y: 0}

	if _, ok := vot.checkHit(m, &CollisionResult{}); ok {
		t.Error("spatially separated objects should not register a hit")
	}
}

func TestViewportObjectsTarget_Hit_IndexPointsFragments(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectShip})

	vot := viewportObjectsTarget{vp: vp}
	m := playerMissile{x: 100, y: 50}

	hit, ok := vot.checkHit(m, &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != 0 {
		t.Errorf("objectIdx = %d, want 0", hit.objectIdx)
	}
	if hit.points != PointsShip {
		t.Errorf("points = %d, want %d", hit.points, PointsShip)
	}
	if len(hit.explosionFragments) != 2 {
		t.Errorf("fragment count = %d, want 2 (ship has two fragments)", len(hit.explosionFragments))
	}
}

func TestViewportObjectsTarget_ReturnsFirstMatch(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects,
		&state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectShip},
		&state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectHelicopterReg},
	)

	vot := viewportObjectsTarget{vp: vp}
	m := playerMissile{x: 100, y: 50}

	hit, ok := vot.checkHit(m, &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != 0 {
		t.Errorf("objectIdx = %d, want 0 (first object should be hit)", hit.objectIdx)
	}
}

func TestCheckFirstHit_BridgeBeforeObjects(t *testing.T) {
	t.Parallel()

	// Both bridge and an object overlap the missile; bridge must win.
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 126, Y: 84, Type: domain.ObjectHelicopterReg})

	targets := [2]target{
		bridgeTarget{vp: vp, y: 100, active: true, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	m := playerMissile{x: 128, y: 85}

	hit, ok := checkFirstHit(m, targets[:], &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != -1 {
		t.Errorf("objectIdx = %d, want -1 (bridge should be hit first)", hit.objectIdx)
	}
}

func TestCheckFirstHit_FallsThroughToObjects(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectShip})

	targets := [2]target{
		bridgeTarget{vp: vp, y: 100, active: false, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	m := playerMissile{x: 100, y: 50}

	hit, ok := checkFirstHit(m, targets[:], &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != 0 {
		t.Errorf("objectIdx = %d, want 0 (object should be hit)", hit.objectIdx)
	}
}

func TestCheckFirstHit_NothingHit(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	targets := [2]target{
		bridgeTarget{vp: vp, y: 100, active: false, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	m := playerMissile{x: 100, y: 50}

	if _, ok := checkFirstHit(m, targets[:], &CollisionResult{}); ok {
		t.Error("expected no hit")
	}
}

func TestCheckCollisions_PlaneVsTerrain(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 130 }
	rightX := func(_ int) int { return 200 }

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, state.NewViewport(), leftX, rightX, false, 0, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied from terrain collision")
	}
}

func TestCheckCollisions_PlaneVsFuelDepot_Refueling(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118, Y: domain.PlaneY - 10, Type: domain.ObjectFuel})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false, 0)

	if !result.Refueling {
		t.Error("expected Refueling")
	}
	if result.PlayerDied {
		t.Error("fuel depot should not kill the player")
	}
}

func TestCheckCollisions_PlaneVsBridge(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	var m state.PlayerMissile
	var hm state.HeliMissile

	// bridgeY=145: bridgeTop=123, plane rows [128,136) overlap.
	result := CheckCollisions(120, &m, &hm, state.NewViewport(), leftX, rightX, true, 145, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied")
	}
	if !result.BridgeHit {
		t.Error("expected BridgeHit")
	}
	if result.PointsScored != PointsBridge {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsBridge)
	}
	if len(result.ExplosionFragments) != 6 {
		t.Errorf("fragment count = %d, want 6", len(result.ExplosionFragments))
	}
}

func TestCheckCollisions_PlaneVsEnemy(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118, Y: domain.PlaneY, Type: domain.ObjectHelicopterReg})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied")
	}
	if result.PointsScored != PointsHelicopterReg {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsHelicopterReg)
	}
	if len(result.DestroyObjects) != 1 || result.DestroyObjects[0] != 0 {
		t.Errorf("DestroyObjects = %v, want [0]", result.DestroyObjects)
	}
	if len(result.ExplosionFragments) == 0 {
		t.Error("expected explosion fragments")
	}
}

func TestCheckCollisions_PlaneVsTank_PassesThrough(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118, Y: domain.PlaneY, Type: domain.ObjectTank, TankLocation: domain.TankLocationRoad})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false, 0)

	if result.PlayerDied {
		t.Error("plane should pass through tank without dying")
	}
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty (tank must not be destroyed by plane)", result.DestroyObjects)
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
}

func TestCheckCollisions_MissileVsBridge(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	m := state.PlayerMissile{X: 128, Y: 50, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, state.NewViewport(), leftX, rightX, true, 60, false, 0)

	if !result.BridgeHit {
		t.Error("expected BridgeHit")
	}
	if result.PointsScored != PointsBridge {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsBridge)
	}
	if m.Active {
		t.Error("missile should be deactivated after bridge hit")
	}
}

func TestCheckCollisions_MissileVsObject(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectShip})

	m := state.PlayerMissile{X: 100, Y: 50, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false, 0)

	if result.PointsScored != PointsShip {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsShip)
	}
	if len(result.DestroyObjects) != 1 || result.DestroyObjects[0] != 0 {
		t.Errorf("DestroyObjects = %v, want [0]", result.DestroyObjects)
	}
	if m.Active {
		t.Error("missile should be deactivated after hit")
	}
}

func TestCheckCollisions_MissileVsTank_PassesThrough(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98, Y: 48, Type: domain.ObjectTank})

	m := state.PlayerMissile{X: 100, Y: 50, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, vp, leftX, rightX, false, 0, false, 0)

	if !m.Active {
		t.Error("missile should remain active when passing through tank")
	}
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty", result.DestroyObjects)
	}
}

func TestCheckCollisions_MissileVsDestroyedBridge_PassesThrough(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	m := state.PlayerMissile{X: 128, Y: 50, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120, &m, &hm, state.NewViewport(), leftX, rightX, true, 60, true, 0)

	if result.BridgeHit {
		t.Error("destroyed bridge should not register a hit")
	}
	if !m.Active {
		t.Error("missile should remain active when bridge is destroyed")
	}
}

func TestCheckCollisions_HeliMissileVsPlane(t *testing.T) {
	t.Parallel()

	leftX, rightX := openTerrain()

	var m state.PlayerMissile
	hm := state.HeliMissile{X: 121, Y: domain.PlaneY + 2, Active: true}

	result := CheckCollisions(120, &m, &hm, state.NewViewport(), leftX, rightX, false, 0, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied from helicopter missile")
	}
}

func TestApplyBridgeDestroyedTanks_InGap(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	// X = $70 - spriteWidth: X+spriteWidth = $70 → exactly at the left edge of the gap.
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: tankGapLeftEdge - assets.SpriteTankWidth, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	})

	var result CollisionResult
	bridgeTarget{vp: vp, bridgeIndex: 8}.onHit(&result)

	if len(result.DestroyObjects) != 1 || result.DestroyObjects[0] != 0 {
		t.Errorf("DestroyObjects = %v, want [0]", result.DestroyObjects)
	}
	if result.PointsScored != PointsTank {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsTank)
	}
	if len(result.ExplosionFragments) != 1 {
		t.Fatalf("ExplosionFragments count = %d, want 1", len(result.ExplosionFragments))
	}
}

func TestApplyBridgeDestroyedTanks_OnBank_LateLevel(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	obj := &state.ViewportObject{
		X: 0x20, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	vp.Objects = append(vp.Objects, obj)

	var result CollisionResult
	bridgeTarget{vp: vp, bridgeIndex: bridgeEarlyLevel + 1}.onHit(&result)

	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty (tank should become bank-tank)", result.DestroyObjects)
	}
	if obj.TankLocation != domain.TankLocationBank {
		t.Errorf("TankLocation = %v, want TankLocationBank", obj.TankLocation)
	}
}

func TestApplyBridgeDestroyedTanks_OnBank_EarlyLevel(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	obj := &state.ViewportObject{
		X: 0x20, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	vp.Objects = append(vp.Objects, obj)

	var result CollisionResult
	bridgeTarget{vp: vp, bridgeIndex: bridgeEarlyLevel}.onHit(&result)

	// Early level: tank freezes in place (not removed, not converted).
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty (tank should freeze in place)", result.DestroyObjects)
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
	if obj.TankLocation != domain.TankLocationRoad {
		t.Errorf("TankLocation = %v, want TankLocationRoad (should stay frozen)", obj.TankLocation)
	}
}

func TestMoveTank_Road_FrozenWhenBridgeDestroyed(t *testing.T) {
	t.Parallel()

	obj := &state.ViewportObject{
		X: 128, Orientation: domain.OrientationRight,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	}
	ts := &state.TankShell{}

	moveTank(obj, 0, ts, true)

	if obj.X != 128 {
		t.Errorf("frozen road tank moved: X = %d, want 128", obj.X)
	}
}

func TestApplyBridgeDestroyedTanks_JustOutsideGap_NotDestroyed(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	// X+spriteWidth = $70-1: sprite ends one pixel before the gap — on the bank.
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: tankGapLeftEdge - assets.SpriteTankWidth - 1, Y: 50,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
	})

	var result CollisionResult
	bridgeTarget{vp: vp, bridgeIndex: bridgeEarlyLevel + 1}.onHit(&result)

	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty (tank just outside gap should not be destroyed)", result.DestroyObjects)
	}
	if len(result.ExplosionFragments) != 0 {
		t.Errorf("ExplosionFragments count = %d, want 0 (no explosion for bank tank)", len(result.ExplosionFragments))
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
}
