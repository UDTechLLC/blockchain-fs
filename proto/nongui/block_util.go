package nongui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	walletFilename = "wallet.json"
)

type WalletCreateRequest struct {
}

type WalletCreateInfo struct {
	Success bool
	Address string
	PrivKey string
	PubKey  string

	CpkZeroIndex string   `json:"-"`
	Raft         *RaftApi `json:"-"`
}

type WalletListResponse struct {
	Success     bool
	ListWallets []string
}

type WalletHashInfo struct {
	Success bool
	Credit  int
}

func (walletInfo *WalletCreateInfo) Save() (err error) {
	// Marshal
	walletJson, err := json.MarshalIndent(&walletInfo, "", "  ")
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

func (walletInfo *WalletCreateInfo) Load() error {
	// Read from file
	js, err := ioutil.ReadFile(walletFilename)
	if err != nil {
		fmt.Printf("Load %s: ReadFile: %#v\n", walletFilename, err)
		return err
	}

	// Unmarshal
	err = json.Unmarshal(js, &walletInfo)
	if err != nil {
		fmt.Printf("Failed to unmarshal wallet file")
		return err
	}

	return nil
}

func (walletInfo *WalletCreateInfo) IsEmpty() bool {
	if walletInfo.Address == "" {
		return true
	}
	return false
}

func (walletInfo *WalletCreateInfo) Update(info *WalletCreateInfo) {
	walletInfo.Success = info.Success
	walletInfo.Address = info.Address
	walletInfo.PrivKey = info.PrivKey
	walletInfo.PubKey = info.PubKey
}
