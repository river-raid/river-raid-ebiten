package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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

// IsPausePressed returns true if the pause key (H) is just pressed.
func IsPausePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyH)
}

// IsUnpausePressed returns true if any key other than H is just pressed (unpause trigger).
func IsUnpausePressed() bool {
	var keys []ebiten.Key
	keys = inpututil.AppendJustPressedKeys(keys)

	for _, k := range keys {
		if k != ebiten.KeyH {
			return true
		}
	}

	return false
}

// IsEnterPressed returns true if the Enter key is pressed.
func IsEnterPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

// ScanMenuNumber returns the number key (1–count) just pressed this frame, or 0 if none.
func ScanMenuNumber(count int) int {
	keys := []ebiten.Key{
		ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4,
		ebiten.KeyDigit5, ebiten.KeyDigit6, ebiten.KeyDigit7, ebiten.KeyDigit8,
	}

	for i, k := range keys {
		if i >= count {
			break
		}

		if inpututil.IsKeyJustPressed(k) {
			return i + 1
		}
	}

	return 0
}
