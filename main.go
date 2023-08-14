package main

import (
	"github.com/blckfalcon/go-ytdlp-mngr/ui"
)

func main() {
	var app = ui.NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}

	app.CleanUp()
}
