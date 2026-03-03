package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestMoveFighter_WrapsLeft(t *testing.T) {
	t.Parallel()

	slot := state.ViewportSlot{X: 2, Orientation: domain.OrientationLeft, Activated: true}
	moveFighter(&slot)

	if slot.X != fighterResetLeftX {
		t.Errorf("fighter wrap left: got X=%d, want %d", slot.X, fighterResetLeftX)
	}
}

func TestMoveFighter_WrapsRight(t *testing.T) {
	t.Parallel()

	slot := state.ViewportSlot{X: 232, Orientation: domain.OrientationRight, Activated: true}
	moveFighter(&slot)

	if slot.X != fighterResetRightX {
		t.Errorf("fighter wrap right: got X=%d, want %d", slot.X, fighterResetRightX)
	}
}

func TestMoveShipOrHelicopter_EvenTickOnly(t *testing.T) {
	t.Parallel()

	slot := state.ViewportSlot{X: 100, Orientation: domain.OrientationRight, Activated: true}

	// Odd tick: no movement.
	moveShipOrHelicopter(&slot, 1)

	if slot.X != 100 {
		t.Errorf("odd tick: got X=%d, want 100", slot.X)
	}

	// Even tick: moves right.
	moveShipOrHelicopter(&slot, 2)

	if slot.X != 102 {
		t.Errorf("even tick: got X=%d, want 102", slot.X)
	}
}

func TestMoveBalloon_Every4thFrame(t *testing.T) {
	t.Parallel()

	slot := state.ViewportSlot{X: 100, Orientation: domain.OrientationRight, Activated: true}

	// tick & 3 != 1: no movement.
	moveBalloon(&slot, 0)

	if slot.X != 100 {
		t.Errorf("tick 0: got X=%d, want 100", slot.X)
	}

	// tick & 3 == 1: moves.
	moveBalloon(&slot, 1)

	if slot.X != 102 {
		t.Errorf("tick 1: got X=%d, want 102", slot.X)
	}
}
