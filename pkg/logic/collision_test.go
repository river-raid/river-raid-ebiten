package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// openTerrain returns a terrain function that leaves the full screen width open.
func openTerrain() TerrainEdgesFunc {
	return func(_ domain.Px, _ int) (domain.Px, domain.Px) { return 0, 256 }
}

// bankedTerrain returns a terrain function with banks outside [leftX, rightX).
func bankedTerrain(leftX, rightX domain.Px) TerrainEdgesFunc {
	return func(_ domain.Px, _ int) (domain.Px, domain.Px) { return leftX, rightX }
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

	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: false, destroyed: false}
	m := playerMissile{x: 128, y: 90}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("inactive bridge should not register a hit")
	}
}

func TestBridgeTarget_Destroyed(t *testing.T) {
	t.Parallel()

	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: true}
	m := playerMissile{x: 128, y: 90}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("destroyed bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileAbove(t *testing.T) {
	t.Parallel()

	// Bridge at 100 px = 800 sp. bridgeTop=78; missile bottom = 70 + 1 = 71 ≤ 78 → just outside.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 70}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile above bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileBelow(t *testing.T) {
	t.Parallel()

	// Bridge at 100 px = 800 sp. The band ends at bridgeY inclusive, so Y=101 is past it.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 101}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile below bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileLeftOfBridge(t *testing.T) {
	t.Parallel()

	// Missile Y is inside the bridge band, but its X is over the left bank.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: bridgeLeftX - assets.SpritePlayerMissileWidth, y: 85}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile left of the bridge should not register a hit")
	}
}

func TestBridgeTarget_MissileRightOfBridge(t *testing.T) {
	t.Parallel()

	// Missile Y is inside the bridge band, but its X is over the right bank.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: bridgeRightX, y: 85}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("missile right of the bridge should not register a hit")
	}
}

func TestBridgeTarget_StrikerOriginAtBridgeY(t *testing.T) {
	t.Parallel()

	// Bridge at 100. The missile's origin row lands exactly on bridgeY, the inclusive
	// end of the band.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 100}

	if _, ok := bt.checkHit(m, &CollisionResult{}); !ok {
		t.Error("striker origin at bridgeY should register a hit")
	}
}

func TestBridgeTarget_StrikerStraddlingBridgeTop(t *testing.T) {
	t.Parallel()

	// Bridge at 100, band (78,100]. The missile spans [75,81): part of its box is inside
	// the band, but its origin row is still above it, so it misses.
	bt := bridgeTarget{vp: state.NewViewport(), y: 100 * domain.SubpixelScale, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 75}

	if _, ok := bt.checkHit(m, &CollisionResult{}); ok {
		t.Error("striker whose origin row is above the band should not register a hit")
	}
}

func TestBridgeTarget_Hit_PointsAndFragments(t *testing.T) {
	t.Parallel()

	// Bridge at 100 px = 800 sp. bridgeYPx=100, bridgeTop=78. Missile y=85 is in [78,100).
	const by = 100 * domain.SubpixelScale
	bt := bridgeTarget{vp: state.NewViewport(), y: by, active: true, destroyed: false}
	m := playerMissile{x: 128, y: 85} // screen pixels; inside [78, 100)

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

	wantXs := map[domain.SP]bool{bridgeFragX0: true, bridgeFragX1: true}
	wantYs := map[domain.SP]bool{
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
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectShip})

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
		&state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectShip},
		&state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectHelicopterReg},
	)

	vot := viewportObjectsTarget{vp: vp}
	m := playerMissile{x: 100, y: 50}

	hit, ok := vot.checkHit(m, &CollisionResult{})

	if !ok {
		t.Fatal("expected a hit")
	}
	if hit.objectIdx != 0 {
		t.Errorf("objectIdx = %d, want 0 (the strike lands on one target)", hit.objectIdx)
	}
}

func TestApplyStrike_SpentOnTheBridge(t *testing.T) {
	t.Parallel()

	// The bridge and a helicopter both overlap the plane. The strike lands on the
	// bridge and is spent: the helicopter survives and does not score.
	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 126 * domain.SubpixelScale, Y: domain.PlaneY * domain.SubpixelScale, Type: domain.ObjectHelicopterReg})

	targets := [2]target{
		bridgeTarget{vp: vp, y: 140 * domain.SubpixelScale, active: true, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	p := playerPlane{x: 128}

	var result CollisionResult
	if !applyStrike(p, targets[:], &result) {
		t.Fatal("expected a hit")
	}

	if !result.BridgeHit {
		t.Error("expected BridgeHit")
	}
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty (the strike is spent on the bridge)", result.DestroyObjects)
	}
	if result.PointsScored != PointsBridge {
		t.Errorf("PointsScored = %d, want %d (the bridge only)", result.PointsScored, PointsBridge)
	}
}

