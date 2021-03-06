package main

import (
	"fmt"
	"time"

	"bitbucket.org/udt/wizefs/cmd/wizebit_proto/nongui"
	"bitbucket.org/udt/wizefs/internal/config"
	"github.com/leedark/ui"
)

const (
	//BucketOriginName = "Bucket1.zip"
	BucketOriginName = "Bucket1"
)

type MainWindow struct {
	window     *ui.Window
	walletTab  *WalletTab
	storageTab *StorageTab

	blockApi   *nongui.BlockApi
	raftApi    *nongui.RaftApi
	timeTicker *time.Ticker

	walletInfo *nongui.WalletCreateInfo
}

func NewMainWindow() *MainWindow {
	main := &MainWindow{}

	main.window = ui.NewWindow("Wize Client "+config.ProgramVersion, 1000, 600, false)
	main.window.SetMargined(true)
	main.window.Center()

	main.Init()
	gui := main.buildGUI()

	main.window.SetChild(gui)
	main.window.OnClosing(main.OnClosing)

	return main
}

func (main *MainWindow) NewTimer(seconds int, action func()) {
	main.timeTicker = time.NewTicker(time.Duration(seconds) * time.Second)
	go action()
}

func (main *MainWindow) MountStorage() {
	cerr := nongui.MountStorage(BucketOriginName)
	if cerr != nil {
		ui.MsgBoxError(main.window, "Mount Storage Error", fmt.Sprintf("%v", cerr))
		return
	}
}

func (main *MainWindow) UnmountStorage() {
	cerr := nongui.UnmountStorage(BucketOriginName)
	if cerr != nil {
		ui.MsgBoxError(main.window, "Unmount Storage Error", fmt.Sprintf("%v", cerr))
	}
}

func (main *MainWindow) Init() {
	main.blockApi = nongui.NewBlockApi()
	main.raftApi = nongui.NewRaftApi()

	main.NewTimer(60, main.ApiTicker)
}

func (main *MainWindow) ApiTicker() {
	for {
		select {
		case <-main.timeTicker.C:
			main.blockApi.CheckApi()
			main.raftApi.CheckApi()
		}
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

	//tab.Append("   Debug   ", NewDebugTab().Control())
	//tab.SetMargined(2, true)

	main.walletTab.Init()
	main.storageTab.Init()

	return tab
}

func (main *MainWindow) OnClosing(window *ui.Window) bool {
	main.timeTicker.Stop()

	if main.walletInfo != nil {
		main.walletInfo.Save()

		main.UnmountStorage()
	}

	ui.Quit()
	return true
}
