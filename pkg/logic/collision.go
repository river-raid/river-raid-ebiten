package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Collision bounding box dimensions per object type.
type objectBounds struct {
	Width      int
	Height     int
	Explosions int
	Points     int
}

var objectBoundsTable = map[domain.ObjectType]objectBounds{ //nolint:gochecknoglobals // constant table
	domain.ObjectHelicopterReg: {10, 8, 1, PointsHelicopterReg},
	domain.ObjectShip:          {19, 8, 3, PointsShip},
	domain.ObjectHelicopterAdv: {10, 8, 2, PointsHelicopterAdv},
	domain.ObjectTank:          {10, 8, 2, PointsTank},
	domain.ObjectFighter:       {10, 8, 2, PointsFighter},
	domain.ObjectBalloon:       {10, 17, 2, PointsBalloon},
	domain.ObjectFuel:          {10, 25, 2, PointsFuel},
}

// Plane dimensions.
const (
	planeWidth  = 8
	planeHeight = 8
)

// Bridge dimensions.
const (
	bridgeVerticalExtent = 22 // vertical height of the bridge in pixels
)

// CollisionResult describes what happened during collision checks.
type CollisionResult struct {
	DestroySlots []int // indices of viewport slots to remove
	PointsScored int
	PlayerDied   bool
	Refueling    bool
	BridgeHit    bool
}

// CheckCollisions runs the full per-frame collision sequence.
func CheckCollisions(
	planeX int,
	missile *state.PlayerMissile,
	heliMissile *state.HeliMissile,
	vp *state.Viewport,
	terrainLeftX, terrainRightX func(y int) int, // terrain edge lookups
	bridgeActive bool,
	bridgeY int,
	bridgeDestroyed bool,
) CollisionResult {
	var result CollisionResult

	// 1. Plane vs terrain edges.
	for row := range planeHeight {
		y := domain.PlaneY + row
		leftEdge := terrainLeftX(y)
		rightEdge := terrainRightX(y)

		if planeX < leftEdge || planeX+planeWidth > rightEdge {
			result.PlayerDied = true

			return result
		}
	}

	// 2. Plane vs bridge.
	if bridgeActive && !bridgeDestroyed {
		bridgeTop := bridgeY - bridgeVerticalExtent
		if domain.PlaneY+planeHeight > bridgeTop && domain.PlaneY < bridgeY {
			result.PlayerDied = true

			return result
		}
	}

	// 3. Plane vs viewport objects.
	for i := range vp.Slots {
		slot := &vp.Slots[i]
		bounds, ok := objectBoundsTable[slot.Type]

		if !ok {
			continue
		}

		if !boxOverlap(planeX, domain.PlaneY, planeWidth, planeHeight,
			slot.X, slot.Y, bounds.Width, bounds.Height) {
			continue
		}

		if slot.Type != domain.ObjectFuel {
			result.PlayerDied = true

			return result
		}

		result.Refueling = true
	}

	// 4. PlayerMissile vs bridge and viewport objects.
	if missile.Active {
		// PlayerMissile vs bridge.
		if bridgeActive && !bridgeDestroyed {
			bridgeTop := bridgeY - bridgeVerticalExtent
			if missile.Y >= bridgeTop && missile.Y < bridgeY {
				result.BridgeHit = true
				result.PointsScored += PointsBridge
				missile.Active = false
			}
		}

		// PlayerMissile vs viewport objects.
		if missile.Active {
			s := assets.SpritePlayerMissile
			missileW := s.Width
			missileH := s.Height()

			for i := range vp.Slots {
				slot := &vp.Slots[i]
				bounds, ok := objectBoundsTable[slot.Type]

				if !ok {
					continue
				}

				if boxOverlap(missile.X, missile.Y, missileW, missileH,
					slot.X, slot.Y, bounds.Width, bounds.Height) {
					result.PointsScored += bounds.Points
					result.DestroySlots = append(result.DestroySlots, i)
					missile.Active = false

					break
				}
			}
		}
	}

	// 5. Helicopter missile vs player.
	if heliMissile.Active {
		dx := heliMissile.X - planeX
		if dx >= -1 && dx <= planeWidth &&
			heliMissile.Y >= domain.PlaneY && heliMissile.Y < domain.PlaneY+planeHeight {
			result.PlayerDied = true

			return result
		}
	}

	return result
}

// boxOverlap returns true if two axis-aligned bounding boxes overlap.
func boxOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
