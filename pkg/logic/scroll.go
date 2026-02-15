package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// Scroll and terrain generation constants.
const (
	bridgeLoopStart  = 33
	bridgeLoopLength = 15
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
	s.ScrollY = bufferHeight - domain.ViewportHeight
	// New terrain will be rendered just above the viewport.
	s.NextRenderY = s.ScrollY - domain.NumLinesPerTerrainProfile
}

// InitFromStartingBridge sets the scroll state to begin at the given starting bridge.
func (s *ScrollState) InitFromStartingBridge(sb domain.StartingBridge) {
	s.BridgeIndex = StartingBridgeValues[sb]
}

// NextFragment returns the terrain fragment at the current scroll position
// and advances the cursor to the next fragment.
func (s *ScrollState) NextFragment() assets.TerrainFragment {
	frag := assets.LevelTerrain[s.BridgeIndex][s.FragmentNum]

	s.FragmentNum++
	if s.FragmentNum >= domain.NumFragmentsPerLevel {
		s.FragmentNum = 0
		s.BridgeIndex++

		if s.BridgeIndex >= domain.NumLevels {
			s.BridgeIndex = (s.BridgeIndex-domain.NumLevels)%bridgeLoopLength + bridgeLoopStart
		}
	}

	return frag
}

// FragmentToRender holds information about a terrain fragment that needs rendering.
type FragmentToRender struct {
	Fragment assets.TerrainFragment
	Y        int
}

// AdvanceLines advances the scroll by the given number of lines.
// ScrollY decreases (viewport moves up in buffer), revealing new terrain at the top.
// Returns a slice of fragments that need to be rendered.
func (s *ScrollState) AdvanceLines(count int) []FragmentToRender {
	var toRender []FragmentToRender

	for range count {
		s.ScrollY--

		// If the viewport top has reached the next render position, generate a fragment.
		if s.ScrollY <= s.NextRenderY+domain.NumLinesPerTerrainProfile {
			frag := s.NextFragment()
			toRender = append(toRender, FragmentToRender{
				Fragment: frag,
				Y:        s.NextRenderY,
			})
			s.NextRenderY -= domain.NumLinesPerTerrainProfile
		}

		s.LineInFrag++
		if s.LineInFrag >= domain.NumLinesPerTerrainProfile {
			s.LineInFrag = 0
		}
	}

	return toRender
}
