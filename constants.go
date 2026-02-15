package main

// Color represents a ZX Spectrum palette color.
type Color int

const (
	ColorBlack Color = iota
	ColorBlue
	ColorRed
	ColorMagenta
	ColorGreen
	ColorCyan
	ColorYellow
	ColorWhite
)

// Game color aliases.
const (
	ColorRiver      = ColorBlue
	ColorBank       = ColorGreen
	ColorPlayer1    = ColorYellow
	ColorPlayer2    = ColorCyan
	ColorHelicopter = ColorYellow
	ColorShip       = ColorCyan
	ColorBalloon    = ColorYellow
	ColorFuel       = ColorMagenta
	ColorRock       = ColorRed
	ColorMissile    = ColorGreen
	ColorExplosion  = ColorGreen
)

// Fuel constants.
const (
	FuelCheckInterval   = 3
	FuelIntakeAmount    = 4
	FuelLevelLow        = 192
	FuelLevelAlmostFull = 252
	FuelLevelFull       = 255
)

// Scoring constants.
const (
	PointsShip          = 30
	PointsHelicopterReg = 60
	PointsBalloon       = 60
	PointsFuel          = 80
	PointsFighter       = 100
	PointsHelicopterAdv = 150
	PointsTank          = 250
	PointsBridge        = 500
)

const (
	ActivationIntervalNormal = 31
)

// Player constants.
const (
	LivesInitial = 4
	PlaneStartX  = 120
	PlaneY       = 128
)

// Screen constants.
const (
	ScreenWidth    = 256
	ScreenHeight   = 192
	ViewportHeight = 136
	WindowScale    = 3
)
