package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/assets"
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/state"
)

// Enemy movement constants.
const (
	EnemyMoveStep      = 2
	fighterMoveStep    = 4
	fighterWrapLeftX   = 0
	fighterWrapRightX  = 232
	fighterResetLeftX  = 232
	fighterResetRightX = 4
	tankFireX          = 128
	balloonTickMask    = 3
	balloonTickMatch   = 1
	evenTickMask       = 1
)

// Boundary calculation constants.
const (
	balloonProbeTopY    = 0   // Y offset for balloon top probe
	balloonProbeBottomY = 8   // Y offset for balloon bottom probe
	defaultProbeY       = 0   // Y offset for single-probe enemies
	screenMinX          = 0   // minimum X coordinate (left edge)
	screenMaxX          = 255 // maximum X coordinate (right edge)
)

// TerrainBuffer is an interface for querying terrain edges.
// This allows us to avoid importing the render package directly.
type TerrainBuffer interface {
	GetEdges(x, y int) (leftX, rightX int)
}

// getProbeYOffsets returns the Y offsets for terrain probe points for a given enemy type.
// Ships, helicopters, tanks, and fighters use a single probe point at Y+0.
// Balloons use two probe points (top and bottom) due to their 16px height.
func getProbeYOffsets(objectType domain.ObjectType) []int {
	switch objectType {
	case domain.ObjectBalloon:
		// Balloons are 16px tall (2 spawn slots), check top and bottom.
		return []int{balloonProbeTopY, balloonProbeBottomY}
	default:
		// All other enemies are 8px tall, single probe at Y+0.
		return []int{defaultProbeY}
	}
}

// InitializeEnemyBoundaries calculates movement boundaries for newly spawned enemies.
// This should be called after spawning enemies to set their MinX/MaxX based on terrain.
// Only calculates boundaries for enemies at Y=0 (newly spawned).
// scrollY is the current scroll position in the terrain buffer.
func InitializeEnemyBoundaries(vp *state.Viewport, terrain TerrainBuffer, scrollY int) {
	for i := range vp.Objects {
		obj := vp.Objects[i]

		// Only initialize boundaries for newly spawned objects (Y=0).
		// Once set, boundaries remain fixed as the object scrolls down.
		if obj.Y != 0 {
			continue
		}

		// Skip rocks and fuel depots (they don't move).
		if obj.IsRock || obj.Type == domain.ObjectFuel {
			continue
		}

		// Get sprite width and probe Y offsets for this enemy type.
		sprite := assets.SpriteObjects[obj.Type]
		spriteWidth := sprite.Width
		probeYOffsets := getProbeYOffsets(obj.Type)

		// Calculate movement boundaries based on terrain at spawn position.
		// The enemy spawns at buffer position scrollY and remains at that position.
		// As scrollY decreases, the viewport moves up, making the enemy appear to
		// scroll down, but the enemy's buffer Y position doesn't change.
		// Therefore, we only need to check the terrain at the spawn position.

		// Query terrain edges at the spawn position for each probe point.
		minX := screenMinX
		maxX := screenMaxX

		for _, yOffset := range probeYOffsets {
			// Calculate buffer Y position for this probe point.
			// Enemy is at buffer scrollY, probe at scrollY + yOffset.
			bufferY := scrollY + yOffset

			// Query terrain edges at the probe position.
			// Pass enemy's spawn X position to determine which shoulder it's in.
			leftEdge, rightEdge := terrain.GetEdges(obj.X, bufferY)

			// Adjust edges for sprite width.
			// leftEdge is the rightmost pixel of the left bank (first river pixel).
			// rightEdge is the leftmost pixel of the right bank (last river pixel + 1).
			// Just account for sprite width on the right side.
			adjustedLeft := leftEdge
			adjustedRight := rightEdge - spriteWidth

			// Use most restrictive bounds across all probe points.
			if adjustedLeft > minX {
				minX = adjustedLeft
			}
			if adjustedRight < maxX {
				maxX = adjustedRight
			}
		}

		obj.MinX = minX
		obj.MaxX = maxX
	}
}

// moveEnemies updates all activated enemy positions based on their type-specific AI.
func moveEnemies(vp *state.Viewport) {
	for i := range vp.Objects {
		obj := vp.Objects[i]
		if !obj.Activated {
			continue
		}

		switch obj.Type {
		case domain.ObjectHelicopterReg, domain.ObjectHelicopterAdv, domain.ObjectShip:
			moveShipOrHelicopter(obj, vp.Tick)
		case domain.ObjectFighter:
			moveFighter(obj)
		case domain.ObjectTank:
			moveTank(obj, vp.Tick)
		case domain.ObjectBalloon:
			moveBalloon(obj, vp.Tick)
		case domain.ObjectFuel:
			// Fuel depots are static.
		}
	}
}

// moveShipOrHelicopter moves 2px on even ticks, reversing at terrain boundaries.
func moveShipOrHelicopter(obj *state.ViewportObject, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if obj.Orientation == domain.OrientationLeft {
		obj.X -= EnemyMoveStep
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += EnemyMoveStep
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}

// moveFighter moves 4px every frame, wrapping at screen edges.
func moveFighter(obj *state.ViewportObject) {
	if obj.Orientation == domain.OrientationLeft {
		obj.X -= fighterMoveStep
		if obj.X <= fighterWrapLeftX {
			obj.X = fighterResetLeftX
		}
	} else {
		obj.X += fighterMoveStep
		if obj.X >= fighterWrapRightX {
			obj.X = fighterResetRightX
		}
	}
}

// moveTank moves 2px on even ticks (road tanks only for now).
func moveTank(obj *state.ViewportObject, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if obj.Orientation == domain.OrientationLeft {
		obj.X -= EnemyMoveStep
	} else {
		obj.X += EnemyMoveStep
	}
}

// moveBalloon moves 2px every 4th frame, reversing at terrain boundaries.
func moveBalloon(obj *state.ViewportObject, tick int) {
	if tick&balloonTickMask != balloonTickMatch {
		return
	}

	if obj.Orientation == domain.OrientationLeft {
		obj.X -= EnemyMoveStep
		if obj.X <= obj.MinX {
			obj.Orientation = domain.OrientationRight
		}
	} else {
		obj.X += EnemyMoveStep
		if obj.X >= obj.MaxX {
			obj.Orientation = domain.OrientationLeft
		}
	}
}
