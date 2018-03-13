package main

import (
	"github.com/leedark/ui"
)

func main() {
	err := ui.Main(func() {
		NewMainWindow().Show()
	})

	if err != nil {
		panic(err)
	}
}
