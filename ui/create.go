package main

import (
	"fmt"

	"github.com/leedark/ui"
)

func (app *App) buildCreateDialog() ui.Control {
	createwindow := ui.NewWindow("Dialog", 300, 120, false)
	createwindow.SetMargined(true)
	createwindow.Center()

	mainBox := ui.NewVerticalBox()

	topBox := ui.NewVerticalBox()

	originLabel := ui.NewLabel("Input filesystem label:")
	originEdit := ui.NewEntry()
	topBox.Append(originLabel, false)
	topBox.Append(originEdit, false)
	topBox.SetPadded(true)

	mainBox.Append(topBox, false)
	mainBox.Append(ui.NewHorizontalSeparator(), false)

	buttonBox := ui.NewHorizontalBox()
	okButton := ui.NewButton("Ok")
	cancelButton := ui.NewButton("Cancel")

	buttonBox.Append(okButton, true)
	buttonBox.Append(cancelButton, true)
	buttonBox.SetPadded(true)

	mainBox.Append(buttonBox, false)

	mainBox.SetPadded(true)

	createwindow.SetChild(mainBox)

	createwindow.OnClosing(func(*ui.Window) bool {
		createwindow.Destroy()
		return false
	})

	okButton.OnClicked(func(*ui.Button) {
		origin := originEdit.Text()
		fmt.Println("OK Origin: ", origin)

		createwindow.Destroy()

		cerr := RunCommand("create", origin)
		if cerr != nil {
			fmt.Println(cerr)
			ui.MsgBoxError(window, "Error", fmt.Sprintf("Error: %v", cerr))
		} else {
			app.rethink()
			app.updateModel()
		}
	})

	cancelButton.OnClicked(func(*ui.Button) {
		createwindow.Destroy()
	})

	return createwindow
}
