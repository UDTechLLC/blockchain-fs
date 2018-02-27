package main

import (
	"fmt"
	"sort"

	"bitbucket.org/udt/wizefs/internal/config"
	"github.com/leedark/ui"
)

type App struct {
	db           FilesystemDB
	createDialog *CreateDialog

	listView      *ui.Table
	listViewModel *ui.TableModel
	createButton  *ui.Button
	deleteButton  *ui.Button
}

func (app *App) Init() {
	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}
	for origin, filesystem := range config.CommonConfig.Filesystems {
		fs := Filesystem{
			Index:      len(app.db.Filesystems) + 1,
			Origin:     origin,
			OriginPath: filesystem.OriginPath,
			Type:       int(filesystem.Type),
			Mountpoint: filesystem.MountpointKey,
		}
		app.db.Filesystems = append(app.db.Filesystems, fs)
	}
}

func (app *App) updateModel() {
	config.CommonConfig.Load()
	for origin, filesystem := range config.CommonConfig.Filesystems {
		// TODO: HACK simple update by checking all map - is not quick solution
		if !app.db.HasOrigin(origin) {
			fs := Filesystem{
				Index:      len(app.db.Filesystems) + 1,
				Origin:     origin,
				OriginPath: filesystem.OriginPath,
				Type:       int(filesystem.Type),
				Mountpoint: filesystem.MountpointKey,
			}
			app.db.Filesystems = append(app.db.Filesystems, fs)
			app.listViewModel.RowInserted(len(app.db.Filesystems) - 1)
		}
	}
}

func (app *App) buildGUI() ui.Control {
	mainBox := ui.NewHorizontalBox()

	listBox := ui.NewVerticalBox()

	app.listViewModel = ui.NewTableModel(&app.db)
	listView := ui.NewTable(app.listViewModel, ui.TableStyleMultiSelect)
	listView.AppendTextColumn("Index", 0)
	listView.AppendTextColumn("Origin", 1)
	listView.AppendTextColumn("Path", 2)
	listView.AppendTextColumn("Type", 3)
	listView.AppendTextColumn("Mount", 4)

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

	app.listView.OnSelectionChanged(app.OnListViewSelectionChanged)
	app.createButton.OnClicked(app.OnCreateClicked)
	app.deleteButton.OnClicked(app.OnDeleteClicked)

	app.rethink()
	return mainBox
}

func (app *App) OnListViewSelectionChanged(table *ui.Table) {
	app.HandleSelectionChanged()
}

func (app *App) OnCreateClicked(button *ui.Button) {
	app.buildCreateDialog()
	app.createDialog.window.Show()
}

func (app *App) OnDeleteClicked(button *ui.Button) {
	app.DeleteSelected()
}

func (app *App) rethink() {
	sel := app.listView.GetSelection()
	fmt.Printf("selected: %v\n", sel)

	//if len(sel) == 1 {
	//	dbitem := app.db.Filesystems[sel[0]]
	//	fmt.Println("dbitem: ", dbitem)
	//}

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
		if len(sel) == 1 {
			dbitem := app.db.Filesystems[sel[0]]
			cerr := RunCommand("delete", dbitem.Origin)
			if cerr != nil {
				fmt.Println(cerr)
				ui.MsgBoxError(window, "Error", fmt.Sprintf("Error: %v", cerr))
			} else {
				//app.rethink()
				//app.updateModel()
			}
		}

		app.db.Filesystems = append(app.db.Filesystems[:idx], app.db.Filesystems[idx+1:]...)
		app.listViewModel.RowDeleted(idx)
	}
	app.HandleSelectionChanged()
}
