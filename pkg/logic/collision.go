package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Collision bounding box dimensions and explosion fragment layout per object type.
type collisionProfile struct {
	fragments []explosionFragmentOffset // relative (dX, dY) offsets for explosion fragment spawning
	width     int
	height    int
	points    int
}

// explosionFragmentOffset is a relative (dX, dY) pixel offset used when spawning
// explosion fragments from a destroyed object's position.
type explosionFragmentOffset struct {
	x domain.Px
	y domain.Px
}

// Explosion fragment layout constants.
// Each explosion fragment sprite is 16×8 px; offsets are multiples of the fragment height (8).
const (
	shipFragLateralOff domain.Px = 8 // X offset for ship's second fragment (one tile right)
)

// collisionProfiles maps each object type to its collision bounding box and hit outcome.
var collisionProfiles = map[domain.ObjectType]collisionProfile{
	domain.ObjectHelicopterReg: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}},
		width:     assets.SpriteHelicopterWidth,
		height:    assets.SpriteHelicopterHeight,
		points:    PointsHelicopterReg,
	},
	domain.ObjectHelicopterAdv: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}},
		width:     assets.SpriteHelicopterWidth,
		height:    assets.SpriteHelicopterHeight,
		points:    PointsHelicopterAdv,
	},
	domain.ObjectShip: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}, {x: shipFragLateralOff, y: 0}},
		width:     assets.SpriteShipWidth,
		height:    assets.SpriteShipHeight,
		points:    PointsShip,
	},
	domain.ObjectFighter: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}},
		width:     assets.SpriteFighterWidth,
		height:    assets.SpriteFighterHeight,
		points:    PointsFighter,
	},
	domain.ObjectBalloon: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}, {x: 0, y: assets.SpriteExplosionHeight}},
		width:     assets.SpriteBalloonWidth,
		height:    assets.SpriteBalloonHeight,
		points:    PointsBalloon,
	},
	domain.ObjectFuel: {
		fragments: []explosionFragmentOffset{{x: 0, y: 0}, {x: 0, y: assets.SpriteExplosionHeight}, {x: 0, y: assets.SpriteExplosionHeight * 2}},
		width:     assets.SpriteFuelWidth,
		height:    assets.SpriteFuelHeight,
		points:    PointsFuel,
	},
}

// Plane dimensions.
const (
	planeWidth  domain.Px = assets.SpritePlayerWidth
	planeHeight domain.Px = assets.SpritePlayerHeight
)

// Bridge dimensions and explosion layout.
const (
	bridgeVerticalExtent domain.Px = 22 // vertical height of the bridge in pixels

	// Horizontal span of the bridge structure: the 32-pixel band over the river,
	// between the road surfaces on either bank.
	bridgeLeftX  domain.Px = 0x70 // left edge of the bridge structure
	bridgeRightX domain.Px = 0x90 // one past the right edge of the bridge structure

	// Bridge explosion fragment X positions (fixed, independent of bridge X).
	bridgeFragX0 domain.SP = 0x70 * domain.SubpixelScale // the left column of the 2×3 grid
	bridgeFragX1 domain.SP = 0x80 * domain.SubpixelScale // the right column of the 2×3 grid

	// Bridge explosion fragment Y offsets relative to bridgeY (bottom of bridge).
	bridgeFragRow0 domain.SP = 4 * domain.SubpixelScale  // bottom row: bridgeY - 4
	bridgeFragRow1 domain.SP = 12 * domain.SubpixelScale // middle row: bridgeY - 12
	bridgeFragRow2 domain.SP = 20 * domain.SubpixelScale // top row:    bridgeY - 20
)

// striker is an entity that can hit the bridge and viewport objects.
// It provides its bounding box and reports whether a given object type is a valid target.
// Both the player plane and the player missile implement this interface.
type striker interface {
	bounds() (x, y, w, h domain.Px)
	canHit(domain.ObjectType) bool
}

// playerPlane is the player's aircraft as a striker.
type playerPlane struct {
	x domain.Px
}

func (p playerPlane) bounds() (x, y, w, h domain.Px) {
	return p.x, domain.PlaneY, planeWidth, planeHeight
}

