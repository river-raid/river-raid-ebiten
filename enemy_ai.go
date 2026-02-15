package main

// Enemy movement constants.
const (
	enemyMoveStep      = 2
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
func MoveEnemies(vp *Viewport) {
	for i := range vp.Slots {
		slot := &vp.Slots[i]
		if !slot.Activated {
			continue
		}

		switch slot.Type {
		case ObjectHelicopterReg, ObjectHelicopterAdv, ObjectShip:
			moveShipOrHelicopter(slot, vp.Tick)
		case ObjectFighter:
			moveFighter(slot)
		case ObjectTank:
			moveTank(slot, vp.Tick)
		case ObjectBalloon:
			moveBalloon(slot, vp.Tick)
		case ObjectFuel:
			// Fuel depots are static.
		}
	}
}

// moveShipOrHelicopter moves 2px on even ticks, bouncing off screen edges.
func moveShipOrHelicopter(slot *Slot, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if slot.Orientation == OrientationLeft {
		slot.X -= enemyMoveStep
		if slot.X <= terrainCheckAheadL {
			slot.Orientation = OrientationRight
		}
	} else {
		slot.X += enemyMoveStep
		if slot.X >= ScreenWidth-terrainCheckAheadR {
			slot.Orientation = OrientationLeft
		}
	}
}

// moveFighter moves 4px every frame, wrapping at screen edges.
func moveFighter(slot *Slot) {
	if slot.Orientation == OrientationLeft {
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
func moveTank(slot *Slot, tick int) {
	if tick&evenTickMask != 0 {
		return
	}

	if slot.Orientation == OrientationLeft {
		slot.X -= enemyMoveStep
	} else {
		slot.X += enemyMoveStep
	}
}

// moveBalloon moves 2px every 4th frame, bouncing off screen edges.
func moveBalloon(slot *Slot, tick int) {
	if tick&balloonTickMask != balloonTickMatch {
		return
	}

	if slot.Orientation == OrientationLeft {
		slot.X -= enemyMoveStep
		if slot.X <= terrainCheckAheadL {
			slot.Orientation = OrientationRight
		}
	} else {
		slot.X += enemyMoveStep
		if slot.X >= ScreenWidth-terrainCheckAheadR {
			slot.Orientation = OrientationLeft
		}
	}
}
