package state

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

func TestViewport_SpawnFromScroll(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 12 is a fuel depot at X=120.
	v.SpawnFromScroll(0, 12)

	if len(v.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(v.Slots))
	}

	if v.Slots[0].Type != domain.ObjectFuel || v.Slots[0].X != 120 {
		t.Errorf("slot = %+v, want fuel at X=120", v.Slots[0])
	}
}

func TestViewport_SpawnSkipsEmptySlots(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 0 is empty.
	v.SpawnFromScroll(0, 0)

	if len(v.Slots) != 0 {
		t.Errorf("expected 0 slots for empty spawn, got %d", len(v.Slots))
	}
}

func TestViewport_SpawnRocks(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 19 is a rock.
	v.SpawnFromScroll(0, 19)

	if len(v.Slots) != 1 {
		t.Fatalf("expected 1 slot for rock, got %d", len(v.Slots))
	}

	if !v.Slots[0].IsRock {
		t.Errorf("expected slot to be a rock, got IsRock=%v", v.Slots[0].IsRock)
	}
}

func TestViewport_ActivateObjects(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Slots = append(v.Slots, ViewportSlot{X: 100, Type: domain.ObjectShip})

	// Tick 0 & mask 31 == 0, so activation should happen.
	v.ActivateObjects()

	if !v.Slots[0].Activated {
		t.Error("expected slot to be activated at tick 0")
	}
}

func TestViewport_ScrollRemovesOffscreen(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Slots = append(v.Slots,
		ViewportSlot{X: 100, Y: domain.ViewportHeight - 2, Type: domain.ObjectShip},
		ViewportSlot{X: 50, Y: 0, Type: domain.ObjectHelicopterReg},
	)

	v.ScrollObjects(4)

	if len(v.Slots) != 1 {
		t.Fatalf("expected 1 slot after scroll, got %d", len(v.Slots))
	}

	if v.Slots[0].Type != domain.ObjectHelicopterReg {
		t.Errorf("remaining slot = %+v, want helicopter", v.Slots[0])
	}
}

func TestViewport_Clear(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Slots = append(v.Slots, ViewportSlot{X: 1}, ViewportSlot{X: 2})
	v.Clear()

	if len(v.Slots) != 0 {
		t.Errorf("expected 0 slots after clear, got %d", len(v.Slots))
	}
}
