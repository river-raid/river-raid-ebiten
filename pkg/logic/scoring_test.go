package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestAddScore_AddsPoints(t *testing.T) {
	t.Parallel()

	player := state.PlayerState{Score: 100}
	controls := state.ControlFlags{}
	addScore(&player, &controls, 30)

	if player.Score != 130 {
		t.Errorf("score = %d, want 130", player.Score)
	}
}

func TestAddScore_BonusLifeAwarded(t *testing.T) {
	t.Parallel()

	player := state.PlayerState{Score: 9_990, Lives: 4}
	controls := state.ControlFlags{}
	addScore(&player, &controls, 30)

	if player.Lives != 5 {
		t.Errorf("lives = %d, want 5", player.Lives)
	}

	if !controls.BonusLife {
		t.Error("BonusLife flag should be set")
	}
}

func TestAddScore_NoBonusLifeWithinSameThreshold(t *testing.T) {
	t.Parallel()

	player := state.PlayerState{Score: 5_000, Lives: 4}
	controls := state.ControlFlags{}
	addScore(&player, &controls, 100)

	if player.Lives != 4 {
		t.Errorf("lives = %d, want 4", player.Lives)
	}

	if controls.BonusLife {
		t.Error("BonusLife flag should not be set")
	}
}

func TestUpdateHighScore_ReplacesIfHigher(t *testing.T) {
	t.Parallel()

	var hs [4]int
	hs[0] = 1000
	updateHighScore(&hs, 0, 2000)

	if hs[0] != 2000 {
		t.Errorf("high score = %d, want 2000", hs[0]) //nolint:gosec // G602: fixed-size [4]int, index 0 is always valid
	}
}

func TestUpdateHighScore_NoChangeIfLower(t *testing.T) {
	t.Parallel()

	var hs [4]int
	hs[0] = 5000
	updateHighScore(&hs, 0, 3000)

	if hs[0] != 5000 {
		t.Errorf("high score = %d, want 5000", hs[0]) //nolint:gosec // G602: fixed-size [4]int, index 0 is always valid
	}
}

func TestUpdateHighScore_UsesCorrectSlot(t *testing.T) {
	t.Parallel()

	var hs [4]int
	updateHighScore(&hs, 2, 9999)

	if hs[2] != 9999 {
		t.Errorf("hs[2] = %d, want 9999", hs[2]) //nolint:gosec // G602: fixed-size [4]int, index 2 is always valid
	}

	if hs[0] != 0 || hs[1] != 0 || hs[3] != 0 { //nolint:gosec // G602: fixed-size [4]int, indices 0/1/3 are always valid
		t.Error("other slots should be unchanged")
	}
}
