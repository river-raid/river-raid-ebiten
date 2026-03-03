package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// DrawGameplay renders the current gameplay frame to the screen.
func DrawGameplay(screen *ebiten.Image, s *state.GameState, terrain *TerrainBuffer) {
	screenWrapper := newEbitenScreen(screen)
	drawTerrainBuffer(screenWrapper, terrain, s.ScrollY)

	// Draw viewport objects.
	drawViewportSlots(screen, s.Viewport)

	// Draw projectiles.
	drawPlayerMissile(screen, s.Missile)
	drawTankShell(screen, s.TankShell)
	drawHeliMissile(screen, s.HeliMissile)

	// Draw player.
	drawPlayer(screen, s.CurrentPlayer, s.PlaneX, s.PlaneSpriteBank)
}