// canHit returns true for all profiled objects except fuel depots; fuel is
// checked separately via checkFuelOverlap before the bridge/object scan.
func (p playerPlane) canHit(t domain.ObjectType) bool {
	if t == domain.ObjectFuel {
		return false
	}

	_, ok := collisionProfiles[t]
	return ok
}

// playerMissile is the player's missile as a striker.
type playerMissile struct {
	x, y domain.Px
}

func (m playerMissile) bounds() (x, y, w, h domain.Px) {
	return m.x, m.y, assets.SpritePlayerMissileWidth, assets.SpritePlayerMissileHeight
}

// canHit returns true for all profiled objects.
func (m playerMissile) canHit(t domain.ObjectType) bool {
	_, ok := collisionProfiles[t]
	return ok
}

// hitResult is the outcome of a striker hitting a target.
type hitResult struct {
	explosionFragments []state.ExplosionFragment
	objectIdx          int // index into vp.Objects; -1 for bridge hits
	points             int
}

// target is something a striker can hit: either the bridge or the viewport objects.
// checkHit detects a hit and applies any side effects of that hit to r.
type target interface {
	checkHit(s striker, r *CollisionResult) (hitResult, bool)
}

// Tank gap X bounds, per spec/07-enemies.md.
// A road tank is in the river gap when X+10 >= $70 and X <= $90.
const (
	tankGapLeftEdge  domain.Px = 0x70 // X+10 must be >= this to be in the gap
	tankGapRightEdge domain.Px = 0x90 // X must be <= this to be in the gap
	tankGapProbe     domain.Px = 10   // added to X before comparing with left edge
	bridgeEarlyLevel           = 7    // bridge index threshold: <= this → freeze tank; > this → bank-tank
)

// bridgeTarget references bridge state and checks the bridge on each hit test.
type bridgeTarget struct {
	vp          *state.Viewport
	y           domain.SP
	bridgeIndex int
	active      bool
	destroyed   bool
}

func (b bridgeTarget) checkHit(s striker, r *CollisionResult) (hitResult, bool) {
	if !b.active || b.destroyed {
		return hitResult{}, false
	}

	px, py, pw, _ := s.bounds()
	bridgeYPx := b.y.ToPx()
	bridgeTop := bridgeYPx - bridgeVerticalExtent

	// Vertically the striker meets the bridge at one row, its render origin, tested
	// against the half-open band (bridgeTop, bridgeY].
	if py > bridgeYPx || py <= bridgeTop {
		return hitResult{}, false
	}

	if px+pw <= bridgeLeftX || px >= bridgeRightX {
		return hitResult{}, false
	}

	b.onHit(r)

	return hitResult{
		objectIdx:          -1,
		points:             PointsBridge,
		explosionFragments: bridgeExplosionFragments(b.y),
	}, true
}

// onHit runs the frozen road-tank gap check. For each road tank:
//   - If in the river gap (X+10 >= $70 and X <= $90): destroy it, award 250 pts, spawn 1 fragment.
//   - Otherwise: freeze in place (bridge ≤ 7) or convert to bank-tank (bridge > 7).
func (b bridgeTarget) onHit(r *CollisionResult) {
	for i, obj := range b.vp.Objects {
		if obj.Type != domain.ObjectTank || obj.TankLocation != domain.TankLocationRoad {
			continue
		}

		// Shift SP coordinate to screen pixels for the gap boundary check.
		objXPx := obj.X.ToPx()
		if objXPx+tankGapProbe >= tankGapLeftEdge && objXPx <= tankGapRightEdge {
			// Tank is over the river gap: destroy it.
			r.DestroyObjects = append(r.DestroyObjects, i)
			r.ExplosionFragments = append(r.ExplosionFragments, state.ExplosionFragment{X: obj.X, Y: obj.Y})
			r.PointsScored += PointsTank
		} else if b.bridgeIndex > bridgeEarlyLevel {
			// Tank is on the bank, late level: convert to bank-tank behavior.
			convertToBankTank(obj)
		}
		// Early level: tank remains frozen (moveTank skips road tanks
		// while bridgeDestroyed is true).
	}
}

