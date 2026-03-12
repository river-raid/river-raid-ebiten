package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
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

	// Draw explosion fragments.
	drawExplosionFragments(vc, s.Explosion)

	// Draw player — suppressed during dying and overview modes.
	if s.GameplayMode != domain.GameplayDying && s.GameplayMode != domain.GameplayOverview {
		drawPlayer(vc, s.CurrentPlayer, s.PlaneX, s.PlaneSpriteBank)
	}

	// Draw HUD (scores, lives, fuel gauge, bridge count).
	DrawHUD(screen, s)
}
