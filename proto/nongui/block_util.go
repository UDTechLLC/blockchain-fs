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
	Success      bool
	Address      string
	PrivKey      string
	PubKey       string
	CpkZeroIndex string
}

type WalletListResponse struct {
	Success     bool
	ListWallets []string
}

type WalletHashInfo struct {
	Success bool
	Credit  int
}

func SaveWalletInfo(wallet *WalletCreateInfo) (err error) {
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

func LoadWalletInfo() (wallet *WalletCreateInfo, err error) {
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
