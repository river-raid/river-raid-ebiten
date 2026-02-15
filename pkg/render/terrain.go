package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

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
	RenderIdx   int // current scanline counter (0–23)
	ProfileIdx  int // which profile shape to use
	LineIdx     int // current line index into the profile (wraps at 16)
	WidthOffset int
	EdgeMode    assets.EdgeMode
}

// TerrainBuffer manages an off-screen image for incremental terrain rendering.
type TerrainBuffer struct {
	image  *ebiten.Image
	Island IslandState
}

// NewTerrainBuffer creates a terrain buffer tall enough for the given height.
func NewTerrainBuffer(height int) *TerrainBuffer {
	return &TerrainBuffer{
		image: ebiten.NewImage(platform.ScreenWidth, height),
	}
}

// renderRegularLine renders a single scanline of a regular terrain profile.
// leftX is the left bank edge X, rightX is the right bank edge X.
// y is the destination Y in the buffer.
func (tb *TerrainBuffer) renderRegularLine(y, leftX, rightX int, bankColor, riverColor color.RGBA) {
	// Fill left bank (green) from x=0 to left edge.
	if leftX > 0 {
		fillRect(tb.image, 0, y, leftX, bankColor)
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
		fillRect(tb.image, riverStart, y, riverEnd-riverStart, riverColor)
	}

	// Fill right bank (green) from right edge to screen boundary.
	if rightX < platform.ScreenWidth {
		fillRect(tb.image, rightX, y, platform.ScreenWidth-rightX, bankColor)
	}
}

// bridgeGapStart and bridgeGapEnd define the byte range cleared when a bridge
// is destroyed (4 bytes in the center of the 32-byte scanline pattern).
const (
	bridgeGapStart = 14
	bridgeGapEnd   = 18
)

// RenderFragment renders a single terrain fragment (16 scanlines) into the buffer.
// bufY is the starting Y position in the buffer.
// bridgeDestroyed controls whether the bridge gap is rendered for road/bridge profiles.
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
			coordinateLeft := int(p.Values[line]) + frag.Byte3
			leftX := coordinateLeft - edgeOffsetAdjust
			rightX := calculateRightEdge(coordinateLeft, frag.Byte2, frag.EdgeMode)
			tb.renderRegularLine(bufY+line, leftX, rightX, bankColor, riverColor)
			tb.renderIslandLine(bufY+line, bankColor)
		}
	case assets.CanalProfile:
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
	leftX := tb.Island.WidthOffset + int(profile.Values[lineIdx]) + islandCenterOffset
	rightX := calculateIslandRightEdge(leftX, tb.Island.EdgeMode)

	// Island draws green (bank) between leftX and rightX.
	if leftX < 0 {
		leftX = 0
	}
	if rightX > platform.ScreenWidth {
		rightX = platform.ScreenWidth
	}
	if rightX > leftX {
		fillRect(tb.image, leftX, y, rightX-leftX, bankColor)
	}

	tb.Island.LineIdx++
	tb.Island.RenderIdx++

	if tb.Island.RenderIdx >= domain.NumLinesPerTerrainProfile {
		tb.Island.Active = false
	}
}

// calculateIslandRightEdge computes the right edge of an island.
func calculateIslandRightEdge(leftX int, mode assets.EdgeMode) int {
	switch mode {
	case assets.EdgeMirrored:
		return 2*islandCenterOffset - leftX //nolint:mnd // formula: rightX = 2*center - leftX
	case assets.EdgeOffset:
		return islandDefaultHalf + leftX
	default:
		panic(fmt.Sprintf("calculateRightEdge: unsupported edge mode %d", mode))
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

	for line := range lines {
		y := bufY + line
		for byteIdx := range bridgeRoadBytes {
			attr := attrPattern[byteIdx]
			paper := palette[(attr>>3)&0x07] //nolint:mnd // ZX attribute: bits 5-3 = paper color
			ink := palette[attr&0x07]        //nolint:mnd // ZX attribute: bits 2-0 = ink color

			px := pixelPattern[byteIdx]
			baseX := byteIdx * bitsPerByte

			for bit := range bitsPerByte {
				if px&(1<<(7-bit)) != 0 { //nolint:mnd // MSB first
					tb.image.Set(baseX+bit, y, ink)
				} else {
					tb.image.Set(baseX+bit, y, paper)
				}
			}
		}
	}
}

// fillRect fills a horizontal strip of pixels with the given color.
func fillRect(img *ebiten.Image, x, y, w int, c color.RGBA) {
	for px := range w {
		img.Set(x+px, y, c)
	}
}

// DrawTerrainBuffer draws the visible portion of the terrain buffer to the screen.
// scrollY is the buffer Y coordinate of the top of the visible viewport.
// As scrollY decreases, the viewport moves up in the buffer, revealing newer terrain
// (rendered at lower Y) at the top of the screen — terrain scrolls downward.
func DrawTerrainBuffer(screen *ebiten.Image, tb *TerrainBuffer, scrollY int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(-scrollY))
	screen.DrawImage(tb.image, op)
}
