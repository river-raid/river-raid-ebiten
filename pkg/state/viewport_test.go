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

	if len(v.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(v.Objects))
	}

	if v.Objects[0].Type != domain.ObjectFuel || v.Objects[0].X != 120 {
		t.Errorf("object = %+v, want fuel at X=120", v.Objects[0])
	}
}

func TestViewport_SpawnSkipsEmptySlots(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 0 is empty.
	v.SpawnFromScroll(0, 0)

	if len(v.Objects) != 0 {
		t.Errorf("expected 0 objects for empty spawn, got %d", len(v.Objects))
	}
}

func TestViewport_SpawnRocks(t *testing.T) {
	t.Parallel()

	v := NewViewport()

	// Level 0, slot 19 is a rock.
	v.SpawnFromScroll(0, 19)

	if len(v.Objects) != 1 {
		t.Fatalf("expected 1 object for rock, got %d", len(v.Objects))
	}

	if !v.Objects[0].IsRock {
		t.Errorf("expected object to be a rock, got IsRock=%v", v.Objects[0].IsRock)
	}
}

func TestViewport_ActivateObjects(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Objects = append(v.Objects, &ViewportObject{X: 100, Type: domain.ObjectShip})

	// Tick 0 & mask 31 == 0, so activation should happen.
	v.ActivateObjects()

	if !v.Objects[0].Activated {
		t.Error("expected object to be activated at tick 0")
	}
}

func TestViewport_ScrollRemovesOffscreen(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Objects = append(v.Objects,
		&ViewportObject{X: 100, Y: domain.ViewportHeight - 2, Type: domain.ObjectShip},
		&ViewportObject{X: 50, Y: 0, Type: domain.ObjectHelicopterReg},
	)

	v.ScrollObjects(4)

	if len(v.Objects) != 1 {
		t.Fatalf("expected 1 object after scroll, got %d", len(v.Objects))
	}

	if v.Objects[0].Type != domain.ObjectHelicopterReg {
		t.Errorf("remaining object = %+v, want helicopter", v.Objects[0])
	}
}

func TestViewport_Clear(t *testing.T) {
	t.Parallel()

	v := NewViewport()
	v.Objects = append(v.Objects, &ViewportObject{X: 1}, &ViewportObject{X: 2})
	v.Clear()

	if len(v.Objects) != 0 {
		t.Errorf("expected 0 objects after clear, got %d", len(v.Objects))
	}
}
