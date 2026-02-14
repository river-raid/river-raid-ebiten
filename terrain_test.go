package main

import "testing"

func TestResolveTerrainProfiles(t *testing.T) {
	raw := [15][16]byte{
		{0x10},                       // < 0x80 → Regular
		{profileMarkerCanal1},        // 0x80 → Canal
		{profileMarkerRoadAndBridge}, // 0xC0 → RoadAndBridge
		{profileMarkerCanal2},        // 0xE0 → Canal
		{0x02, 0x04, 0x06, 0x08},     // < 0x80 → Regular, values preserved
	}

	profiles := resolveTerrainProfiles(&raw)

	assertProfileType[RegularProfile](t, profiles[0], "profile 0")
	assertProfileType[CanalProfile](t, profiles[1], "profile 1")
	assertProfileType[RoadAndBridgeProfile](t, profiles[2], "profile 2")
	assertProfileType[CanalProfile](t, profiles[3], "profile 3")

	// Regular profiles preserve all 16 bytes.
	rp, ok := profiles[4].(RegularProfile)
	if !ok {
		t.Fatal("profile 4: not RegularProfile")
	}

	if rp.Values != raw[4] {
		t.Errorf("profile 4: values not preserved")
	}
}

func assertProfileType[T TerrainProfile](t *testing.T, p TerrainProfile, label string) {
	t.Helper()

	if _, ok := p.(T); !ok {
		t.Errorf("%s: got %T, want %T", label, p, *new(T))
	}
}

func TestLevelTerrain_ProfileIndicesInRange(t *testing.T) {
	for level := range LevelTerrain {
		for frag := range LevelTerrain[level] {
			idx := LevelTerrain[level][frag].ProfileIndex
			if idx < 0 || idx > 14 {
				t.Errorf("level %d fragment %d: profile index %d out of range",
					level, frag, idx)
			}
		}
	}
}

func TestIslands_EdgeModes(t *testing.T) {
	for i, island := range Islands {
		if island.EdgeMode < EdgeMirrored || island.EdgeMode > EdgeOffset {
			t.Errorf("island %d: edge mode %d out of range", i, island.EdgeMode)
		}
	}
}
