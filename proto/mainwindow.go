package main

import (
	"fmt"
	"time"

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

	blockApi   *BlockApi
	raftApi    *RaftApi
	timeTicker *time.Ticker

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

func NewTimer(seconds int, action func()) *time.Ticker {
	timeTicker := time.NewTicker(time.Duration(seconds) * time.Second)
	go action()
	return timeTicker
}

func (main *MainWindow) Init() {
	main.blockApi = NewBlockApi()
	main.raftApi = NewRaftApi()

	main.timeTicker = NewTimer(60, main.ApiTicker)
}

func (main *MainWindow) ApiTicker() {
	for t := range main.timeTicker.C {
		fmt.Println("Tick at", t)
		main.blockApi.CheckApi()
		main.raftApi.CheckApi()
	}
}

func (main *MainWindow) Show() {
	main.window.Show()
}

func (main *MainWindow) buildGUI() ui.Control {
	tab := ui.NewTab()

	main.walletTab = NewWalletTab(main)
	tab.Append("   Wallet  ", main.walletTab.Control())
	tab.SetMargined(0, true)

	main.storageTab = NewStorageTab(main)
	tab.Append("  Storage  ", main.storageTab.Control())
	tab.SetMargined(1, true)

	tab.Append("   Debug   ", NewDebugTab().Control())
	tab.SetMargined(2, true)

	main.walletTab.init()
	main.storageTab.init()

	return tab
}

func (main *MainWindow) OnClosing(window *ui.Window) bool {
	main.timeTicker.Stop()
	fmt.Println("Ticker stopped")

	// FIXME:
	if main.walletInfo != nil {
		saveWalletInfo(main.walletInfo)
	}

	ui.Quit()
	return true
}
