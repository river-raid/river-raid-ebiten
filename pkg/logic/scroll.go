package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TerrainRenderer is the interface for rendering terrain fragments and querying edges.
type TerrainRenderer interface {
	RenderFragment(frag assets.TerrainFragment, bufY int, bridgeDestroyed bool)
	GetEdges(x, y int) (leftX, rightX int)
}

// FragmentToRender holds information about a terrain fragment that needs rendering.
type FragmentToRender struct {
	Fragment assets.TerrainFragment
	Y        int
}

// Scroll and terrain generation constants.
const (
	bridgeLoopStart  = 33
	bridgeLoopLength = 15
)

// advanceAndRender advances the scroll by the given number of lines and renders
// all necessary terrain fragments. This is the high-level API for scroll operations.
// It handles all scroll state updates, terrain rendering, and viewport updates atomically.
func advanceAndRender(
	s *state.GameState,
	count int,
	terrain TerrainRenderer,
) {
	frags, spawnIdx := advanceLines(s, count)

	// Render all fragments that need to be drawn.
	for _, f := range frags {
		terrain.RenderFragment(f.Fragment, f.Y, s.BridgeDestroyed)
	}

	// Update viewport atomically: spawn, scroll, and activate objects.
	s.Viewport.UpdateForScroll(s.BridgeIndex, spawnIdx, count)

	// Initialize movement boundaries for newly spawned enemies.
	// Pass the terrain buffer and current scroll position so boundaries can be
	// queried from the rendered terrain (single source of truth).
	InitializeEnemyBoundaries(s.Viewport, terrain, s.ScrollY)
}

// advanceLines advances the scroll by the given number of lines.
// ScrollY decreases (viewport moves up in buffer), revealing new terrain at the top.
// Returns a slice of fragments that need to be rendered and the current spawn index.
func advanceLines(s *state.GameState, count int) (fragments []FragmentToRender, spawnIndex int) {
	var toRender []FragmentToRender

	// We need bufferHeight for wrapping.
	// Since bufferHeight is not currently in GameState, we'll assume it's calculated from ScrollY and NextRenderY
	// but wait, the original code used a passed-in bufferHeight.
	// In the new architecture, where does bufferHeight live?
	// For now, let's assume it's a constant or we can get it from somewhere.
	// Looking at game.go, terrainBufferHeight = domain.ViewportHeight + domain.NumLinesPerTerrainProfile = 136 + 16 = 152.
	const bufferHeight = domain.ViewportHeight + domain.NumLinesPerTerrainProfile

	for range count {
		s.ScrollY--
		s.ScrollOffset++

		// If the viewport top has reached the next render position, generate a fragment.
		if s.ScrollY <= s.NextRenderY+domain.NumLinesPerTerrainProfile {
			frag := nextFragment(s)

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

	// Calculate spawn index from scroll offset
	spawnIdx := (int(s.ScrollOffset) / domain.NumLinesPerSpawnSlot) % domain.NumSpawnSlotsPerLevel //nolint:mnd // formula

	return toRender, spawnIdx
}

func nextFragment(s *state.GameState) assets.TerrainFragment {
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
