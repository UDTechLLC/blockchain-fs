package main

import (
	"fmt"

	"github.com/leedark/ui"
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

func (t *WalletTab) init() {
	// TODO: load wallet.json or wizeBlockAPI: get wallet info
	walletInfo := false

	// TODO: update controls
	if walletInfo {

	}

	// wallets list
	t.refreshWalletsView()
}

func (t *WalletTab) refreshWalletsView() {
	// TODO: refactor to just add new wallet to the end

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

func (t *WalletTab) Control() ui.Control {
	return t.tab
}

func (t *WalletTab) onCreateWalletClicked(button *ui.Button) {
	// wizeBlockAPI: create wallet
	result, err := t.main.blockApi.PostWalletCreate(&WalletCreateRequest{})
	if err != nil {
		fmt.Println(err)
	}

	if result == nil {
		ui.MsgBoxError(t.main.window, "Error", "Result is nil")
		return
	}

	// update controls
	t.walletAddressEntry.SetText(result.Address)
	t.walletPrivateKeyEntry.SetText(result.PrivKey)

	idx := len(result.PubKey) / 2
	t.walletPublicKeyEntry1.SetText(result.PubKey[:idx])
	t.walletPublicKeyEntry2.SetText(result.PubKey[idx:])

	t.createWalletButton.Disable()

	// update table
	//t.refreshWalletsView()

	w := Wallet{
		Index:   len(t.db.Wallets) + 1,
		Address: result.Address,
	}
	t.db.Wallets = append(t.db.Wallets, w)
	t.walletsModel.RowInserted(len(t.db.Wallets) - 1)
}
