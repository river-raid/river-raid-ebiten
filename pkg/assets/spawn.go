package assets

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

// SpawnSlot is a level object entry with all fields directly represented.
type SpawnSlot struct {
	Type         domain.ObjectType
	IsRock       bool
	RockVariant  int
	TankLocation domain.TankLocation
	Orientation  domain.Orientation
	X            int
}
