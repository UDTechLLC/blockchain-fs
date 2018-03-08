package main

import (
	"fmt"
	"time"

	"bitbucket.org/udt/wizefs/internal/util"
	"bitbucket.org/udt/wizefs/proto/nongui"
	"github.com/leedark/ui"
)

type StorageTab struct {
	main             *MainWindow
	tab              *ui.Box
	timeTicker       *time.Ticker
	alreadyAvailable bool

	putFileButton    *ui.Button
	getFileButton    *ui.Button
	removeFileButton *ui.Button
	logBuffer        *nongui.StringBuffer
	logBox           *ui.MultilineEntry

	db         FileDB
	filesView  *ui.Table
	filesModel *ui.TableModel
}

func NewStorageTab(mainWindow *MainWindow) *StorageTab {
	makeTab := &StorageTab{
		main: mainWindow,
	}
	makeTab.buildGUI()
	return makeTab
}

func (t *StorageTab) NewTimer(seconds int, action func()) {
	t.timeTicker = time.NewTicker(time.Duration(seconds) * time.Second)
	go action()
}

func (t *StorageTab) buildGUI() {
	t.tab = ui.NewHorizontalBox()

	vbox1 := ui.NewVerticalBox()
	t.putFileButton = ui.NewButton("Put file")
	t.getFileButton = ui.NewButton("Get file")
	t.removeFileButton = ui.NewButton("Remove file")

	t.putFileButton.OnClicked(t.onPutFileClicked)
	t.getFileButton.OnClicked(t.onGetFileClicked)
	t.removeFileButton.OnClicked(t.onRemoveFileClicked)
	vbox1.SetPadded(true)
	vbox1.Append(t.putFileButton, false)
	vbox1.Append(t.getFileButton, false)
	vbox1.Append(t.removeFileButton, false)

	vbox2 := ui.NewVerticalBox()

	hbox2a := ui.NewHorizontalBox()

	listBox := ui.NewVerticalBox()
	t.filesModel = ui.NewTableModel(&t.db)
	t.filesView = ui.NewTable(t.filesModel, ui.TableStyleMultiSelect)
	t.filesView.AppendTextColumn("Index", 0)
	t.filesView.AppendTextColumn("RaftIndex", 1)
	t.filesView.AppendTextColumn("Name", 2)
	t.filesView.AppendTextColumn("Time", 3)
	listBox.Append(t.filesView, true)

	hbox2a.Append(listBox, true)

	hbox2b := ui.NewHorizontalBox()

	t.logBuffer = nongui.NewStringBuffer()
	t.logBox = ui.NewMultilineEntry()
	t.logBox.SetReadOnly(true)
	hbox2b.SetPadded(true)
	hbox2b.Append(t.logBox, true)

	vbox2.SetPadded(true)
	vbox2.Append(hbox2a, true)
	vbox2.Append(hbox2b, true)

	t.tab.SetPadded(true)
	t.tab.Append(vbox1, false)
	t.tab.Append(ui.NewVerticalSeparator(), false)
	t.tab.Append(vbox2, true)
}

func (t *StorageTab) Control() ui.Control {
	return t.tab
}

func (t *StorageTab) ApiTicker() {
	for {
		select {
		case <-t.timeTicker.C:
			if t.alreadyAvailable != t.main.raftApi.Available {
				if t.main.blockApi.Available {
					t.reloadFilesView()
					t.alreadyAvailable = true
				} else {
					// just clear files listview
					for i := 0; i < len(t.db.Files); i++ {
						t.filesModel.RowDeleted(0)
					}
					t.db.Files = nil

					t.buttonEnabled(false)
					t.alreadyAvailable = false
				}
			}
		}
	}
}

func (t *StorageTab) Init() {
	if t.main.raftApi.Available {
		t.reloadFilesView()
		t.alreadyAvailable = true
	} else {
		t.buttonEnabled(false)
		t.alreadyAvailable = false
	}

	t.NewTimer(60, t.ApiTicker)
}

func (t *StorageTab) reloadFilesView() {
	// check wallet info existing
	if t.main.walletInfo.IsEmpty() {
		//fmt.Printf("walletInfo is empty\n")
		return
	}

	// clear db and Model
	for i := 0; i < len(t.db.Files); i++ {
		t.filesModel.RowDeleted(0)
	}
	t.db.Files = nil

	// CHECKIT:
	_, cpkIndexLastInt64, err := t.main.walletInfo.GetZeroIndex()
	if err != nil {

	}

	// list cycle
	var index int64 = 0
	for index < cpkIndexLastInt64 {
		index++

		// CHECKIT:
		fileRaft, err := t.main.walletInfo.GetFileIndex(index)
		if err != nil {
			continue
		}

		f := File{
			Index:     len(t.db.Files) + 1,
			RaftIndex: int(index),
			Name:      fileRaft.Filename,
			Timestamp: fileRaft.TimeStamp,
			shaKey:    fileRaft.ShaKey,
			cpkIndex:  fileRaft.CpkIndex,
		}
		t.db.Files = append(t.db.Files, f)
		t.filesModel.RowInserted(len(t.db.Files) - 1)
	}

	// TODO: how to select row?
	//if len(t.db.Files) > 0 {
	//	t.filesModel.RowChanged(0)
	//}

	//
	if len(t.db.Files) > 0 {
		t.buttonEnabled(true)
	} else {
		t.buttonEnabled(false)
		// just Put file
		t.putFileButton.Enable()
	}
}

