package logic

import (
	"testing"

	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
)

func TestShellSpawnX(t *testing.T) {
	t.Parallel()

	const tankX domain.SP = 100 * domain.SubpixelScale

	tests := []struct {
		name   string
		orient domain.Orientation
		wantPx domain.Px
	}{
		{
			name:   "left-facing: 16px left of left edge",
			orient: domain.OrientationLeft,
			wantPx: 100 - shellSpawnOffset,
		},
		{
			name:   "right-facing: 16px right of right edge",
			orient: domain.OrientationRight,
			wantPx: 100 + assets.SpriteTankWidth + shellSpawnOffset,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := shellSpawnX(tankX, tc.orient)
			want := tc.wantPx.ToSP()

			if got != want {
				t.Errorf("ShellSpawnX(%d, %v) = %d, want %d", tankX, tc.orient, got, want)
			}
		})
	}
}
