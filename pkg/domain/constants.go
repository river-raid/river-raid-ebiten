package domain

// Asset constants.
const (
	NumLevels                 = 48
	NumFragmentsPerLevel      = 64
	NumLinesPerTerrainProfile = 16
	NumSpawnSlotsPerLevel     = 128
	NumTerrainProfiles        = 15
	NumLinesPerSpawnSlot      = NumFragmentsPerLevel * NumLinesPerTerrainProfile / NumSpawnSlotsPerLevel
)

// Fuel constants.
const (
	FuelCheckInterval   = 3
	FuelIntakeAmount    = 4
	FuelLevelEmpty      = 0
	FuelLevelLow        = 192
	FuelLevelAlmostFull = 252
	FuelLevelFull       = 255
)

// Timing constants.
const (
	ActivationIntervalNormal = 31
)

// Player constants.
const (
	LivesInitial = 4
	PlaneStartX  = 120
	PlaneY       = 120
)

// Viewport height constants.
// VisibleViewportHeight is the number of rows actually shown on screen.
// ViewportBlankZone is the number of hidden top rows used for the scroll-in effect.
// TotalViewportHeight is the full logical game height including the blank zone.
const (
	VisibleViewportHeight = 136
	ViewportBlankZone     = 8
	TotalViewportHeight   = VisibleViewportHeight + ViewportBlankZone
)
