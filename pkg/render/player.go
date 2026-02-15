package render

import (
	"image/draw"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// DrawPlayer renders the player's plane.
func DrawPlayer(screen draw.Image, player domain.Player, x int, isBanked bool) {
	// Draw player plane.
	var s assets.Sprite
	if isBanked {
		s = assets.SpritePlayerBanked
	} else {
		s = assets.SpritePlayerLevel
	}

	color := playerColors[player]

	drawSprite(screen, s, x, domain.PlaneY, color, false)
}
