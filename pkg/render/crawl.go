package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// CrawlPixels stores the raw pixel bits for the bottom-row text crawl.
// It covers GlyphSize scanlines × (ScreenWidth + GlyphSize) pixels: one extra character
// column beyond the visible area so new characters enter smoothly from the right.
type CrawlPixels = [assets.GlyphSize][platform.ScreenWidth + assets.GlyphSize]bool

// crawlRow is the screen character row used for the scrolling text crawl.
const crawlRow = 22

// DrawCrawl renders the text crawl pixel buffer at row 23 (y=184–191).
// Set pixels are drawn in white.
func DrawCrawl(screen draw.Image, pixels *CrawlPixels) {
	white := palette[platform.ColorWhite]
	crawlY := crawlRow * assets.GlyphSize

	for row := range assets.GlyphSize {
		for x := range platform.ScreenWidth {
			if pixels[row][x] {
				screen.Set(x, crawlY+row, white)
			}
		}
	}
}
