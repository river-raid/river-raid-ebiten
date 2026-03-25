package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
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

func TestRegisterScore_ReplacesIfHigher(t *testing.T) {
	t.Parallel()

	hs := map[domain.StartingBridge]int{domain.StartingBridge01: 1000}
	registerScore(hs, domain.StartingBridge01, 2000)

	if hs[domain.StartingBridge01] != 2000 {
		t.Errorf("high score = %d, want 2000", hs[domain.StartingBridge01])
	}
}

func TestRegisterScore_NoChangeIfLower(t *testing.T) {
	t.Parallel()

	hs := map[domain.StartingBridge]int{domain.StartingBridge01: 5000}
	registerScore(hs, domain.StartingBridge01, 3000)

	if hs[domain.StartingBridge01] != 5000 {
		t.Errorf("high score = %d, want 5000", hs[domain.StartingBridge01])
	}
}

func TestRegisterScore_UsesCorrectKey(t *testing.T) {
	t.Parallel()

	hs := make(map[domain.StartingBridge]int)
	registerScore(hs, domain.StartingBridge20, 9999)

	if hs[domain.StartingBridge20] != 9999 {
		t.Errorf("hs[Bridge20] = %d, want 9999", hs[domain.StartingBridge20])
	}

	if hs[domain.StartingBridge01] != 0 || hs[domain.StartingBridge05] != 0 || hs[domain.StartingBridge30] != 0 {
		t.Error("other keys should be unchanged")
	}
}
