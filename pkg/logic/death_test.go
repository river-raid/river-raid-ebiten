package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// noopTerrain is a no-op TerrainRenderer for death tests that do not exercise rendering.
// mockTerrainBuffer (defined in enemy_ai_test.go) also satisfies TerrainRenderer and is
// used when edge queries matter; noopTerrain is sufficient for state-transition tests.
var noopTerrain = newMockTerrainBuffer()

// newDeathTestState returns a minimal GameState suitable for death tests.
func newDeathTestState() *state.GameState {
	s := state.NewGameState()
	ResetPerLife(s, noopTerrain)
	s.GameplayMode = domain.GameplayNormal
	s.Config.StartingBridge = domain.StartingBridge01
	return s
}

// TestTriggerDeath_SetsMode checks GameplayDying is set.
func TestTriggerDeath_SetsMode(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	triggerDeath(s)

	if s.GameplayMode != domain.GameplayDying {
		t.Errorf("GameplayMode = %v, want GameplayDying", s.GameplayMode)
	}
}

// TestTriggerDeath_SetsDyingFrame checks DyingFrame is initialised to DyingFrameCount.
func TestTriggerDeath_SetsDyingFrame(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	triggerDeath(s)

	if s.DyingFrame != domain.DyingFrameCount {
		t.Errorf("DyingFrame = %d, want %d", s.DyingFrame, domain.DyingFrameCount)
	}
}

// TestTriggerDeath_StopsPlane checks speed is set to 0.
func TestTriggerDeath_StopsPlane(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Speed = domain.SpeedFast
	triggerDeath(s)

	if s.Speed != 0 {
		t.Errorf("Speed = %d, want 0", s.Speed)
	}
}

// TestTriggerDeath_ClearsMissile checks the missile is deactivated.
func TestTriggerDeath_ClearsMissile(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Missile.Active = true
	triggerDeath(s)

	if s.Missile.Active {
		t.Error("Missile.Active should be false after death trigger")
	}
}

// TestTriggerDeath_SpawnsFragments checks two fragments are spawned centered on the plane.
func TestTriggerDeath_SpawnsFragments(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneX = 120
	triggerDeath(s)

	if len(s.Explosion.Fragments) != 2 {
		t.Fatalf("len(Explosion.Fragments) = %d, want 2", len(s.Explosion.Fragments))
	}

	f1 := s.Explosion.Fragments[0]
	f2 := s.Explosion.Fragments[1]

	wantX := s.PlaneX + (assets.SpritePlayerWidth-assets.SpriteExplosionWidth)/2
	wantY := domain.PlaneY + (assets.SpritePlayerHeight-assets.SpriteExplosionHeight*2)/2
	if f1.X != wantX {
		t.Errorf("fragment[0].X = %d, want %d", f1.X, wantX)
	}
	if f1.Y != wantY {
		t.Errorf("fragment[0].Y = %d, want %d", f1.Y, wantY)
	}
	if f2.X != wantX {
		t.Errorf("fragment[1].X = %d, want %d", f2.X, wantX)
	}
	if f2.Y != wantY+assets.SpriteExplosionHeight {
		t.Errorf("fragment[1].Y = %d, want %d", f2.Y, wantY+assets.SpriteExplosionHeight)
	}
}

// TestUpdateDying_DecrementsFrame checks DyingFrame decrements each call.
func TestUpdateDying_DecrementsFrame(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	triggerDeath(s)
	initial := s.DyingFrame

	updateDying(s, noopTerrain)

	if s.DyingFrame != initial-1 {
		t.Errorf("DyingFrame = %d, want %d", s.DyingFrame, initial-1)
	}

	if s.GameplayMode != domain.GameplayDying {
		t.Error("GameplayMode should still be GameplayDying before frame reaches 0")
	}
}

// TestUpdateDying_TransitionsAfterFrameZero checks transition to scroll-in when DyingFrame hits 0.
func TestUpdateDying_TransitionsAfterFrameZero(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	triggerDeath(s)
	// Give the player lives so it restarts (not game over); pre-decrement value.
	s.Players[domain.Player1].Lives = 1

	// Drive DyingFrame to zero.
	for s.DyingFrame > 0 {
		updateDying(s, noopTerrain)
	}

	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
}

// TestHandlePostDeath_SinglePlayerRestart checks restart when lives remain (pre-decrement > 0).
func TestHandlePostDeath_SinglePlayerRestart(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Lives = 1
	handlePostDeath(s, noopTerrain)

	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
}

// TestHandlePostDeath_SinglePlayerGameOver checks game over when no lives remain (pre-decrement = 0).
func TestHandlePostDeath_SinglePlayerGameOver(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Lives = 0
	handlePostDeath(s, noopTerrain)

	if s.Screen != domain.ScreenGameOver {
		t.Errorf("Screen = %v, want ScreenGameOver", s.Screen)
	}
}

