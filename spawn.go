package main

// SpawnSlot is a level object entry with all fields directly represented.
type SpawnSlot struct {
	Type         ObjectType
	IsRock       bool
	RockVariant  int
	TankLocation TankLocation
	Orientation  Orientation
	X            int
}
