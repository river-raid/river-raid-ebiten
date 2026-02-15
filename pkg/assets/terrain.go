package assets

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// TerrainProfile represents a terrain profile variant.
// One of RegularProfile, CanalProfile, or RoadAndBridgeProfile.
type TerrainProfile interface {
	isTerrainProfile()
}

// ProfileValues holds per-scanline left-edge X offsets for a regular terrain profile.
type ProfileValues = [domain.NumLinesPerTerrainProfile]byte

// RegularProfile stores per-scanline left-edge X offsets.
type RegularProfile struct {
	Values ProfileValues
}

func (RegularProfile) isTerrainProfile() {}

// CanalProfile is a marker type for canal terrain (fixed geometry).
type CanalProfile struct{}

func (CanalProfile) isTerrainProfile() {}

// RoadAndBridgeProfile is a marker type for road/bridge terrain (fixed geometry).
type RoadAndBridgeProfile struct{}

func (RoadAndBridgeProfile) isTerrainProfile() {}

// EdgeMode controls how the right terrain edge is calculated from the left edge.
type EdgeMode int

// Edge modes.
const (
	EdgeMirrored = iota + 1 // rightX = 2*center - leftX
	EdgeOffset              // rightX = width + leftX
)

// TerrainFragment is a decoded 4-byte terrain entry.
//   - Byte3 is added to profile values for left edge: leftX = profileValue + Byte3
//   - Byte2 is the center/width param for right edge calculation
type TerrainFragment struct {
	ProfileIndex int
	Byte2        int
	Byte3        int
	EdgeMode     EdgeMode
	IslandNum    int
}

// IslandDefinition is a decoded 3-byte island entry.
type IslandDefinition struct {
	ProfileIndex int
	WidthOffset  int
	EdgeMode     EdgeMode
}
