package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// GameplayState provides all data needed to render a single frame of gameplay.
type GameplayState struct {
	Terrain       *TerrainBuffer
	Viewport      *state.Viewport
	Missile       *logic.PlayerMissile
	TankShell     *logic.TankShell
	HeliMissile   *logic.HeliMissile
	PlaneX        int
	PlaneBanked   bool
	CurrentPlayer domain.Player
	ScrollY       int
}

// DrawGameplay renders the current gameplay frame to the screen.
func DrawGameplay(screen *ebiten.Image, gs GameplayState) {
	screenWrapper := newEbitenScreen(screen)
	drawTerrainBuffer(screenWrapper, gs.Terrain, gs.ScrollY)

	// Draw viewport objects.
	drawViewportSlots(screen, gs.Viewport)

	// Draw projectiles.
	drawPlayerMissile(screen, gs.Missile)
	drawTankShell(screen, gs.TankShell)
	drawHeliMissile(screen, gs.HeliMissile)

	// Draw player.
	drawPlayer(screen, gs.CurrentPlayer, gs.PlaneX, gs.PlaneBanked)
}
