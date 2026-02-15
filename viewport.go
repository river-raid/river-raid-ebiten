package main

// Viewport manages the active object slots on screen.
type Viewport struct {
	Slots          []Slot
	SpawnIndex     int // current index into LevelObjects for spawning
	ActivationMask int // 31 normally, 15 after bridge destruction
	Tick           int // frame counter for activation timing
}

// NewViewport creates a viewport with default activation timing.
func NewViewport() Viewport {
	return Viewport{
		ActivationMask: ActivationIntervalNormal,
	}
}

// SpawnFromScroll checks the current level's spawn data and adds new objects
// to the viewport. Called when the scroll advances past a new spawn slot.
func (v *Viewport) SpawnFromScroll(bridgeIndex, spawnIdx int) {
	if spawnIdx == v.SpawnIndex {
		return // already spawned this slot
	}

	v.SpawnIndex = spawnIdx

	slot := LevelObjects[bridgeIndex][spawnIdx]
	if slot.X == 0 {
		return // empty spawn slot
	}

	// Rocks are rendered inline during terrain drawing, not as active slots.
	if slot.IsRock {
		return
	}

	v.Slots = append(v.Slots, Slot{
		X:            slot.X,
		Y:            0, // spawns at top of viewport
		Type:         slot.Type,
		TankLocation: slot.TankLocation,
		Orientation:  slot.Orientation,
	})
}

// ActivateObjects marks inactive objects as activated based on the tick counter.
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

// ScrollObjects moves all objects down by the given number of pixels
// and removes any that have scrolled past the viewport bottom.
func (v *Viewport) ScrollObjects(speed int) {
	v.Tick++

	kept := v.Slots[:0]

	for i := range v.Slots {
		v.Slots[i].Y += speed

		if v.Slots[i].Y < ViewportHeight {
			kept = append(kept, v.Slots[i])
		}
	}

	v.Slots = kept
}

// Clear removes all active objects from the viewport.
func (v *Viewport) Clear() {
	v.Slots = v.Slots[:0]
}
