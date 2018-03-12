package main

import (
	"fmt"

	"github.com/leedark/ui"
)

type CreateDialog struct {
	window       *ui.Window
	originLabel  *ui.Label
	originEdit   *ui.Entry
	okButton     *ui.Button
	cancelButton *ui.Button
}

func (app *App) buildCreateDialog() {
	app.createDialog = &CreateDialog{}

	app.createDialog.window = ui.NewWindow("Dialog", 300, 120, false)
	app.createDialog.window.SetMargined(true)
	app.createDialog.window.Center()

	mainBox := ui.NewVerticalBox()

	topBox := ui.NewVerticalBox()

	app.createDialog.originLabel = ui.NewLabel("Input filesystem label:")
	app.createDialog.originEdit = ui.NewEntry()
	topBox.Append(app.createDialog.originLabel, false)
	topBox.Append(app.createDialog.originEdit, false)
	topBox.SetPadded(true)

	mainBox.Append(topBox, false)
	mainBox.Append(ui.NewHorizontalSeparator(), false)

	buttonBox := ui.NewHorizontalBox()
	app.createDialog.okButton = ui.NewButton("Ok")
	app.createDialog.cancelButton = ui.NewButton("Cancel")

	buttonBox.Append(app.createDialog.okButton, true)
	buttonBox.Append(app.createDialog.cancelButton, true)
	buttonBox.SetPadded(true)

	mainBox.Append(buttonBox, false)
	mainBox.SetPadded(true)

	app.createDialog.window.SetChild(mainBox)

	app.createDialog.window.OnClosing(app.OnCreateDialogClosing)
	app.createDialog.okButton.OnClicked(app.OnCreadeOkClicked)
	app.createDialog.cancelButton.OnClicked(app.OnCreateCancelClicked)

	//return app.createDialog
}

func (app *App) OnCreateDialogClosing(window *ui.Window) bool {
	app.createDialog.window.Destroy()
	return false
}

func (app *App) OnCreadeOkClicked(button *ui.Button) {
	origin := app.createDialog.originEdit.Text()
	fmt.Println("OK Origin: ", origin)

	app.createDialog.window.Destroy()

	cerr := RunCommand("create", origin)
	if cerr != nil {
		fmt.Println(cerr)
		ui.MsgBoxError(window, "Error", fmt.Sprintf("%v", cerr))
	} else {
		app.rethink()
		app.updateModel()
	}
}

func (app *App) OnCreateCancelClicked(button *ui.Button) {
	app.createDialog.window.Destroy()
}
