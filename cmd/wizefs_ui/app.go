package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	//"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/core"
	"bitbucket.org/udt/wizefs/internal/util"
	"github.com/leedark/ui"
)

type App struct {
	storage      *core.Storage
	db           FilesystemDB
	createDialog *CreateDialog

	listView      *ui.Table
	listViewModel *ui.TableModel
	createButton  *ui.Button
	deleteButton  *ui.Button
	mountButton   *ui.Button
	unmountButton *ui.Button
	putfileButton *ui.Button
	getfileButton *ui.Button
}

func (app *App) Init() {
	app.storage = core.NewStorage()

	//if config.CommonConfig == nil {
	//	config.InitWizeConfig()
	//}
	for origin, filesystem := range app.storage.Config.Filesystems {
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
	app.storage.Config.Load()
	for origin, filesystem := range app.storage.Config.Filesystems {
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

func (app *App) updateModelItem(idx int) {
	time.Sleep(200 * time.Millisecond)

	filesystem := &app.db.Filesystems[idx]
	fsinfo, _, _ := app.storage.Config.GetMountpointInfoByOrigin(filesystem.Origin)
	filesystem.Mountpoint = fsinfo.MountpointKey
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
	app.mountButton = ui.NewButton("Mount")
	app.unmountButton = ui.NewButton("Unmount")
	app.putfileButton = ui.NewButton("Put file")
	app.getfileButton = ui.NewButton("Get file")
	buttonBox.Append(app.createButton, false)
	buttonBox.Append(app.deleteButton, false)
	buttonBox.Append(app.mountButton, false)
	buttonBox.Append(app.unmountButton, false)
	buttonBox.Append(app.putfileButton, false)
	buttonBox.Append(app.getfileButton, false)
	buttonBox.SetPadded(true)

	mainBox.Append(listBox, true)
	mainBox.Append(buttonBox, false)
	mainBox.SetPadded(true)

	app.listView.OnSelectionChanged(app.OnListViewSelectionChanged)
	app.createButton.OnClicked(app.OnCreateClicked)
	app.deleteButton.OnClicked(app.OnDeleteClicked)
	app.mountButton.OnClicked(app.OnMountClicked)
	app.unmountButton.OnClicked(app.OnUnmountClicked)
	app.putfileButton.OnClicked(app.OnPutFileClicked)
	app.getfileButton.OnClicked(app.OnGetFileClicked)

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

func (app *App) OnMountClicked(button *ui.Button) {
	var origin string = ""
	sel := app.listView.GetSelection()
	if len(sel) != 1 {
		return
	}

	idx := sel[0]
	dbitem := app.db.Filesystems[idx]
	origin = dbitem.Origin

	cerr := RunCommand("mount", origin)
	if cerr != nil {
		ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
	} else {
		app.updateModelItem(idx)
		app.listViewModel.RowChanged(idx)

		app.rethink()
	}
}

func (app *App) OnUnmountClicked(button *ui.Button) {
	var origin string = ""
	sel := app.listView.GetSelection()
	if len(sel) != 1 {
		return
	}

	idx := sel[0]
	dbitem := app.db.Filesystems[idx]
	origin = dbitem.Origin

	cerr := RunCommand("unmount", origin)
	if cerr != nil {
		ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
	} else {
		app.updateModelItem(idx)
		app.listViewModel.RowChanged(idx)

		app.rethink()
	}
}

func (app *App) OnPutFileClicked(button *ui.Button) {
	var origin string = ""
	sel := app.listView.GetSelection()
	if len(sel) != 1 {
		return
	}

	idx := sel[0]
	dbitem := app.db.Filesystems[idx]
	origin = dbitem.Origin

	file := ui.OpenFile(window, util.UserHomeDir()+"/*.*")
	//fmt.Println("file: ", file)

	if file == "" {
		//ui.MsgBoxError(window, "Error",
		//	fmt.Sprintf("Please, select file for putting it to filesystem"))
		return
	}

	cerr := RunCommand("put", file, origin)
	if cerr != nil {
		ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
	} else {
		//app.updateModelItem(idx)
		//app.listViewModel.RowChanged(idx)

		//app.rethink()
	}
}

func (app *App) OnGetFileClicked(button *ui.Button) {
	var origin string = ""
	sel := app.listView.GetSelection()
	if len(sel) != 1 {
		return
	}

	idx := sel[0]
	dbitem := app.db.Filesystems[idx]
	origin = dbitem.Origin

	// get mountpoint path
	_, mpinfo, _ := app.storage.Config.GetMountpointInfoByOrigin(origin)
	if mpinfo.MountpointPath == "" {
		return
	}

	// open file from mountpoint
	file := ui.OpenFile(window, mpinfo.MountpointPath+"/*.*")
	//fmt.Println("file: ", file)
	if file == "" {
		//ui.MsgBoxError(window, "Error",
		//	fmt.Sprintf("Please, select file for gettig it from filesystem"))
		return
	}

	fileBase := filepath.Base(file)
	//fmt.Println("fileBase: ", fileBase)
	cerr := RunCommand("get", fileBase, origin)
	if cerr != nil {
		ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
	} else {
		//
	}
}

func (app *App) rethink() {
	sel := app.listView.GetSelection()
	//fmt.Printf("selected: %v\n", sel)

	invalid := len(sel) > 0

	app.deleteButton.Disable()
	app.mountButton.Disable()
	app.unmountButton.Disable()
	app.putfileButton.Disable()
	app.getfileButton.Disable()

	if invalid {

		if len(sel) == 1 {
			idx := sel[0]
			dbitem := app.db.Filesystems[idx]
			// check error

			if dbitem.Mountpoint == "" {
				// is not mounted
				app.deleteButton.Enable()
				app.mountButton.Enable()
			} else {
				// is mounted
				app.unmountButton.Enable()
				app.putfileButton.Enable()
				app.getfileButton.Enable()
			}
		}
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
				ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
			} else {
				app.db.Filesystems = append(app.db.Filesystems[:idx], app.db.Filesystems[idx+1:]...)
				app.listViewModel.RowDeleted(idx)
			}
		}
	}
	app.HandleSelectionChanged()
}
