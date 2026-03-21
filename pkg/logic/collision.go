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
	x int
	y int
}

// Explosion fragment layout constants.
// Each explosion fragment sprite is 16×8 px; offsets are multiples of the fragment height (8).
const (
	shipFragLateralOff = 8 // X offset for ship's second fragment (one tile right)
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
	planeWidth  = assets.SpritePlayerWidth
	planeHeight = assets.SpritePlayerHeight
)

// Bridge dimensions and explosion layout.
const (
	bridgeVerticalExtent = 22 // vertical height of the bridge in pixels

	// Bridge explosion fragment X positions (fixed, independent of bridge X).
	bridgeFragX0 = 0x70 // the left column of the 2×3 grid
	bridgeFragX1 = 0x80 // the right column of the 2×3 grid

	// Bridge explosion fragment Y offsets relative to bridgeY (bottom of bridge).
	bridgeFragRow0 = 4  // bottom row: bridgeY - 4
	bridgeFragRow1 = 12 // middle row: bridgeY - 12
	bridgeFragRow2 = 20 // top row:    bridgeY - 20
)

// striker is an entity that can hit the bridge and viewport objects.
// It provides its bounding box and reports whether a given object type is a valid target.
// Both the player plane and the player missile implement this interface.
type striker interface {
	bounds() (x, y, w, h int)
	canHit(domain.ObjectType) bool
}

// playerPlane is the player's aircraft as a striker.
type playerPlane struct {
	x int
}

func (p playerPlane) bounds() (x, y, w, h int) {
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
	x, y int
}

func (m playerMissile) bounds() (x, y, w, h int) {
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
	tankGapLeftEdge  = 0x70 // X+10 must be >= this to be in the gap
	tankGapRightEdge = 0x90 // X must be <= this to be in the gap
	tankGapProbe     = 10   // added to X before comparing with left edge
	bridgeEarlyLevel = 7    // bridge index threshold: <= this → freeze tank; > this → bank-tank
)

// bridgeTarget references bridge state and checks the bridge on each hit test.
type bridgeTarget struct {
	vp          *state.Viewport
	y           int
	bridgeIndex int
	active      bool
	destroyed   bool
}

func (b bridgeTarget) checkHit(s striker, r *CollisionResult) (hitResult, bool) {
	if !b.active || b.destroyed {
		return hitResult{}, false
	}

	_, py, _, ph := s.bounds()
	bridgeTop := b.y - bridgeVerticalExtent

	if py+ph <= bridgeTop || py >= b.y {
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

		if obj.X+tankGapProbe >= tankGapLeftEdge && obj.X <= tankGapRightEdge {
			// Tank is over the river gap: destroy it.
			r.DestroyObjects = append(r.DestroyObjects, i)
			r.ExplosionFragments = append(r.ExplosionFragments, state.ExplosionFragment{X: obj.X, Y: obj.Y})
			r.PointsScored += PointsTank
		} else if b.bridgeIndex > bridgeEarlyLevel {
			// Tank is on the bank, late level: convert to bank-tank behavior.
			obj.TankLocation = domain.TankLocationBank
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

		if !boxOverlap(px, py, pw, ph, obj.X, obj.Y, profile.width, profile.height) {
			continue
		}

		var frags []state.ExplosionFragment
		for _, off := range profile.fragments {
			frags = append(frags, state.ExplosionFragment{X: obj.X + off.x, Y: obj.Y + off.y})
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

// checkFirstHit returns the first target in targets that striker s hits.
// Side effects of the hit are applied to r. ok is false when nothing is hit.
func checkFirstHit(s striker, targets []target, r *CollisionResult) (hitResult, bool) {
	for _, t := range targets {
		if hit, ok := t.checkHit(s, r); ok {
			return hit, true
		}
	}

	return hitResult{}, false
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

		if boxOverlap(px, py, pw, ph, obj.X, obj.Y, profile.width, profile.height) {
			return true
		}
	}

	return false
}

// CheckCollisions runs the full per-frame collision sequence.
func CheckCollisions(
	planeX int,
	missile *state.PlayerMissile,
	heliMissile *state.HeliMissile,
	vp *state.Viewport,
	terrainLeftX, terrainRightX func(y int) int,
	bridgeActive bool,
	bridgeY int,
	bridgeDestroyed bool,
	bridgeIndex int,
) CollisionResult {
	var result CollisionResult

	// 1. Plane vs. terrain.
	if planeHitsTerrain(planeX, terrainLeftX, terrainRightX) {
		result.PlayerDied = true
		return result
	}

	plane := playerPlane{x: planeX}

	// 2. Plane vs. fuel depot (refueling; does not kill the player).
	result.Refueling = checkFuelOverlap(plane, vp)

	// Bridge must be checked before viewport objects. Road tanks occupy the same X span
	// as the bridge, so any striker that reaches a tank's position would have already
	// hit the bridge.
	bt := bridgeTarget{
		active:      bridgeActive,
		y:           bridgeY,
		destroyed:   bridgeDestroyed,
		vp:          vp,
		bridgeIndex: bridgeIndex,
	}
	targets := [2]target{bt, viewportObjectsTarget{vp: vp}}

	// 3. Plane vs. bridge and objects.
	if hit, ok := checkFirstHit(plane, targets[:], &result); ok {
		result.applyHit(hit)
		result.PlayerDied = true
	}

	// 4. Missile vs. bridge and objects.
	if missile.Active {
		m := playerMissile{x: missile.X, y: missile.Y}

		if hit, ok := checkFirstHit(m, targets[:], &result); ok {
			result.applyHit(hit)
			missile.Active = false
		}
	}

	// 5. Helicopter missile vs. plane.
	if heliMissileHitsPlane(heliMissile, planeX) {
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

// planeHitsTerrain returns true if any row of the plane overlaps a terrain bank.
func planeHitsTerrain(planeX int, terrainLeftX, terrainRightX func(y int) int) bool {
	for row := range planeHeight {
		y := domain.PlaneY + row
		if planeX < terrainLeftX(y) || planeX+planeWidth > terrainRightX(y) {
			return true
		}
	}

	return false
}

// heliMissileHitsPlane returns true if the helicopter missile overlaps the player plane.
func heliMissileHitsPlane(heliMissile *state.HeliMissile, planeX int) bool {
	if !heliMissile.Active {
		return false
	}

	dx := heliMissile.X - planeX
	return dx >= -1 && dx <= planeWidth &&
		heliMissile.Y >= domain.PlaneY && heliMissile.Y < domain.PlaneY+planeHeight
}

// bridgeExplosionFragments returns the 6 explosion fragments for a destroyed bridge.
func bridgeExplosionFragments(bridgeY int) []state.ExplosionFragment {
	var frags []state.ExplosionFragment
	for _, row := range [3]int{bridgeFragRow0, bridgeFragRow1, bridgeFragRow2} {
		y := bridgeY - row
		frags = append(frags,
			state.ExplosionFragment{X: bridgeFragX0, Y: y},
			state.ExplosionFragment{X: bridgeFragX1, Y: y},
		)
	}

	return frags
}

// boxOverlap returns true if two axis-aligned bounding boxes overlap.
func boxOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
