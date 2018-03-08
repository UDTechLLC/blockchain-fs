package main

import (
	"fmt"
	"time"

	"bitbucket.org/udt/wizefs/proto/nongui"
	"github.com/leedark/ui"
)

type WalletTab struct {
	main             *MainWindow
	tab              *ui.Box
	timeTicker       *time.Ticker
	alreadyAvailable bool

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
	return makeTab
}

func (t *WalletTab) NewTimer(seconds int, action func()) {
	t.timeTicker = time.NewTicker(time.Duration(seconds) * time.Second)
	go action()
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
	t.walletsView.AppendTextColumn("Credit", 2)
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

func (t *WalletTab) ApiTicker() {
	for {
		select {
		case <-t.timeTicker.C:
			if t.alreadyAvailable != t.main.blockApi.Available {
				if t.main.blockApi.Available {
					t.reloadWalletsView()
					t.alreadyAvailable = true
				} else {
					// just clear wallets listview
					for i := 0; i < len(t.db.Wallets); i++ {
						t.walletsModel.RowDeleted(0)
					}
					t.db.Wallets = nil

					t.createWalletButton.Disable()
					t.alreadyAvailable = false
				}
			}
		}
	}
}

func (t *WalletTab) Init() {
	t.main.walletInfo = &nongui.WalletCreateInfo{
		Raft: t.main.raftApi,
	}
	err := t.main.walletInfo.Load()
	if err != nil {
		//fmt.Println("Load wallet error: ", err.Error())
		//fmt.Println("walletInfo is nil")
		t.main.storageTab.buttonEnabled(false)
	} else {
		// update controls
		t.updateWalletInfo()

		// mount!?
		t.main.MountStorage()
		t.main.storageTab.buttonEnabled(true)
	}

	// wallets list
	if t.main.blockApi.Available {
		t.reloadWalletsView()
		t.alreadyAvailable = true
	} else {
		t.createWalletButton.Disable()
		t.alreadyAvailable = false
	}

	t.NewTimer(60, t.ApiTicker)
}

func (t *WalletTab) updateWalletInfo() {
	if t.main.walletInfo.IsEmpty() {
		return
	}

	t.walletAddressEntry.SetText(t.main.walletInfo.Address)
	t.walletPrivateKeyEntry.SetText(t.main.walletInfo.PrivKey)

	idx := len(t.main.walletInfo.PubKey) / 2
	t.walletPublicKeyEntry1.SetText(t.main.walletInfo.PubKey[:idx])
	t.walletPublicKeyEntry2.SetText(t.main.walletInfo.PubKey[idx:])

	t.createWalletButton.Disable()
}

func (t *WalletTab) reloadWalletsView() {
	for i := 0; i < len(t.db.Wallets); i++ {
		t.walletsModel.RowDeleted(0)
	}
	t.db.Wallets = nil

	// wizeBlockAPI: Get Wallets List
	walletsList, err := t.main.blockApi.GetWalletsList()
	if err != nil {
		fmt.Println(err)
	}

	// Fill Wallets List View
	for _, address := range walletsList {
		// wizeBlockAPI: Get Wallet Info for every Wallet by address
		var credit int = -1
		info, err := t.main.blockApi.GetWalletInfo(address)
		if err == nil {
			if info.Success {
				credit = info.Credit
			}
		}

		w := Wallet{
			Index:   len(t.db.Wallets) + 1,
			Address: address,
			Credit:  credit,
		}
		t.db.Wallets = append(t.db.Wallets, w)
		t.walletsModel.RowInserted(len(t.db.Wallets) - 1)
	}
}

func (t *WalletTab) onCreateWalletClicked(button *ui.Button) {
	// wizeBlockAPI: Create Wallet
	walletInfo, err := t.main.blockApi.PostWalletCreate(&nongui.WalletCreateRequest{})
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

	walletInfo.CpkZeroIndex = "0"

	// save to wallet.json
	err = walletInfo.Save()
	if err != nil {
		//ui.MsgBoxError(t.main.window, "Error", "Save wallet error: "+err.Error())
		fmt.Println("Save wallet error: ", err.Error())
	}

	t.main.walletInfo.Update(walletInfo)
	t.main.storageTab.buttonEnabled(true)

	// update controls
	t.updateWalletInfo()

	// update table
	//t.reloadWalletsView()
	w := Wallet{
		Index:   len(t.db.Wallets) + 1,
		Address: walletInfo.Address,
		Credit:  0,
	}
	t.db.Wallets = append(t.db.Wallets, w)
	t.walletsModel.RowInserted(len(t.db.Wallets) - 1)

	t.afterCreateWallet()
}

func (t *WalletTab) afterCreateWallet() {
	// create single bucket (directory)
	cerr := nongui.CreateStorage(BucketOriginName)
	if cerr != nil {
		fmt.Println("Create bucket error:", cerr.Error())
		//ui.MsgBoxError(t.main.window, "Error", fmt.Sprintf("%v", cerr))
	}

	// mount?
	t.main.MountStorage()
}
