package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// newScrollInTestState returns a GameState in the scroll-in phase.
func newScrollInTestState() *state.GameState {
	s := state.NewGameState(0)
	s.GameplayMode = domain.GameplayScrollIn
	s.ScrollInState = scrollInScrolling
	return s
}

// TestScrollIn_DecrementsLivesOnCompletion checks that lives are decremented
// when the scroll-in sequence finishes.
func TestScrollIn_DecrementsLivesOnCompletion(t *testing.T) {
	t.Parallel()

	s := newScrollInTestState()
	s.Players[domain.Player1].Lives = 4
	terrain := newMockTerrainBuffer()

	// Drive scroll-in to completion.
	for s.ScrollInState == scrollInScrolling {
		updateScrollIn(s, terrain)
	}

	if s.Players[domain.Player1].Lives != 3 {
		t.Errorf("Lives = %d after scroll-in, want 3", s.Players[domain.Player1].Lives)
	}
}

// TestScrollIn_DecrementHappensOnce checks that repeated waits do not decrement again.
func TestScrollIn_DecrementHappensOnce(t *testing.T) {
	t.Parallel()

	s := newScrollInTestState()
	s.Players[domain.Player1].Lives = 4
	terrain := newMockTerrainBuffer()

	// Drive to completion.
	for s.ScrollInState == scrollInScrolling {
		updateScrollIn(s, terrain)
	}

	livesAfterScrollIn := s.Players[domain.Player1].Lives

	// A few more calls in waiting state should not change lives.
	for range 5 {
		updateScrollIn(s, terrain)
	}

	if s.Players[domain.Player1].Lives != livesAfterScrollIn {
		t.Errorf("Lives changed during waiting phase: got %d, want %d",
			s.Players[domain.Player1].Lives, livesAfterScrollIn)
	}
}
