package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// Terrain rendering constants.
const (
	edgeOffsetAdjust   = 6   // subtracted from left edge for edge sprite width
	islandCenterOffset = 138 // added to island left edge to center on screen
	islandDefaultHalf  = 60  // default half-width for island right edge calculation
	centerDivisor      = 2   // divisor for calculating center point
	// the terrain buffer is sized as total viewport height plus one-fragment lookahead.
	terrainBufferHeight = domain.TotalViewportHeight + domain.NumLinesPerTerrainProfile

	// bridgeStartX and bridgeEndX are the pixel boundaries of the canal gap and road/bridge span.
	// The canal/road is 32 pixels wide, centered on the 256-pixel screen.
	bridgeStartX = 112
	bridgeEndX   = 144
)

// TerrainEdges stores the left and right river edges for a single scanline.
type TerrainEdges struct {
	LeftX        int  // leftmost X coordinate of the river (right edge of left bank)
	RightX       int  // rightmost X coordinate of the river (left edge of right bank)
	HasIsland    bool // true if this scanline has an island
	IslandLeftX  int  // left edge of island (if HasIsland)
	IslandRightX int  // right edge of island (if HasIsland)
}

// TerrainBuffer manages an off-screen image for incremental terrain rendering.
// It also stores queryable edge data for each scanline to support O(1) collision detection.
type TerrainBuffer struct {
	buffer PixelBuffer
	image  *CircularImage // kept for drawTerrainBuffer access
	edges  []TerrainEdges // edge data for each scanline (same height as buffer)
}

// NewTerrainBuffer creates a terrain buffer.
func NewTerrainBuffer() *TerrainBuffer {
	circImg := NewCircularImage(platform.ScreenWidth, terrainBufferHeight)
	return &TerrainBuffer{
		buffer: circImg,
		image:  circImg,
		edges:  make([]TerrainEdges, terrainBufferHeight),
	}
}

// Clear fills the entire terrain buffer with black and zeroes all edge data.
// Called on respawn so the scroll-in begins from a blank screen.
func (tb *TerrainBuffer) Clear() {
	tb.image.Clear()
	for i := range tb.edges {
		tb.edges[i] = TerrainEdges{}
	}
}

// EdgeAt returns the TerrainEdges for a single buffer row, wrapping the circular index.
func (tb *TerrainBuffer) EdgeAt(bufY int) TerrainEdges {
	height := len(tb.edges)
	wrappedY := ((bufY % height) + height) % height

	return tb.edges[wrappedY]
}

// GetEdges returns the left and right river boundaries for a sprite at position (x, y) with given height.
// Y coordinates are automatically wrapped to buffer bounds (circular buffer).
// The method checks all scanlines from y to y+spriteHeight-1 and returns the most restrictive
// (narrowest) boundaries across all those scanlines.
// If any scanline has an island, the X coordinate determines which shoulder (left or right)
// the position is in, and returns boundaries for that shoulder only.
// Returns (leftX, rightX) representing the navigable river boundaries for this sprite.
func (tb *TerrainBuffer) GetEdges(x, y, spriteHeight int) (leftX, rightX int) {
	height := len(tb.edges)

	// Initialize with the widest possible boundaries
	leftX = 0
	rightX = platform.ScreenWidth

	// Check all scanlines the sprite overlaps
	for dy := range spriteHeight {
		scanlineY := ((y+dy)%height + height) % height
		edges := tb.edges[scanlineY]

		var scanlineLeft, scanlineRight int

		// If there's no island, use the full river edges.
		if !edges.HasIsland {
			scanlineLeft = edges.LeftX
			scanlineRight = edges.RightX
		} else {
			// Island present: determine which shoulder based on X position.
			// Calculate island center to determine left vs right shoulder.
			islandCenter := (edges.IslandLeftX + edges.IslandRightX) / centerDivisor

			if x < islandCenter {
				// Left shoulder: bounded by left bank and left island edge.
				scanlineLeft = edges.LeftX
				scanlineRight = edges.IslandLeftX
			} else {
				// Right shoulder: bounded by right island edge and right bank.
				scanlineLeft = edges.IslandRightX
				scanlineRight = edges.RightX
			}
		}

		// Use the most restrictive (narrowest) boundaries
		if scanlineLeft > leftX {
			leftX = scanlineLeft
		}
		if scanlineRight < rightX {
			rightX = scanlineRight
		}
	}

	return leftX, rightX
}

