package state

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// ViewportObject represents a single object in the viewport.
type ViewportObject struct {
	X            int
	Y            int
	MinX         int
	MaxX         int
	Type         domain.ObjectType
	RockVariant  int
	TankLocation domain.TankLocation
	Orientation  domain.Orientation
	IsRock       bool
	Activated    bool
}

// NewViewportObject creates a viewport object from a spawn slot. Returns nil if the slot is empty.
func NewViewportObject(slot assets.SpawnSlot) *ViewportObject {
	if slot.X == 0 {
		// the slot is empty
		return nil
	}
	return &ViewportObject{
		X:            slot.X,
		Type:         slot.Type,
		TankLocation: slot.TankLocation,
		Orientation:  slot.Orientation,
		IsRock:       slot.IsRock,
		RockVariant:  slot.RockVariant,
	}
}

// Viewport manages the active objects on screen.
type Viewport struct {
	Objects        []*ViewportObject
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

// ScrollObjects moves all objects down by the given number of pixels
// and removes any that have scrolled past the viewport bottom.
func (v *Viewport) ScrollObjects(speed int) {
	kept := v.Objects[:0]

	for i := range v.Objects {
		v.Objects[i].Y += speed

		if v.Objects[i].Y < domain.ViewportHeight {
			kept = append(kept, v.Objects[i])
		}
	}

	v.Objects = kept
}

// ActivateObjects marks inactive objects as activated based on the tick counter.
func (v *Viewport) ActivateObjects() {
	if v.Tick&v.ActivationMask != 0 {
		return
	}

	for i := range v.Objects {
		if !v.Objects[i].Activated {
			v.Objects[i].Activated = true
		}
	}
}

// Clear removes all active objects from the viewport.
func (v *Viewport) Clear() {
	v.Objects = v.Objects[:0]
}