// viewportObjectsTarget references the viewport and iterates its objects on each hit test.
type viewportObjectsTarget struct {
	vp *state.Viewport
}

func (v viewportObjectsTarget) checkHit(s striker, _ *CollisionResult) (hitResult, bool) {
	px, py, pw, ph := s.bounds()

	for i, obj := range v.vp.Objects {
		if !s.canHit(obj.Type) {
			continue
		}

		profile := collisionProfiles[obj.Type]

		objXPx := obj.X.ToPx()
		objYPx := obj.Y.ToPx()
		if !boxOverlap(px, py, pw, ph, objXPx, objYPx, domain.Px(profile.width), domain.Px(profile.height)) {
			continue
		}

		var frags []state.ExplosionFragment
		for _, off := range profile.fragments {
			frags = append(frags, state.ExplosionFragment{
				X: obj.X + off.x.ToSP(),
				Y: obj.Y + off.y.ToSP(),
			})
		}

		return hitResult{objectIdx: i, points: profile.points, explosionFragments: frags}, true
	}

	return hitResult{}, false
}

// CollisionResult describes what happened during collision checks.
type CollisionResult struct {
	DestroyObjects     []int // indices of viewport objects to remove
	ExplosionFragments []state.ExplosionFragment
	PointsScored       int
	PlayerDied         bool
	Refueling          bool
	BridgeHit          bool
}

// applyStrike folds the striker's hit into r and reports whether it hit anything. A
// strike is spent on the target it lands on — the missile is destroyed by it and the
// plane crashes into it — so a striker reaches one target however many its box
// overlaps. Side effects of the hit are applied to r.
func applyStrike(s striker, targets []target, r *CollisionResult) bool {
	for _, t := range targets {
		if hit, ok := t.checkHit(s, r); ok {
			r.applyHit(hit)

			return true
		}
	}

	return false
}

// checkFuelOverlap returns true and the object index when the plane overlaps a
// fuel depot. FuelState is not a valid striker target (playerPlane.canHit returns
// false for ObjectFuel), so it must be checked separately.
func checkFuelOverlap(plane playerPlane, vp *state.Viewport) bool {
	px, py, pw, ph := plane.bounds()

	for _, obj := range vp.Objects {
		if obj.Type != domain.ObjectFuel {
			continue
		}

		profile := collisionProfiles[domain.ObjectFuel]

		if boxOverlap(px, py, pw, ph, obj.X.ToPx(), obj.Y.ToPx(), domain.Px(profile.width), domain.Px(profile.height)) {
			return true
		}
	}

	return false
}

// CheckCollisions runs the full per-frame collision sequence.
func CheckCollisions(
	planeX domain.SP,
	missile *state.PlayerMissile,
	heliMissile *state.HeliMissile,
	vp *state.Viewport,
	terrainEdges TerrainEdgesFunc,
	bridgeActive bool,
	bridgeY domain.SP,
	bridgeDestroyed bool,
	bridgeIndex int,
) CollisionResult {
	var result CollisionResult

	planeXPx := planeX.ToPx()
	plane := playerPlane{x: planeXPx}

	// 1. Plane vs. fuel depot (refueling; does not kill the player).
	result.Refueling = checkFuelOverlap(plane, vp)

	// The bridge comes first so its road-tank side effects run before the object scan
	// walks the same slots.
	bt := bridgeTarget{
		active:      bridgeActive,
		y:           bridgeY,
		destroyed:   bridgeDestroyed,
		vp:          vp,
		bridgeIndex: bridgeIndex,
	}
	targets := [2]target{bt, viewportObjectsTarget{vp: vp}}

	// 2. Plane vs. bridge and objects, else vs. terrain. Terrain is the fall-through:
	// a striker that overlaps something no target claims is touching land. A plane
	// straddling the road beside a bridge therefore takes the bridge with it.
	//
	// Refueling is not a strike and costs the plane nothing, so it does not shield the
	// plane from the bank under it: the plane refuels and dies.
	if applyStrike(plane, targets[:], &result) {
		result.PlayerDied = true
	} else if planeHitsTerrain(planeXPx, terrainEdges) {
		result.PlayerDied = true
	}

	// The plane's collision resolves ahead of the missile's and the helicopter missile's,
	// and death ends the frame where it happens.
	if result.PlayerDied {
		return result
	}

	// 3. Missile vs. bridge and objects, else vs. terrain.
	if missile.Active {
		m := playerMissile{x: missile.X.ToPx(), y: missile.Y.ToPx()}

		if applyStrike(m, targets[:], &result) {
			missile.Active = false
		} else if missileHitsTerrain(m, terrainEdges) {
			missile.Active = false
		}
	}

	// 4. Helicopter missile vs. plane.
	if heliMissileHitsPlane(heliMissile, planeXPx) {
		result.PlayerDied = true
	}

	return result
}

