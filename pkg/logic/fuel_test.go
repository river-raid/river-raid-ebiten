package logic

import "testing"

func TestUpdateFuel_ConsumesOnEvenTick(t *testing.T) {
	t.Parallel()

	fuel, _ := UpdateFuel(100, 0, false)
	if fuel != 99 {
		t.Errorf("even tick: fuel = %d, want 99", fuel)
	}

	fuel, _ = UpdateFuel(100, 1, false)
	if fuel != 100 {
		t.Errorf("odd tick: fuel = %d, want 100", fuel)
	}
}

func TestUpdateFuel_Refueling(t *testing.T) {
	t.Parallel()

	fuel, _ := UpdateFuel(100, 0, true)
	if fuel != 104 {
		t.Errorf("refuel: fuel = %d, want 104", fuel)
	}
}

func TestUpdateFuel_RefuelCap(t *testing.T) {
	t.Parallel()

	fuel, _ := UpdateFuel(250, 0, true)
	if fuel != 252 {
		t.Errorf("refuel cap: fuel = %d, want 252", fuel)
	}
}

func TestUpdateFuel_EmptyTriggersDeath(t *testing.T) {
	t.Parallel()

	fuel, result := UpdateFuel(1, 0, false)
	if fuel != 0 {
		t.Errorf("fuel = %d, want 0", fuel)
	}

	if result != FuelResultNoFuel {
		t.Error("expected FuelResultNoFuel")
	}
}

func TestUpdateFuel_LowFuelWarning(t *testing.T) {
	t.Parallel()

	_, result := UpdateFuel(63, 1, false)
	if result != FuelResultLowFuel {
		t.Error("expected FuelResultLowFuel at 63")
	}

	_, result = UpdateFuel(64, 1, false)
	if result != FuelResultNormal {
		t.Error("expected no FuelResultNormal at 64")
	}
}

func TestUpdateFuel_FuelFullOnCapTransition(t *testing.T) {
	t.Parallel()

	// Fuel just below cap → transitions to cap: FuelFull should be set.
	_, result := UpdateFuel(250, 0, true)
	if result != FuelResultFullFuel {
		t.Error("expected FuelResultFullFuel when fuel reaches cap")
	}

	// Not refueling with enough fuel.
	_, result = UpdateFuel(250, 0, false)
	if result != FuelResultNormal {
		t.Error("expected no FuelResultNormal when not refueling")
	}
}
