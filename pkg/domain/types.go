package domain

// GameplayMode represents the current gameplay sub-mode.
type GameplayMode int

// Gameplay modes.
const (
	GameplayNormal GameplayMode = iota
	GameplayScrollIn
	GameplayOverview
	GameplayRefuel
	GameplayDying
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
	StartingBridge01 StartingBridge = 1
	StartingBridge05 StartingBridge = 5
	StartingBridge20 StartingBridge = 20
	StartingBridge30 StartingBridge = 30
)

// GameConfig holds game configuration options.
type GameConfig struct {
	IsTwoPlayer    bool
	StartingBridge StartingBridge
}