func TestCheckCollisions_MissileSpentOnFirstTarget(t *testing.T) {
	t.Parallel()

	// A fuel depot and a helicopter both overlap the missile. The missile strikes one
	// and is spent: the other survives, and only one score is awarded.
	terrainEdges := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects,
		&state.ViewportObject{X: 100 * domain.SubpixelScale, Y: 46 * domain.SubpixelScale, Type: domain.ObjectFuel},
		&state.ViewportObject{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Type: domain.ObjectHelicopterReg},
	)

	m := state.PlayerMissile{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

	if len(result.DestroyObjects) != 1 {
		t.Errorf("DestroyObjects = %v, want exactly one", result.DestroyObjects)
	}
	if result.PointsScored != PointsFuel {
		t.Errorf("PointsScored = %d, want %d (the first target only)", result.PointsScored, PointsFuel)
	}
	if m.Active {
		t.Error("missile should be spent")
	}
}

func TestApplyStrike_ObjectsWithoutBridge(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectShip})

	targets := [2]target{
		bridgeTarget{vp: vp, y: 100 * domain.SubpixelScale, active: false, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	m := playerMissile{x: 100, y: 50}

	var result CollisionResult
	if !applyStrike(m, targets[:], &result) {
		t.Fatal("expected a hit")
	}

	if len(result.DestroyObjects) != 1 || result.DestroyObjects[0] != 0 {
		t.Errorf("DestroyObjects = %v, want [0]", result.DestroyObjects)
	}
}

func TestApplyStrike_NothingHit(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	targets := [2]target{
		bridgeTarget{vp: vp, y: 100, active: false, destroyed: false},
		viewportObjectsTarget{vp: vp},
	}
	m := playerMissile{x: 100, y: 50}

	if applyStrike(m, targets[:], &CollisionResult{}) {
		t.Error("expected no hit")
	}
}

func TestCheckCollisions_PlaneVsTerrain(t *testing.T) {
	t.Parallel()

	terrainEdges := bankedTerrain(130, 200)

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, false, 0, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied from terrain collision")
	}
}

func TestCheckCollisions_PlaneVsFuelDepot_Refueling(t *testing.T) {
	t.Parallel()

	terrainEdges := openTerrain()

	vp := state.NewViewport()
	// planeX=120 (screen px). Fuel depot at screen X=118 (SP) overlaps plane.
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118 * domain.SubpixelScale, Y: (domain.PlaneY - 10) * domain.SubpixelScale, Type: domain.ObjectFuel})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

	if !result.Refueling {
		t.Error("expected Refueling")
	}
	if result.PlayerDied {
		t.Error("fuel depot should not kill the player")
	}
}

func TestCheckCollisions_PlaneVsBridge(t *testing.T) {
	t.Parallel()

	terrainEdges := openTerrain()

	var m state.PlayerMissile
	var hm state.HeliMissile

	// bridgeY=140 px = 1120 sp: bridgeYPx=140, bridgeTop=118. The plane's collision row is
	// its origin, PlaneY=120, which is inside (118,140] → hit.
	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 140*domain.SubpixelScale, false, 0)

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

	terrainEdges := openTerrain()

	vp := state.NewViewport()
	// Heli at screen X=118 (SP), PlaneY (SP). Plane at screen X=120. Overlaps.
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118 * domain.SubpixelScale, Y: domain.PlaneY * domain.SubpixelScale, Type: domain.ObjectHelicopterReg})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

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

	terrainEdges := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 118, Y: domain.PlaneY, Type: domain.ObjectTank, TankLocation: domain.TankLocationRoad})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

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

	terrainEdges := openTerrain()

	// Missile in SP; bridge at 60 px = 480 sp → bridgeYPx=60, bridgeTop=38. Missile y=50 is in [38,60).
	m := state.PlayerMissile{X: 128 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 60*domain.SubpixelScale, false, 0)

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

	terrainEdges := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectShip})

	m := state.PlayerMissile{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

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

	terrainEdges := openTerrain()

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{X: 98 * domain.SubpixelScale, Y: 48 * domain.SubpixelScale, Type: domain.ObjectTank})

	m := state.PlayerMissile{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

	if !m.Active {
		t.Error("missile should remain active when passing through tank")
	}
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty", result.DestroyObjects)
	}
}

