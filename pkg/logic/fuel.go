package logic

import "github.com/morozov/river-raid-ebiten/pkg/state"

// FuelState system constants.
const (
	fuelConsumeTickMask = 1   // consume on even frames
	fuelConsumeAmount   = 1   // fuel consumed per eligible frame
	fuelIntakeAmount    = 4   // fuel added per frame while over a depot
	fuelLowThreshold    = 64  // low fuel warning threshold
	fuelRefuelCap       = 252 // "tank full" cap during refueling
)

// UpdateFuel processes fuel consumption and refueling for one frame.
// tick is the current frame counter. refueling is true if the plane is over a depot.
// Returns the new fuel level and any triggered events.
func UpdateFuel(fuel, tick int, refueling bool) (int, state.FuelState) {
	if refueling {
		fuel += fuelIntakeAmount

		if fuel >= fuelRefuelCap {
			if fuel > fuelRefuelCap {
				fuel = fuelRefuelCap
			}
			return fuel, state.FuelStateFull
		}
	} else if tick&fuelConsumeTickMask == 0 {
		fuel -= fuelConsumeAmount
	}

	if fuel <= 0 {
		return 0, state.FuelStateEmpty
	}

	if fuel < fuelLowThreshold {
		return fuel, state.FuelStateLow
	}

	return fuel, state.FuelStateNormal
}