func (t *StorageTab) buttonEnabled(enable bool) {
	if enable {
		t.putFileButton.Enable()
		t.getFileButton.Enable()
		t.removeFileButton.Enable()
	} else {
		t.putFileButton.Disable()
		t.getFileButton.Disable()
		t.removeFileButton.Disable()
	}
}

func (t *StorageTab) logMessage(message string) {
	t.logBuffer = t.logBuffer.Append(
		time.Now().Format("15:04:05.999999") + ": " + message + "\n")
	t.logBox.SetText(t.logBuffer.ToString())
}

func (t *StorageTab) putFile(filename string) {
	var cerr error

	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Open file: " + filename)
	})

	origin := BucketOriginName

	// mount?!
	//cerr = nongui.MountStorage(origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Mount bucket " + origin)
	//	})
	//}

	putSuccess := false
	cerr = nongui.PutFile(filename, origin)
	if cerr != nil {
		ui.QueueMain(func() {
			t.logMessage("Put error: " + cerr.Error())
		})
	} else {
		putSuccess = true
		// TODO: should we wait after put command?
		//time.Sleep(500 * time.Millisecond)
		ui.QueueMain(func() {
			t.logMessage("Put file [" + filename + "] to bucket " + origin)
		})
	}

	// unmount?!
	//cerr = nongui.UnmountStorage(origin)
	//if cerr != nil {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount bucket " + origin)
	//	})
	//}

	if putSuccess {
		t.main.walletInfo.SaveFileToRaft(filename)
		t.reloadFilesView()
	}

	ui.QueueMain(func() {
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) getFile(source string, destination string) {
	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Save file: " + source + " to " + destination)
	})

	origin := BucketOriginName

	// mount?!
	//cerr = nongui.MountStorage(origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Mount bucket " + origin)
	//	})
	//}

	cerr := nongui.GetFile(source, origin, destination)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.logMessage("Get error: " + cerr.Error())
		})
	} else {
		//time.Sleep(500 * time.Millisecond)
		ui.QueueMain(func() {
			t.logMessage("Get file [" + source + "] from bucket " + origin +
				" to path [" + destination + "]")
		})
	}

	// unmount?!
	//cerr = nongui.UnmountStorage(origin)
	//if cerr != nil {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount bucket " + origin)
	//	})
	//}

	ui.QueueMain(func() {
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) removeFile(file File) {
	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Open file: " + file.Name)
	})

	origin := BucketOriginName

	// mount?!
	//cerr = nongui.MountStorage(origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Mount bucket " + origin)
	//	})
	//}

	removeSuccess := false
	cerr := nongui.RemoveFile(file.Name, origin)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.logMessage("Remove error: " + cerr.Error())
		})
	} else {
		removeSuccess = true
		//time.Sleep(500 * time.Millisecond)
		ui.QueueMain(func() {
			t.logMessage("Remove file [" + file.Name + "] from bucket " + origin)
		})
	}

	// unmount?!
	//cerr = nongui.UnmountStorage(origin)
	//if cerr != nil {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//	return
	//} else {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount bucket " + origin)
	//	})
	//}

	if removeSuccess {
		fileRaft := &nongui.FileRaftValue{
			Filename:  file.Name,
			TimeStamp: file.Timestamp,
			ShaKey:    file.shaKey,
			CpkIndex:  file.cpkIndex,
		}

		err := t.main.walletInfo.RemoveFileFromRaft(fileRaft)
		if err != nil {

		}
		t.reloadFilesView()
	}

	ui.QueueMain(func() {
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) onPutFileClicked(button *ui.Button) {
	// check wallet info existing
	if t.main.walletInfo.IsEmpty() {
		fmt.Printf("walletInfo is empty\n")
		return
	}

	file := ui.OpenFile(t.main.window, util.UserHomeDir()+"/*.*")
	if file == "" {
		ui.MsgBoxError(t.main.window, "Error",
			fmt.Sprintf("Please, select file for putting it to storage"))
		return
	}

	go t.putFile(file)
}

func (t *StorageTab) onGetFileClicked(button *ui.Button) {
	// check wallet info existing
	if t.main.walletInfo.IsEmpty() {
		fmt.Printf("walletInfo is empty\n")
		return
	}

	sel := t.filesView.GetSelection()
	if len(sel) != 1 {
		fmt.Println("Nothing is selected!")
		return
	}

	idx := sel[0]
	dbitem := t.db.Files[idx]
	filename := dbitem.Name

	// save file to path
	filenameSave := ui.SaveFile(t.main.window, util.UserHomeDir()+"/"+filename)
	if filenameSave == "" {
		ui.MsgBoxError(t.main.window, "Error",
			fmt.Sprintf("Please, select file for gettig it from storage"))
		return
	}

	go t.getFile(filename, filenameSave)
}

func (t *StorageTab) onRemoveFileClicked(button *ui.Button) {
	// check wallet info existing
	if t.main.walletInfo.IsEmpty() {
		fmt.Printf("walletInfo is empty\n")
		return
	}

	sel := t.filesView.GetSelection()
	if len(sel) != 1 {
		fmt.Println("Nothing is selected!")
		return
	}

	idx := sel[0]
	dbitem := t.db.Files[idx]

	go t.removeFile(dbitem)
}
