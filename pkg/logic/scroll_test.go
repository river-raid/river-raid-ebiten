package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

func TestNextFragment_AdvancesWithinLevel(t *testing.T) {
	t.Parallel()

	s := state.GameState{}
	_ = nextFragment(&s)

	if s.FragmentNum != 1 {
		t.Errorf("FragmentNum = %d, want 1", s.FragmentNum)
	}

	if s.BridgeIndex != 0 {
		t.Errorf("BridgeIndex = %d, want 0", s.BridgeIndex)
	}
}

func TestNextFragment_AdvancesToNextLevel(t *testing.T) {
	t.Parallel()

	s := state.GameState{FragmentNum: domain.NumFragmentsPerLevel - 1}
	_ = nextFragment(&s)

	if s.FragmentNum != 0 {
		t.Errorf("FragmentNum = %d, want 0", s.FragmentNum)
	}

	if s.BridgeIndex != 1 {
		t.Errorf("BridgeIndex = %d, want 1", s.BridgeIndex)
	}
}

func TestNextFragment_WrapsAfterLevel47(t *testing.T) {
	t.Parallel()

	s := state.GameState{BridgeIndex: domain.NumLevels - 1, FragmentNum: domain.NumFragmentsPerLevel - 1}
	_ = nextFragment(&s)

	if s.BridgeIndex < bridgeLoopStart || s.BridgeIndex >= bridgeLoopStart+bridgeLoopLength {
		t.Errorf("BridgeIndex = %d, want in range [%d, %d)",
			s.BridgeIndex, bridgeLoopStart, bridgeLoopStart+bridgeLoopLength)
	}
}

// bridgeFragmentIndex is the index within level 0 (bridge 1) of the RoadAndBridgeProfile fragment.
// assets.LevelTerrain[0][2] has ProfileIndex=2, which resolves to RoadAndBridgeProfile.
const bridgeFragmentIndex = 2

// TestNextFragment_SetsBridgeSectionForBridgeProfile verifies that BridgeSection is set to true
// and BridgeYPosition is computed correctly when a RoadAndBridgeProfile fragment is emitted.
func TestNextFragment_SetsBridgeSectionForBridgeProfile(t *testing.T) {
	t.Parallel()

	// Position GameState to emit the bridge fragment (index 2, level 0).
	// Use arbitrary non-zero ScrollY and NextRenderY so we can verify the position formula.
	const scrollY = 10
	const nextRenderY = -5

	s := state.GameState{
		BridgeIndex: 0,
		FragmentNum: bridgeFragmentIndex,
		ScrollY:     scrollY,
		NextRenderY: nextRenderY,
	}

	// Confirm the fragment at that index is indeed RoadAndBridgeProfile.
	frag := assets.LevelTerrain[0][bridgeFragmentIndex]
	if _, ok := assets.TerrainProfiles[frag.ProfileIndex].(assets.RoadAndBridgeProfile); !ok {
		t.Fatalf("test assumption broken: fragment %d is not RoadAndBridgeProfile", bridgeFragmentIndex)
	}

	_ = nextFragment(&s)

	if !s.BridgeSection {
		t.Error("BridgeSection = false, want true after RoadAndBridgeProfile fragment")
	}

	wantY := nextRenderY - scrollY + domain.NumLinesPerTerrainProfile
	if s.BridgeYPosition != wantY {
		t.Errorf("BridgeYPosition = %d, want %d", s.BridgeYPosition, wantY)
	}
}

// TestNextFragment_DoesNotClearBridgeSectionForNonBridgeProfile verifies that
// nextFragment does not clear BridgeSection when a non-bridge fragment is emitted;
// clearing is done by the scroll-off logic in updateViewportForScroll.
func TestNextFragment_DoesNotClearBridgeSectionForNonBridgeProfile(t *testing.T) {
	t.Parallel()

	// Fragment index 0 in level 0 has ProfileIndex=11 (RegularProfile).
	s := state.GameState{
		BridgeIndex:   0,
		FragmentNum:   0,
		BridgeSection: true, // previously set; must NOT be cleared by nextFragment
	}

	frag := assets.LevelTerrain[0][0]
	if _, ok := assets.TerrainProfiles[frag.ProfileIndex].(assets.RoadAndBridgeProfile); ok {
		t.Fatalf("test assumption broken: fragment 0 is unexpectedly a RoadAndBridgeProfile")
	}

	_ = nextFragment(&s)

	if !s.BridgeSection {
		t.Error("BridgeSection = false: nextFragment must not clear BridgeSection; scroll-off clears it")
	}
}

// TestUpdateViewportForScroll_ClearsBridgeSectionWhenScrolledOff verifies that
// BridgeSection is cleared once BridgeYPosition scrolls past TotalViewportHeight.
func TestUpdateViewportForScroll_ClearsBridgeSectionWhenScrolledOff(t *testing.T) {
	t.Parallel()

	s := &state.GameState{
		BridgeSection:   true,
		BridgeYPosition: domain.TotalViewportHeight - 1, // one pixel before edge
	}
	vp := state.NewViewport()
	// Set SpawnIndex to match advanceLines formula so spawnFromScroll is a no-op.
	vp.SpawnIndex = 0
	s.Viewport = vp
	s.Missile = &state.PlayerMissile{}
	s.TankShell = &state.TankShell{}
	s.HeliMissile = &state.HeliMissile{}

	// Advance by 2 pixels so BridgeYPosition crosses TotalViewportHeight.
	// spawnIdx matches s.Viewport.SpawnIndex so no spawn occurs (terrain=nil is safe).
	updateViewportForScroll(s, vp.SpawnIndex, 2, nil)

	if s.BridgeSection {
		t.Errorf("BridgeSection = true after bridge scrolled off viewport; want false (BridgeYPosition=%d)", s.BridgeYPosition)
	}
}

// TestUpdateViewportForScroll_ResetsBridgeDestroyedWhenScrolledOff verifies that
// BridgeDestroyed is cleared when the bridge scrolls off the viewport, so the next
// bridge is not pre-destroyed.
func TestUpdateViewportForScroll_ResetsBridgeDestroyedWhenScrolledOff(t *testing.T) {
	t.Parallel()

	s := &state.GameState{
		BridgeSection:   true,
		BridgeYPosition: domain.TotalViewportHeight - 1, // one pixel before edge
		BridgeDestroyed: true,
	}
	vp := state.NewViewport()
	vp.SpawnIndex = 0
	s.Viewport = vp
	s.Missile = &state.PlayerMissile{}
	s.TankShell = &state.TankShell{}
	s.HeliMissile = &state.HeliMissile{}

	updateViewportForScroll(s, vp.SpawnIndex, 2, nil)

	if s.BridgeDestroyed {
		t.Error("BridgeDestroyed = true after bridge scrolled off viewport; want false")
	}
}
