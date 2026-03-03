package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// maxExplosionFragments is the maximum number of active explosion fragments.
const maxExplosionFragments = 16

// explosionFrameMax is the last animation frame; fragments are removed after this frame.
const explosionFrameMax = 6

// animateExplosionFragments advances each fragment's animation by one frame and removes
// fragments that have completed their animation (Frame > explosionFrameMax).
// Returns the updated slice.
func animateExplosionFragments(fragments []state.ExplodingFragment) []state.ExplodingFragment {
	out := fragments[:0]

	for _, f := range fragments {
		f.Frame++
		if f.Frame > explosionFrameMax {
			continue // remove completed fragment
		}

		out = append(out, f)
	}

	return out
}

// scrollExplosionFragments adds the current scroll speed to every fragment's Y offset,
// keeping fragments stationary relative to the terrain as the screen scrolls.
func scrollExplosionFragments(fragments []state.ExplodingFragment, speed int) {
	for i := range fragments {
		fragments[i].Y += speed
	}
}

// spawnExplosionFragments appends new fragments, sets the Exploding control flag, clears
// FireSound, and caps the total at maxExplosionFragments (excess oldest are dropped).
// If incoming is empty, the slice and flags are unchanged.
func spawnExplosionFragments(
	existing []state.ExplodingFragment,
	incoming []state.ExplodingFragment,
	controls *state.ControlFlags,
) []state.ExplodingFragment {
	if len(incoming) == 0 {
		return existing
	}

	controls.Exploding = true
	controls.FireSound = false

	existing = append(existing, incoming...)
	if len(existing) > maxExplosionFragments {
		existing = existing[len(existing)-maxExplosionFragments:]
	}

	return existing
}
