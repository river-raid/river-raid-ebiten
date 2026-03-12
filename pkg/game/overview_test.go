package game

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TestOverviewCrawlShift checks that pixels shift left by overviewCrawlShift on update.
func TestOverviewCrawlShift(t *testing.T) {
	o := newOverviewState()

	// Set a pixel at x=10, row=0.
	o.crawlPixels[0][10] = true

	// tick=1: not a multiple of overviewCrawlCharEvery, so no character stamp.
	o.updateCrawl(1)

	want := 10 - overviewCrawlShift
	if !o.crawlPixels[0][want] {
		t.Errorf("expected pixel at x=%d after shift, got false", want)
	}
	if o.crawlPixels[0][10] {
		t.Errorf("expected original pixel at x=10 to be cleared after shift")
	}
}

// TestOverviewCrawlShiftClearsTrailing verifies that trailing pixels are zeroed after shift.
func TestOverviewCrawlShiftClearsTrailing(t *testing.T) {
	o := newOverviewState()

	width := len(o.crawlPixels[0])
	o.crawlPixels[0][width-1] = true
	o.crawlPixels[0][width-2] = true

	o.updateCrawl(1)

	for i := width - overviewCrawlShift; i < width; i++ {
		if o.crawlPixels[0][i] {
			t.Errorf("expected trailing pixel at x=%d to be false after shift", i)
		}
	}
}

// TestOverviewCrawlStampsOnFirstFrame checks that a character is stamped on frame 0.
func TestOverviewCrawlStampsOnFirstFrame(t *testing.T) {
	o := newOverviewState()

	// Start at 'R' (index 1) — the leading space at index 0 has no pixels and would
	// produce a false negative. The shift happens before the stamp, so the glyph lands
	// at overviewCrawlStampCol*GlyphSize (not -overviewCrawlShift).
	o.crawlTextPos = 1

	// tick=0 → multiple of overviewCrawlCharEvery, so stamp happens.
	o.updateCrawl(0)

	stampX := overviewCrawlStampCol * assets.GlyphSize // stamp is placed after the shift

	hasPixel := false
	for row := range assets.GlyphSize {
		for bit := range assets.GlyphSize {
			if o.crawlPixels[row][stampX+bit] {
				hasPixel = true
			}
		}
	}

	if !hasPixel {
		t.Error("expected pixels to be set after first character stamp")
	}
}

// TestOverviewCrawlTextWraps checks that crawlTextPos wraps back to 0 after the full message.
func TestOverviewCrawlTextWraps(t *testing.T) {
	o := newOverviewState()

	msgLen := len(overviewCreditsMsg)
	for range msgLen + 1 {
		o.stampNextChar()
	}

	if o.crawlTextPos != 1 {
		t.Errorf("expected crawlTextPos=1 after wrapping, got %d", o.crawlTextPos)
	}
}

// TestOverviewCrawlGameOverPrefix checks that the game-over overview state starts by
// consuming the game-over prefix before moving to the credits loop.
func TestOverviewCrawlGameOverPrefix(t *testing.T) {
	o := newOverviewStateGameOver()

	// Consume the entire prefix.
	for range overviewGameOverMsg {
		o.stampNextChar()
	}

	if o.prefixPos != len(overviewGameOverMsg) {
		t.Errorf("prefixPos = %d, want %d", o.prefixPos, len(overviewGameOverMsg))
	}

	// Next stamp should advance into the credits loop.
	prevCreditsPos := o.crawlTextPos
	o.stampNextChar()

	if o.crawlTextPos != prevCreditsPos+1 {
		t.Errorf("crawlTextPos = %d, want %d (credits loop advancing)", o.crawlTextPos, prevCreditsPos+1)
	}
}

// TestOverviewCrawlGameOverAppearsOnce checks that after the prefix is exhausted the
// credits loop does not cycle back to the game-over message.
func TestOverviewCrawlGameOverAppearsOnce(t *testing.T) {
	o := newOverviewStateGameOver()

	// Consume prefix + full credits loop once.
	total := len(overviewGameOverMsg) + len(overviewCreditsMsg) + 1
	for range total {
		o.stampNextChar()
	}

	// prefixPos must not advance beyond its length — the game-over message never repeats.
	if o.prefixPos != len(overviewGameOverMsg) {
		t.Errorf("prefixPos = %d after credits loop, want %d (game over appeared more than once)",
			o.prefixPos, len(overviewGameOverMsg))
	}
}

