package main

import (
	"github.com/leedark/ui"
)

type DebugTab struct {
	tab *ui.Box
}

func NewDebugTab() *DebugTab {
	makeTab := &DebugTab{}
	makeTab.buildGUI()
	return makeTab
}

func (t *DebugTab) buildGUI() {
	t.tab = ui.NewHorizontalBox()
	t.tab.Append(ui.NewLabel("Debug Page will be soon!"), false)
}

func (t *DebugTab) Control() ui.Control {
	return t.tab
}
