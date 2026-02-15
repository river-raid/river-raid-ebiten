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
// The buffer is filled from the bottom up: new terrain is rendered at decreasing Y.
// ScrollY is the buffer Y of the viewport top; it decreases as the player advances.
type ScrollState struct {
	BridgeIndex     int // current bridge (level) index, 0-based
	FragmentNum     int // current fragment within the bridge (0–63)
	LineInFrag      int // current scanline within the fragment (0–15)
	NextRenderY     int // next Y position (top of next fragment) to render into
	ScrollY         int // buffer Y of the viewport top; decreases over time
	BridgeYPosition int // Y position of the current bridge in the viewport
}

// InitScroll sets up initial scroll positions for a given buffer height.
func (s *ScrollState) InitScroll(bufferHeight int) {
	// Start the viewport at the bottom of the buffer.
	s.ScrollY = bufferHeight - ViewportHeight
	// New terrain will be rendered just above the viewport.
	s.NextRenderY = s.ScrollY - fragmentLines
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

// advanceLines advances the scroll by the given number of lines.
// ScrollY decreases (viewport moves up in buffer), revealing new terrain at the top.
// When the viewport reaches the next render position, a new fragment is generated.
func (s *ScrollState) advanceLines(tb *TerrainBuffer, count int) {
	for range count {
		s.ScrollY--

		// If the viewport top has reached the next render position, generate a fragment.
		if s.ScrollY <= s.NextRenderY+fragmentLines {
			frag := s.nextFragment()
			tb.renderFragment(frag, s.NextRenderY, false)
			s.NextRenderY -= fragmentLines
		}

		s.LineInFrag++
		if s.LineInFrag >= fragmentLines {
			s.LineInFrag = 0
		}
	}
}
