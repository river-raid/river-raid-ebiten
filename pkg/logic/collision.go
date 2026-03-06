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

// canHit returns true for all profiled objects except tanks, which missiles pass through.
func (m playerMissile) canHit(t domain.ObjectType) bool {
	if t == domain.ObjectTank {
		return false
	}

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
type target interface {
	checkHit(s striker) (hitResult, bool)
}

// bridgeTarget references bridge state and checks the bridge on each hit test.
type bridgeTarget struct {
	y         int
	active    bool
	destroyed bool
}

func (b bridgeTarget) checkHit(s striker) (hitResult, bool) {
	if !b.active || b.destroyed {
		return hitResult{}, false
	}

	_, py, _, ph := s.bounds()
	bridgeTop := b.y - bridgeVerticalExtent

	if py+ph <= bridgeTop || py >= b.y {
		return hitResult{}, false
	}

	return hitResult{
		objectIdx:          -1,
		points:             PointsBridge,
		explosionFragments: bridgeExplosionFragments(b.y),
	}, true
}

// viewportObjectsTarget references the viewport and iterates its objects on each hit test.
type viewportObjectsTarget struct {
	vp *state.Viewport
}

func (v viewportObjectsTarget) checkHit(s striker) (hitResult, bool) {
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
// ok is false when nothing is hit.
func checkFirstHit(s striker, targets []target) (hitResult, bool) {
	for _, t := range targets {
		if hit, ok := t.checkHit(s); ok {
			return hit, true
		}
	}

	return hitResult{}, false
}

// checkFuelOverlap returns true and the object index when the plane overlaps a
// fuel depot. Fuel is not a valid striker target (playerPlane.canHit returns
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

	// Build the two static targets that are shared between the plane and missile checks.
	targets := [2]target{
		bridgeTarget{active: bridgeActive, y: bridgeY, destroyed: bridgeDestroyed},
		viewportObjectsTarget{vp: vp},
	}

	// 3. Plane vs. bridge and objects.
	if hit, ok := checkFirstHit(plane, targets[:]); ok {
		result.applyHit(hit)
		result.PlayerDied = true
		return result
	}

	// 4. Missile vs. bridge and objects.
	if missile.Active {
		m := playerMissile{x: missile.X, y: missile.Y}

		if hit, ok := checkFirstHit(m, targets[:]); ok {
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
