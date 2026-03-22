package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// newScrollInTestState returns a GameState in the scroll-in phase.
func newScrollInTestState() *state.GameState {
	s := state.NewGameState()
	ResetPerLife(s, noopTerrain)
	s.GameplayMode = domain.GameplayScrollIn
	s.ScrollInState = scrollInScrolling
	s.InputInterface = input.InterfaceFor(0)
	return s
}

// TestStep_BridgeHit_SetsActivationFast verifies that hitting a bridge raises the
// activation mask to ActivationIntervalFast for the remainder of the life.
func TestStep_BridgeHit_SetsActivationFast(t *testing.T) {
	t.Parallel()

	s := state.NewGameState()
	s.GameplayMode = domain.GameplayNormal
	s.InputInterface = input.InterfaceFor(0)
	s.PlaneX = domain.PlaneStartX
	s.Missile = &state.PlayerMissile{}
	s.TankShell = &state.TankShell{}
	s.HeliMissile = &state.HeliMissile{}
	s.Viewport = state.NewViewport()

	// Position the missile to hit the bridge (bridgeY=60, extent=22: missile at Y=45 overlaps).
	const bridgeY = 60
	s.BridgeSection = true
	s.BridgeYPosition = bridgeY
	s.Missile.Active = true
	s.Missile.X = domain.PlaneStartX
	s.Missile.Y = 45

	step(s, s.InputInterface, newMockTerrainBuffer())

	if s.Viewport.ActivationMask != domain.ActivationIntervalFast {
		t.Errorf("ActivationMask = %d after bridge hit, want %d (ActivationIntervalFast)",
			s.Viewport.ActivationMask, domain.ActivationIntervalFast)
	}
}

// TestScrollIn_DecrementsLivesOnCompletion checks that lives are decremented
// when the scroll-in sequence finishes.
func TestScrollIn_DecrementsLivesOnCompletion(t *testing.T) {
	t.Parallel()

	s := newScrollInTestState()
	s.Players[domain.Player1].Lives = 4
	terrain := newMockTerrainBuffer()

	// Drive scroll-in to completion.
	for s.GameplayMode == domain.GameplayScrollIn {
		updateScrollIn(s, terrain)
	}

	if s.Players[domain.Player1].Lives != 3 {
		t.Errorf("Lives = %d after scroll-in, want 3", s.Players[domain.Player1].Lives)
	}
}