func TestCheckCollisions_PlaneStraddlingRoadTakesBridge(t *testing.T) {
	t.Parallel()

	// Banks at [112,144). The plane straddles the left road edge while level with the
	// bridge: it touches both, so the bridge is claimed before terrain kills it.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(111*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 140*domain.SubpixelScale, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied")
	}
	if !result.BridgeHit {
		t.Error("plane straddling the road at the bridge should take the bridge with it")
	}
	if result.PointsScored != PointsBridge {
		t.Errorf("PointsScored = %d, want %d", result.PointsScored, PointsBridge)
	}
}

func TestCheckCollisions_PlaneOnRoadClearOfBridge_DiesOnly(t *testing.T) {
	t.Parallel()

	// Fully out over the road, clear of the bridge span: terrain claims it, no points.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(160*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 140*domain.SubpixelScale, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied")
	}
	if result.BridgeHit {
		t.Error("plane clear of the bridge span should not hit it")
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
}

func TestCheckCollisions_PlaneOnFuelDepotOverBank_RefuelsAndDies(t *testing.T) {
	t.Parallel()

	// The plane overlaps a fuel depot while also over the bank. It overlaps both, so
	// both outcomes apply: it refuels and it dies.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: 160 * domain.SubpixelScale, Y: domain.PlaneY * domain.SubpixelScale, Type: domain.ObjectFuel,
	})

	var m state.PlayerMissile
	var hm state.HeliMissile

	result := CheckCollisions(160*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

	if !result.Refueling {
		t.Fatal("expected Refueling")
	}
	if !result.PlayerDied {
		t.Error("a fuel depot does not shield the plane from the bank underneath it")
	}
}

func TestCheckCollisions_PlaneDeathEndsTheFrame(t *testing.T) {
	t.Parallel()

	// The plane crashes into the bank on the same frame the missile is lined up on a
	// helicopter. The plane's collision resolves first, so the missile never scores.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	vp := state.NewViewport()
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: 126 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Type: domain.ObjectHelicopterReg,
	})

	m := state.PlayerMissile{X: 128 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(160*domain.SubpixelScale, &m, &hm, vp, terrainEdges, false, 0, false, 0)

	if !result.PlayerDied {
		t.Fatal("expected PlayerDied")
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
	if len(result.DestroyObjects) != 0 {
		t.Errorf("DestroyObjects = %v, want empty", result.DestroyObjects)
	}
	if !m.Active {
		t.Error("missile should be untouched when the frame ends at the plane's death")
	}
}

func TestCheckCollisions_MissileVsTerrain_Destroyed(t *testing.T) {
	t.Parallel()

	// Banks at [112,144). The missile is over the left bank.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	m := state.PlayerMissile{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, false, 0, false, 0)

	if m.Active {
		t.Error("missile over the bank should be destroyed by terrain")
	}
	if result.PointsScored != 0 {
		t.Errorf("PointsScored = %d, want 0", result.PointsScored)
	}
	if len(result.ExplosionFragments) != 0 {
		t.Errorf("fragment count = %d, want 0", len(result.ExplosionFragments))
	}
}

func TestCheckCollisions_MissileVsTerrain_BlocksBridgeHit(t *testing.T) {
	t.Parallel()

	// The missile is inside the bridge's Y band but over the left bank: the terrain
	// stops it before it can reach the bridge.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	m := state.PlayerMissile{X: 100 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 60*domain.SubpixelScale, false, 0)

	if result.BridgeHit {
		t.Error("missile stopped by terrain should not hit the bridge")
	}
	if m.Active {
		t.Error("missile should be destroyed by terrain")
	}
}

func TestCheckCollisions_MissileBlankRowsDoNotHitTerrain(t *testing.T) {
	t.Parallel()

	// The bank cuts in only across the missile frame's blank trailing rows. Those rows
	// have no set pixels, so they collide with nothing.
	const missileY = 50

	terrainEdges := func(_ domain.Px, y int) (domain.Px, domain.Px) {
		if y >= missileY+assets.SpritePlayerMissileOpaqueHeight && y < missileY+assets.SpritePlayerMissileHeight {
			return 200, 256
		}
		return 0, 256
	}

	m := state.PlayerMissile{X: 128 * domain.SubpixelScale, Y: missileY * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, false, 0, false, 0)

	if !m.Active {
		t.Error("terrain meeting only the missile's blank rows should not remove it")
	}
}

func TestCheckCollisions_MissileOverRiver_ReachesBridge(t *testing.T) {
	t.Parallel()

	// Same banks, but the missile is over the river: it reaches the bridge.
	terrainEdges := bankedTerrain(bridgeLeftX, bridgeRightX)

	m := state.PlayerMissile{X: 124 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 60*domain.SubpixelScale, false, 0)

	if !result.BridgeHit {
		t.Error("missile over the river should hit the bridge")
	}
	if m.Active {
		t.Error("missile should be deactivated after a bridge hit")
	}
}

func TestCheckCollisions_MissileVsDestroyedBridge_PassesThrough(t *testing.T) {
	t.Parallel()

	terrainEdges := openTerrain()

	m := state.PlayerMissile{X: 128 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale, Active: true}
	var hm state.HeliMissile

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, true, 60*domain.SubpixelScale, true, 0)

	if result.BridgeHit {
		t.Error("destroyed bridge should not register a hit")
	}
	if !m.Active {
		t.Error("missile should remain active when bridge is destroyed")
	}
}