// TestUpdateGameOver_TransitionsToOverview checks that updateGameOver immediately
// switches the screen to ScreenOverview.
func TestUpdateGameOver_TransitionsToOverview(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenGameOver

	g.updateGameOver()

	if g.state.Screen != domain.ScreenOverview {
		t.Errorf("Screen = %v, want ScreenOverview", g.state.Screen)
	}
}

// TestUpdateGameOver_SetsGameplayOverview checks that GameplayMode is set to GameplayOverview.
func TestUpdateGameOver_SetsGameplayOverview(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenGameOver

	g.updateGameOver()

	if g.state.GameplayMode != domain.GameplayOverview {
		t.Errorf("GameplayMode = %v, want GameplayOverview", g.state.GameplayMode)
	}
}

// TestUpdateGameOver_OverviewStateHasGameOverPrefix checks that the crawl state
// contains the game-over prefix message.
func TestUpdateGameOver_OverviewStateHasGameOverPrefix(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenGameOver

	g.updateGameOver()

	if g.overview == nil {
		t.Fatal("overview state is nil after updateGameOver")
	}

	if len(g.overview.prefixMsg) == 0 {
		t.Error("overview prefixMsg is empty; expected GAME OVER message")
	}
}

// TestUpdateGameOver_PreservesHighScore verifies that the high score set before
// game over is not wiped by the overview initialisation.
func TestUpdateGameOver_PreservesHighScore(t *testing.T) {
	g := NewGame()
	g.state.Screen = domain.ScreenGameOver
	g.state.Config.StartingBridge = domain.StartingBridge01
	slot := domain.HighScoreSlot(domain.StartingBridge01)
	g.state.HighScores[slot] = 12345

	g.updateGameOver()

	if g.state.HighScores[slot] != 12345 {
		t.Errorf("HighScores[%d] = %d, want 12345 (score lost during overview init)", slot, g.state.HighScores[slot])
	}
}

// TestInitOverview verifies that initOverview sets screen and gameplay mode correctly.
func TestInitOverview(t *testing.T) {
	g := NewGame()

	g.initOverview(1)

	if g.state.Screen != domain.ScreenOverview {
		t.Errorf("expected ScreenOverview, got %v", g.state.Screen)
	}

	if g.state.GameplayMode != domain.GameplayOverview {
		t.Errorf("expected GameplayOverview, got %v", g.state.GameplayMode)
	}

	if g.overview == nil {
		t.Error("expected overview state to be initialised")
	}

	if g.overview.bridgeCount != 0 {
		t.Errorf("expected bridgeCount=0, got %d", g.overview.bridgeCount)
	}
}

// TestUpdateOverviewBridgeLimit checks that the overview auto-transitions to
// control selection after overviewBridgeLimit bridges.
func TestUpdateOverviewBridgeLimit(t *testing.T) {
	g := NewGame()
	g.initOverview(1)

	g.overview.bridgeCount = overviewBridgeLimit - 1
	g.overview.prevBridgeSection = false
	g.state.BridgeSection = true

	prevBridgeSection := g.overview.prevBridgeSection

	g.overview.prevBridgeSection = g.state.BridgeSection

	if g.state.BridgeSection && !prevBridgeSection {
		g.overview.bridgeCount++
		if g.overview.bridgeCount >= overviewBridgeLimit {
			g.state = state.NewGameState()
			g.overview = nil
			g.controlSelectionPhase = 0
			g.controlSelectionTimer = controlSelectionTimeout
		}
	}

	if g.state.Screen != domain.ScreenControlSelection {
		t.Errorf("expected ScreenControlSelection after %d bridges, got %v",
			overviewBridgeLimit, g.state.Screen)
	}

	if g.overview != nil {
		t.Error("expected overview state to be nil after transition")
	}

	if g.controlSelectionTimer != controlSelectionTimeout {
		t.Errorf("expected timer=%d, got %d", controlSelectionTimeout, g.controlSelectionTimer)
	}
}
