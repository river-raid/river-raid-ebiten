package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TestAnimateExplosionFragments_AdvancesFrame verifies that each call increments Frame by 1.
func TestAnimateExplosionFragments_AdvancesFrame(t *testing.T) {
	t.Parallel()

	ex := state.Explosion{Fragments: []state.ExplosionFragment{{X: 10, Y: 20}}}
	ex = animateExplosion(ex)

	if len(ex.Fragments) != 1 {
		t.Fatalf("len = %d, want 1", len(ex.Fragments))
	}

	if ex.Frame != 1 {
		t.Errorf("Frame = %d, want 1", ex.Frame)
	}
}

// TestAnimateExplosionFragments_RemovesAfterFrameMax verifies that fragments are removed
// once their frame advances past the last sprite frame.
func TestAnimateExplosionFragments_RemovesAfterFrameMax(t *testing.T) {
	t.Parallel()

	// Start at the last valid frame; one more call must clear the fragments.
	ex := state.Explosion{Fragments: []state.ExplosionFragment{{X: 0, Y: 0}}, Frame: domain.NumExplosionSpriteFrames - 1}
	ex = animateExplosion(ex)

	if len(ex.Fragments) != 0 {
		t.Errorf("len = %d, want 0 after last frame expires", len(ex.Fragments))
	}
}

// TestAnimateExplosionFragments_PreservesPosition verifies that X/Y are not touched.
func TestAnimateExplosionFragments_PreservesPosition(t *testing.T) {
	t.Parallel()

	ex := state.Explosion{Fragments: []state.ExplosionFragment{{X: 42, Y: 99}}, Frame: 3}
	ex = animateExplosion(ex)

	if ex.Fragments[0].X != 42 || ex.Fragments[0].Y != 99 {
		t.Errorf("position changed: got (%d,%d), want (42,99)", ex.Fragments[0].X, ex.Fragments[0].Y)
	}
}

// TestAnimateExplosionFragments_NoOpOnEmpty verifies that animating an empty Explosion
// returns it unchanged.
func TestAnimateExplosionFragments_NoOpOnEmpty(t *testing.T) {
	t.Parallel()

	ex := state.Explosion{}
	ex = animateExplosion(ex)

	if len(ex.Fragments) != 0 || ex.Frame != 0 {
		t.Errorf("non-empty result on empty input: fragments=%d frame=%d", len(ex.Fragments), ex.Frame)
	}
}

// TestScrollExplosionFragments_AddsSpeedToY verifies that Y is shifted by the scroll speed.
func TestScrollExplosionFragments_AddsSpeedToY(t *testing.T) {
	t.Parallel()

	cases := []struct {
		startY int
		wantY  int
	}{
		{50, 52},
		{80, 82},
	}

	frags := make([]state.ExplosionFragment, len(cases))
	for i, c := range cases {
		frags[i] = state.ExplosionFragment{X: 10, Y: c.startY}
	}

	ex := state.Explosion{Fragments: frags}
	scrollExplosionFragments(&ex, 2)

	for i, c := range cases {
		if ex.Fragments[i].Y != c.wantY {
			t.Errorf("frags[%d].Y = %d, want %d", i, ex.Fragments[i].Y, c.wantY)
		}
	}
}

// TestScrollExplosionFragments_DoesNotChangeX verifies that X is unchanged by scrolling.
func TestScrollExplosionFragments_DoesNotChangeX(t *testing.T) {
	t.Parallel()

	ex := state.Explosion{Fragments: []state.ExplosionFragment{{X: 33, Y: 10}}}
	scrollExplosionFragments(&ex, 4)

	for _, f := range ex.Fragments {
		if f.X != 33 {
			t.Errorf("X changed: got %d, want 33", f.X)
		}
	}
}

// TestSpawnExplosionFragments_SetsExplodingFlag verifies that Controls.Exploding is set
// and Controls.FireSound is cleared when new fragments are spawned.
func TestSpawnExplosionFragments_SetsExplodingFlag(t *testing.T) {
	t.Parallel()

	ctrl := &state.ControlFlags{FireSound: true, Exploding: false}
	incoming := []state.ExplosionFragment{{X: 0, Y: 0}}
	spawnExplosionFragments(state.Explosion{}, incoming, ctrl)

	if !ctrl.Exploding {
		t.Error("Exploding should be set")
	}

	if ctrl.FireSound {
		t.Error("FireSound should be cleared")
	}
}

// TestSpawnExplosionFragments_NoOpOnEmpty verifies that no flags are changed when
// the incoming slice is empty.
func TestSpawnExplosionFragments_NoOpOnEmpty(t *testing.T) {
	t.Parallel()

	ctrl := &state.ControlFlags{FireSound: true, Exploding: false}
	spawnExplosionFragments(state.Explosion{}, nil, ctrl)

	if ctrl.Exploding {
		t.Error("Exploding should not be set for empty incoming")
	}

	if !ctrl.FireSound {
		t.Error("FireSound should be unchanged for empty incoming")
	}
}

// TestSpawnExplosionFragments_CapsAtMax verifies that the fragment slice never exceeds
// maxExplosionFragments.
func TestSpawnExplosionFragments_CapsAtMax(t *testing.T) {
	t.Parallel()

	existing := make([]state.ExplosionFragment, maxExplosionFragments)
	for i := range existing {
		existing[i] = state.ExplosionFragment{X: i, Y: 0}
	}

	incoming := []state.ExplosionFragment{{X: 100, Y: 1}, {X: 101, Y: 2}}
	ctrl := &state.ControlFlags{}
	result := spawnExplosionFragments(state.Explosion{Fragments: existing}, incoming, ctrl)

	if len(result.Fragments) != maxExplosionFragments {
		t.Errorf("len = %d, want %d", len(result.Fragments), maxExplosionFragments)
	}

	// The newest fragments (incoming) should be at the tail.
	if result.Fragments[len(result.Fragments)-1].X != 101 {
		t.Errorf("last X = %d, want 101 (newest fragment)", result.Fragments[len(result.Fragments)-1].X)
	}

	if result.Fragments[len(result.Fragments)-2].X != 100 {
		t.Errorf("second-to-last X = %d, want 100", result.Fragments[len(result.Fragments)-2].X)
	}
}

// TestSpawnExplosionFragments_AppendsToExisting verifies that new fragments are added
// after existing ones when the cap is not exceeded.
func TestSpawnExplosionFragments_AppendsToExisting(t *testing.T) {
	t.Parallel()

	existing := state.Explosion{Fragments: []state.ExplosionFragment{{X: 1, Y: 1}}}
	incoming := []state.ExplosionFragment{{X: 2, Y: 2}}
	ctrl := &state.ControlFlags{}
	result := spawnExplosionFragments(existing, incoming, ctrl)

	if len(result.Fragments) != 2 {
		t.Fatalf("len = %d, want 2", len(result.Fragments))
	}

	if result.Fragments[0].X != 1 || result.Fragments[1].X != 2 {
		t.Errorf("wrong order: [%d, %d], want [1, 2]", result.Fragments[0].X, result.Fragments[1].X)
	}
}
