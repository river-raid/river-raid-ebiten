package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Collision bounding box dimensions and explosion fragment layout per object type.
type objectBounds struct {
	Fragments []explosionFragmentOffset // relative (dX, dY) offsets for explosion fragment spawning
	Width     int
	Height    int
	Points    int
}

// explosionFragmentOffset is a relative (dX, dY) pixel offset used when spawning
// explosion fragments from a destroyed object's position.
type explosionFragmentOffset struct {
	X int
	Y int
}

// Explosion fragment layout constants.
// Each explosion fragment sprite is 16×8 px; offsets are multiples of the fragment height (8)
// or the intentional Z80 INC-B quirk offset (17) for the fuel depot.
const (
	fragRow1Offset     = 8  // vertical offset of the first (0-base) row of explosion fragments
	fragRow2Offset     = 16 // vertical offset of the second (0-base) row of explosion fragments
	shipFragLateralOff = 8  // X offset for ship's second fragment (one tile right)
	shipFragVertOff    = 4  // Y offset for ship's third fragment (half a fragment down)
	fuelFragRow3Offset = 17 // vertical offset of the third (0-base) row of fuel depot explosion fragments
)

// Explosion fragment offsets per object type.
var objectBoundsTable = map[domain.ObjectType]objectBounds{ //nolint:gochecknoglobals // constant table
	domain.ObjectHelicopterReg: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}},
		Width:     assets.SpriteHelicopterWidth,
		Height:    assets.SpriteHelicopterHeight,
		Points:    PointsHelicopterReg,
	},
	domain.ObjectHelicopterAdv: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}},
		Width:     assets.SpriteHelicopterWidth,
		Height:    assets.SpriteHelicopterHeight,
		Points:    PointsHelicopterAdv,
	},
	domain.ObjectShip: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}, {X: shipFragLateralOff, Y: 0}, {X: 0, Y: shipFragVertOff}},
		Width:     assets.SpriteShipWidth,
		Height:    assets.SpriteShipHeight,
		Points:    PointsShip,
	},
	domain.ObjectFighter: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}},
		Width:     assets.SpriteFighterWidth,
		Height:    assets.SpriteFighterHeight,
		Points:    PointsFighter,
	},
	domain.ObjectBalloon: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}, {X: 0, Y: fragRow1Offset}},
		Width:     assets.SpriteBalloonWidth,
		Height:    assets.SpriteBalloonHeight,
		Points:    PointsBalloon,
	},
	domain.ObjectFuel: {
		Fragments: []explosionFragmentOffset{{X: 0, Y: 0}, {X: 0, Y: fragRow1Offset}, {X: 0, Y: fragRow2Offset}, {X: 0, Y: fuelFragRow3Offset}},
		Width:     assets.SpriteFuelWidth,
		Height:    assets.SpriteFuelHeight,
		Points:    PointsFuel,
	},
}

// Plane dimensions.
const (
	planeWidth  = assets.SpritePlayerWidth
	planeHeight = assets.SpritePlayerHeight
)

// Bridge dimensions and explosion layout.
const (
	bridgeVerticalExtent = 22 // vertical height of the bridge in pixels

	// Bridge explosion fragment X positions (fixed, independent of bridge X).
	bridgeFragX0 = 0x70 // left column of the 2×3 grid
	bridgeFragX1 = 0x80 // right column of the 2×3 grid

	// Bridge explosion fragment Y offsets relative to bridgeY (bottom of bridge).
	bridgeFragRow0 = 4  // bottom row: bridgeY - 4
	bridgeFragRow1 = 12 // middle row: bridgeY - 12
	bridgeFragRow2 = 20 // top row:    bridgeY - 20
)

// CollisionResult describes what happened during collision checks.
type CollisionResult struct {
	DestroyObjects     []int // indices of viewport objects to remove
	ExplosionFragments []state.ExplodingFragment
	PointsScored       int
	PlayerDied         bool
	Refueling          bool
	BridgeHit          bool
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
	for i := range vp.Objects {
		obj := vp.Objects[i]
		bounds, ok := objectBoundsTable[obj.Type]

		if !ok {
			continue
		}

		if !boxOverlap(planeX, domain.PlaneY, planeWidth, planeHeight,
			obj.X, obj.Y, bounds.Width, bounds.Height) {
			continue
		}

		if obj.Type != domain.ObjectFuel {
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

				// Spawn 6 explosion fragments in a 2×3 grid.
				for _, row := range [3]int{bridgeFragRow0, bridgeFragRow1, bridgeFragRow2} {
					y := bridgeY - row
					result.ExplosionFragments = append(result.ExplosionFragments,
						state.ExplodingFragment{X: bridgeFragX0, Y: y, Frame: 1},
						state.ExplodingFragment{X: bridgeFragX1, Y: y, Frame: 1},
					)
				}
			}
		}

		// PlayerMissile vs viewport objects.
		if missile.Active {
			for i := range vp.Objects {
				obj := vp.Objects[i]
				bounds, ok := objectBoundsTable[obj.Type]

				if !ok {
					continue
				}

				if boxOverlap(missile.X, missile.Y, assets.SpritePlayerMissileWidth, assets.SpritePlayerMissileHeight,
					obj.X, obj.Y, bounds.Width, bounds.Height) {
					result.PointsScored += bounds.Points
					result.DestroyObjects = append(result.DestroyObjects, i)
					missile.Active = false

					for _, off := range bounds.Fragments {
						result.ExplosionFragments = append(result.ExplosionFragments, state.ExplodingFragment{
							X:     obj.X + off.X,
							Y:     obj.Y + off.Y,
							Frame: 1,
						})
					}

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
