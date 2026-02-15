package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/morozov/river-raid-ebiten/pkg/game"
)

const WindowScale = 3

func main() {
	ebiten.SetWindowTitle(game.Title)
	ebiten.SetWindowSize(game.Width*WindowScale, game.Height*WindowScale)
	ebiten.SetTPS(game.Tps)

	if err := ebiten.RunGame(&game.Game{}); err != nil {
		log.Fatal(err)
	}
}
