package main

import (
	"fmt"
	"time"

	"bitbucket.org/udt/wizefs/internal/util"
	"github.com/leedark/ui"
)

type StorageTab struct {
	main *MainWindow
	tab  *ui.Box

	putFileButton *ui.Button
	logBuffer     *StringBuffer
	logBox        *ui.MultilineEntry
}

func NewStorageTab(mainWindow *MainWindow) *StorageTab {
	makeTab := &StorageTab{
		main: mainWindow,
	}
	makeTab.buildGUI()
	return makeTab
}

func (t *StorageTab) buildGUI() {
	t.tab = ui.NewHorizontalBox()
	//t.tab.Append(ui.NewLabel("Storage Page will be soon!"), false)

	vbox1 := ui.NewVerticalBox()
	t.putFileButton = ui.NewButton("Put file")
	t.putFileButton.OnClicked(t.onPutFileClicked)
	vbox1.SetPadded(true)
	vbox1.Append(t.putFileButton, false)

	vbox2 := ui.NewVerticalBox()
	t.logBuffer = NewStringBuffer()
	t.logBox = ui.NewMultilineEntry()
	t.logBox.SetReadOnly(true)
	vbox2.SetPadded(true)
	vbox2.Append(t.logBox, true)

	t.tab.SetPadded(true)
	t.tab.Append(vbox1, false)
	t.tab.Append(ui.NewVerticalSeparator(), false)
	t.tab.Append(vbox2, true)
}

func (t *StorageTab) Control() ui.Control {
	return t.tab
}

func (t *StorageTab) logMessage(message string) {
	t.logBuffer = t.logBuffer.Append(
		time.Now().Format("15:04:05.999999") + ": " + message + "\n")
	t.logBox.SetText(t.logBuffer.ToString())
}

func (t *StorageTab) putFile(file string) {
	ui.QueueMain(func() {
		t.putFileButton.Disable()
		t.logMessage("Open file: " + file)
	})

	origin := BucketOriginName

	// mount
	cerr := RunCommand("mount", origin)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.putFileButton.Enable()
			t.logMessage("Mount error: " + cerr.Error())
		})
		return
	}

	time.Sleep(500 * time.Millisecond)
	ui.QueueMain(func() {
		t.logMessage("Mount bucket " + origin)
	})

	cerr = RunCommand("put", file, origin)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.logMessage("Put error: " + cerr.Error())
		})
	} else {
		//time.Sleep(500 * time.Millisecond)
		ui.QueueMain(func() {
			t.logMessage("Put file [" + file + "] to bucket " + origin)
		})
	}

	// unmount
	cerr = RunCommand("unmount", origin)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.logMessage("Unmount error: " + cerr.Error())
		})
	}

	ui.QueueMain(func() {
		t.logMessage("Unmount bucket " + origin)
		t.putFileButton.Enable()
	})
}

func (t *StorageTab) onPutFileClicked(button *ui.Button) {
	file := ui.OpenFile(t.main.window, util.UserHomeDir()+"/Downloads/*.*")
	//fmt.Println("file: ", file)

	if file == "" {
		ui.MsgBoxError(t.main.window, "Error",
			fmt.Sprintf("Please, select file for putting it to filesystem"))
		return
	}

	go t.putFile(file)
}
