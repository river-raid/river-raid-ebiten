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
	edgeOffsetAdjust   = 6  // subtracted from left edge for edge sprite width
	bridgeRoadBytes    = 32 // bytes per full-width scanline pattern
	bitsPerByte        = 8
	islandCenterOffset = 128 // added to island left edge to center on screen
	islandDefaultHalf  = 60  // default half-width for island right edge calculation
)

// IslandState tracks the rendering state of an active island.
type IslandState struct {
	Active      bool
	RenderIdx   int // current scanline counter (0–15)
	ProfileIdx  int // which profile shape to use
	LineIdx     int // current line index into the profile (wraps at 16)
	WidthOffset int
	EdgeMode    assets.EdgeMode
}

// TerrainBuffer manages an off-screen image for incremental terrain rendering.
type TerrainBuffer struct {
	buffer PixelBuffer
	image  *CircularImage // kept for DrawTerrainBuffer access
	Island IslandState
}

// NewTerrainBuffer creates a terrain buffer tall enough for the given height.
func NewTerrainBuffer(height int) *TerrainBuffer {
	circImg := NewCircularImage(platform.ScreenWidth, height)
	return &TerrainBuffer{
		buffer: circImg,
		image:  circImg,
	}
}

// renderRegularLine renders a single scanline of a regular terrain profile.
// leftX is the left bank edge X, rightX is the right bank edge X.
// y is the destination Y in the buffer.
func (tb *TerrainBuffer) renderRegularLine(y, leftX, rightX int, bankColor, riverColor color.RGBA) {
	// Fill left bank (green) from x=0 to left edge.
	if leftX > 0 {
		fillRect(tb.buffer, 0, y, leftX, bankColor)
	}

	// Fill river (blue) between banks.
	riverStart := leftX
	riverEnd := rightX

	if riverStart < 0 {
		riverStart = 0
	}

	if riverEnd > platform.ScreenWidth {
		riverEnd = platform.ScreenWidth
	}

	if riverEnd > riverStart {
		fillRect(tb.buffer, riverStart, y, riverEnd-riverStart, riverColor)
	}

	// Fill right bank (green) from right edge to screen boundary.
	if rightX < platform.ScreenWidth {
		fillRect(tb.buffer, rightX, y, platform.ScreenWidth-rightX, bankColor)
	}
}

// bridgeGapStart and bridgeGapEnd define the byte range cleared when a bridge
// is destroyed (4 bytes in the center of the 32-byte scanline pattern).
const (
	bridgeGapStart = 14
	bridgeGapEnd   = 18
)

// RenderFragment renders a single terrain fragment (16 scanlines) into the buffer.
// bufY is the starting Y position in the buffer (top of the fragment).
// bridgeDestroyed controls whether the bridge gap is rendered for road/bridge profiles.
// Scanlines are rendered bottom-to-top: line 0 at bufY+15, line 15 at bufY.
func (tb *TerrainBuffer) RenderFragment(frag assets.TerrainFragment, bufY int, bridgeDestroyed bool) {
	// Trigger a new island if the fragment references one.
	if frag.IslandNum > 0 && !tb.Island.Active {
		island := assets.Islands[frag.IslandNum-1]
		tb.Island = IslandState{
			Active:      true,
			ProfileIdx:  island.ProfileIndex,
			WidthOffset: island.WidthOffset,
			EdgeMode:    island.EdgeMode,
		}
	}

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
			rightX := calculateRightEdge(coordinateLeft, frag.Byte2, frag.EdgeMode)
			tb.renderRegularLine(y, leftX, rightX, bankColor, riverColor)
			tb.renderIslandLine(y, bankColor)
		}
	case assets.CanalProfile:
		// Render canal pattern (handles its own Y iteration)
		tb.renderBridgeRoadLine(bufY, domain.NumLinesPerTerrainProfile, assets.BridgeRoadData[:bridgeRoadBytes])
	case assets.RoadAndBridgeProfile:
		pattern := assets.BridgeRoadData[bridgeRoadBytes : 2*bridgeRoadBytes]
		if bridgeDestroyed {
			// Copy pattern and clear the bridge bytes to create the destruction gap.
			var destroyed [bridgeRoadBytes]byte
			copy(destroyed[:], pattern)
			for i := bridgeGapStart; i < bridgeGapEnd; i++ {
				destroyed[i] = 0
			}
			pattern = destroyed[:]
		}
		// Render road/bridge pattern (handles its own Y iteration)
		tb.renderBridgeRoadLine(bufY, domain.NumLinesPerTerrainProfile, pattern)
	}
}

