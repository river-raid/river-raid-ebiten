package main

import "testing"

func TestDecodeSpawnSlot_EmptySlots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		definition byte
		x          byte
	}{
		{"zero definition and zero x", 0x00, 0x00},
		{"zero definition nonzero x", 0x00, 0x50},
		{"nonzero definition zero x", 0x07, 0x00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := decodeSpawnSlot(tt.definition, tt.x)
			if got != (SpawnSlot{}) {
				t.Errorf("decodeSpawnSlot(%#02x, %#02x) = %+v, want empty SpawnSlot", tt.definition, tt.x, got)
			}
		})
	}
}

func TestDecodeSpawnSlot_ObjectTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		definition byte
		x          byte
		wantType   ObjectType
		wantOrient Orientation
		wantTank   TankLocation
	}{
		{"helicopter reg right", 0x01, 0x60, ObjectHelicopterReg, OrientationRight, TankLocationRoad},
		{"ship right", 0x02, 0x70, ObjectShip, OrientationRight, TankLocationRoad},
		{"helicopter adv right", 0x03, 0x80, ObjectHelicopterAdv, OrientationRight, TankLocationRoad},
		{"tank on road right", 0x04, 0x90, ObjectTank, OrientationRight, TankLocationRoad},
		{"fighter right", 0x05, 0xA0, ObjectFighter, OrientationRight, TankLocationRoad},
		{"balloon right", 0x06, 0xB0, ObjectBalloon, OrientationRight, TankLocationRoad},
		{"fuel depot right", 0x07, 0xC0, ObjectFuel, OrientationRight, TankLocationRoad},
		{"helicopter left", 0x41, 0x60, ObjectHelicopterReg, OrientationLeft, TankLocationRoad},
		{"ship left", 0x42, 0xA0, ObjectShip, OrientationLeft, TankLocationRoad},
		{"tank on bank right", 0x24, 0x50, ObjectTank, OrientationRight, TankLocationBank},
		{"tank on bank left", 0x64, 0x50, ObjectTank, OrientationLeft, TankLocationBank},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := decodeSpawnSlot(tt.definition, tt.x)
			if got.Type != tt.wantType {
				t.Errorf("Type = %d, want %d", got.Type, tt.wantType)
			}
			if got.IsRock {
				t.Error("IsRock = true, want false")
			}
			if got.Orientation != tt.wantOrient {
				t.Errorf("Orientation = %d, want %d", got.Orientation, tt.wantOrient)
			}
			if got.TankLocation != tt.wantTank {
				t.Errorf("TankLocation = %d, want %d", got.TankLocation, tt.wantTank)
			}
			if got.X != int(tt.x) {
				t.Errorf("X = %d, want %d", got.X, tt.x)
			}
		})
	}
}

func TestDecodeSpawnSlot_Rocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		definition  byte
		x           byte
		wantVariant int
	}{
		{"rock variant 0", 0x08, 0x78, 0},
		{"rock variant 1", 0x09, 0x80, 1},
		{"rock variant 2", 0x0A, 0x90, 2},
		{"rock variant 3", 0x0B, 0xE0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := decodeSpawnSlot(tt.definition, tt.x)
			if !got.IsRock {
				t.Error("IsRock = false, want true")
			}
			if got.RockVariant != tt.wantVariant {
				t.Errorf("RockVariant = %d, want %d", got.RockVariant, tt.wantVariant)
			}
			if got.X != int(tt.x) {
				t.Errorf("X = %d, want %d", got.X, tt.x)
			}
			if got.Type != 0 {
				t.Errorf("Type = %d, want 0 for rock", got.Type)
			}
		})
	}
}
