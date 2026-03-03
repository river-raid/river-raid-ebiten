package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Scoring constants.
const (
	bonusLifeStep = 10_000 // bonus life awarded every this many points
)

// addScore adds points to the given player's score.
// Awards a bonus life for each 10,000-point threshold crossed.
func addScore(player *state.PlayerState, controls *state.ControlFlags, points int) {
	prev := player.Score
	player.Score += points

	if player.Score/bonusLifeStep > prev/bonusLifeStep {
		player.Lives++
		controls.BonusLife = true
	}
}

// updateHighScore replaces the high score for the given starting bridge slot if
// the provided score exceeds it.
func updateHighScore(highScores *[4]int, slot, score int) {
	if score > highScores[slot] {
		highScores[slot] = score
	}
}
