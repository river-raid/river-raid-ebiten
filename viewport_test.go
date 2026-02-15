package main

import "testing"

func TestViewport_SpawnFromScroll(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 12 is a fuel depot at X=120.
	v.SpawnFromScroll(0, 12)

	if len(v.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(v.Slots))
	}

	if v.Slots[0].Type != ObjectFuel || v.Slots[0].X != 120 {
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

func TestViewport_SpawnSkipsRocks(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 19 is a rock.
	v.SpawnFromScroll(0, 19)

	if len(v.Slots) != 0 {
		t.Errorf("expected 0 slots for rock, got %d", len(v.Slots))
	}
}

func TestViewport_ActivateObjects(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Slots = append(v.Slots, Slot{X: 100, Type: ObjectShip})

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
		Slot{X: 100, Y: ViewportHeight - 2, Type: ObjectShip},
		Slot{X: 50, Y: 0, Type: ObjectHelicopterReg},
	)

	v.ScrollObjects(4)

	if len(v.Slots) != 1 {
		t.Fatalf("expected 1 slot after scroll, got %d", len(v.Slots))
	}

	if v.Slots[0].Type != ObjectHelicopterReg {
		t.Errorf("remaining slot = %+v, want helicopter", v.Slots[0])
	}
}

func TestViewport_Clear(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Slots = append(v.Slots, Slot{X: 1}, Slot{X: 2})
	v.Clear()

	if len(v.Slots) != 0 {
		t.Errorf("expected 0 slots after clear, got %d", len(v.Slots))
	}
}
