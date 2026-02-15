package assets

import (
	"testing"
)

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
