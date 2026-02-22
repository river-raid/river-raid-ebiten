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
