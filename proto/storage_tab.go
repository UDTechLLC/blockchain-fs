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
