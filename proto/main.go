package main

import (
	"github.com/leedark/ui"
)

// FIXME
//var window *ui.Window

func main() {
	err := ui.Main(func() {
		NewMainWindow().Show()
	})

	if err != nil {
		panic(err)
	}
}
