package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// TestAnimateExplosionFragments_AdvancesFrame verifies that each call increments Frame by 1.
func TestAnimateExplosionFragments_AdvancesFrame(t *testing.T) {
	t.Parallel()

	frags := []state.ExplodingFragment{{X: 10, Y: 20, Frame: 1}}
	frags = animateExplosionFragments(frags)

	if len(frags) != 1 {
		t.Fatalf("len = %d, want 1", len(frags))
	}

	if frags[0].Frame != 2 {
		t.Errorf("Frame = %d, want 2", frags[0].Frame)
	}
}

// TestAnimateExplosionFragments_RemovesAfterFrameMax verifies that fragments are removed
// once their frame advances past explosionFrameMax (6).
func TestAnimateExplosionFragments_RemovesAfterFrameMax(t *testing.T) {
	t.Parallel()

	frags := []state.ExplodingFragment{{Frame: explosionFrameMax}} // will become 7 → removed
	frags = animateExplosionFragments(frags)

	if len(frags) != 0 {
		t.Errorf("len = %d, want 0 after last frame expires", len(frags))
	}
}

// TestAnimateExplosionFragments_KeepsFrameSix verifies that frame 6 (erase frame) survives
// one call and is only removed on the following call.
func TestAnimateExplosionFragments_KeepsFrameSix(t *testing.T) {
	t.Parallel()

	// Frame 5 → advances to 6, still kept.
	frags := []state.ExplodingFragment{{Frame: explosionFrameMax - 1}}
	frags = animateExplosionFragments(frags)

	if len(frags) != 1 {
		t.Fatalf("len = %d, want 1 at frame 6", len(frags))
	}

	if frags[0].Frame != explosionFrameMax {
		t.Errorf("Frame = %d, want %d", frags[0].Frame, explosionFrameMax)
	}

	// Frame 6 → advances to 7, removed.
	frags = animateExplosionFragments(frags)
	if len(frags) != 0 {
		t.Errorf("len = %d, want 0 after frame 6 expires", len(frags))
	}
}

// TestAnimateExplosionFragments_PreservesPosition verifies that X/Y are not touched.
func TestAnimateExplosionFragments_PreservesPosition(t *testing.T) {
	t.Parallel()

	frags := []state.ExplodingFragment{{X: 42, Y: 99, Frame: 3}}
	frags = animateExplosionFragments(frags)

	if frags[0].X != 42 || frags[0].Y != 99 {
		t.Errorf("position changed: got (%d,%d), want (42,99)", frags[0].X, frags[0].Y)
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

	frags := make([]state.ExplodingFragment, len(cases))
	for i, c := range cases {
		frags[i] = state.ExplodingFragment{X: 10, Y: c.startY, Frame: 1}
	}

	scrollExplosionFragments(frags, 2)

	for i, c := range cases {
		if frags[i].Y != c.wantY {
			t.Errorf("frags[%d].Y = %d, want %d", i, frags[i].Y, c.wantY)
		}
	}
}

// TestScrollExplosionFragments_DoesNotChangeX verifies that X is unchanged by scrolling.
func TestScrollExplosionFragments_DoesNotChangeX(t *testing.T) {
	t.Parallel()

	frags := []state.ExplodingFragment{{X: 33, Y: 10, Frame: 1}}
	scrollExplosionFragments(frags, 4)

	for _, f := range frags {
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
	incoming := []state.ExplodingFragment{{X: 0, Y: 0, Frame: 1}}
	spawnExplosionFragments(nil, incoming, ctrl)

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
	spawnExplosionFragments(nil, nil, ctrl)

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

	existing := make([]state.ExplodingFragment, maxExplosionFragments)
	for i := range existing {
		existing[i] = state.ExplodingFragment{Frame: 1}
	}

	incoming := []state.ExplodingFragment{{Frame: 2}, {Frame: 3}}
	ctrl := &state.ControlFlags{}
	result := spawnExplosionFragments(existing, incoming, ctrl)

	if len(result) != maxExplosionFragments {
		t.Errorf("len = %d, want %d", len(result), maxExplosionFragments)
	}

	// The newest fragments (incoming) should be at the tail.
	if result[len(result)-1].Frame != 3 {
		t.Errorf("last frame = %d, want 3 (newest fragment)", result[len(result)-1].Frame)
	}

	if result[len(result)-2].Frame != 2 {
		t.Errorf("second-to-last frame = %d, want 2", result[len(result)-2].Frame)
	}
}

// TestSpawnExplosionFragments_AppendsToExisting verifies that new fragments are added
// after existing ones when the cap is not exceeded.
func TestSpawnExplosionFragments_AppendsToExisting(t *testing.T) {
	t.Parallel()

	existing := []state.ExplodingFragment{{X: 1, Y: 1, Frame: 1}}
	incoming := []state.ExplodingFragment{{X: 2, Y: 2, Frame: 1}}
	ctrl := &state.ControlFlags{}
	result := spawnExplosionFragments(existing, incoming, ctrl)

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}

	if result[0].X != 1 || result[1].X != 2 {
		t.Errorf("wrong order: [%d, %d], want [1, 2]", result[0].X, result[1].X)
	}
}
