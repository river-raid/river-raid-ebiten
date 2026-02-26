package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// DrawGameplay renders the current gameplay frame to the screen.
func DrawGameplay(screen *ebiten.Image, s *state.GameState, terrain *TerrainBuffer) {
	screenWrapper := newEbitenScreen(screen)
	drawTerrainBuffer(screenWrapper, terrain, s.ScrollY)

	vc := newViewportCanvas(screen)

	// Draw viewport objects.
	drawViewportSlots(vc, s.Viewport, s.GameplayMode)

	// Draw projectiles.
	drawPlayerMissile(vc, s.Missile)
	drawTankShell(vc, s.TankShell)
	drawHeliMissile(vc, s.HeliMissile)

	// Draw player.
	drawPlayer(vc, s.CurrentPlayer, s.PlaneX, s.PlaneSpriteBank)
}
