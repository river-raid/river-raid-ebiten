package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// planeXSP converts a screen-pixel plane X to subpixels for test readability.
const testPlaneXSP = 120 * domain.SubpixelScale // 960 sp

func TestMissile_Fire(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, testPlaneXSP)

	if !m.Active {
		t.Fatal("expected missile to be active after fire")
	}

	wantX := testPlaneXSP + missileSpawnOffX
	wantY := domain.PlaneYSP - missileSpawnOffY
	if m.X != wantX || m.Y != wantY {
		t.Errorf("missile position: got (%d,%d), want (%d,%d)", m.X, m.Y, wantX, wantY)
	}
}

func TestMissile_FireOnlyOnce(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, testPlaneXSP)
	wantX := m.X
	FireMissile(&m, 200*domain.SubpixelScale) // should be ignored

	if m.X != wantX {
		t.Errorf("second fire changed X: got %d, want %d", m.X, wantX)
	}
}

func TestMissile_Update(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, testPlaneXSP)
	startY := m.Y
	updateMissile(&m, testPlaneXSP)

	if m.Y != startY-missileSpeed {
		t.Errorf("after 1 update: Y=%d, want %d", m.Y, startY-missileSpeed)
	}
}

func TestMissile_TracksPlaneX(t *testing.T) {
	t.Parallel()

	const movedPlaneXSP = 140 * domain.SubpixelScale

	var m state.PlayerMissile
	FireMissile(&m, testPlaneXSP)
	updateMissile(&m, movedPlaneXSP)

	wantX := movedPlaneXSP + missileSpawnOffX
	if m.X != wantX {
		t.Errorf("missile X after plane move: got %d, want %d", m.X, wantX)
	}
}

func TestMissile_DeactivatesAtTop(t *testing.T) {
	t.Parallel()

	// Y just above missileTopY: one step brings it below.
	startY := missileTopY + missileSpeed - 1
	m := state.PlayerMissile{X: testPlaneXSP, Y: startY, Active: true}
	updateMissile(&m, testPlaneXSP)

	if m.Active {
		t.Error("expected missile to deactivate at top of screen")
	}
}
