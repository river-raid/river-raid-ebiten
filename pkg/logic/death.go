package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// triggerDeath initiates the player death sequence.
// It stops the plane, clears active projectiles, spawns two explosion fragments,
// and transitions to GameplayDying.
func triggerDeath(s *state.GameState) {
	// Stop scrolling.
	s.Speed = 0

	// Clear active projectiles.
	s.Missile.Active = false
	s.TankShell.IsFlying = false
	s.TankShell.IsExploding = false

	// Spawn two explosion fragments stacked vertically, centred on the plane sprite.
	// The pair produces a 16×16 explosion over the 8×8 plane.
	fragX := s.PlaneX + (assets.SpritePlayerWidth-assets.SpriteExplosionWidth)/2
	frag1Y := domain.PlaneY + (assets.SpritePlayerHeight-assets.SpriteExplosionHeight*2)/2
	frag2Y := frag1Y + assets.SpriteExplosionHeight
	frag1 := state.ExplosionFragment{X: fragX, Y: frag1Y}
	frag2 := state.ExplosionFragment{X: fragX, Y: frag2Y}
	s.Explosion.Fragments = append(s.Explosion.Fragments, frag1, frag2)
	s.Controls.Exploding = true

	// Enter dying mode.
	s.GameplayMode = domain.GameplayDying
	s.DyingFrame = domain.DyingFrameCount
}

// updateDying advances the dying animation one frame.
// When DyingFrame reaches zero the post-death logic runs.
func updateDying(s *state.GameState, terrain TerrainRenderer) {
	s.Explosion = animateExplosion(s.Explosion)

	s.DyingFrame--
	if s.DyingFrame > 0 {
		return
	}

	// Animation complete: clear control flags then process post-death.
	s.Controls = state.ControlFlags{}
	handlePostDeath(s, terrain)
}

// handlePostDeath determines the next state after dying.
// Lives are read before the scroll-in decrement: a value of 0
// means the player has no sessions left; > 0 means one more session remains and the
// upcoming scroll-in will consume it.
func handlePostDeath(s *state.GameState, terrain TerrainRenderer) {
	if s.Config.IsTwoPlayer {
		other := domain.Player2
		if s.CurrentPlayer == domain.Player2 {
			other = domain.Player1
		}
		if s.Players[other].Lives > 0 {
			s.CurrentPlayer = other
		}
	}

	if s.Players[s.CurrentPlayer].Lives > 0 {
		ResetPerLife(s, terrain)
		s.GameplayMode = domain.GameplayScrollIn
	} else {
		triggerGameOver(s)
	}
}

// triggerGameOver updates the high score and transitions to the game over screen.
// In two-player mode the higher of both players' scores is used.
func triggerGameOver(s *state.GameState) {
	slot := domain.HighScoreSlot(s.Config.StartingBridge)
	score := s.Players[s.CurrentPlayer].Score
	if s.Config.IsTwoPlayer {
		other := domain.Player1
		if s.CurrentPlayer == domain.Player1 {
			other = domain.Player2
		}
		if s.Players[other].Score > score {
			score = s.Players[other].Score
		}
	}
	updateHighScore(&s.HighScores, slot, score)
	s.Screen = domain.ScreenGameOver
}

// ResetPerLife resets all per-life state in preparation for a new scroll-in.
// Per-player state (score, bridge index/counter) is preserved; lives are decremented
// by updateScrollIn at the end of the scroll-in, not here.
// The terrain buffer is cleared to black so scroll-in starts from a blank screen.
// Called both on the initial game start (from game.NewGame) and on every respawn
// (from handlePostDeath) so there is a single code path for all life starts.
func ResetPerLife(s *state.GameState, terrain TerrainRenderer) {
	s.Fuel = fuelRefuelCap
	s.PlaneX = domain.PlaneStartX
	s.PlaneSpriteBank = 0
	s.Speed = domain.SpeedNormal

	// Reset viewport to empty.
	s.Viewport = state.NewViewport()

	// Clear explosion fragments.
	s.Explosion = state.Explosion{}

	// Clear active projectiles.
	s.Missile = &state.PlayerMissile{}
	s.TankShell = &state.TankShell{}
	s.HeliMissile = &state.HeliMissile{}

	// Clear control flags.
	s.Controls = state.ControlFlags{}

	// Reset terrain scroll state.
	s.ScrollY = domain.NumLinesPerTerrainProfile
	s.ScrollOffset = domain.NumLinesPerTerrainProfile
	s.FragmentNum = 1
	s.LineInFrag = 0
	s.NextRenderY = 0
	s.ScrollInCount = 0
	s.ScrollInState = 0

	// Align SpawnIndex to the initial ScrollOffset so the first scroll step does not
	// spuriously spawn a mid-sequence object. SpawnIndex must equal the spawn index that
	// advanceLines will compute for the current ScrollOffset; otherwise the inequality
	// check in spawnFromScroll fires immediately and spawns an out-of-context object.
	s.Viewport.SpawnIndex = (int(s.ScrollOffset) / domain.NumLinesPerSpawnSlot) % domain.NumSpawnSlotsPerLevel

	// Pre-set bridge destroyed flag so the first bridge renders with the destruction
	// gap during scroll-in. It is cleared at the end of scroll-in (updateScrollIn).
	s.BridgeDestroyed = true

	// Clear bridge section tracking so stale bridge collision windows do not persist.
	s.BridgeSection = false
	s.BridgeYPosition = 0
	s.BridgeFragBufY = 0
	s.BridgeFragment = assets.TerrainFragment{}

	// Clear the terrain buffer to black so the scroll-in begins from a blank screen.
	// Without this, stale pixels from the previous life remain visible until overwritten.
	terrain.Clear()
}
