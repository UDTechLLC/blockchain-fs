package main

import (
	"bitbucket.org/udt/wizefs/internal/config"
	"github.com/leedark/ui"
)

const (
	BucketOriginName = "Bucket1.zip"
)

type MainWindow struct {
	window   *ui.Window
	blockApi *BlockApi
}

func NewMainWindow() *MainWindow {
	main := &MainWindow{}
	main.Init()

	main.window = ui.NewWindow("Wize Client "+config.ProgramVersion, 1000, 600, false)
	main.window.SetMargined(true)
	main.window.Center()

	gui := main.buildGUI()

	main.window.SetChild(gui)

	main.window.OnClosing(main.OnClosing)

	return main
}

func (main *MainWindow) Init() {
	main.blockApi = NewBlockApi()
}

func (main *MainWindow) Show() {
	main.window.Show()
}

func (main *MainWindow) buildGUI() ui.Control {
	//mainBox := ui.NewHorizontalBox()

	tab := ui.NewTab()

	tab.Append("  Wallet  ", NewWalletTab(main).Control())
	tab.SetMargined(0, true)

	tab.Append("  Storage  ", NewStorageTab().Control())
	tab.SetMargined(1, true)

	tab.Append("  Debug  ", NewDebugTab().Control())
	tab.SetMargined(2, true)

	return tab

	//return mainBox
}

func (main *MainWindow) OnClosing(window *ui.Window) bool {
	ui.Quit()
	return true
}
