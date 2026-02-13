package main

import "github.com/hajimehoshi/ebiten/v2"

// Game implements the ebiten.Game interface.
type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(_ *ebiten.Image) {}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
