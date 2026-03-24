package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestFireHeliMissile_LaunchesWhenInactive(t *testing.T) {
	t.Parallel()

	// heliX=512 sp (64 px), heliY=160 sp (20 px) — both already on 64-sp boundary.
	const heliX = 64 * domain.SubpixelScale
	const heliY = 20 * domain.SubpixelScale

	hm := &state.HeliMissile{}
	FireHeliMissile(hm, heliX, heliY, domain.OrientationLeft)

	if !hm.Active {
		t.Fatal("missile not active after firing")
	}
	if hm.Orientation != domain.OrientationLeft {
		t.Errorf("orientation: got %v, want Left", hm.Orientation)
	}
	// X aligned to heliMissileAlignSP boundary.
	wantX := heliX &^ (heliMissileAlignSP - 1)
	if hm.X != wantX {
		t.Errorf("X: got %d, want %d", hm.X, wantX)
	}
	// Y starts heliMissileSpawnOffYSP below helicopter.
	wantY := heliY + heliMissileSpawnOffYSP
	if hm.Y != wantY {
		t.Errorf("Y: got %d, want %d", hm.Y, wantY)
	}
}

func TestFireHeliMissile_NoOpWhenAlreadyActive(t *testing.T) {
	t.Parallel()

	hm := &state.HeliMissile{Active: true, X: 10, Y: 10}
	FireHeliMissile(hm, 100, 50, domain.OrientationRight)

	// State must be unchanged.
	if hm.X != 10 || hm.Y != 10 {
		t.Errorf("state changed while missile already active: X=%d Y=%d", hm.X, hm.Y)
	}
}

func TestUpdateHeliMissile_MovesHorizontallyOnly(t *testing.T) {
	t.Parallel()

	const startX = 100 * domain.SubpixelScale
	const startY = 50 * domain.SubpixelScale

	hm := &state.HeliMissile{Active: true, X: startX, Y: startY, Orientation: domain.OrientationLeft}
	updateHeliMissile(hm, newMockTerrainBuffer(), 0)

	if hm.X != startX-heliMissileHorizSP {
		t.Errorf("X after left move: got %d, want %d", hm.X, startX-heliMissileHorizSP)
	}
	// Y must not change — scroll system handles that.
	if hm.Y != startY {
		t.Errorf("Y changed in updateHeliMissile: got %d, want %d", hm.Y, startY)
	}
	if !hm.Active {
		t.Error("missile deactivated prematurely")
	}
}

func TestUpdateHeliMissile_DeactivatesAtLeftEdge(t *testing.T) {
	t.Parallel()

	// Start just inside screen-left so one step takes it off screen.
	startX := heliMissileHorizSP - domain.SubpixelScale // goes negative after one step
	hm := &state.HeliMissile{Active: true, X: startX, Y: 50 * domain.SubpixelScale, Orientation: domain.OrientationLeft}
	updateHeliMissile(hm, newMockTerrainBuffer(), 0)

	if hm.Active {
		t.Error("missile still active after going off left edge")
	}
}

func TestUpdateHeliMissile_DeactivatesAtRightEdge(t *testing.T) {
	t.Parallel()

	// Start just inside screen-right so one step takes it off screen.
	startX := domain.SP(platform.ScreenWidth*domain.SubpixelScale - domain.SubpixelScale)
	hm := &state.HeliMissile{Active: true, X: startX, Y: 50 * domain.SubpixelScale, Orientation: domain.OrientationRight}
	updateHeliMissile(hm, newMockTerrainBuffer(), 0)

	if hm.Active {
		t.Error("missile still active after going off right edge")
	}
}

func TestUpdateHeliMissile_NoOpWhenInactive(t *testing.T) {
	t.Parallel()

	const startX = 100 * domain.SubpixelScale
	const startY = 50 * domain.SubpixelScale

	hm := &state.HeliMissile{Active: false, X: startX, Y: startY}
	updateHeliMissile(hm, newMockTerrainBuffer(), 0)

	if hm.X != startX || hm.Y != startY {
		t.Errorf("inactive missile state changed: X=%d Y=%d", hm.X, hm.Y)
	}
}

func TestUpdateHeliMissile_DeactivatesOnTerrainHit(t *testing.T) {
	t.Parallel()

	terrain := newMockTerrainBuffer()
	// River from x=50 to x=200 screen pixels at buffer Y=50.
	terrain.setEdges(50, 50, 200)

	// Right-facing missile approaching right bank (in SP).
	// After move: X = 194*sp + 16sp = 194*8+16 = 1568. startX >> 3 = 196, spans [196,203] → hits bank.
	startX := domain.SP(194 * domain.SubpixelScale)
	hm := &state.HeliMissile{Active: true, X: startX, Y: 50 * domain.SubpixelScale, Orientation: domain.OrientationRight}
	updateHeliMissile(hm, terrain, 0)

	if hm.Active {
		t.Error("missile still active after hitting right bank")
	}
}

func TestUpdateHeliMissile_StaysActiveOverRiver(t *testing.T) {
	t.Parallel()

	terrain := newMockTerrainBuffer()
	terrain.setEdges(50, 50, 200)

	// Right-facing missile well within the river (in SP).
	// After move: X = 100*8+16 = 816. startX >> 3 = 102, spans [102,109] → clear of bank at 200.
	startX := domain.SP(100 * domain.SubpixelScale)
	hm := &state.HeliMissile{Active: true, X: startX, Y: 50 * domain.SubpixelScale, Orientation: domain.OrientationRight}
	updateHeliMissile(hm, terrain, 0)

	if !hm.Active {
		t.Error("missile deactivated over open river")
	}
}
