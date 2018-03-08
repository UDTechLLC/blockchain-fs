package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"bitbucket.org/udt/wizefs/internal/util"
	"bitbucket.org/udt/wizefs/proto/nongui"
	jwt "github.com/dgrijalva/jwt-go"
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
		case <-t.timeTicker.C: // t := <-t.timeTicker.C:
			//fmt.Println("Tick at", t)
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
	// chech wallet info existing
	if t.main.walletInfo == nil {
		fmt.Printf("walletInfo is nil\n")
		return
	}

	// clear db and Model
	for i := 0; i < len(t.db.Files); i++ {
		t.filesModel.RowDeleted(0)
	}
	t.db.Files = nil

	// get last CPKIndex = CPK + 0000000000000000 (8 bytes)
	// prepare key for Get
	cpkIndex0 := []byte(t.main.walletInfo.PubKey)
	index0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(index0, uint64(0))
	index0 = []byte(hex.EncodeToString(index0))
	//fmt.Printf("index0: %s\n", string(index0))

	cpkIndex0 = append(cpkIndex0, index0...)

	// Get last CPIIndex
	// TODO: try to use wallet.CpkZeroIndex instead of this
	cpkIndexLast, err := t.main.raftApi.GetKey(string(cpkIndex0))
	if err != nil {
		fmt.Printf("Try to get last CPKIndex was failed with error: %s\n", err.Error())
		return
	}
	//fmt.Printf("cpkIndexLast: %s\n", cpkIndexLast)

	// casting string to int64
	var cpkIndexLastInt64 int64
	if cpkIndexLast == "" {
		// if last CPKIndex is not existing, just set it to 0
		cpkIndexLastInt64 = int64(0)
	} else {
		cpkIndexLastDecode, err := hex.DecodeString(cpkIndexLast)
		if err != nil {
			fmt.Printf("Try to decode last CPKIndex was failed with error: %s\n", err.Error())
			return
		}
		cpkIndexLastInt64 = int64(binary.LittleEndian.Uint64(cpkIndexLastDecode))
	}
	//fmt.Printf("cpkIndexLastInt64: %d\n", cpkIndexLastInt64)

	// list cycle
	var index int64 = 0
	for index < cpkIndexLastInt64 {
		index++

		cpkIndex := []byte(t.main.walletInfo.PubKey)

		cpkIndexNew := make([]byte, 8)
		binary.LittleEndian.PutUint64(cpkIndexNew, uint64(index))
		cpkIndexNew = []byte(hex.EncodeToString(cpkIndexNew))
		//fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

		cpkIndex = append(cpkIndex, cpkIndexNew...)
		//fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

		shaKeyString, err := t.main.raftApi.GetKey(string(cpkIndex))
		if err != nil {
			// TODO:
			fmt.Println("Error when getting SHA256:", err)
			continue
		}
		// CPKIndex is absent
		if shaKeyString == "" {
			//fmt.Printf("Skip this index: %d\n", index)
			continue
		}
		//fmt.Printf("shaKeyString: %s\n", shaKeyString)

		value, err := t.main.raftApi.GetKey(shaKeyString)
		if err != nil {
			// TODO:
			fmt.Println("Error when getting FileInfo:", err)
			continue
		}
		//fmt.Printf("value: %s\n", value)

		// CPK
		cpkTest := string(value[0:128])
		if cpkTest != t.main.walletInfo.PubKey {
			// TODO:
			fmt.Println("CPK was not matched!")
			continue
		}

		// Info (Base64(File.Basename) + Timestamp)
		info := value[128:]
		infoLen := len(info)

		// TODO: Base64 signed with CSK - Parse & Verify with CPK
		signed64 := info[0 : infoLen-16]
		fmt.Printf("signed64: %s\n", signed64)
		basename64, err := t.ecdsaParseVerifyWithCPK(signed64)
		if err != nil {
			// if we got error then we don't add this file to list
			continue
		}
		fmt.Printf("basename64: %s\n", basename64)

		fileBasename, err := base64.RawURLEncoding.DecodeString(string(basename64))
		if err != nil {
			// if we got error then we don't add this file to list
			continue
		}
		fmt.Println("Filename:", string(fileBasename))

		timeStamp := info[infoLen-16:]
		timeStampDecode, err := hex.DecodeString(timeStamp)
		if err != nil {
			// if we got error then we don't add this file to list
			continue
		}
		timeStampInt64 := int64(binary.LittleEndian.Uint64(timeStampDecode))
		timeStampTime := time.Unix(timeStampInt64, 0)
		//fmt.Println("Timestamp: ", timeStampTime.Format(time.RFC1123))

		f := File{
			Index:     len(t.db.Files) + 1,
			RaftIndex: int(index),
			Name:      string(fileBasename),
			Timestamp: timeStampTime,
			shaKey:    shaKeyString,
			cpkIndex:  string(cpkIndex),
		}
		t.db.Files = append(t.db.Files, f)
		t.filesModel.RowInserted(len(t.db.Files) - 1)
	}

	// TODO: how to select row?
	//if len(t.db.Files) > 0 {
	//	t.filesModel.RowChanged(0)
	//}

	//
	//fmt.Println("len:", len(t.db.Files))
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
	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Open file: " + filename)
	})

	origin := BucketOriginName

	// mount
	//cerr := RunCommand("mount", origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//}

	// TODO: we must wait until mount finishes its actions
	// TODO: check ORIGIN? every 100 milliseconds
	//time.Sleep(500 * time.Millisecond)
	//ui.QueueMain(func() {
	//	t.logMessage("Mount bucket " + origin)
	//})

	putSuccess := false
	cerr := nongui.RunCommand("put", filename, origin)
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

	// unmount
	//cerr = RunCommand("unmount", origin)
	//if cerr != nil {
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//}

	//ui.QueueMain(func() {
	//	t.logMessage("Unmount bucket " + origin)
	//})

	// TODO: add Key/Value to Raft here
	if putSuccess {
		t.saveFileToRaft(filename)
		t.reloadFilesView()
	}

	ui.QueueMain(func() {
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) ecdsaSignWithCSK(basename64 string) (string, error) {
	// TODO: check walletInfo and Keys
	if t.main.walletInfo == nil {
		return "", fmt.Errorf("Wallet Info is nil. We can't get Keys.")
	}
	if len(t.main.walletInfo.PrivKey) != 64 {
		return "", fmt.Errorf("Private Key is wrong!")
	}
	if len(t.main.walletInfo.PubKey) != 128 {
		return "", fmt.Errorf("Public Key is wrong!")
	}
	ECDSAKeyD := t.main.walletInfo.PrivKey
	ECDSAKeyX := t.main.walletInfo.PubKey[:64]
	ECDSAKeyY := t.main.walletInfo.PubKey[64:]

	keyD := new(big.Int)
	keyX := new(big.Int)
	keyY := new(big.Int)
	keyD.SetString(ECDSAKeyD, 16)
	keyX.SetString(ECDSAKeyX, 16)
	keyY.SetString(ECDSAKeyY, 16)

	//fmt.Println("basename64:", basename64)
	claims := &jwt.MapClaims{
		"basename64": basename64,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keyX,
		Y:     keyY,
	}

	privateKey := ecdsa.PrivateKey{D: keyD, PublicKey: publicKey}
	signed64, err := token.SignedString(&privateKey)
	if err != nil {
		return "", err
	}
	//fmt.Println("signed64:", signed64)

	return signed64, nil
}

func (t *StorageTab) ecdsaParseVerifyWithCPK(signed64 string) (string, error) {
	// TODO: check walletInfo and Keys
	if t.main.walletInfo == nil {
		return "", fmt.Errorf("Wallet Info is nil. We can't get Keys")
	}
	if len(t.main.walletInfo.PubKey) != 128 {
		return "", fmt.Errorf("Public Key is wrong!")
	}
	ECDSAKeyX := t.main.walletInfo.PubKey[:64]
	ECDSAKeyY := t.main.walletInfo.PubKey[64:]

	keyX := new(big.Int)
	keyY := new(big.Int)
	keyX.SetString(ECDSAKeyX, 16)
	keyY.SetString(ECDSAKeyY, 16)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keyX,
		Y:     keyY,
	}

	token, err := jwt.Parse(signed64, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return &publicKey, nil
	})
	// TODO: err != nil
	if err != nil {
		return "", err
	}

	var basename64 string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		basename64 = claims["basename64"].(string)
	} else {
		// TODO: not ok?
	}

	return basename64, nil
}

