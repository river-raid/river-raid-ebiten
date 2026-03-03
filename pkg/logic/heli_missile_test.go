package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestFireHeliMissile_LaunchesWhenInactive(t *testing.T) {
	t.Parallel()

	hm := &state.HeliMissile{}
	FireHeliMissile(hm, 64, 20, domain.OrientationLeft)

	if !hm.Active {
		t.Fatal("missile not active after firing")
	}
	if hm.Orientation != domain.OrientationLeft {
		t.Errorf("orientation: got %v, want Left", hm.Orientation)
	}
	// X aligned to 8-pixel boundary: 64 & 0xF8 = 64.
	if hm.X != 64 {
		t.Errorf("X: got %d, want 64", hm.X)
	}
	// Y starts 4 pixels below helicopter.
	if hm.Y != 24 {
		t.Errorf("Y: got %d, want 24", hm.Y)
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

	hm := &state.HeliMissile{Active: true, X: 100, Y: 50, Orientation: domain.OrientationLeft}
	updateHeliMissile(hm)

	if hm.X != 100-heliMissileHorizSpeed {
		t.Errorf("X after left move: got %d, want %d", hm.X, 100-heliMissileHorizSpeed)
	}
	// Y must not change — scroll system handles that.
	if hm.Y != 50 {
		t.Errorf("Y changed in updateHeliMissile: got %d, want 50", hm.Y)
	}
	if !hm.Active {
		t.Error("missile deactivated prematurely")
	}
}

func TestUpdateHeliMissile_DeactivatesAtLeftEdge(t *testing.T) {
	t.Parallel()

	hm := &state.HeliMissile{Active: true, X: 4, Y: 50, Orientation: domain.OrientationLeft}
	updateHeliMissile(hm) // X becomes 4-8 = -4 < 0

	if hm.Active {
		t.Error("missile still active after going off left edge")
	}
}

func TestUpdateHeliMissile_DeactivatesAtRightEdge(t *testing.T) {
	t.Parallel()

	hm := &state.HeliMissile{Active: true, X: platform.ScreenWidth - 4, Y: 50, Orientation: domain.OrientationRight}
	updateHeliMissile(hm) // X becomes ScreenWidth+4 >= ScreenWidth

	if hm.Active {
		t.Error("missile still active after going off right edge")
	}
}

func TestUpdateHeliMissile_NoOpWhenInactive(t *testing.T) {
	t.Parallel()

	hm := &state.HeliMissile{Active: false, X: 100, Y: 50}
	updateHeliMissile(hm)

	if hm.X != 100 || hm.Y != 50 {
		t.Errorf("inactive missile state changed: X=%d Y=%d", hm.X, hm.Y)
	}
}
