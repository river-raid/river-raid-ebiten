package render

import (
	"image"
	"image/draw"
)

// Screen represents a drawable surface that supports both pixel-level drawing
// and efficient image blitting operations.
type Screen interface {
	draw.Image
	// DrawImageRegion draws a rectangular region from src onto this screen at the given position.
	// srcRect defines the region to copy from src.
	// dstX, dstY define where to draw the region on this screen.
	DrawImageRegion(src image.Image, srcRect image.Rectangle, dstX, dstY int)
}
