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

	var m PlayerMissile
	var hm HeliMissile
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, &vp, leftX, rightX, false, 0, false)

	if !result.PlayerDied {
		t.Error("expected player death from terrain collision")
	}
}

func TestCheckCollisions_PlaneVsFuelDepot(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	var m PlayerMissile
	var hm HeliMissile
	vp := state.NewViewport()
	vp.Slots = append(vp.Slots, domain.Slot{X: 118, Y: domain.PlaneY - 10, Type: domain.ObjectFuel})

	result := CheckCollisions(120, &m, &hm, &vp, leftX, rightX, false, 0, false)

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

	m := PlayerMissile{X: 100, Y: 50, Active: true}
	var hm HeliMissile
	vp := state.NewViewport()
	vp.Slots = append(vp.Slots, domain.Slot{X: 98, Y: 48, Type: domain.ObjectShip})

	result := CheckCollisions(120, &m, &hm, &vp, leftX, rightX, false, 0, false)

	if result.PointsScored != domain.PointsShip {
		t.Errorf("points = %d, want %d", result.PointsScored, domain.PointsShip)
	}

	if len(result.DestroySlots) != 1 || result.DestroySlots[0] != 0 {
		t.Errorf("DestroySlots = %v, want [0]", result.DestroySlots)
	}

	if m.Active {
		t.Error("missile should be deactivated after hit")
	}
}

func TestCheckCollisions_HeliMissileVsPlayer(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	var m PlayerMissile
	hm := HeliMissile{X: 121, Y: domain.PlaneY + 2, Active: true}
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, &vp, leftX, rightX, false, 0, false)

	if !result.PlayerDied {
		t.Error("expected player death from helicopter missile")
	}
}

func TestCheckCollisions_MissileVsBridge(t *testing.T) {
	t.Parallel()

	leftX := func(_ int) int { return 0 }
	rightX := func(_ int) int { return 256 }

	m := PlayerMissile{X: 128, Y: 50, Active: true}
	var hm HeliMissile
	vp := state.NewViewport()

	result := CheckCollisions(120, &m, &hm, &vp, leftX, rightX, true, 60, false)

	if !result.BridgeHit {
		t.Error("expected bridge hit")
	}

	if result.PointsScored != domain.PointsBridge {
		t.Errorf("points = %d, want %d", result.PointsScored, domain.PointsBridge)
	}
}
