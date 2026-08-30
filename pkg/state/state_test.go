package state

import "testing"

// TestResetForNewGameClearsPause covers restarting from a paused game. Restart
// is reachable while paused, and a new game that begins paused needs a second,
// unrelated keypress before it will move.
func TestResetForNewGameClearsPause(t *testing.T) {
	t.Parallel()

	s := NewGameState()
	s.Paused = true
	s.ResetForNewGame()

	if s.Paused {
		t.Error("a restarted game is still paused")
	}
}
