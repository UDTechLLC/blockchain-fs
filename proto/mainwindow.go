package main

import (
	"bitbucket.org/udt/wizefs/internal/config"
	"github.com/leedark/ui"
)

const (
	BucketOriginName = "Bucket1.zip"
)

type MainWindow struct {
	window     *ui.Window
	walletTab  *WalletTab
	storageTab *StorageTab

	blockApi *BlockApi
	raftApi  *RaftApi

	walletInfo *WalletCreateInfo
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
	main.raftApi = NewRaftApi()
}

func (main *MainWindow) Show() {
	main.window.Show()
}

func (main *MainWindow) buildGUI() ui.Control {
	tab := ui.NewTab()

	main.walletTab = NewWalletTab(main)
	tab.Append("  Wallet  ", main.walletTab.Control())
	tab.SetMargined(0, true)

	main.storageTab = NewStorageTab(main)
	tab.Append("  Storage  ", main.storageTab.Control())
	tab.SetMargined(1, true)

	tab.Append("  Debug  ", NewDebugTab().Control())
	tab.SetMargined(2, true)

	main.walletTab.init()
	main.storageTab.init()

	return tab
}

func (main *MainWindow) OnClosing(window *ui.Window) bool {
	// FIXME:
	if main.walletInfo != nil {
		saveWalletInfo(main.walletInfo)
	}

	ui.Quit()
	return true
}
