package assets

import (
	"github.com/morozov/river-raid-ebiten/pkg/platform"
)

// TextSpan defines a run of text at a given character-cell position with
// ink (foreground) and paper (background) colors.
type TextSpan struct {
	Text     string
	Row, Col int
	Ink      platform.Color
}
