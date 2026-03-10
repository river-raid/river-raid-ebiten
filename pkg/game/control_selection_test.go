package game

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

func TestGameModeConfig(t *testing.T) {
	tests := []struct {
		mode           int
		isTwoPlayer    bool
		startingBridge domain.StartingBridge
	}{
		{1, false, domain.StartingBridge01},
		{2, true, domain.StartingBridge01},
		{3, false, domain.StartingBridge05},
		{4, true, domain.StartingBridge05},
		{5, false, domain.StartingBridge20},
		{6, true, domain.StartingBridge20},
		{7, false, domain.StartingBridge30},
		{8, true, domain.StartingBridge30},
	}

	for _, tt := range tests {
		cfg := ModeConfig(tt.mode)
		if cfg.IsTwoPlayer != tt.isTwoPlayer {
			t.Errorf("GameModeConfig(%d).IsTwoPlayer = %v, want %v", tt.mode, cfg.IsTwoPlayer, tt.isTwoPlayer)
		}

		if cfg.StartingBridge != tt.startingBridge {
			t.Errorf("GameModeConfig(%d).StartingBridge = %v, want %v", tt.mode, cfg.StartingBridge, tt.startingBridge)
		}
	}
}