// renderRegularLine renders a single scanline of a regular terrain profile.
// leftX is the left bank edge X, rightX is the right bank edge X.
// y is the destination Y in the buffer.
func (tb *TerrainBuffer) renderRegularLine(y, leftX, rightX int, bankColor, riverColor color.RGBA) {
	// Store edge data for collision detection.
	height := len(tb.edges)
	wrappedY := ((y % height) + height) % height
	tb.edges[wrappedY] = TerrainEdges{
		LeftX:     leftX,
		RightX:    rightX,
		HasIsland: false, // will be updated by renderIslandFragment if needed
	}

	// Fill left bank (green) from x=0 to left edge.
	fillRect(tb.buffer, 0, y, leftX, bankColor)

	// Fill river (blue) between banks.
	fillRect(tb.buffer, leftX, y, rightX-leftX, riverColor)

	// Fill right bank (green) from right edge to screen boundary.
	fillRect(tb.buffer, rightX, y, platform.ScreenWidth-rightX, bankColor)
}

// RenderFragment renders a single terrain fragment (16 scanlines) into the buffer.
// bufY is the starting Y position in the buffer (top of the fragment).
// bridgeDestroyed controls whether the bridge gap is rendered for road/bridge profiles.
// Scanlines are rendered bottom-to-top: line 0 at bufY+15, line 15 at bufY.
func (tb *TerrainBuffer) RenderFragment(frag assets.TerrainFragment, bufY int, bridgeDestroyed bool) {
	profile := assets.TerrainProfiles[frag.ProfileIndex]

	bankColor := palette[colorBank]
	riverColor := palette[colorRiver]

	switch p := profile.(type) {
	case assets.RegularProfile:
		for line := range domain.NumLinesPerTerrainProfile {
			// Bottom-to-top rendering: line 0 at bottom (bufY+15), line 15 at top (bufY)
			y := bufY + (domain.NumLinesPerTerrainProfile - 1 - line)
			coordinateLeft := int(p.Values[line]) + frag.Byte3
			leftX := coordinateLeft - edgeOffsetAdjust
			rightX := calculateOtherEdge(frag.Byte2, coordinateLeft, frag.EdgeMode)
			tb.renderRegularLine(y, leftX, rightX, bankColor, riverColor)
		}

		if frag.IslandNum > 0 {
			island := assets.Islands[frag.IslandNum-1]
			tb.renderIslandFragment(bufY, island, bankColor)
		}
	case assets.CanalProfile:
		tb.renderBandedLines(bufY, domain.NumLinesPerTerrainProfile, colorBank, colorRiver)
	case assets.RoadAndBridgeProfile:
		innerColor := colorBridge
		if bridgeDestroyed {
			innerColor = colorRiver
		}
		tb.renderBandedLines(bufY, domain.NumLinesPerTerrainProfile, colorRoad, innerColor)
	}
}

// renderIslandFragment renders all 16 scanlines of an island fragment.
func (tb *TerrainBuffer) renderIslandFragment(bufY int, island assets.IslandDefinition, bankColor color.RGBA) {
	profile, ok := assets.TerrainProfiles[island.ProfileIndex].(assets.RegularProfile)
	if !ok {
		return
	}

	height := len(tb.edges)

	for line := range domain.NumLinesPerTerrainProfile {
		// Bottom-to-top rendering: line 0 at bottom (bufY+15), line 15 at top (bufY)
		y := bufY + (domain.NumLinesPerTerrainProfile - 1 - line)

		coordinateLeft := int(profile.Values[line]) + island.WidthOffset
		rX := islandCenterOffset + coordinateLeft
		lX := calculateOtherEdge(islandDefaultHalf, coordinateLeft, assets.EdgeMirrored)

		// Update edge data to include island information.
		wrappedY := ((y % height) + height) % height
		tb.edges[wrappedY].HasIsland = true
		tb.edges[wrappedY].IslandLeftX = lX
		tb.edges[wrappedY].IslandRightX = rX

		fillRect(tb.buffer, lX, y, rX-lX, bankColor)
	}
}

// calculateOtherEdge computes the other bank edge X from the center/width parameter,
// the given edge X, and the edge mode.
func calculateOtherEdge(param, edgeX int, mode assets.EdgeMode) int {
	switch mode {
	case assets.EdgeMirrored:
		const mirroredEdgeMultiplier = 2

		return mirroredEdgeMultiplier*param - edgeX
	case assets.EdgeOffset:
		return param + edgeX
	default:
		panic(fmt.Sprintf("calculateOtherEdge: unsupported edge mode %d", mode))
	}
}

