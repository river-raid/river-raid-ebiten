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
	// boundaryPadding defines the horizontal padding between the river banks
	// and the bounding box of moving objects in pixels
	boundaryPadding = 8
)

// TerrainBuffer is an interface for querying terrain edges.
// This allows us to avoid importing the render package directly.
type TerrainBuffer interface {
	GetEdges(x, y, spriteHeight int) (leftX, rightX int)
}

// initializeObjectBoundaries calculates movement boundaries for a newly spawned object.
// This is called once at spawn time to set MinX/MaxX based on terrain.
// scrollY is the current scroll position in the terrain buffer.
func initializeObjectBoundaries(obj *state.ViewportObject, terrain TerrainBuffer, scrollY int) {
	// Skip rocks and fuel depots (they don't move).
	if obj.IsRock || obj.Type == domain.ObjectFuel {
		return
	}

	// Get sprite dimensions for this enemy type.
	sprite := assets.SpriteObjects[obj.Type]
	spriteWidth := sprite.Width
	spriteHeight := sprite.Height()

	// Query terrain edges across all scanlines the sprite overlaps.
	// Pass enemy's spawn X position to determine which shoulder it's in.
	leftEdge, rightEdge := terrain.GetEdges(obj.X, scrollY, spriteHeight)

	obj.MinX = leftEdge + boundaryPadding
	obj.MaxX = rightEdge - spriteWidth - boundaryPadding
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
