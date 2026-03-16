package assets

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

func TestLevelObjects_EmptySlots(t *testing.T) {
	t.Parallel()

	// Level 0, slot 0 should be empty (zero value).
	got := SpawnSlots[0][0]
	if got != (SpawnSlot{}) {
		t.Errorf("SpawnSlots[0][0] = %+v, want empty SpawnSlot", got)
	}
}

func TestLevelObjects_KnownObjects(t *testing.T) {
	t.Parallel()

	// Level 0, slot 12: original byte 0x87 = fuel depot (type 7), bit 7 set (unused), X=0x78.
	got := SpawnSlots[0][12]
	want := SpawnSlot{Type: domain.ObjectFuel, X: 0x7b}

	if got != want {
		t.Errorf("SpawnSlots[0][12] = %+v, want %+v", got, want)
	}

	// Level 0, slot 17: original byte 0x42 = ship (type 2), orientation left, X=0xa0.
	got = SpawnSlots[0][17]
	want = SpawnSlot{Type: domain.ObjectShip, Orientation: domain.OrientationLeft, X: 0xa0}

	if got != want {
		t.Errorf("SpawnSlots[0][17] = %+v, want %+v", got, want)
	}
}

func TestLevelObjects_KnownRocks(t *testing.T) {
	t.Parallel()

	// Level 0, slot 19: original byte 0x0b = rock (bit 3 set), variant 3 (bits 0-1 = 0x03), X=0x08.
	got := SpawnSlots[0][19]
	want := SpawnSlot{IsRock: true, RockVariant: 3, X: 0x08}

	if got != want {
		t.Errorf("SpawnSlots[0][19] = %+v, want %+v", got, want)
	}

	// Level 0, slot 32: original byte 0x08 = rock variant 0, X=0x10.
	got = SpawnSlots[0][32]
	want = SpawnSlot{IsRock: true, RockVariant: 0, X: 0x10}

	if got != want {
		t.Errorf("SpawnSlots[0][32] = %+v, want %+v", got, want)
	}
}
