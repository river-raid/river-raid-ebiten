package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/input"
	"github.com/morozov/river-raid-ebiten/pkg/logic"
	"github.com/morozov/river-raid-ebiten/pkg/render"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Overview mode constants.
const (
	overviewBridgeLimit   = 5  // bridges before auto-transition to control selection
	overviewCrawlShift    = 2  // pixels shifted left per frame
	overviewCrawlStampCol = 32 // character column where glyphs are stamped (one beyond visible area)
	overviewCrawlHighBit  = 0x80

	// overviewCrawlCharEvery is the number of frames between character stamps: one stamp
	// per character width of scrolling.
	overviewCrawlCharEvery = assets.GlyphSize / overviewCrawlShift
)

// overviewCreditsMsg is the scrolling credits text displayed during attract mode.
// Uses GlyphTrademark for ™ and rune 127 for © (ZX Spectrum ROM character).
var overviewCreditsMsg = []rune( //nolint:gochecknoglobals // module-level constant slice
	" RIVER RAID" + string(assets.GlyphTrademark) +
		"      by Carol Shaw       \x7f 1983 " +
		string([]rune{assets.GlyphLogo0, assets.GlyphLogo1, assets.GlyphLogo2, assets.GlyphLogo3, assets.GlyphLogo4, assets.GlyphLogo5, assets.GlyphLogo6}) +
		" Inc. All rights reserved                     ",
)

// overviewGameOverMsg is the scrolling text shown once after game over, before the
// credits loop begins. It precedes msg_credits in memory with no separator, so GAME
// OVER scrolls exactly once.
var overviewGameOverMsg = []rune( //nolint:gochecknoglobals // module-level constant slice
	" .....GAME OVER.....                           ",
)

// OverviewState holds all state specific to the overview/attract screen.
// prefixMsg is placed first so GC pointer scanning covers only 8 bytes instead of
// scanning the entire struct to reach the slice pointer.
type OverviewState struct {
	// prefixMsg, when non-nil, is consumed character-by-character before the normal
	// credits loop begins. Used to prepend the GAME OVER message after a game ends.
	prefixMsg         []rune
	crawlPixels       render.CrawlPixels
	crawlTextPos      int
	bridgeCount       int
	prefixPos         int
	prevBridgeSection bool
}

func newOverviewState() *OverviewState {
	return &OverviewState{}
}

// newOverviewStateGameOver returns an OverviewState whose crawl starts with the
// GAME OVER message, falling through seamlessly into the credits loop.
func newOverviewStateGameOver() *OverviewState {
	return &OverviewState{prefixMsg: overviewGameOverMsg}
}

// updateCrawl advances the text crawl one frame: shifts pixels left then stamps
// the next character every overviewCrawlCharEvery frames.
func (o *OverviewState) updateCrawl(tick uint8) {
	for row := range assets.GlyphSize {
		copy(o.crawlPixels[row][:], o.crawlPixels[row][overviewCrawlShift:])
		width := len(o.crawlPixels[row])
		for i := width - overviewCrawlShift; i < width; i++ {
			o.crawlPixels[row][i] = false
		}
	}

	if tick%overviewCrawlCharEvery == 0 {
		o.stampNextChar()
	}
}

// stampNextChar writes the next glyph at the right edge of the crawl pixel buffer.
// If a prefix message (e.g. GAME OVER) is still being consumed it is used first;
// once exhausted the normal credits loop takes over and loops independently.
func (o *OverviewState) stampNextChar() {
	var r rune
	if o.prefixPos < len(o.prefixMsg) {
		r = o.prefixMsg[o.prefixPos]
		o.prefixPos++
	} else {
		r = overviewCreditsMsg[o.crawlTextPos]
		o.crawlTextPos++
		if o.crawlTextPos >= len(overviewCreditsMsg) {
			o.crawlTextPos = 0
		}
	}

	glyph := assets.GlyphData(r)
	stampX := overviewCrawlStampCol * assets.GlyphSize

	for row := range assets.GlyphSize {
		b := glyph[row]
		for bit := range assets.GlyphSize {
			o.crawlPixels[row][stampX+bit] = b&(overviewCrawlHighBit>>bit) != 0
		}
	}
}

// initOverview sets up the game state for attract mode using the given game mode
// number (1–8) and transitions to ScreenOverview.
func (g *Game) initOverview(mode int) {
	g.applyConfig(ModeConfig(mode))
	g.state.GameplayMode = domain.GameplayOverview
	g.state.Screen = domain.ScreenOverview
	g.overview = newOverviewState()
}

// initOverviewGameOver transitions to overview mode after game over.
// It preserves the high scores already updated by triggerGameOver, resets terrain
// for the attract demo using the same game config, and starts the crawl with the
// GAME OVER message before falling through to the credits loop.
func (g *Game) initOverviewGameOver() {
	g.applyConfig(g.state.Config)
	g.state.GameplayMode = domain.GameplayOverview
	g.state.Screen = domain.ScreenOverview
	g.overview = newOverviewStateGameOver()
}

func (g *Game) updateOverview() {
	// Enter key → reset to a fresh game and begin scroll-in.
	if input.IsEnterPressed() {
		g.state.ResetForNewGame()
		logic.ResetPerLife(g.state, g.terrain)
		g.state.Screen = domain.ScreenGameplay

		return
	}

	prevBridgeSection := g.overview.prevBridgeSection

	// Advance terrain scroll (enemies appear and move; no collision or fuel).
	logic.UpdateGameplay(g.state, g.terrain)

	g.overview.prevBridgeSection = g.state.BridgeSection
	g.overview.updateCrawl(g.state.Tick)

	// Count each new bridge that scrolls into view.
	if g.state.BridgeSection && !prevBridgeSection {
		g.overview.bridgeCount++
		if g.overview.bridgeCount >= overviewBridgeLimit {
			// Auto-transition back to control selection after 5 bridges.
			g.state = state.NewGameState()
			g.overview = nil
			g.controlSelectionPhase = 0
			g.controlSelectionTimer = controlSelectionTimeout
		}
	}
}

func (g *Game) drawOverview(screen *ebiten.Image) {
	render.DrawGameplay(screen, g.state, g.terrain)
	render.DrawCrawl(screen, &g.overview.crawlPixels)
}
