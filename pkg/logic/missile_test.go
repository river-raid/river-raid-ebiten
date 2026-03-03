package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestMissile_Fire(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, 120)

	if !m.Active {
		t.Fatal("expected missile to be active after fire")
	}

	if m.X != 123 || m.Y != 112 {
		t.Errorf("missile position: got (%d,%d), want (123,112)", m.X, m.Y)
	}
}

func TestMissile_FireOnlyOnce(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, 120)
	FireMissile(&m, 200) // should be ignored

	if m.X != 123 {
		t.Errorf("second fire changed X: got %d, want 123", m.X)
	}
}

func TestMissile_Update(t *testing.T) {
	t.Parallel()

	var m state.PlayerMissile
	FireMissile(&m, 120)
	updateMissile(&m)

	if m.Y != 106 {
		t.Errorf("after 1 update: Y=%d, want 106", m.Y)
	}
}

func TestMissile_DeactivatesAtTop(t *testing.T) {
	t.Parallel()

	m := state.PlayerMissile{X: 100, Y: 10, Active: true}
	updateMissile(&m) // Y = 4, below missileTopY

	if m.Active {
		t.Error("expected missile to deactivate at top of screen")
	}
}
