package main

import (
	"github.com/leedark/ui"
)

var window *ui.Window

func main() {

	err := ui.Main(func() {

		window = ui.NewWindow("Hello", 500, 500, false)
		window.SetMargined(true)
		window.Center()

		box := ui.NewHorizontalBox()

		listbox := ui.NewVerticalBox()
		listview := ui.NewMultilineEntry()
		listbox.Append(listview, true)

		buttonbox := ui.NewVerticalBox()
		button := ui.NewButton("Create")
		buttonbox.Append(button, false)

		box.Append(listbox, true)
		box.Append(buttonbox, false)

		window.SetChild(box)

		button.OnClicked(func(*ui.Button) {
			ui.MsgBox(window, "Title", "Hello")
		})

		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})

		window.Show()

	})
	if err != nil {
		panic(err)
	}
}