// applyHit folds a hitResult into the CollisionResult.
func (r *CollisionResult) applyHit(hit hitResult) {
	r.PointsScored += hit.points
	r.ExplosionFragments = append(r.ExplosionFragments, hit.explosionFragments...)

	if hit.objectIdx >= 0 {
		r.DestroyObjects = append(r.DestroyObjects, hit.objectIdx)
	} else {
		r.BridgeHit = true
	}
}

// TerrainEdgesFunc reports the navigable river edges at viewport row y for a sprite
// positioned at x. The x argument selects the shoulder when an island splits the river.
type TerrainEdgesFunc func(x domain.Px, y int) (leftX, rightX domain.Px)

// planeHitsTerrain returns true if any row of the plane overlaps a terrain bank.
func planeHitsTerrain(planeX domain.Px, terrainEdges TerrainEdgesFunc) bool {
	return spriteHitsTerrain(planeX, domain.PlaneY, planeWidth, planeHeight, terrainEdges)
}

// missileHitsTerrain returns true if any set-pixel row of the missile overlaps a terrain
// bank. The missile is destroyed by terrain just as the plane is, which is what keeps it
// from reaching a target it has no clear line of flight to. Only the opaque rows count:
// the frame's blank trailing rows collide with nothing.
func missileHitsTerrain(m playerMissile, terrainEdges TerrainEdgesFunc) bool {
	x, y, w, _ := m.bounds()
	return spriteHitsTerrain(x, int(y), w, assets.SpritePlayerMissileOpaqueHeight, terrainEdges)
}

// spriteHitsTerrain returns true if any row of a sprite box overlaps a terrain bank.
func spriteHitsTerrain(x domain.Px, topY int, w, h domain.Px, terrainEdges TerrainEdgesFunc) bool {
	for row := range h {
		leftX, rightX := terrainEdges(x, topY+int(row))
		if x < leftX || x+w > rightX {
			return true
		}
	}

	return false
}

// heliMissileHitsPlane returns true if the helicopter missile overlaps the player plane.
// planeXPx is the plane X in screen pixels.
func heliMissileHitsPlane(heliMissile *state.HeliMissile, planeXPx domain.Px) bool {
	if !heliMissile.Active {
		return false
	}

	hx := heliMissile.X.ToPx()
	hy := heliMissile.Y.ToPx()
	dx := hx - planeXPx
	return dx >= -1 && dx <= planeWidth &&
		hy >= domain.PlaneY && hy < domain.PlaneY+planeHeight
}

// bridgeExplosionFragments returns the 6 explosion fragments for a destroyed bridge.
func bridgeExplosionFragments(bridgeY domain.SP) []state.ExplosionFragment {
	var frags []state.ExplosionFragment
	for _, row := range [3]domain.SP{bridgeFragRow0, bridgeFragRow1, bridgeFragRow2} {
		y := bridgeY - row
		frags = append(frags,
			state.ExplosionFragment{X: bridgeFragX0, Y: y},
			state.ExplosionFragment{X: bridgeFragX1, Y: y},
		)
	}

	return frags
}

// boxOverlap returns true if two axis-aligned bounding boxes overlap.
func boxOverlap(ax, ay, aw, ah, bx, by, bw, bh domain.Px) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
