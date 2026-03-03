package logic

import (
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

	// Align plane X to an 8-pixel boundary.
	alignedX := s.PlaneX & domain.PlaneXAlignMask

	// Spawn two explosion fragments at (alignedX, DeathFragmentY) and (alignedX, DeathFragmentY+DeathFragmentSpacing).
	frag1 := state.ExplodingFragment{X: alignedX, Y: domain.DeathFragmentY, Frame: 1}
	frag2 := state.ExplodingFragment{X: alignedX, Y: domain.DeathFragmentY + domain.DeathFragmentSpacing, Frame: 1}
	s.ExplodingFragments = append(s.ExplodingFragments, frag1, frag2)
	s.Controls.Exploding = true

	// Enter dying mode.
	s.GameplayMode = domain.GameplayDying
	s.DyingFrame = domain.DyingFrameCount
}

// updateDying advances the dying animation one frame.
// When DyingFrame reaches zero the post-death logic runs.
func updateDying(s *state.GameState) {
	// Advance explosion fragment animation.
	s.ExplodingFragments = animateExplosionFragments(s.ExplodingFragments)

	s.DyingFrame--
	if s.DyingFrame > 0 {
		return
	}

	// Animation complete: clear control flags then process post-death.
	s.Controls = state.ControlFlags{}
	handlePostDeath(s)
}

// handlePostDeath decrements lives and determines the next state after dying.
func handlePostDeath(s *state.GameState) {
	s.Players[s.CurrentPlayer].Lives--

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
		resetPerLife(s)
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

// resetPerLife resets all per-life state in preparation for a new scroll-in.
// Per-player state (score, lives, bridge index/counter) is preserved.
func resetPerLife(s *state.GameState) {
	s.Fuel = domain.FuelLevelFull
	s.PlaneX = domain.PlaneStartX
	s.Speed = domain.SpeedNormal

	// Reset viewport to empty.
	s.Viewport = state.NewViewport()

	// Clear explosion fragments.
	s.ExplodingFragments = nil

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

	// Clear bridge destroyed flag.
	s.BridgeDestroyed = false
}
