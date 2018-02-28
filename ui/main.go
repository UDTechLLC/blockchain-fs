package main

import (
	"github.com/leedark/ui"

	"bitbucket.org/udt/wizefs/internal/config"
)

var window *ui.Window

func main() {
	err := ui.Main(func() {
		app := &App{}
		app.Init()

		gui := app.buildGUI()

		window = ui.NewWindow("Wize Client "+config.ProgramVersion, 700, 500, false)
		window.SetMargined(true)
		window.Center()

		window.SetChild(gui)

		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})

		window.Show()
	})

	if err != nil {
		panic(err)
	}
}
