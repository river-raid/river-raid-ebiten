package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

func TestScrollState_NextFragment_AdvancesWithinLevel(t *testing.T) {
	t.Parallel()

	s := ScrollState{}
	_ = s.NextFragment()

	if s.FragmentNum != 1 {
		t.Errorf("FragmentNum = %d, want 1", s.FragmentNum)
	}

	if s.BridgeIndex != 0 {
		t.Errorf("BridgeIndex = %d, want 0", s.BridgeIndex)
	}
}

func TestScrollState_NextFragment_AdvancesToNextLevel(t *testing.T) {
	t.Parallel()

	s := ScrollState{FragmentNum: domain.NumFragmentsPerLevel - 1}
	_ = s.NextFragment()

	if s.FragmentNum != 0 {
		t.Errorf("FragmentNum = %d, want 0", s.FragmentNum)
	}

	if s.BridgeIndex != 1 {
		t.Errorf("BridgeIndex = %d, want 1", s.BridgeIndex)
	}
}

func TestScrollState_NextFragment_WrapsAfterLevel47(t *testing.T) {
	t.Parallel()

	s := ScrollState{BridgeIndex: domain.NumLevels - 1, FragmentNum: domain.NumFragmentsPerLevel - 1}
	_ = s.NextFragment()

	if s.BridgeIndex < bridgeLoopStart || s.BridgeIndex >= bridgeLoopStart+bridgeLoopLength {
		t.Errorf("BridgeIndex = %d, want in range [%d, %d)",
			s.BridgeIndex, bridgeLoopStart, bridgeLoopStart+bridgeLoopLength)
	}
}
