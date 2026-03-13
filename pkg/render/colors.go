package render

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// playerColors contains colors for player sprites.
var playerColors = [2]platform.Color{
	domain.Player1: platform.ColorYellow,
	domain.Player2: platform.ColorCyan,
}

// Game color aliases.
const (
	colorRiver        = platform.ColorBlue
	colorBank         = platform.ColorGreen
	colorHelicopter   = platform.ColorYellow
	colorShip         = platform.ColorCyan
	colorBalloon      = platform.ColorYellow
	colorFuel         = platform.ColorMagenta
	colorFuelBlinking = platform.ColorWhite
	colorRock         = platform.ColorRed
	colorMissile      = platform.ColorGreen
	colorExplosion    = platform.ColorGreen
	colorRoad         = platform.ColorWhite
	colorBridge       = platform.ColorYellow
)

// objectColors contains default colors for objects on the river.
var objectColors = [8]platform.Color{
	domain.ObjectHelicopterReg: colorHelicopter,
	domain.ObjectShip:          colorShip,
	domain.ObjectHelicopterAdv: colorHelicopter,
	domain.ObjectTank:          platform.ColorBlue, // XOR on a green bank → blue
	domain.ObjectFighter:       platform.ColorBlue, // XOR on a green bank → blue
	domain.ObjectBalloon:       colorBalloon,
	domain.ObjectFuel:          colorFuel,
}
