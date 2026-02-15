package main

// GameplayMode represents the current gameplay sub-mode.
type GameplayMode int

const (
	GameplayNormal GameplayMode = iota
	GameplayScrollIn
	GameplayOverview
	GameplayRefuel
)

// Speed represents the scroll speed in pixels per frame.
type Speed int

const (
	SpeedSlow   Speed = 1
	SpeedNormal Speed = 2
	SpeedFast   Speed = 4
)

// ObjectType identifies what kind of object occupies a viewport slot.
type ObjectType int

const (
	ObjectHelicopterReg ObjectType = iota + 1
	ObjectShip
	ObjectHelicopterAdv
	ObjectTank
	ObjectFighter
	ObjectBalloon
	ObjectFuel
)

// CollisionMode tracks what the player is currently colliding with.
type CollisionMode int

const (
	CollisionNone CollisionMode = iota
	CollisionFuelDepot
	CollisionMissile
	CollisionFighter
	CollisionHelicopterMissile
)

// InputInterface identifies the selected input method.
type InputInterface int

const (
	InputKeyboard InputInterface = iota
	InputSinclair
	InputKempston
	InputCursor
)

// Player identifies which player is active.
type Player int

const (
	Player1 Player = iota
	Player2
)

// GameScreen represents the top-level screen state.
type GameScreen int

const (
	ScreenControlSelection GameScreen = iota
	ScreenInstructions
	ScreenOverview
	ScreenGameplay
	ScreenGameOver
)

// Orientation represents the facing direction of an object.
type Orientation int

const (
	OrientationRight Orientation = iota
	OrientationLeft
)

// TankLocation indicates whether a tank is on the road or the river bank.
type TankLocation int

const (
	TankLocationRoad TankLocation = iota
	TankLocationBank
)

// EdgeMode controls how the right terrain edge is calculated from the left edge.
type EdgeMode int

const (
	EdgeMirrored = 1 // rightX = 2*center - leftX
	EdgeOffset   = 2 // rightX = width + leftX
)

// StartingBridge selects which bridge the game starts at.
type StartingBridge int

const (
	StartingBridge01 StartingBridge = iota // Bridge 1
	StartingBridge05                       // Bridge 5
	StartingBridge20                       // Bridge 20
	StartingBridge30                       // Bridge 30
)

// ControlFlags holds the expanded state of the original control byte.
type ControlFlags struct {
	Speed     Speed
	FireSound bool
	LowFuel   bool
	BonusLife bool
	Exploding bool
}

// GameConfig holds game configuration options.
type GameConfig struct {
	IsTwoPlayer    bool
	StartingBridge StartingBridge
}

// PlayerState holds a per-player state that persists across lives.
type PlayerState struct {
	Score       int
	Lives       int
	BridgeIndex int
}

// Slot represents a single object in the viewport.
type Slot struct {
	X            int
	Y            int
	Type         ObjectType
	TankLocation TankLocation
	Orientation  Orientation
	IsRock       bool
	Activated    bool
}

// ExplodingFragment represents an active explosion fragment.
type ExplodingFragment struct {
	X     int
	Y     int
	Frame int
}
