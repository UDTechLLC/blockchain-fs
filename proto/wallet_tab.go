package main

import (
	"github.com/leedark/ui"
)

type WalletTab struct {
	tab *ui.Box
}

func NewWalletTab() *WalletTab {
	makeTab := &WalletTab{}
	makeTab.buildGUI()
	return makeTab
}

func (t *WalletTab) buildGUI() {
	t.tab = ui.NewHorizontalBox()

	t.tab.Append(ui.NewLabel("Wallet Page will be soon!"), false)

	//return mainBox
}

func (t *WalletTab) Control() ui.Control {
	return t.tab
}