// renderBandedLines renders scanlines with outerColor on both sides and innerColor
// in the inner band [bridgeStartX, bridgeEndX). Used for canals and road/bridge sections.
// Edge data is set to full screen width (no bank collision) for all rendered rows.
func (tb *TerrainBuffer) renderBandedLines(bufY, lines int, outerColor, innerColor platform.Color) {
	height := len(tb.edges)

	outerInk := palette[outerColor]
	innerInk := palette[innerColor]

	for line := range lines {
		y := bufY + (lines - 1 - line)
		for x := range platform.ScreenWidth {
			if x >= bridgeStartX && x < bridgeEndX {
				tb.buffer.Set(x, y, innerInk)
			} else {
				tb.buffer.Set(x, y, outerInk)
			}
		}
		wrappedY := ((y % height) + height) % height
		tb.edges[wrappedY] = TerrainEdges{LeftX: 0, RightX: platform.ScreenWidth}
	}
}

// fillRect fills a horizontal strip of pixels with the given color.
func fillRect(buf PixelBuffer, x, y, w int, c color.RGBA) {
	for px := range w {
		buf.Set(x+px, y, c)
	}
}

// LevelRenderPosition specifies where to start rendering a level.
type LevelRenderPosition struct {
	LevelIndex    int // which level to render (0-47)
	StartFragment int // which fragment to start from (0-63)
	NumFragments  int // how many fragments to render
}

// DrawLevel renders terrain fragments into a buffer, iterating bottom-to-top
// (matching the game's progression direction). The buffer is populated starting
// at Y=0 (bottom) and progressing upward.
//
// This is the high-level API for static level rendering. It handles all iteration
// and coordinate calculation internally, ensuring the output matches game behavior.
func DrawLevel(buf PixelBuffer, pos LevelRenderPosition) {
	levelFragments := assets.LevelTerrain[pos.LevelIndex]

	// Create a temporary TerrainBuffer to track island state across fragments.
	// We don't need the circular buffer, just the island state tracking.
	tb := &TerrainBuffer{
		buffer: buf,
	}

	// Render fragments bottom-to-top (as the game progresses).
	// Fragment 0 renders at the bottom (highest Y), subsequent fragments above it.
	for i := range pos.NumFragments {
		fragIdx := (pos.StartFragment + i) % domain.NumFragmentsPerLevel
		frag := levelFragments[fragIdx] //nolint:gosec // G602: fragIdx is bounded by modulo NumFragmentsPerLevel

		// Calculate Y position for bottom-to-top rendering.
		// RenderFragment expects bufY to be the TOP of the fragment (where line 15 renders).
		// Last fragment (i = NumFragments-1) at Y=0 (top of image).
		// First fragment (i = 0) at Y=(NumFragments-1)*16 (bottom of image).
		// Since RenderFragment renders bottom-to-top (line 0 at bufY+15, line 15 at bufY),
		// we pass the top Y coordinate of each fragment.
		fragmentY := (pos.NumFragments - 1 - i) * domain.NumLinesPerTerrainProfile

		// bridgeDestroyed=false (no gameplay state in static rendering).
		tb.RenderFragment(frag, fragmentY, false)
	}
}

// drawTerrainBuffer draws the visible portion of the terrain buffer to the screen.
// scrollY is the buffer Y coordinate of game row 0 (top of the logical viewport).
// The blank zone (game rows 0–7) is skipped: only game rows [ViewportBlankZone, TotalViewportHeight)
// are drawn, mapping to screen rows [0, VisibleViewportHeight).
func drawTerrainBuffer(screen Screen, tb *TerrainBuffer, scrollY int) {
	img := tb.image.Image()
	height := img.Bounds().Dy()

	// Skip the blank zone: start reading from game row ViewportBlankZone.
	start := scrollY + domain.ViewportBlankZone
	wrappedStart := ((start % height) + height) % height
	drawHeight := domain.VisibleViewportHeight

	// Check if the draw region spans the wrap boundary.
	if wrappedStart+drawHeight > height {
		// Draw in two parts: bottom of buffer, then top of buffer.
		bottomHeight := height - wrappedStart
		screen.DrawImageRegion(img, image.Rect(0, wrappedStart, platform.ScreenWidth, height), 0, 0)
		screen.DrawImageRegion(img, image.Rect(0, 0, platform.ScreenWidth, drawHeight-bottomHeight), 0, bottomHeight)
	} else {
		screen.DrawImageRegion(img, image.Rect(0, wrappedStart, platform.ScreenWidth, wrappedStart+drawHeight), 0, 0)
	}
}
