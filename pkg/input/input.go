package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Interface is implemented by each input method (keyboard layouts, gamepad).
type Interface interface {
	IsLeftPressed() bool
	IsRightPressed() bool
	IsUpPressed() bool
	IsDownPressed() bool
	IsFirePressed() bool
}

// keyInterface implements Interface for a keyboard layout.
type keyInterface struct {
	fire  []ebiten.Key
	left  ebiten.Key
	up    ebiten.Key
	right ebiten.Key
	down  ebiten.Key
}

func (k keyInterface) IsLeftPressed() bool  { return ebiten.IsKeyPressed(k.left) }
func (k keyInterface) IsRightPressed() bool { return ebiten.IsKeyPressed(k.right) }
func (k keyInterface) IsUpPressed() bool    { return ebiten.IsKeyPressed(k.up) }
func (k keyInterface) IsDownPressed() bool  { return ebiten.IsKeyPressed(k.down) }

func (k keyInterface) IsFirePressed() bool {
	for _, key := range k.fire {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}

	return false
}

// kempstonInterface implements Interface for the gamepad (Kempston joystick).
type kempstonInterface struct{}

func (kempstonInterface) IsLeftPressed() bool {
	return isGamepadButtonPressed(ebiten.StandardGamepadButtonLeftLeft)
}

func (kempstonInterface) IsRightPressed() bool {
	return isGamepadButtonPressed(ebiten.StandardGamepadButtonLeftRight)
}

func (kempstonInterface) IsUpPressed() bool {
	return isGamepadButtonPressed(ebiten.StandardGamepadButtonLeftTop)
}

func (kempstonInterface) IsDownPressed() bool {
	return isGamepadButtonPressed(ebiten.StandardGamepadButtonLeftBottom)
}

func (kempstonInterface) IsFirePressed() bool {
	return isGamepadButtonPressed(ebiten.StandardGamepadButtonRightBottom)
}

func isGamepadButtonPressed(btn ebiten.StandardGamepadButton) bool {
	var ids []ebiten.GamepadID

	ids = ebiten.AppendGamepadIDs(ids)
	if len(ids) == 0 {
		return false
	}

	return ebiten.IsStandardGamepadButtonPressed(ids[0], btn)
}

// Predefined interfaces for each control type.
var (
	InterfaceArrowKeys Interface = keyInterface{
		fire:  []ebiten.Key{ebiten.KeySpace},
		left:  ebiten.KeyArrowLeft,
		right: ebiten.KeyArrowRight,
		up:    ebiten.KeyArrowUp,
		down:  ebiten.KeyArrowDown,
	}

	InterfaceKeyboard Interface = keyInterface{
		fire: []ebiten.Key{
			ebiten.KeyZ, ebiten.KeyX, ebiten.KeyC, ebiten.KeyV, ebiten.KeyB,
			ebiten.KeyN, ebiten.KeyM, ebiten.KeySpace,
		},
		left:  ebiten.KeyO,
		right: ebiten.KeyP,
		up:    ebiten.KeyQ,
		down:  ebiten.KeyA,
	}

	InterfaceSinclair Interface = keyInterface{
		fire:  []ebiten.Key{ebiten.KeyDigit0},
		left:  ebiten.KeyDigit6,
		right: ebiten.KeyDigit7,
		up:    ebiten.KeyDigit9,
		down:  ebiten.KeyDigit8,
	}

	InterfaceCursor Interface = keyInterface{
		fire:  []ebiten.Key{ebiten.KeyDigit0},
		left:  ebiten.KeyDigit5,
		right: ebiten.KeyDigit8,
		up:    ebiten.KeyDigit7,
		down:  ebiten.KeyDigit6,
	}

	InterfaceKempston Interface = kempstonInterface{}
)

// compositeInterface combines multiple interfaces into one, short-circuiting on first match.
type compositeInterface struct {
	ifaces []Interface
}

func (c compositeInterface) IsLeftPressed() bool {
	for _, iface := range c.ifaces {
		if iface.IsLeftPressed() {
			return true
		}
	}

	return false
}

func (c compositeInterface) IsRightPressed() bool {
	for _, iface := range c.ifaces {
		if iface.IsRightPressed() {
			return true
		}
	}

	return false
}

func (c compositeInterface) IsUpPressed() bool {
	for _, iface := range c.ifaces {
		if iface.IsUpPressed() {
			return true
		}
	}

	return false
}

func (c compositeInterface) IsDownPressed() bool {
	for _, iface := range c.ifaces {
		if iface.IsDownPressed() {
			return true
		}
	}

	return false
}

func (c compositeInterface) IsFirePressed() bool {
	for _, iface := range c.ifaces {
		if iface.IsFirePressed() {
			return true
		}
	}

	return false
}

// Menu positions for the optionally enabled interfaces.
const (
	selSinclair = 1
	selCursor   = 3
)

// InterfaceFor returns a composite interface for the given 0-based menu selection.
// ArrowKeys, Keyboard, and Kempston are always included. Sinclair and Cursor are
// mutually exclusive (they share the same keys) and are appended only when selected.
func InterfaceFor(sel int) Interface {
	alwaysEnabled := []Interface{InterfaceArrowKeys, InterfaceKeyboard, InterfaceKempston}

	switch sel {
	case selSinclair:
		return compositeInterface{append(alwaysEnabled, InterfaceSinclair)}
	case selCursor:
		return compositeInterface{append(alwaysEnabled, InterfaceCursor)}
	default:
		return compositeInterface{alwaysEnabled}
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

// IsRestartPressed returns true if Shift+Enter is just pressed (title screen restart).
func IsRestartPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) &&
		(ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight))
}

// IsControlSelectPressed returns true if Ctrl+Enter is just pressed (control selection screen).
func IsControlSelectPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) &&
		(ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight))
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
