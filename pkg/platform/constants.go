package platform

// Screen constants.
const (
	ScreenWidth  = 256
	ScreenHeight = 192
)

// Color represents a ZX Spectrum palette color.
type Color int

// Standard ZX Spectrum palette palette colors.
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

// ZX Spectrum platform constants.
const (
	BitsPerByte = 8
)
