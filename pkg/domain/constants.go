package domain

// Tps is the number of game ticks per second.
const Tps = 48

// SubpixelShift is the number of fractional bits in subpixel coordinates
// (i.e., log₂ of subpixel per screen pixel). Chosen so that every game speed
// in px/sec is an integer number of subpixel per tick at the current Tps.
const SubpixelShift = 3

// SubpixelScale is the number of subpixel per screen pixel (1 << SubpixelShift).
// Shift a subpixel coordinate right by SubpixelShift to obtain screen pixels.
const SubpixelScale = 1 << SubpixelShift

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

// Timing constants.
// vp.Tick increments once per pixel scrolled.
const (
	ActivationIntervalNormal = 32 // activate every 32 scroll-ticks → 32 px window
	ActivationIntervalFast   = 16 // activate every 16 scroll-ticks → 16 px window (after bridge)
)

// Player constants.
const (
	LivesInitial = 4
	PlaneStartX  = 120
	PlaneY       = 120
)

const (
	// DyingFrameCount is the number of frames the dying animation runs (~1.33 s).
	DyingFrameCount = Tps * 4 / 3
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

// TotalViewportHeightSP is the total viewport height in subpixel.
const TotalViewportHeightSP SP = TotalViewportHeight * SubpixelScale

// Player start position in subpixel.
const (
	PlaneStartXSP SP = PlaneStartX * SubpixelScale
	PlaneYSP      SP = PlaneY * SubpixelScale
)