func (t *StorageTab) saveFileToRaft(file string) {
	basename := filepath.Base(file)
	//fmt.Printf("basename: %s\n", basename)
	fi, err := os.Stat(file)
	if err != nil || fi == nil {
		fmt.Printf("os.Stat error: %s\n", err.Error())
		return
	}

	// Key = SHA256 [ Base64(File.Basename) + File.Size + Timestamp ]
	key := []byte{}

	// Base64(File.Basename)
	basename64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(basename)))
	base64.RawURLEncoding.Encode(basename64, []byte(basename))
	//fmt.Printf("basename64: %s\n", string(basename64))
	key = append(key, basename64...)

	// File.Size, int64 to []byte
	fileSize := make([]byte, 8)
	binary.LittleEndian.PutUint64(fileSize, uint64(fi.Size()))
	fileSize = []byte(hex.EncodeToString(fileSize))
	//fmt.Printf("fi.Size(): %d\n", fi.Size())
	//fmt.Printf("fileSize: %s\n", string(fileSize))
	key = append(key, fileSize...)

	// Timestamp
	timeStamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeStamp, uint64(time.Now().Unix()))
	timeStamp = []byte(hex.EncodeToString(timeStamp))
	//fmt.Printf("timeStamp: %s\n", string(timeStamp))
	key = append(key, timeStamp...)

	//fmt.Printf("key: %s\n", string(key))

	shaKey := sha256.Sum256(key)
	////shaKey2 := sha256.Sum256([]byte(string(key)))
	//fmt.Printf("shaKey: %x\n", shaKey[:])
	shaKeyString := hex.EncodeToString(shaKey[:])

	// Value = CPK + Base64(File.Basename) + Timestamp
	value := []byte(t.main.walletInfo.PubKey) // CPK

	// TODO: Base64 signed with CSK
	fmt.Printf("basename64: %s\n", string(basename64))
	signed64, err := t.ecdsaSignWithCSK(string(basename64))
	if err != nil {
		// if we got error then we don't save this file to Raft
		return
	}
	fmt.Printf("signed64: %s\n", signed64)
	basename64test, err := t.ecdsaParseVerifyWithCPK(signed64)
	if err != nil {
		// if we got error then we don't save this file to Raft
		return
	}
	fmt.Printf("basename64test: %s\n", basename64test)

	value = append(value, signed64...)  // Base64
	value = append(value, timeStamp...) // Timestamp
	//fmt.Printf("value: %s\n", string(value))

	////

	// CPK Index
	cpkIndex := []byte(t.main.walletInfo.PubKey)

	// get last CPKIndex = CPK + 0000000000000000 (8 bytes)
	// prepare key for Get
	cpkIndex0 := []byte(t.main.walletInfo.PubKey)
	index0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(index0, uint64(0))
	index0 = []byte(hex.EncodeToString(index0))
	//fmt.Printf("index0: %s\n", string(index0))

	cpkIndex0 = append(cpkIndex0, index0...)

	// Get last CPIIndex
	// TODO: try to use wallet.CpkZeroIndex instead of this
	cpkIndexLast, err := t.main.raftApi.GetKey(string(cpkIndex0))
	if err != nil {
		fmt.Printf("Try to get last CPKIndex was failed with error: %s\n", err.Error())
		return
	}
	//fmt.Printf("cpkIndexLast: %s\n", cpkIndexLast)

	// casting string to int64
	var cpkIndexLastInt64 int64
	if cpkIndexLast == "" {
		// if last CPKIndex is not existing, just set it to 0
		cpkIndexLastInt64 = int64(0)
	} else {
		cpkIndexLastDecode, err := hex.DecodeString(cpkIndexLast)
		if err != nil {
			fmt.Printf("Try to decode last CPKIndex was failed with error: %s\n", err.Error())
			return
		}
		cpkIndexLastInt64 = int64(binary.LittleEndian.Uint64(cpkIndexLastDecode))
	}
	//fmt.Printf("cpkIndexLastInt64: %d\n", cpkIndexLastInt64)

	// create new CPKIndex
	cpkIndexNew := make([]byte, 8)
	binary.LittleEndian.PutUint64(cpkIndexNew, uint64(cpkIndexLastInt64+1))
	cpkIndexNew = []byte(hex.EncodeToString(cpkIndexNew))
	//fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

	cpkIndex = append(cpkIndex, cpkIndexNew...)
	//fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

	// Main Key/Value Store
	t.main.raftApi.SetKey(shaKeyString, string(value))

	// CPKIndex Key/Value Store
	t.main.raftApi.SetKey(string(cpkIndex), shaKeyString)
	t.main.raftApi.SetKey(string(cpkIndex0), string(cpkIndexNew))

	// TODO: save last cpkIndex to wallet
	t.main.walletInfo.CpkZeroIndex = string(cpkIndexNew)
}

