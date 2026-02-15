package main

// Scroll and terrain generation constants.
const (
	fragmentsPerLevel = 64
	numLevels         = 48
	bridgeLoopStart   = 33
	bridgeLoopLength  = 15
)

// StartingBridgeValues maps the StartingBridge enum to actual bridge indices.
var StartingBridgeValues = [4]int{1, 5, 20, 30} //nolint:gochecknoglobals // constant table

// ScrollState tracks the terrain generation cursor and scroll position.
type ScrollState struct {
	BridgeIndex     int // current bridge (level) index, 0-based
	FragmentNum     int // current fragment within the bridge (0–63)
	LineInFrag      int // current scanline within the fragment (0–15)
	GeneratedY      int // next Y position in the terrain buffer to render
	ScrollY         int // current scroll offset (top of visible area in buffer)
	BridgeYPosition int // Y position of the current bridge in the viewport
}

// InitFromStartingBridge sets the scroll state to begin at the given starting bridge.
func (s *ScrollState) InitFromStartingBridge(sb StartingBridge) {
	s.BridgeIndex = StartingBridgeValues[sb]
}

// nextFragment returns the terrain fragment at the current scroll position
// and advances the cursor to the next fragment.
func (s *ScrollState) nextFragment() TerrainFragment {
	frag := LevelTerrain[s.BridgeIndex][s.FragmentNum]

	s.FragmentNum++
	if s.FragmentNum >= fragmentsPerLevel {
		s.FragmentNum = 0
		s.BridgeIndex++

		if s.BridgeIndex >= numLevels {
			s.BridgeIndex = (s.BridgeIndex-numLevels)%bridgeLoopLength + bridgeLoopStart
		}
	}

	return frag
}

// advanceLines advances the scroll state by the given number of terrain lines,
// rendering new lines into the terrain buffer as needed.
func (s *ScrollState) advanceLines(tb *TerrainBuffer, count int) {
	for range count {
		// If we need more generated terrain, render the next fragment.
		if s.ScrollY+ViewportHeight >= s.GeneratedY {
			frag := s.nextFragment()
			tb.renderFragment(frag, s.GeneratedY, false)
			s.GeneratedY += fragmentLines
		}

		s.ScrollY++
		s.LineInFrag++

		if s.LineInFrag >= fragmentLines {
			s.LineInFrag = 0
		}
	}
}
