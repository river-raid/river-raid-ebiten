package main

// Bit positions and masks for the spawn slot definition byte.
const (
	spawnBitRock         = 3
	spawnBitTankLocation = 5
	spawnBitOrientation  = 6

	spawnMaskRockVariant = 0x03
	spawnMaskObjectType  = 0x07
)

// SpawnSlot is a decoded 2-byte level object entry.
type SpawnSlot struct {
	Type         ObjectType
	IsRock       bool
	RockVariant  int
	TankLocation TankLocation
	Orientation  Orientation
	X            int
}

// decodeSpawnSlot decodes a 2-byte spawn slot (definition + X position).
// Definition=0 or X=0 both indicate an empty slot.
func decodeSpawnSlot(definition, x byte) SpawnSlot {
	if definition == 0 || x == 0 {
		return SpawnSlot{}
	}

	s := SpawnSlot{X: int(x)}

	if definition&(1<<spawnBitRock) != 0 {
		s.IsRock = true
		s.RockVariant = int(definition & spawnMaskRockVariant)

		return s
	}

	s.Type = ObjectType(definition & spawnMaskObjectType)

	if definition&(1<<spawnBitTankLocation) != 0 {
		s.TankLocation = TankLocationBank
	}

	if definition&(1<<spawnBitOrientation) != 0 {
		s.Orientation = OrientationLeft
	}

	return s
}
