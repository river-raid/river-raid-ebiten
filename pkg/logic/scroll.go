package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TerrainRenderer is the interface for rendering terrain fragments and querying edges.
type TerrainRenderer interface {
	RenderFragment(frag assets.TerrainFragment, bufY int, bridgeDestroyed bool)
	GetEdges(x, y, spriteHeight int) (leftX, rightX int)
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

	// Update viewport atomically: scroll, spawn, and activate objects.
	updateViewportForScroll(s, spawnIdx, count, terrain)
}

// updateViewportForScroll performs all viewport updates for a scroll event.
// This includes scrolling existing objects, spawning new objects, and activating objects.
// Boundaries are initialized for newly spawned enemies at spawn time.
func updateViewportForScroll(s *state.GameState, spawnIdx, speed int, terrain TerrainRenderer) {
	// Step 1: Scroll all objects down and remove those off-screen.
	s.Viewport.ScrollObjects(speed)

	// Step 2: Spawn new objects based on scroll position.
	spawnFromScroll(s, spawnIdx, terrain)

	// Step 3: Increment tick counter.
	s.Viewport.Tick++

	// Step 4: Activate objects based on tick counter.
	s.Viewport.ActivateObjects()
}

// spawnFromScroll spawns new objects based on the current spawn index.
// Initializes movement boundaries for enemies at spawn time.
func spawnFromScroll(s *state.GameState, spawnIdx int, terrain TerrainRenderer) {
	if spawnIdx == s.Viewport.SpawnIndex {
		return // already spawned this object
	}

	s.Viewport.SpawnIndex = spawnIdx

	obj := state.NewViewportObject(assets.SpawnSlots[s.BridgeIndex][spawnIdx])
	if obj == nil {
		return // empty spawn slot
	}

	// Initialize movement boundaries for enemies at spawn time.
	initializeObjectBoundaries(obj, terrain, s.ScrollY)

	s.Viewport.Objects = append(s.Viewport.Objects, obj)
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
	// terrainBufferHeight = domain.TotalViewportHeight + domain.NumLinesPerTerrainProfile = 144 + 16 = 160.
	const bufferHeight = domain.TotalViewportHeight + domain.NumLinesPerTerrainProfile

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
