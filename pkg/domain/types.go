package domain

// GameplayMode represents the current gameplay sub-mode.
type GameplayMode int

// Gameplay modes.
const (
	GameplayNormal GameplayMode = iota
	GameplayScrollIn
	GameplayOverview
	GameplayRefuel
)

// Speed represents the scroll speed in pixels per frame.
type Speed int

// Speeds.
const (
	SpeedSlow   Speed = 1
	SpeedNormal Speed = 2
	SpeedFast   Speed = 4
)

// ObjectType identifies what kind of object occupies a viewport slot.
type ObjectType int

// Object types.
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

// Collision modes.
const (
	CollisionNone CollisionMode = iota
	CollisionFuelDepot
	CollisionMissile
	CollisionFighter
	CollisionHelicopterMissile
)

// InputInterface identifies the selected input method.
type InputInterface int

// Input interfaces.
const (
	InputKeyboard InputInterface = iota
	InputSinclair
	InputKempston
	InputCursor
)

// Player identifies which player is active.
type Player int

// Players.
const (
	Player1 Player = iota
	Player2
)

// GameScreen represents the top-level screen state.
type GameScreen int

// Top-level screen states.
const (
	ScreenControlSelection GameScreen = iota
	ScreenInstructions
	ScreenOverview
	ScreenGameplay
	ScreenGameOver
)

// Orientation represents the facing direction of an object.
type Orientation int

// Orientations.
const (
	OrientationRight Orientation = iota
	OrientationLeft
)

// TankLocation indicates whether a tank is on the road or the river bank.
type TankLocation int

// Tank locations.
const (
	TankLocationRoad TankLocation = iota
	TankLocationBank
)

// StartingBridge selects which bridge the game starts at.
type StartingBridge int

// Starting bridge options.
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
	RockVariant  int
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
