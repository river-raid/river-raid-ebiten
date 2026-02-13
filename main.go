package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetTPS(50)
	ebiten.SetWindowSize(ScreenWidth*WindowScale, ScreenHeight*WindowScale)
	ebiten.SetWindowTitle("River Raid")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
