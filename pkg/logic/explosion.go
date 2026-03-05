package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// animateExplosion advances the shared animation frame by one and removes all
// fragments once the animation has completed (all frames exhausted).
// Returns the updated Explosion.
func animateExplosion(ex state.Explosion) state.Explosion {
	if len(ex.Fragments) == 0 {
		return ex
	}

	ex.Frame++
	if ex.Frame >= domain.NumExplosionSpriteFrames {
		ex.Fragments = nil
		ex.Frame = 0
	}

	return ex
}

// scrollExplosionFragments adds the current scroll speed to every fragment's Y offset,
// keeping fragments stationary relative to the terrain as the screen scrolls.
func scrollExplosionFragments(ex *state.Explosion, speed int) {
	for i := range ex.Fragments {
		ex.Fragments[i].Y += speed
	}
}

// spawnExplosionFragments appends new fragments, sets the Exploding control flag, and
// clears FireSound. If incoming is empty, the struct and flags are unchanged.
func spawnExplosionFragments(
	ex state.Explosion,
	incoming []state.ExplosionFragment,
	controls *state.ControlFlags,
) state.Explosion {
	if len(incoming) == 0 {
		return ex
	}

	controls.Exploding = true
	controls.FireSound = false

	ex.Fragments = append(ex.Fragments, incoming...)

	return ex
}
