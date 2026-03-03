package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// newDeathTestState returns a minimal GameState suitable for death tests.
func newDeathTestState() *state.GameState {
	s := state.NewGameState(0)
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

// TestTriggerDeath_SpawnsFragments checks two fragments are spawned.
func TestTriggerDeath_SpawnsFragments(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneX = 120
	triggerDeath(s)

	if len(s.ExplodingFragments) != 2 {
		t.Fatalf("len(ExplodingFragments) = %d, want 2", len(s.ExplodingFragments))
	}

	f1 := s.ExplodingFragments[0]
	f2 := s.ExplodingFragments[1]

	wantX := 120 & domain.PlaneXAlignMask
	if f1.X != wantX {
		t.Errorf("fragment[0].X = %d, want %d", f1.X, wantX)
	}
	if f1.Y != domain.DeathFragmentY {
		t.Errorf("fragment[0].Y = %d, want %d", f1.Y, domain.DeathFragmentY)
	}
	if f2.Y != domain.DeathFragmentY+domain.DeathFragmentSpacing {
		t.Errorf("fragment[1].Y = %d, want %d", f2.Y, domain.DeathFragmentY+domain.DeathFragmentSpacing)
	}
}

// TestTriggerDeath_AlignsPlaneX checks 8-pixel alignment of the spawn X.
func TestTriggerDeath_AlignsPlaneX(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneX = 125 // not 8-pixel aligned
	triggerDeath(s)

	wantX := 125 & domain.PlaneXAlignMask // 120
	if s.ExplodingFragments[0].X != wantX {
		t.Errorf("fragment X = %d, want %d (aligned)", s.ExplodingFragments[0].X, wantX)
	}
}

// TestUpdateDying_DecrementsFrame checks DyingFrame decrements each call.
func TestUpdateDying_DecrementsFrame(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	triggerDeath(s)
	initial := s.DyingFrame

	updateDying(s)

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
	// Give the player lives so it restarts (not game over).
	s.Players[domain.Player1].Lives = 2

	// Drive DyingFrame to zero.
	for s.DyingFrame > 0 {
		updateDying(s)
	}

	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
}

// TestHandlePostDeath_DecrementsLives checks life is decremented.
func TestHandlePostDeath_DecrementsLives(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Lives = 3
	handlePostDeath(s)

	if s.Players[domain.Player1].Lives != 2 {
		t.Errorf("Lives = %d, want 2", s.Players[domain.Player1].Lives)
	}
}

// TestHandlePostDeath_SinglePlayerRestart checks restart when lives remain.
func TestHandlePostDeath_SinglePlayerRestart(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Lives = 2
	handlePostDeath(s)

	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
}

// TestHandlePostDeath_SinglePlayerGameOver checks game over when no lives remain.
func TestHandlePostDeath_SinglePlayerGameOver(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Lives = 1 // will become 0 after decrement
	handlePostDeath(s)

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
	handlePostDeath(s)

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
	handlePostDeath(s)

	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
	if s.GameplayMode != domain.GameplayScrollIn {
		t.Errorf("GameplayMode = %v, want GameplayScrollIn", s.GameplayMode)
	}
}

// TestHandlePostDeath_TwoPlayer_GameOverWhenBothDead checks game over in 2P when both have no lives.
func TestHandlePostDeath_TwoPlayer_GameOverWhenBothDead(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Config.IsTwoPlayer = true
	s.CurrentPlayer = domain.Player1
	s.Players[domain.Player1].Lives = 1 // will become 0
	s.Players[domain.Player2].Lives = 0
	handlePostDeath(s)

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
	handlePostDeath(s)

	if s.CurrentPlayer != domain.Player1 {
		t.Errorf("CurrentPlayer = %v, want Player1", s.CurrentPlayer)
	}
}

// TestResetPerLife_FuelRestored checks fuel is reset to full.
func TestResetPerLife_FuelRestored(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Fuel = 50
	resetPerLife(s)

	if s.Fuel != domain.FuelLevelFull {
		t.Errorf("Fuel = %d, want %d", s.Fuel, domain.FuelLevelFull)
	}
}

// TestResetPerLife_PlaneXCentered checks plane X is reset to center.
func TestResetPerLife_PlaneXCentered(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.PlaneX = 50
	resetPerLife(s)

	if s.PlaneX != domain.PlaneStartX {
		t.Errorf("PlaneX = %d, want %d", s.PlaneX, domain.PlaneStartX)
	}
}

// TestResetPerLife_FragmentsCleared checks explosion fragments are cleared.
func TestResetPerLife_FragmentsCleared(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.ExplodingFragments = []state.ExplodingFragment{{X: 10, Y: 20, Frame: 3}}
	resetPerLife(s)

	if len(s.ExplodingFragments) != 0 {
		t.Errorf("ExplodingFragments len = %d, want 0", len(s.ExplodingFragments))
	}
}

// TestResetPerLife_ScorePreserved checks per-player score is not reset.
func TestResetPerLife_ScorePreserved(t *testing.T) {
	t.Parallel()

	s := newDeathTestState()
	s.Players[domain.Player1].Score = 5000
	resetPerLife(s)

	if s.Players[domain.Player1].Score != 5000 {
		t.Errorf("Score = %d, want 5000", s.Players[domain.Player1].Score)
	}
}
