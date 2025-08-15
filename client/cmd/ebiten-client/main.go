package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"go-ddd-architecture/client/internal/api/gameclient"
	"go-ddd-architecture/client/internal/ui"
)

func main() {
	api := gameclient.New("http://127.0.0.1:8080")
	state := &ui.State{}
	app := ui.NewApp(api, state)
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Intellect Client")
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
