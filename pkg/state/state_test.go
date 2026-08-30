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

// TestHaltedCoversEveryStop pins which states stop the game. The game tick, the
// world step and the mixer all read the predicate, so widening or narrowing it
// moves them together.
func TestHaltedCoversEveryStop(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		gs   GameState
		want bool
	}{
		{"running", GameState{}, false},
		{"paused", GameState{Paused: true}, true},
		{"awaiting control", GameState{ScrollInState: ScrollInWaiting}, true},
		{"both", GameState{Paused: true, ScrollInState: ScrollInWaiting}, true},
	}

	for _, c := range cases {
		if got := c.gs.Halted(); got != c.want {
			t.Errorf("%s: Halted() = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestAdvanceStopsWhenHalted ties the game clock to Halted. A halted game that
// keeps ticking animates the blades and the fuel station while the world and the
// sound are stopped.
func TestAdvanceStopsWhenHalted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		gs   GameState
		want int
	}{
		{"running", GameState{}, 1},
		{"paused", GameState{Paused: true}, 0},
		{"awaiting control", GameState{ScrollInState: ScrollInWaiting}, 0},
	}

	for _, c := range cases {
		c.gs.Advance()

		if c.gs.Tick != c.want {
			t.Errorf("%s: Tick = %d after Advance, want %d", c.name, c.gs.Tick, c.want)
		}
	}
}

// TestHaltedPanicsOnUnknownPhase covers the project rule that an invalid state
// must panic rather than be silently read as a valid one. Without it any value
// but ScrollInWaiting reads as "running".
func TestHaltedPanicsOnUnknownPhase(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Error("Halted accepted an out-of-range ScrollInPhase")
		}
	}()

	gs := GameState{ScrollInState: ScrollInPhase(99)}
	_ = gs.Halted()
}
