package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// DrawGameplay renders the current gameplay frame to the screen.
func DrawGameplay(screen *ebiten.Image, s *state.GameState, terrain *TerrainBuffer) {
	scrollYPx := s.ScrollY.ToPx()

	screenWrapper := newEbitenScreen(screen)
	drawTerrainBuffer(screenWrapper, terrain, scrollYPx)

	vc := newViewportCanvas(screen)

	// Draw viewport objects.
	drawViewportSlots(vc, s.Viewport, int(s.Tick), s.GameplayMode, terrain, scrollYPx)

	// Draw projectiles.
	drawPlayerMissile(vc, s.Missile)
	drawTankShell(vc, s.TankShell)
	drawHeliMissile(vc, s.HeliMissile)

	// Draw explosion fragments.
	drawExplosionFragments(vc, s.Explosion)

	// Draw player only during normal gameplay and refueling.
	if s.GameplayMode == domain.GameplayNormal || s.GameplayMode == domain.GameplayRefuel {
		drawPlayer(vc, s.CurrentPlayer, s.PlaneX.ToPx(), s.PlaneSpriteBank)
	}

	// Draw HUD (scores, lives, fuel gauge, bridge count).
	DrawHUD(screen, s)
}
