package main

import (
	"fmt"
	"sort"

	"github.com/leedark/ui"
)

type App struct {
	db FilesystemDB

	listView      *ui.Table
	listViewModel *ui.TableModel
	createButton  *ui.Button
	deleteButton  *ui.Button
}

func (app *App) Init() {

}

func (app *App) buildGUI() ui.Control {
	mainBox := ui.NewHorizontalBox()

	listBox := ui.NewVerticalBox()

	app.listViewModel = ui.NewTableModel(&app.db)
	listView := ui.NewTable(app.listViewModel, ui.TableStyleMultiSelect)
	listView.AppendTextColumn("Index", 0)
	listView.AppendTextColumn("Filesystem", 1)
	listView.AppendTextColumn("Path", 2)

	listView.OnSelectionChanged(func(t *ui.Table) {
		app.HandleSelectionChanged()
	})

	listBox.Append(listView, true)
	app.listView = listView

	buttonBox := ui.NewVerticalBox()
	app.createButton = ui.NewButton("Create")
	app.deleteButton = ui.NewButton("Delete")
	buttonBox.Append(app.createButton, false)
	buttonBox.Append(app.deleteButton, false)
	buttonBox.SetPadded(true)

	mainBox.Append(listBox, true)
	mainBox.Append(buttonBox, false)
	mainBox.SetPadded(true)

	app.createButton.OnClicked(func(*ui.Button) {
		fs := Filesystem{
			Index: len(app.db.Filesystems) + 1,
			Name:  "filesystem Name",
			Path:  "filesystem Path",
		}
		app.db.Filesystems = append(app.db.Filesystems, fs)
		app.listViewModel.RowInserted(len(app.db.Filesystems) - 1)
		app.rethink()
	})

	app.deleteButton.OnClicked(func(*ui.Button) {
		app.DeleteSelected()
	})

	app.rethink()

	return mainBox
}

func (app *App) rethink() {
	sel := app.listView.GetSelection()
	fmt.Printf("selected: %v\n", sel)
	invalid := len(sel) > 0

	if invalid {
		app.deleteButton.Enable()
	} else {
		app.deleteButton.Disable()
	}

	app.createButton.Enable()
}

func (app *App) HandleSelectionChanged() {
	app.rethink()
}

func (app *App) DeleteSelected() {
	sel := app.listView.GetSelection()
	// remove highest-first so we don't screw up our indices
	sort.Sort(sort.Reverse(sort.IntSlice(sel)))
	for _, idx := range sel {
		app.db.Filesystems = append(app.db.Filesystems[:idx], app.db.Filesystems[idx+1:]...)
		app.listViewModel.RowDeleted(idx)
	}
	app.HandleSelectionChanged()
}
