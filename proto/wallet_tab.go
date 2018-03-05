package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/leedark/ui"
)

const (
	walletFilename = "wallet.json"
)

type WalletTab struct {
	main *MainWindow
	tab  *ui.Box

	walletAddressLabel    *ui.Label
	walletPrivateKeyLabel *ui.Label
	walletPublicKeyLabel1 *ui.Label
	walletPublicKeyLabel2 *ui.Label

	walletAddressEntry    *ui.Label
	walletPrivateKeyEntry *ui.Label
	walletPublicKeyEntry1 *ui.Label
	walletPublicKeyEntry2 *ui.Label

	createWalletButton *ui.Button

	db           WalletDB
	walletsView  *ui.Table
	walletsModel *ui.TableModel
}

func NewWalletTab(mainWindow *MainWindow) *WalletTab {
	makeTab := &WalletTab{
		main: mainWindow,
	}
	makeTab.buildGUI()
	makeTab.init()
	return makeTab
}

func (t *WalletTab) buildGUI() {
	vbox := ui.NewVerticalBox()

	hbox1 := ui.NewHorizontalBox()

	vbox1a := ui.NewVerticalBox()
	t.walletAddressLabel = ui.NewLabel("Address:")
	t.walletPrivateKeyLabel = ui.NewLabel("Private Key:")
	t.walletPublicKeyLabel1 = ui.NewLabel("Public Key:")
	t.walletPublicKeyLabel2 = ui.NewLabel("")

	vbox1a.Append(t.walletAddressLabel, false)
	vbox1a.Append(t.walletPrivateKeyLabel, false)
	vbox1a.Append(t.walletPublicKeyLabel1, false)
	vbox1a.Append(t.walletPublicKeyLabel2, false)

	vbox1a.SetPadded(true)

	vbox1b := ui.NewVerticalBox()
	t.walletAddressEntry = ui.NewLabel("")
	t.walletPrivateKeyEntry = ui.NewLabel("")
	t.walletPublicKeyEntry1 = ui.NewLabel("")
	t.walletPublicKeyEntry2 = ui.NewLabel("")
	vbox1b.Append(t.walletAddressEntry, true)
	vbox1b.Append(t.walletPrivateKeyEntry, true)
	vbox1b.Append(t.walletPublicKeyEntry1, true)
	vbox1b.Append(t.walletPublicKeyEntry2, true)
	vbox1b.SetPadded(true)

	hbox1.Append(vbox1a, false)
	hbox1.Append(vbox1b, true)

	t.createWalletButton = ui.NewButton("Create Wallet")
	t.createWalletButton.OnClicked(t.onCreateWalletClicked)
	hbox1.Append(t.createWalletButton, false)

	hbox1.SetPadded(true)

	hbox2 := ui.NewHorizontalBox()

	listBox := ui.NewVerticalBox()
	t.walletsModel = ui.NewTableModel(&t.db)
	t.walletsView = ui.NewTable(t.walletsModel, ui.TableStyleMultiSelect)
	t.walletsView.AppendTextColumn("Index", 0)
	t.walletsView.AppendTextColumn("Address", 1)
	listBox.Append(t.walletsView, true)

	hbox2.Append(listBox, true)

	vbox.Append(hbox1, false)
	vbox.Append(ui.NewHorizontalSeparator(), false)
	vbox.Append(hbox2, true)
	vbox.SetPadded(true)

	t.tab = vbox
}

func (t *WalletTab) Control() ui.Control {
	return t.tab
}

func (t *WalletTab) init() {
	// load wallet.json or
	// TODO: wizeBlockAPI: get wallet info
	walletInfo, err := loadWalletInfo()
	if err != nil {
		//ui.MsgBoxError(t.main.window, "Error", "Load wallet error: "+err.Error())
		fmt.Println("Load wallet error: ", err.Error())
		//return
	}

	// update controls
	if walletInfo != nil {
		t.updateWalletInfo(walletInfo)
		t.main.walletInfo = walletInfo
	} else {
		//ui.MsgBoxError(t.main.window, "Error", "Wallet Info is nil")
		fmt.Println("Wallet Info is nil")
		//return
	}

	// wallets list
	t.reloadWalletsView()
}

