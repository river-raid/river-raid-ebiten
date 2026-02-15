package main

import "github.com/hajimehoshi/ebiten/v2"

// Input holds the current frame's input state.
type Input struct {
	Left  bool
	Right bool
	Up    bool
	Down  bool
	Fire  bool
}

// ScanGameplay reads keyboard and gamepad state for gameplay actions.
func ScanGameplay() Input {
	return Input{
		Left:  ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyO),
		Right: ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyP),
		Up:    ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyQ),
		Down:  ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyA),
		Fire:  ebiten.IsKeyPressed(ebiten.KeySpace),
	}
}

// IsPausePressed returns true if the pause key (H) is pressed.
func IsPausePressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyH)
}

// IsEnterPressed returns true if the Enter key is pressed.
func IsEnterPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyEnter)
}

// ScanMenuNumber returns the number key pressed (1–8), or 0 if none.
func ScanMenuNumber() int {
	keys := []ebiten.Key{
		ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4,
		ebiten.KeyDigit5, ebiten.KeyDigit6, ebiten.KeyDigit7, ebiten.KeyDigit8,
	}

	for i, k := range keys {
		if ebiten.IsKeyPressed(k) {
			return i + 1
		}
	}

	return 0
}
