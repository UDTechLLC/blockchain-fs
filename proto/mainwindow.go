package main

import (
	"bitbucket.org/udt/wizefs/internal/config"
	"github.com/leedark/ui"
)

type MainWindow struct {
	window *ui.Window
}

func NewMainWindow() *MainWindow {
	main := &MainWindow{}
	main.Init()

	gui := main.buildGUI()

	main.window = ui.NewWindow("Wize Client "+config.ProgramVersion, 700, 500, false)
	main.window.SetMargined(true)
	main.window.Center()

	main.window.SetChild(gui)

	main.window.OnClosing(main.OnClosing)

	return main
}

func (main *MainWindow) Init() {
}

func (main *MainWindow) Show() {
	main.window.Show()
}

func (main *MainWindow) buildGUI() ui.Control {
	mainBox := ui.NewHorizontalBox()

	return mainBox
}

func (main *MainWindow) OnClosing(window *ui.Window) bool {
	ui.Quit()
	return true
}