func (t *WalletTab) updateWalletInfo(wallet *WalletCreateInfo) {
	if wallet == nil {
		return
	}

	t.walletAddressEntry.SetText(wallet.Address)
	t.walletPrivateKeyEntry.SetText(wallet.PrivKey)

	idx := len(wallet.PubKey) / 2
	t.walletPublicKeyEntry1.SetText(wallet.PubKey[:idx])
	t.walletPublicKeyEntry2.SetText(wallet.PubKey[idx:])

	t.createWalletButton.Disable()
}

func (t *WalletTab) reloadWalletsView() {
	for i := 0; i < len(t.db.Wallets); i++ {
		t.walletsModel.RowDeleted(0)
	}
	t.db.Wallets = nil

	result, err := t.main.blockApi.GetWalletsList()
	if err != nil {
		fmt.Println(err)
	}
	for _, value := range result {
		w := Wallet{
			Index:   len(t.db.Wallets) + 1,
			Address: value,
		}
		t.db.Wallets = append(t.db.Wallets, w)
		t.walletsModel.RowInserted(len(t.db.Wallets) - 1)
	}
}

func (t *WalletTab) onCreateWalletClicked(button *ui.Button) {
	// wizeBlockAPI: create wallet
	walletInfo, err := t.main.blockApi.PostWalletCreate(&WalletCreateRequest{})
	if err != nil {
		fmt.Println("Create wallet error: ", err.Error())
	}

	if walletInfo == nil {
		ui.MsgBoxError(t.main.window, "Error", "Wallet Info is nil")
		return
	}

	if !walletInfo.Success {
		ui.MsgBoxError(t.main.window, "Error", "Response is not success")
		return
	}

	// save to wallet.json
	err = saveWalletInfo(walletInfo)
	if err != nil {
		//ui.MsgBoxError(t.main.window, "Error", "Save wallet error: "+err.Error())
		fmt.Println("Save wallet error: ", err.Error())
	}

	t.main.walletInfo = walletInfo

	// update controls
	t.updateWalletInfo(walletInfo)

	// update table
	//t.reloadWalletsView()
	w := Wallet{
		Index:   len(t.db.Wallets) + 1,
		Address: walletInfo.Address,
	}
	t.db.Wallets = append(t.db.Wallets, w)
	t.walletsModel.RowInserted(len(t.db.Wallets) - 1)

	t.afterCreateWallet()
}

func (t *WalletTab) afterCreateWallet() {
	// create single bucket (directory)
	origin := BucketOriginName
	cerr := RunCommand("create", origin)
	if cerr != nil {
		fmt.Println("Create bucket error: ", cerr.Error())
		ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	}
}

//

func saveWalletInfo(wallet *WalletCreateInfo) (err error) {
	// Marshal
	walletJson, err := json.MarshalIndent(&wallet, "  ", "  ")
	if err != nil {
		return
	}

	// Write to file
	if walletJson != nil {
		err = ioutil.WriteFile(walletFilename, walletJson, 0644)
		if err != nil {
			fmt.Printf("Save %s: WriteFile: %#v\n", walletFilename, err)
			return
		}
	}

	return
}

func loadWalletInfo() (wallet *WalletCreateInfo, err error) {
	// Read from file
	js, err := ioutil.ReadFile(walletFilename)
	if err != nil {
		fmt.Printf("Load %s: ReadFile: %#v\n", walletFilename, err)
		return nil, err
	}

	// Unmarshal
	wallet = &WalletCreateInfo{}
	err = json.Unmarshal(js, &wallet)
	if err != nil {
		fmt.Printf("Failed to unmarshal wallet file")
		return nil, err
	}

	return wallet, nil
}
