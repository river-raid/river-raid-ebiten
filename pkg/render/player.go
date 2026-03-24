package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

const (
	planeBankLeft  = 1
	planeBankRight = 2
)

// drawPlayer renders the player's plane.
func drawPlayer(screen draw.Image, player domain.Player, x domain.Px, bank int) {
	// Draw player plane.
	var s assets.Sprite
	switch bank {
	case planeBankLeft:
		s = assets.SpritePlayerBanked
		drawSprite(screen, s, x, domain.PlaneY, staticColorFn(playerColors[player]), true) // Mirror for left? Wait, original might have separate sprites.
	case planeBankRight:
		s = assets.SpritePlayerBanked
		drawSprite(screen, s, x, domain.PlaneY, staticColorFn(playerColors[player]), false)
	default: // Level
		s = assets.SpritePlayerLevel
		drawSprite(screen, s, x, domain.PlaneY, staticColorFn(playerColors[player]), false)
	}
}
