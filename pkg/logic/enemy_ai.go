package logic

import (
	"github.com/morozov/river-raid-ebiten/pkg/domain"
	"github.com/morozov/river-raid-ebiten/pkg/platform"
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
	terrainCheckAheadL = 16
	terrainCheckAheadR = 32
	balloonTickMask    = 3
	balloonTickMatch   = 1
	evenTickMask       = 1
)

// MoveEnemies updates all activated enemy positions based on their type-specific AI.
func MoveEnemies(vp *state.Viewport) {
	for i := range vp.Slots {
		slot := &vp.Slots[i]
		if !slot.Activated {
			continue
		}

		switch slot.Type {
		case domain.ObjectHelicopterReg, domain.ObjectHelicopterAdv, domain.ObjectShip:
			moveShipOrHelicopter(slot, vp.Tick)
		case domain.ObjectFighter:
			moveFighter(slot)
		case domain.ObjectTank:
			moveTank(slot, vp.Tick)
		case domain.ObjectBalloon:
			moveBalloon(slot, vp.Tick)
		case domain.ObjectFuel:
			// Fuel depots are static.
		}
	}
}

// moveShipOrHelicopter moves 2px on even ticks, bouncing off screen edges.
func moveShipOrHelicopter(slot *domain.Slot, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if slot.Orientation == domain.OrientationLeft {
		slot.X -= EnemyMoveStep
		if slot.X <= terrainCheckAheadL {
			slot.Orientation = domain.OrientationRight
		}
	} else {
		slot.X += EnemyMoveStep
		if slot.X >= platform.ScreenWidth-terrainCheckAheadR {
			slot.Orientation = domain.OrientationLeft
		}
	}
}

// moveFighter moves 4px every frame, wrapping at screen edges.
func moveFighter(slot *domain.Slot) {
	if slot.Orientation == domain.OrientationLeft {
		slot.X -= fighterMoveStep
		if slot.X <= fighterWrapLeftX {
			slot.X = fighterResetLeftX
		}
	} else {
		slot.X += fighterMoveStep
		if slot.X >= fighterWrapRightX {
			slot.X = fighterResetRightX
		}
	}
}

// moveTank moves 2px on even ticks (road tanks only for now).
func moveTank(slot *domain.Slot, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if slot.Orientation == domain.OrientationLeft {
		slot.X -= EnemyMoveStep
	} else {
		slot.X += EnemyMoveStep
	}
}

// moveBalloon moves 2px every 4th frame, bouncing off screen edges.
func moveBalloon(slot *domain.Slot, tick int) {
	if tick&balloonTickMask != balloonTickMatch {
		return
	}

	if slot.Orientation == domain.OrientationLeft {
		slot.X -= EnemyMoveStep
		if slot.X <= terrainCheckAheadL {
			slot.Orientation = domain.OrientationRight
		}
	} else {
		slot.X += EnemyMoveStep
		if slot.X >= platform.ScreenWidth-terrainCheckAheadR {
			slot.Orientation = domain.OrientationLeft
		}
	}
}
