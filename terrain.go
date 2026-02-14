package main

// Profile marker bytes used to identify terrain profile variants.
const (
	profileMarkerCanal1        = 0x80
	profileMarkerCanal2        = 0xE0
	profileMarkerRoadAndBridge = 0xC0
)

// TerrainProfile represents a terrain profile variant.
// Resolved at load time into one of three concrete types.
type TerrainProfile interface {
	isTerrainProfile()
}

// RegularProfile stores per-scanline left-edge X offsets.
type RegularProfile struct {
	Values [16]byte
}

func (RegularProfile) isTerrainProfile() {}

// CanalProfile is a marker type for canal terrain (fixed geometry).
type CanalProfile struct{}

func (CanalProfile) isTerrainProfile() {}

// RoadAndBridgeProfile is a marker type for road/bridge terrain (fixed geometry).
type RoadAndBridgeProfile struct{}

func (RoadAndBridgeProfile) isTerrainProfile() {}

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

// resolveTerrainProfiles converts raw 16-byte profile data into typed variants.
//
//nolint:gosec // G602: i is bounded by range over [15] array — same size as raw
func resolveTerrainProfiles(raw *[15][16]byte) [15]TerrainProfile {
	var profiles [15]TerrainProfile

	for i := range profiles {
		data := raw[i]

		switch data[0] {
		case profileMarkerCanal1, profileMarkerCanal2:
			profiles[i] = CanalProfile{}
		case profileMarkerRoadAndBridge:
			profiles[i] = RoadAndBridgeProfile{}
		default:
			profiles[i] = RegularProfile{Values: data}
		}
	}

	return profiles
}
