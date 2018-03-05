package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	db         FileDB
	filesView  *ui.Table
	filesModel *ui.TableModel
}

func NewStorageTab(mainWindow *MainWindow) *StorageTab {
	makeTab := &StorageTab{
		main: mainWindow,
	}
	makeTab.buildGUI()
	makeTab.init()
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

	hbox2a := ui.NewHorizontalBox()

	listBox := ui.NewVerticalBox()
	t.filesModel = ui.NewTableModel(&t.db)
	t.filesView = ui.NewTable(t.filesModel, ui.TableStyleMultiSelect)
	t.filesView.AppendTextColumn("Index", 0)
	t.filesView.AppendTextColumn("Name", 1)
	t.filesView.AppendTextColumn("Time", 2)
	listBox.Append(t.filesView, true)

	hbox2a.Append(listBox, true)

	hbox2b := ui.NewHorizontalBox()

	t.logBuffer = NewStringBuffer()
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

func (t *StorageTab) init() {
	// TODO: get files list
	/*
		f := File{
			Index:     len(t.db.Files) + 1,
			Name:      value,
			Timestamp: time.Now(),
		}
		t.db.Files = append(t.db.Files, f)
	*/

	t.reloadFilesView()
}

func (t *StorageTab) reloadFilesView() {
	for i := 0; i < len(t.db.Files); i++ {
		t.filesModel.RowDeleted(0)
	}
	t.db.Files = nil

	// get last index from CPK Index Store
	// CPK + 00000000 (8 bytes)
	// FIXME: is it copy or pointer?
	cpkIndex0 := []byte(t.main.walletInfo.PubKey)
	index0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(index0, uint64(0))
	index0 = []byte(fmt.Sprintf("%x", index0))

	cpkIndex0 = append(cpkIndex0, index0...)

	cpkIndexLast, err := t.main.raftApi.Get(string(cpkIndex0))
	// TODO: check cpkIndexLast - len, value, etc
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Printf("cpkIndexLast: %s\n", cpkIndexLast)
	var cpkIndexLastInt64 int64
	if cpkIndexLast == "" {
		fmt.Println("Empty...")
		return
	} else {
		cpkIndexLastDecode, _ := hex.DecodeString(cpkIndexLast)
		// TODO: check it
		cpkIndexLastInt64 = int64(binary.LittleEndian.Uint64(cpkIndexLastDecode))
	}

	fmt.Printf("cpkIndexLastInt64: %x\n", cpkIndexLastInt64)
	fmt.Printf("cpkIndexLastInt64: %s\n", strconv.Itoa(int(cpkIndexLastInt64)))

	// for
	var index int64 = 0
	for index < cpkIndexLastInt64 {
		index++

		cpkIndex := []byte(t.main.walletInfo.PubKey)

		cpkIndexNew := make([]byte, 8)
		binary.LittleEndian.PutUint64(cpkIndexNew, uint64(index))
		cpkIndexNew = []byte(fmt.Sprintf("%x", cpkIndexNew))

		fmt.Printf("cpkIndexNew: %x\n", cpkIndexNew)
		fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

		cpkIndex = append(cpkIndex, cpkIndexNew...)

		fmt.Printf("cpkIndex: %x\n", cpkIndex)
		fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

		shaKeyString, err := t.main.raftApi.Get(string(cpkIndex))
		if err != nil {
			fmt.Println("Error: ", err)
		}

		fmt.Printf("shaKeyString: %s\n", shaKeyString)

		value, err := t.main.raftApi.Get(shaKeyString)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		fmt.Printf("value: %s\n", value)

		cpkTest := string(value[0:128])
		if cpkTest == t.main.walletInfo.PubKey {
			fmt.Println("Match!")
		} else {
			fmt.Println("NOT Match!")
		}

		info := value[128:]
		fmt.Println("Info: ", string(info))
		infoLen := len(info)

		base64FileBasename := info[0 : infoLen-16]
		fmt.Println("Base64: ", string(base64FileBasename))

		fileBasename, _ := base64.RawURLEncoding.DecodeString(string(base64FileBasename))
		fmt.Println("Filename: ", string(fileBasename))

		timeStamp := info[infoLen-16:]
		fmt.Println("Timestamp: ", string(timeStamp))

		timeStampDecode, _ := hex.DecodeString(timeStamp)
		timeStampInt64 := int64(binary.LittleEndian.Uint64(timeStampDecode))

		timeStampTime := time.Unix(timeStampInt64, 0)
		fmt.Println("Timestamp: ", timeStampTime.Format(time.RFC1123))

		f := File{
			Index:     len(t.db.Files) + 1,
			Name:      string(fileBasename),
			Timestamp: timeStampTime,
		}
		t.db.Files = append(t.db.Files, f)
		t.filesModel.RowInserted(len(t.db.Files) - 1)
	}
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

	// TODO: add Key/Value to Raft here
	t.saveFileToRaft(file)

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

func (t *StorageTab) saveFileToRaft(file string) {
	// TODO: Raft API
	// TODO: Key = SHA256 [ Base64(File.Basename) + File.Size + Timestamp ]
	fileBasename := filepath.Base(file)
	fmt.Printf("fileBasename: %s\n", fileBasename)

	fi, err := os.Stat(file)
	if err != nil || fi == nil {
		// TODO:
		fmt.Printf("os.Stat error: %s\n", err.Error())
	}

	// TODO: Key init
	key := []byte{}

	// Base64(File.Basename)
	base64FileBasename := make([]byte, base64.RawURLEncoding.EncodedLen(len(fileBasename)))
	base64.RawURLEncoding.Encode(base64FileBasename, []byte(fileBasename))
	key = append(key, base64FileBasename...)

	fmt.Printf("base64FileBasename: %x\n", base64FileBasename)
	fmt.Printf("base64FileBasename: %s\n", string(base64FileBasename))

	// File.Size, int64 to []byte
	fileSize := make([]byte, 8)
	binary.LittleEndian.PutUint64(fileSize, uint64(fi.Size()))
	fileSize = []byte(fmt.Sprintf("%x", fileSize))
	key = append(key, fileSize...)

	fmt.Printf("fi.Size(): %d\n", fi.Size())
	fmt.Printf("fileSize: %x\n", fileSize)
	fmt.Printf("fileSize: %s\n", string(fileSize))

	// Timestamp
	timeStamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeStamp, uint64(time.Now().Unix()))
	timeStamp = []byte(fmt.Sprintf("%x", timeStamp))
	key = append(key, timeStamp...)

	fmt.Printf("timeStamp: %x\n", timeStamp)
	fmt.Printf("timeStamp: %s\n", string(timeStamp))

	fmt.Printf("key: %x\n", key)
	fmt.Printf("key: %s\n", string(key))

	shaKey := sha256.Sum256(key)
	shaKeyString := sha256.Sum256([]byte(string(key)))

	fmt.Printf("shaKey: %x\n", shaKey[:])
	fmt.Printf("shaKey: %s\n", string(shaKey[:]))
	fmt.Printf("shaKeyString: %x\n", shaKeyString[:])
	fmt.Printf("shaKeyString: %s\n", string(shaKeyString[:]))

	shaKeyResult := fmt.Sprintf("%x", shaKeyString[:])

	// TODO: Value = CPK + Base64(File.Basename) + Timestamp
	if t.main.walletInfo == nil {
		// TODO:
		fmt.Printf("walletInfo is nil\n")
		return
	}
	value := []byte(t.main.walletInfo.PubKey)    // CPK
	value = append(value, base64FileBasename...) // Base64
	value = append(value, timeStamp...)          // Timestamp

	fmt.Printf("value: %x\n", value)
	fmt.Printf("value: %s\n", string(value))

	// main Key/Value Store
	t.main.raftApi.Set(shaKeyResult, string(value))

	//
	// FIXME: is it copy or pointer?
	cpkIndex := []byte(t.main.walletInfo.PubKey)

	// get last index from CPK Index Store
	// CPK + 00000000 (8 bytes)
	// FIXME: is it copy or pointer?
	cpkIndex0 := []byte(t.main.walletInfo.PubKey)
	index0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(index0, uint64(0))
	index0 = []byte(fmt.Sprintf("%x", index0))

	cpkIndex0 = append(cpkIndex0, index0...)

	cpkIndexLast, err := t.main.raftApi.Get(string(cpkIndex0))
	// TODO: check cpkIndexLast - len, value, etc
	if err != nil {
		return
	}
	fmt.Printf("cpkIndexLast: %s\n", cpkIndexLast)
	var cpkIndexLastInt64 int64
	if cpkIndexLast == "" {
		cpkIndexLastInt64 = int64(0)
	} else {
		cpkIndexLastDecode, _ := hex.DecodeString(cpkIndexLast)
		// TODO: check it
		cpkIndexLastInt64 = int64(binary.LittleEndian.Uint64(cpkIndexLastDecode))
	}

	fmt.Printf("cpkIndexLastInt64: %x\n", cpkIndexLastInt64)
	fmt.Printf("cpkIndexLastInt64: %s\n", strconv.Itoa(int(cpkIndexLastInt64)))

	cpkIndexNew := make([]byte, 8)
	binary.LittleEndian.PutUint64(cpkIndexNew, uint64(cpkIndexLastInt64+1))
	cpkIndexNew = []byte(fmt.Sprintf("%x", cpkIndexNew))

	fmt.Printf("cpkIndexNew: %x\n", cpkIndexNew)
	fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

	cpkIndex = append(cpkIndex, cpkIndexNew...)

	fmt.Printf("cpkIndex: %x\n", cpkIndex)
	fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

	// CPK Index Key/Value Store
	t.main.raftApi.Set(string(cpkIndex), shaKeyResult)

	t.main.raftApi.Set(string(cpkIndex0), string(cpkIndexNew))

	t.reloadFilesView()
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
