package state

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// ViewportSlot represents a single object in the viewport.
type ViewportSlot struct {
	X            int
	Y            int
	Type         domain.ObjectType
	RockVariant  int
	TankLocation domain.TankLocation
	Orientation  domain.Orientation
	IsRock       bool
	Activated    bool
}

// Viewport manages the active object slots on screen.
type Viewport struct {
	Slots          []ViewportSlot
	SpawnIndex     int // current index into spawnSlots for spawning
	ActivationMask int // 31 normally, 15 after bridge destruction
	Tick           int // frame counter for activation timing
}

// NewViewport creates a viewport with default activation timing.
func NewViewport() *Viewport {
	return &Viewport{
		ActivationMask: domain.ActivationIntervalNormal,
	}
}

// UpdateForScroll performs all viewport updates for a scroll event atomically.
// This includes spawning new objects, scrolling existing objects, and activating objects.
// The game should call this once per scroll advance, not individual spawn/scroll/activate methods.
func (v *Viewport) UpdateForScroll(bridgeIndex, spawnIdx, speed int) {
	// Step 1: Spawn new objects based on scroll position.
	v.SpawnFromScroll(bridgeIndex, spawnIdx)

	// Step 2: Increment tick counter.
	v.Tick++

	// Step 3: Scroll all objects down and remove those off-screen.
	v.ScrollObjects(speed)

	// Step 4: Activate objects based on tick counter.
	v.ActivateObjects()
}

// SpawnFromScroll checks the current level's spawn data and adds new objects
// to the viewport. Called when the scroll advances past a new spawn slot.
// Exposed for testing; game code should use UpdateForScroll instead.
func (v *Viewport) SpawnFromScroll(bridgeIndex, spawnIdx int) {
	if spawnIdx == v.SpawnIndex {
		return // already spawned this spawnSlot
	}

	v.SpawnIndex = spawnIdx

	spawnSlot := assets.SpawnSlots[bridgeIndex][spawnIdx]
	if spawnSlot.X == 0 {
		return // empty spawn spawnSlot
	}

	v.Slots = append(v.Slots, ViewportSlot{
		X:            spawnSlot.X,
		Y:            0, // spawns at top of viewport
		Type:         spawnSlot.Type,
		TankLocation: spawnSlot.TankLocation,
		Orientation:  spawnSlot.Orientation,
		IsRock:       spawnSlot.IsRock,
		RockVariant:  spawnSlot.RockVariant,
	})
}

// ScrollObjects moves all objects down by the given number of pixels
// and removes any that have scrolled past the viewport bottom.
// Exposed for testing; game code should use UpdateForScroll instead.
func (v *Viewport) ScrollObjects(speed int) {
	kept := v.Slots[:0]

	for i := range v.Slots {
		v.Slots[i].Y += speed

		if v.Slots[i].Y < domain.ViewportHeight {
			kept = append(kept, v.Slots[i])
		}
	}

	v.Slots = kept
}

// ActivateObjects marks inactive objects as activated based on the tick counter.
// Exposed for testing; game code should use UpdateForScroll instead.
func (v *Viewport) ActivateObjects() {
	if v.Tick&v.ActivationMask != 0 {
		return
	}

	for i := range v.Slots {
		if !v.Slots[i].Activated {
			v.Slots[i].Activated = true
		}
	}
}

// Clear removes all active objects from the viewport.
func (v *Viewport) Clear() {
	v.Slots = v.Slots[:0]
}