func (t *StorageTab) getFile(source string, destination string) {
	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Save file: " + source + " to " + destination)
	})

	origin := BucketOriginName

	// mount
	//cerr := RunCommand("mount", origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//}

	//time.Sleep(500 * time.Millisecond)
	//ui.QueueMain(func() {
	//	t.logMessage("Mount bucket " + origin)
	//})

	cerr := nongui.RunCommand("xget", source, origin, destination)
	if cerr != nil {
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
		ui.QueueMain(func() {
			t.logMessage("Put error: " + cerr.Error())
		})
	} else {
		//time.Sleep(500 * time.Millisecond)
		ui.QueueMain(func() {
			t.logMessage("Get file [" + source + "] from bucket " + origin +
				" to path [" + destination + "]")
		})
	}

	// unmount
	//cerr = RunCommand("unmount", origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//}

	ui.QueueMain(func() {
		//t.logMessage("Unmount bucket " + origin)
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) removeFile(file File) {
	ui.QueueMain(func() {
		t.buttonEnabled(false)
		t.logMessage("Open file: " + file.Name)
	})

	origin := BucketOriginName

	// mount
	//cerr := RunCommand("mount", origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.putFileButton.Enable()
	//		t.logMessage("Mount error: " + cerr.Error())
	//	})
	//	return
	//}

	//time.Sleep(500 * time.Millisecond)
	//ui.QueueMain(func() {
	//	t.logMessage("Mount bucket " + origin)
	//})

	removeSuccess := false
	cerr := nongui.RunCommand("remove", file.Name, origin)
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

	// unmount
	//cerr = RunCommand("unmount", origin)
	//if cerr != nil {
	//	//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	//	ui.QueueMain(func() {
	//		t.logMessage("Unmount error: " + cerr.Error())
	//	})
	//}

	//ui.QueueMain(func() {
	//	t.logMessage("Unmount bucket " + origin)
	//})

	// TODO: add Key/Value to Raft here
	if removeSuccess {
		t.removeFileFromRaft(file)
		t.reloadFilesView()
	}

	ui.QueueMain(func() {
		t.buttonEnabled(true)
	})
}

func (t *StorageTab) removeFileFromRaft(file File) {
	fmt.Println("shaKey:", file.shaKey)
	fmt.Println("cpkIndex:", file.cpkIndex)

	t.main.raftApi.DeleteKey(file.shaKey)
	t.main.raftApi.DeleteKey(file.cpkIndex)

	// TODO: we can fix cpkIndex0 for last cpkIndex?
	lastIndex := t.main.walletInfo.PubKey + t.main.walletInfo.CpkZeroIndex
	if file.cpkIndex == lastIndex {
		// we can update CpkZeroIndex here!
		fmt.Println("We can update CpkZeroIndex here")
	}

	// TODO: or we can rebuild index, yeah!

	// TODO: or we can save file count with cpkIndex!?
	// so cpkIndex will have 2 values: last cpkIndex and file count
}

func (t *StorageTab) onPutFileClicked(button *ui.Button) {
	// chech wallet info existing
	if t.main.walletInfo == nil {
		fmt.Printf("walletInfo is nil\n")
		return
	}

	file := ui.OpenFile(t.main.window, util.UserHomeDir()+"/*.*")
	//fmt.Println("file: ", file)

	if file == "" {
		ui.MsgBoxError(t.main.window, "Error",
			fmt.Sprintf("Please, select file for putting it to storage"))
		return
	}

	go t.putFile(file)
}

func (t *StorageTab) onGetFileClicked(button *ui.Button) {
	// chech wallet info existing
	if t.main.walletInfo == nil {
		fmt.Printf("walletInfo is nil\n")
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
	fmt.Println("filename:", filename)

	// save file to path
	filenameSave := ui.SaveFile(t.main.window, util.UserHomeDir()+"/"+filename)
	fmt.Println("filenameSave: ", filenameSave)
	if filenameSave == "" {
		ui.MsgBoxError(t.main.window, "Error",
			fmt.Sprintf("Please, select file for gettig it from storage"))
		return
	}

	filePath := filepath.Dir(filenameSave)
	fmt.Println("filePath:", filePath)

	go t.getFile(filename, filenameSave)
}

func (t *StorageTab) onRemoveFileClicked(button *ui.Button) {
	// chech wallet info existing
	if t.main.walletInfo == nil {
		fmt.Printf("walletInfo is nil\n")
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
	fmt.Println("filename:", filename)

	go t.removeFile(dbitem)
}
