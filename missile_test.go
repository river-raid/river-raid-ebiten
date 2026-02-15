package main

import "testing"

func TestMissile_Fire(t *testing.T) {
	t.Parallel()

	var m Missile
	m.Fire(120)

	if !m.Active {
		t.Fatal("expected missile to be active after fire")
	}

	if m.X != 124 || m.Y != 126 {
		t.Errorf("missile position: got (%d,%d), want (124,126)", m.X, m.Y)
	}
}

func TestMissile_FireOnlyOnce(t *testing.T) {
	t.Parallel()

	var m Missile
	m.Fire(120)
	m.Fire(200) // should be ignored

	if m.X != 124 {
		t.Errorf("second fire changed X: got %d, want 124", m.X)
	}
}

func TestMissile_Update(t *testing.T) {
	t.Parallel()

	var m Missile
	m.Fire(120)
	m.Update()

	if m.Y != 120 {
		t.Errorf("after 1 update: Y=%d, want 120", m.Y)
	}
}

func TestMissile_DeactivatesAtTop(t *testing.T) {
	t.Parallel()

	m := Missile{X: 100, Y: 10, Active: true}
	m.Update() // Y = 4, below missileTopY

	if m.Active {
		t.Error("expected missile to deactivate at top of screen")
	}
}
