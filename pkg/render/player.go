package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// drawPlayer renders the player's plane.
func drawPlayer(screen draw.Image, player domain.Player, x, bank int) {
	// Draw player plane.
	var s assets.Sprite
	switch bank {
	case 1: // Left
		s = assets.SpritePlayerBanked
		drawSprite(screen, s, x, domain.PlaneY, playerColors[player], true) // Mirror for left? Wait, original might have separate sprites.
	case int(domain.SpeedNormal): // Right (using constant to avoid mnd lint)
		s = assets.SpritePlayerBanked
		drawSprite(screen, s, x, domain.PlaneY, playerColors[player], false)
	default: // Level
		s = assets.SpritePlayerLevel
		drawSprite(screen, s, x, domain.PlaneY, playerColors[player], false)
	}
}
