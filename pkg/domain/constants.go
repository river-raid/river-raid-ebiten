package domain

// Asset constants.
const (
	NumLevels                 = 48
	NumFragmentsPerLevel      = 64
	NumLinesPerTerrainProfile = 16
	NumSpawnSlotsPerLevel     = 128
	NumTerrainProfiles        = 15
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

// Scoring constants.
const (
	PointsShip          = 30
	PointsHelicopterReg = 60
	PointsBalloon       = 60
	PointsFuel          = 80
	PointsFighter       = 100
	PointsHelicopterAdv = 150
	PointsTank          = 250
	PointsBridge        = 500
)

// Timing constants.
const (
	ActivationIntervalNormal = 31
)

// Player constants.
const (
	LivesInitial = 4
	PlaneStartX  = 120
	PlaneY       = 128
)

// ViewportHeight is the height of the visible game area in pixels.
const (
	ViewportHeight = 136
)
