package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Terrain rendering constants.
const (
	fragmentLines    = 16 // scanlines per terrain fragment
	edgeOffsetAdjust = 16 // subtracted from left edge for edge sprite width (two tiles)
	bridgeRoadBytes  = 32 // bytes per full-width scanline pattern
	bitsPerByte      = 8
)

// TerrainBuffer manages an off-screen image for incremental terrain rendering.
type TerrainBuffer struct {
	image *ebiten.Image
}

// newTerrainBuffer creates a terrain buffer tall enough for the given height.
func newTerrainBuffer(height int) *TerrainBuffer {
	return &TerrainBuffer{
		image: ebiten.NewImage(ScreenWidth, height),
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

	if riverEnd > ScreenWidth {
		riverEnd = ScreenWidth
	}

	if riverEnd > riverStart {
		fillRect(tb.image, riverStart, y, riverEnd-riverStart, riverColor)
	}

	// Fill right bank (green) from right edge to screen boundary.
	if rightX < ScreenWidth {
		fillRect(tb.image, rightX, y, ScreenWidth-rightX, bankColor)
	}
}

// bridgeGapStart and bridgeGapEnd define the byte range cleared when a bridge
// is destroyed (4 bytes in the center of the 32-byte scanline pattern).
const (
	bridgeGapStart = 14
	bridgeGapEnd   = 18
)

// renderFragment renders a single terrain fragment (16 scanlines) into the buffer.
// bufY is the starting Y position in the buffer.
// bridgeDestroyed controls whether the bridge gap is rendered for road/bridge profiles.
func (tb *TerrainBuffer) renderFragment(frag TerrainFragment, bufY int, bridgeDestroyed bool) {
	profile := TerrainProfiles[frag.ProfileIndex]

	bankColor := Palette[ColorBank]
	riverColor := Palette[ColorRiver]

	switch p := profile.(type) {
	case RegularProfile:
		for line := range fragmentLines {
			leftX := int(p.Values[line]) + frag.Byte3 - edgeOffsetAdjust
			rightX := calculateRightEdge(leftX, frag.Byte2, frag.EdgeMode)
			tb.renderRegularLine(bufY+line, leftX, rightX, bankColor, riverColor)
		}
	case CanalProfile:
		tb.renderBridgeRoadLine(bufY, fragmentLines, BridgeRoadData[:bridgeRoadBytes])
	case RoadAndBridgeProfile:
		pattern := BridgeRoadData[bridgeRoadBytes : 2*bridgeRoadBytes]
		if bridgeDestroyed {
			// Copy pattern and clear the bridge bytes to create the destruction gap.
			var destroyed [bridgeRoadBytes]byte
			copy(destroyed[:], pattern)
			for i := bridgeGapStart; i < bridgeGapEnd; i++ {
				destroyed[i] = 0
			}
			pattern = destroyed[:]
		}
		tb.renderBridgeRoadLine(bufY, fragmentLines, pattern)
	}
}

// calculateRightEdge computes the right bank edge X from the left edge,
// the center/width parameter, and the edge mode.
func calculateRightEdge(leftX, param int, mode EdgeMode) int {
	switch mode {
	case EdgeMirrored:
		return 2*param - leftX //nolint:mnd // formula: rightX = 2*center - leftX
	case EdgeOffset:
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
	attrPattern := BridgeRoadData[2*bridgeRoadBytes:]

	for line := range lines {
		y := bufY + line
		for byteIdx := range bridgeRoadBytes {
			attr := attrPattern[byteIdx]
			paper := Palette[(attr>>3)&0x07] //nolint:mnd // ZX attribute: bits 5-3 = paper color
			ink := Palette[attr&0x07]        //nolint:mnd // ZX attribute: bits 2-0 = ink color

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

// drawTerrainBuffer draws the visible portion of the terrain buffer to the screen.
func drawTerrainBuffer(screen *ebiten.Image, tb *TerrainBuffer, scrollOffset int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(-scrollOffset))
	screen.DrawImage(tb.image, op)
}
