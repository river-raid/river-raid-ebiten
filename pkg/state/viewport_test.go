package state

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

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
		&ViewportObject{X: 100, Y: domain.TotalViewportHeight - 2, Type: domain.ObjectShip},
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
