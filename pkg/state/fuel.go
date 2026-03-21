package state

// FuelState represents the current fuel condition.
type FuelState int

// FuelState states.
const (
	FuelStateNormal FuelState = iota
	FuelStateLow
	FuelStateEmpty
	FuelStateFull
)
