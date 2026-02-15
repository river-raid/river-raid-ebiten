package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// Fuel system constants.
const (
	fuelConsumeTickMask = 1   // consume on even ticks (tick & 1 == 0)
	fuelLowThreshold    = 64  // low fuel warning threshold
	fuelRefuelCap       = 252 // "tank full" cap during refueling
)

// FuelResult describes what happened during fuel processing.
type FuelResult int

// Fuel results.
const (
	FuelResultNormal FuelResult = iota
	FuelResultLowFuel
	FuelResultNoFuel
)

// UpdateFuel processes fuel consumption and refueling for one frame.
// tick is the current frame counter. refueling is true if the plane is over a depot.
// Returns the new fuel level and any triggered events.
func UpdateFuel(fuel, tick int, refueling bool) (int, FuelResult) {
	if refueling {
		fuel += domain.FuelIntakeAmount
		if fuel > fuelRefuelCap {
			fuel = fuelRefuelCap
		}
	} else if tick&fuelConsumeTickMask == 0 {
		fuel--
	}

	if fuel <= 0 {
		return 0, FuelResultNoFuel
	}

	if fuel < fuelLowThreshold {
		return fuel, FuelResultLowFuel
	}

	return fuel, FuelResultNormal
}
