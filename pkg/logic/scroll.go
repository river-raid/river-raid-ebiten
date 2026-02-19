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
	ScrollOffset    int // 16-bit wrapping scroll counter for spawn index calculation
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

// TerrainRenderer is the interface for rendering terrain fragments.
type TerrainRenderer interface {
	RenderFragment(frag assets.TerrainFragment, bufY int, bridgeDestroyed bool)
}

// ViewportUpdater is the interface for updating viewport state during scroll.
type ViewportUpdater interface {
	UpdateForScroll(bridgeIndex, spawnIdx, speed int)
}

// AdvanceAndRender advances the scroll by the given number of lines and renders
// all necessary terrain fragments. This is the high-level API for scroll operations.
// It handles all scroll state updates, terrain rendering, and viewport updates atomically.
func (s *ScrollState) AdvanceAndRender(
	count, bufferHeight int,
	terrain TerrainRenderer,
	viewport ViewportUpdater,
	bridgeDestroyed bool,
) {
	frags, spawnIdx := s.advanceLines(count, bufferHeight)

	// Render all fragments that need to be drawn.
	for _, f := range frags {
		terrain.RenderFragment(f.Fragment, f.Y, bridgeDestroyed)
	}

	// Update viewport atomically: spawn, scroll, and activate objects.
	viewport.UpdateForScroll(s.BridgeIndex, spawnIdx, count)
}

// advanceLines advances the scroll by the given number of lines.
// ScrollY decreases (viewport moves up in buffer), revealing new terrain at the top.
// Returns a slice of fragments that need to be rendered and the current spawn index.
// bufferHeight is used to wrap buffer Y coordinates to prevent negative values.
// Exposed for testing; game code should use AdvanceAndRender instead.
func (s *ScrollState) advanceLines(count, bufferHeight int) (fragments []FragmentToRender, spawnIndex int) {
	var toRender []FragmentToRender

	for range count {
		s.ScrollY--
		s.ScrollOffset++

		// Wrap ScrollOffset to 16-bit range.
		if s.ScrollOffset >= 0x10000 { //nolint:mnd // 0x10000 = 65536, wraps 16-bit counter
			s.ScrollOffset = 0
		}

		// If the viewport top has reached the next render position, generate a fragment.
		if s.ScrollY <= s.NextRenderY+domain.NumLinesPerTerrainProfile {
			frag := s.NextFragment()

			// Wrap NextRenderY to stay within buffer bounds (circular buffer).
			actualY := s.NextRenderY
			if actualY < 0 || actualY >= bufferHeight {
				actualY = ((actualY % bufferHeight) + bufferHeight) % bufferHeight
			}

			toRender = append(toRender, FragmentToRender{
				Fragment: frag,
				Y:        actualY,
			})
			s.NextRenderY -= domain.NumLinesPerTerrainProfile
		}

		s.LineInFrag++
		if s.LineInFrag >= domain.NumLinesPerTerrainProfile {
			s.LineInFrag = 0
		}
	}

	// Calculate spawn index from scroll offset: (scrollOffset >> 2) & 0x7F
	// Mask to 0x7F to keep within 0-127 range (NumSpawnSlotsPerLevel = 128)
	spawnIdx := (s.ScrollOffset >> 2) & 0x7F //nolint:mnd // 0x7F = 127, masks to valid spawn slot range

	return toRender, spawnIdx
}
