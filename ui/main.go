package main

import (
	"github.com/andlabs/ui"
)

var window *ui.Window

func main() {

	err := ui.Main(func() {

		window = ui.NewWindow("Hello", 500, 500, false)
		window.SetMargined(true)

		input := ui.NewEntry()
		button := ui.NewButton("Greet")
		greeting := ui.NewLabel("")

		box := ui.NewVerticalBox()
		box.Append(ui.NewLabel("Enter your name:"), false)
		box.Append(input, false)
		box.Append(button, false)
		box.Append(greeting, false)

		window.SetChild(box)

		button.OnClicked(func(*ui.Button) {
			greeting.SetText("Hello, " + input.Text() + "!")
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