// TestHandlePostDeath_TwoPlayer_SwitchesToOtherPlayer checks P1→P2 switch.
func TestHandlePostDeath_TwoPlayer_SwitchesToOtherPlayer(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player1
	s.Players[domain.Player1].Lives = 2
	s.Players[domain.Player2].Lives = 2
	handlePostDeath(s, noopTerrain)

	if s.CurrentPlayer != domain.Player2 {
		t.Errorf("CurrentPlayer = %v, want Player2", s.CurrentPlayer)
	}
	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
}

// TestHandlePostDeath_TwoPlayer_RestartsCurrentIfOtherDead checks P1 restart when P2 has no lives.
func TestHandlePostDeath_TwoPlayer_RestartsCurrentIfOtherDead(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player1
	s.Players[domain.Player1].Lives = 3
	s.Players[domain.Player2].Lives = 0
	handlePostDeath(s, noopTerrain)

	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
}

// TestHandlePostDeath_TwoPlayer_GameOverWhenBothDead checks game over in 2P when both have no lives (pre-decrement).
func TestHandlePostDeath_TwoPlayer_GameOverWhenBothDead(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player1
	s.Players[domain.Player1].Lives = 0
	s.Players[domain.Player2].Lives = 0
	handlePostDeath(s, noopTerrain)

	if s.Screen != domain.ScreenGameOver {
		t.Errorf("Screen = %v, want ScreenGameOver", s.Screen)
	}
}

// TestHandlePostDeath_TwoPlayer_P2SwitchesToP1 checks P2→P1 switch.
func TestHandlePostDeath_TwoPlayer_P2SwitchesToP1(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player2
	s.Players[domain.Player1].Lives = 2
	s.Players[domain.Player2].Lives = 2
	handlePostDeath(s, noopTerrain)

	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
}

// TestResetPerLife_FuelRestored checks fuel is reset to full.
func TestResetPerLife_FuelRestored(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Fuel = 50
	ResetPerLife(s, noopTerrain)

	if s.Fuel != fuelRefuelCap {
		t.Errorf("Fuel = %d, want %d", s.Fuel, fuelRefuelCap)
	}
}

// TestResetPerLife_PlaneXCentered checks plane X is reset to center.
func TestResetPerLife_PlaneXCentered(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneX = 50
	ResetPerLife(s, noopTerrain)

	if s.PlaneX != domain.PlaneStartX {
		t.Errorf("PlaneX = %d, want %d", s.PlaneX, domain.PlaneStartX)
	}
}

// TestResetPerLife_PlaneBankCleared checks banking is reset to zero so the
// plane does not respawn in a banked state.
func TestResetPerLife_PlaneBankCleared(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneSpriteBank = 1
	ResetPerLife(s, noopTerrain)

	if s.PlaneSpriteBank != 0 {
		t.Errorf("PlaneSpriteBank = %d, want 0", s.PlaneSpriteBank)
	}
}

// TestResetPerLife_FragmentsCleared checks explosion fragments are cleared.
func TestResetPerLife_FragmentsCleared(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Explosion.Fragments = []state.ExplosionFragment{{X: 10, Y: 20}}
	ResetPerLife(s, noopTerrain)

	if len(s.Explosion.Fragments) != 0 {
		t.Errorf("Explosion.Fragments len = %d, want 0", len(s.Explosion.Fragments))
	}
}

// TestTriggerGameOver_TwoPlayer_UsesHigherScore checks that in 2P mode the higher
// of both players' scores is stored as the high score.
func TestTriggerGameOver_TwoPlayer_UsesHigherScore(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player1
	s.Players[domain.Player1].Score = 3000
	s.Players[domain.Player2].Score = 8000
	s.Config.StartingBridge = domain.StartingBridge01

	triggerGameOver(s)

	slot := domain.HighScoreSlot(domain.StartingBridge01)
	if s.HighScores[slot] != 8000 {
		t.Errorf("HighScores[%d] = %d, want 8000 (P2's higher score)", slot, s.HighScores[slot])
	}
}

func TestResetPerLife_ScorePreserved(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Score = 5000
	ResetPerLife(s, noopTerrain)

	if s.Players[domain.Player1].Score != 5000 {
		t.Errorf("Score = %d, want 5000", s.Players[domain.Player1].Score)
	}
}

// TestResetPerLife_SpawnIndexAligned checks that SpawnIndex is set to match ScrollOffset
// so the first scroll-in step does not spuriously spawn an out-of-context object.
func TestResetPerLife_SpawnIndexAligned(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	// Give the viewport a non-zero SpawnIndex to confirm it gets overwritten.
	s.Viewport.SpawnIndex = 99
	ResetPerLife(s, noopTerrain)

	wantSpawnIndex := int(s.ScrollOffset) / domain.NumLinesPerSpawnSlot
	if s.Viewport.SpawnIndex != wantSpawnIndex {
		t.Errorf("SpawnIndex = %d, want %d (aligned to ScrollOffset %d)",
			s.Viewport.SpawnIndex, wantSpawnIndex, s.ScrollOffset)
	}
}
