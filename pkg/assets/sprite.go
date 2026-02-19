package assets

// Sprite holds a 1-bit-per-pixel bitmap and its pixel dimensions.
// The bitmap is stored as ceil(Width/8) bytes per row, MSB first.
// Height is derived from len(Data) / BytesPerRow.
type Sprite struct {
	Data        []byte
	Width       int
	BytesPerRow int
}

// SpriteID identifies a sprite in the catalog.
type SpriteID int

// Sprite catalog entries.
const (
	SpriteExplosionSmall SpriteID = iota
	SpriteExplosionMedium
	SpriteExplosionLarge
	SpriteShellExplosion0
	SpriteShellExplosion1
	SpriteShellExplosion2
	SpriteShellExplosion3
	SpriteShellExplosion4
	SpriteShellExplosion5
	spriteCount
)

// SpriteCatalog contains all game sprites, indexed by SpriteID.
//
//nolint:mnd // sprite dimensions are inherent to the extracted Data
var SpriteCatalog = [spriteCount]Sprite{
	SpriteExplosionSmall:  newSprite(spriteExplosionSmall[:], 10),
	SpriteExplosionMedium: newSprite(spriteExplosionMedium[:], 10),
	SpriteExplosionLarge:  newSprite(spriteExplosionLarge[:], 10),
	SpriteShellExplosion0: newSprite(spriteShellExplosion0[:], 10),
	SpriteShellExplosion1: newSprite(spriteShellExplosion1[:], 10),
	SpriteShellExplosion2: newSprite(spriteShellExplosion2[:], 10),
	SpriteShellExplosion3: newSprite(spriteShellExplosion3[:], 10),
	SpriteShellExplosion4: newSprite(spriteShellExplosion4[:], 10),
	SpriteShellExplosion5: newSprite(spriteShellExplosion5[:], 10),
}

// Height returns the sprite height in pixels.
func (s Sprite) Height() int {
	return len(s.Data) / s.BytesPerRow
}

// newSprite creates a Sprite from raw 1bpp bitmap Data.
// w is the visual width in pixels; height is derived from len(Data).
func newSprite(data []byte, w int) Sprite {
	bpr := (w + 7) / 8 //nolint:mnd // ceiling division by 8 bits per byte

	return Sprite{
		Data:        data,
		Width:       w,
		BytesPerRow: bpr,
	}
}