func TestCheckCollisions_HeliMissileVsPlane(t *testing.T) {
	t.Parallel()

	terrainEdges := openTerrain()

	var m state.PlayerMissile
	hm := state.HeliMissile{X: 121 * domain.SubpixelScale, Y: (domain.PlaneY + 2) * domain.SubpixelScale, Active: true}

	result := CheckCollisions(120*domain.SubpixelScale, &m, &hm, state.NewViewport(), terrainEdges, false, 0, false, 0)

	if !result.PlayerDied {
		t.Error("expected PlayerDied from helicopter missile")
	}
}

func TestApplyBridgeDestroyedTanks_InGap(t *testing.T) {
	t.Parallel()

	vp := state.NewViewport()
	// X = ($70 - spriteWidth) in SP: (X>>3)+spriteWidth = $70 → exactly at the left edge of the gap.
	vp.Objects = append(vp.Objects, &state.ViewportObject{
		X: (tankGapLeftEdge - assets.SpriteTankWidth).ToSP(), Y: 50 * domain.SubpixelScale,
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

func TestApplyBridgeDestroyedTanks_OnBank_LateLevel_LeftBank(t *testing.T) {
	t.Parallel()

	// Tank is left of the gap (X=0x20 px = 256 sp; (256>>3)+10 = 42 < 0x70 = 112).
	vp := state.NewViewport()
	obj := &state.ViewportObject{
		X: 0x20 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
		MinX:         8,   // road-tank bounds (wide, full screen)
		MaxX:         238, // road-tank bounds
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
	if obj.MinX != 0 {
		t.Errorf("MinX = %d, want 0", obj.MinX)
	}
	wantMaxX := (tankGapLeftEdge - assets.SpriteTankWidth - boundaryPaddingPx).ToSP()
	if obj.MaxX != wantMaxX {
		t.Errorf("MaxX = %d, want %d (gap left edge minus sprite width minus padding, in sp)", obj.MaxX, wantMaxX)
	}
}

func TestApplyBridgeDestroyedTanks_OnBank_LateLevel_RightBank(t *testing.T) {
	t.Parallel()

	// Tank is right of the gap (X=0xA0 px = 1280 sp; (1280>>3) = 160 > 0x90 = 144).
	vp := state.NewViewport()
	obj := &state.ViewportObject{
		X: 0xA0 * domain.SubpixelScale, Y: 50 * domain.SubpixelScale,
		Type:         domain.ObjectTank,
		TankLocation: domain.TankLocationRoad,
		Activated:    true,
		MinX:         8,   // road-tank bounds (wide, full screen)
		MaxX:         238, // road-tank bounds
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
	wantMinX := (tankGapRightEdge + boundaryPaddingPx).ToSP()
	if obj.MinX != wantMinX {
		t.Errorf("MinX = %d, want %d (gap right edge plus padding, in sp)", obj.MinX, wantMinX)
	}
	wantMaxX := (domain.Px(platform.ScreenWidth) - assets.SpriteTankWidth).ToSP()
	if obj.MaxX != wantMaxX {
		t.Errorf("MaxX = %d, want %d (screen width minus sprite width, in sp)", obj.MaxX, wantMaxX)
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
		X: (tankGapLeftEdge - assets.SpriteTankWidth - 1).ToSP(), Y: 50,
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