// renderIslandLine renders one scanline of an active island, drawing green banks
// within the river to narrow it from both sides.
func (tb *TerrainBuffer) renderIslandLine(y int, bankColor color.RGBA) {
	if !tb.Island.Active {
		return
	}

	profile, ok := assets.TerrainProfiles[tb.Island.ProfileIdx].(assets.RegularProfile)
	if !ok {
		return
	}

	lineIdx := tb.Island.LineIdx % domain.NumLinesPerTerrainProfile
	profileValue := int(profile.Values[lineIdx])

	// Calculate island edges based on PHP reference implementation.
	// PHP: side1 = 0x80 + byte2 + value, side2 = calculateOtherSide(0x3C, byte2 + value)
	// Draws from side2 to side1 + 10.
	var leftX, rightX int
	switch tb.Island.EdgeMode {
	case assets.EdgeMirrored:
		// SYMMETRICAL mode: side2 = 2*0x3C - (byte2+value), side1 = 0x80 + byte2 + value
		side2 := 0x78 - tb.Island.WidthOffset - profileValue //nolint:mnd // 0x78 = 2*0x3C
		side1 := 0x80 + tb.Island.WidthOffset + profileValue //nolint:mnd // 0x80 = 128 center
		leftX = side2
		rightX = side1 + 10 //nolint:mnd // constant width addition from PHP
	case assets.EdgeOffset:
		// PARALLEL mode: side2 = 0x3C + (byte2+value), side1 = 0x80 + byte2 + value
		side2 := 0x3C + tb.Island.WidthOffset + profileValue //nolint:mnd // 0x3C = 60 base offset
		side1 := 0x80 + tb.Island.WidthOffset + profileValue //nolint:mnd // 0x80 = 128 center
		leftX = side2
		rightX = side1 + 10 //nolint:mnd // constant width addition from PHP
	default:
		panic(fmt.Sprintf("calculateRightEdge: unsupported edge mode %d", tb.Island.EdgeMode))
	}

	// Island draws green (bank) between leftX and rightX.
	if leftX < 0 {
		leftX = 0
	}
	if rightX > platform.ScreenWidth {
		rightX = platform.ScreenWidth
	}
	if rightX > leftX {
		fillRect(tb.buffer, leftX, y, rightX-leftX, bankColor)
	}

	tb.Island.LineIdx++
	tb.Island.RenderIdx++

	if tb.Island.RenderIdx >= domain.NumLinesPerTerrainProfile {
		tb.Island.Active = false
	}
}

// calculateRightEdge computes the right bank edge X from the left edge,
// the center/width parameter, and the edge mode.
func calculateRightEdge(leftX, param int, mode assets.EdgeMode) int {
	switch mode {
	case assets.EdgeMirrored:
		return 2*param - leftX //nolint:mnd // formula: rightX = 2*center - leftX
	case assets.EdgeOffset:
		return param + leftX
	default:
		panic(fmt.Sprintf("calculateRightEdge: unsupported edge mode %d", mode))
	}
}

// renderBridgeRoadLine renders a full-width scanline pattern for canal or road/bridge
// sections. The pattern is a 32-byte (256-pixel) 1bpp bitmap. Each pixel is colored
// using the corresponding attribute byte from the attribute pattern (bytes 64–95).
// Ink color is used for set bits, paper color for unset bits.
func (tb *TerrainBuffer) renderBridgeRoadLine(bufY, lines int, pixelPattern []byte) {
	attrPattern := assets.BridgeRoadData[2*bridgeRoadBytes:]
	BridgeRoadLines(tb.buffer, bufY, lines, pixelPattern, attrPattern)
}

// fillRect fills a horizontal strip of pixels with the given color.
func fillRect(buf PixelBuffer, x, y, w int, c color.RGBA) {
	for px := range w {
		buf.Set(x+px, y, c)
	}
}

// BridgeRoadLines renders full-width scanline patterns for canal or road/bridge sections.
// bufY is the starting Y position (top of the fragment), lines is the number of scanlines to render.
// pixelPattern is the 32-byte (256-pixel) 1bpp bitmap pattern.
// attrPattern is the 32-byte attribute pattern for coloring.
// Renders bottom-to-top: line 0 at bufY+15, line 15 at bufY (consistent with RenderFragment).
func BridgeRoadLines(buf PixelBuffer, bufY, lines int, pixelPattern, attrPattern []byte) {
	for line := range lines {
		// Bottom-to-top rendering: line 0 at bottom (bufY+15), line 15 at top (bufY)
		y := bufY + (lines - 1 - line)
		for byteIdx := range bridgeRoadBytes {
			attr := attrPattern[byteIdx]
			paper := palette[(attr>>3)&0x07] //nolint:mnd // ZX attribute: bits 5-3 = paper color
			ink := palette[attr&0x07]        //nolint:mnd // ZX attribute: bits 2-0 = ink color

			px := pixelPattern[byteIdx]
			baseX := byteIdx * bitsPerByte

			for bit := range bitsPerByte {
				if px&(1<<(7-bit)) != 0 { //nolint:mnd // MSB first
					buf.Set(baseX+bit, y, ink)
				} else {
					buf.Set(baseX+bit, y, paper)
				}
			}
		}
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

// DrawTerrainBuffer draws the visible portion of the terrain buffer to the screen.
// scrollY is the buffer Y coordinate of the top of the visible viewport.
// As scrollY decreases, the viewport moves up in the buffer, revealing newer terrain
// (rendered at lower Y) at the top of the screen — terrain scrolls downward.
func DrawTerrainBuffer(screen Screen, tb *TerrainBuffer, scrollY int) {
	img := tb.image.Image()
	height := img.Bounds().Dy()

	// Wrap scrollY to buffer bounds (circular buffer).
	wrappedScrollY := ((scrollY % height) + height) % height

	viewportHeight := domain.ViewportHeight

	// Check if viewport spans the wrap boundary.
	if wrappedScrollY+viewportHeight > height {
		// Draw in two parts: bottom of buffer, then top of buffer.
		bottomHeight := height - wrappedScrollY

		// Draw bottom part (from wrappedScrollY to end of buffer).
		bottomRect := image.Rect(0, wrappedScrollY, platform.ScreenWidth, height)
		screen.DrawImageRegion(img, bottomRect, 0, 0)

		// Draw top part (from 0 to remaining viewport height).
		topHeight := viewportHeight - bottomHeight
		topRect := image.Rect(0, 0, platform.ScreenWidth, topHeight)
		screen.DrawImageRegion(img, topRect, 0, bottomHeight)
	} else {
		// Normal case: viewport doesn't wrap, clip to viewport height.
		viewportRect := image.Rect(0, wrappedScrollY, platform.ScreenWidth, wrappedScrollY+viewportHeight)
		screen.DrawImageRegion(img, viewportRect, 0, 0)
	}
}
