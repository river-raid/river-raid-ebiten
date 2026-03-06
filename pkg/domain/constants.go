package domain

// Asset constants.
const (
	NumLevels                 = 48
	NumFragmentsPerLevel      = 64
	NumLinesPerTerrainProfile = 16
	NumSpawnSlotsPerLevel     = 128
	NumTerrainProfiles        = 15
	NumLinesPerSpawnSlot      = NumFragmentsPerLevel * NumLinesPerTerrainProfile / NumSpawnSlotsPerLevel

	// NumExplosionSpriteFrames is the number of distinct sprite frames in the fragment
	// explosion animation (frames 0–4).
	NumExplosionSpriteFrames = 5
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

// High score slot indices (0-based) corresponding to each StartingBridge option.
const (
	HighScoreSlotBridge01 = 0
	HighScoreSlotBridge05 = 1
	HighScoreSlotBridge20 = 2
	HighScoreSlotBridge30 = 3
)

// highScoreSlotTable maps StartingBridge to a 0-based HighScores slot index.
var highScoreSlotTable = map[StartingBridge]int{ //nolint:gochecknoglobals // constant lookup table
	StartingBridge01: HighScoreSlotBridge01,
	StartingBridge05: HighScoreSlotBridge05,
	StartingBridge20: HighScoreSlotBridge20,
	StartingBridge30: HighScoreSlotBridge30,
}

// HighScoreSlot returns the 0-based HighScores slot index for a StartingBridge value.
func HighScoreSlot(sb StartingBridge) int {
	slot, ok := highScoreSlotTable[sb]
	if !ok {
		panic("domain: unknown StartingBridge value")
	}

	return slot
}

const (
	// DyingFrameCount is the number of frames the dying animation runs.
	DyingFrameCount = 16
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
