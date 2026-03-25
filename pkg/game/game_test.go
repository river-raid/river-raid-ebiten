package game

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// TestTickFrozenWhenPaused verifies that GameState.Tick does not advance while
// the game is paused, matching the original where the main loop is blocked by the
// interrupt handler during pause and state_tick never increments.
func TestTickFrozenWhenPaused(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenGameplay
	g.state.Paused = true
	before := g.state.Tick

	if err := g.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if g.state.Tick != before {
		t.Errorf("Tick advanced during pause: got %d, want %d", g.state.Tick, before)
	}
}

// TestTickAdvancesWhenNotPaused verifies that GameState.Tick increments by one
// per Update call during normal (non-paused) operation.
func TestTickAdvancesWhenNotPaused(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenControlSelection
	before := g.state.Tick

	if err := g.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if g.state.Tick != before+1 {
		t.Errorf("Tick = %d, want %d", g.state.Tick, before+1)
	}
}
