package main

import (
	"github.com/leedark/ui"
)

type StorageTab struct {
	tab *ui.Box
}

func NewStorageTab() *StorageTab {
	makeTab := &StorageTab{}
	makeTab.buildGUI()
	return makeTab
}

func (t *StorageTab) buildGUI() {
	t.tab = ui.NewHorizontalBox()

	t.tab.Append(ui.NewLabel("Storage Page will be soon!"), false)

	//return mainBox
}

func (t *StorageTab) Control() ui.Control {
	return t.tab
}
