package nongui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const (
	baseURL        = "http://localhost:4000"
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

type BlockApi struct {
	Available bool
	http      *http.Client
}

func NewBlockApi() *BlockApi {
	blockApi := &BlockApi{
		Available: true,
		http:      &http.Client{},
	}

	blockApi.CheckApi()
	return blockApi
}

func (c *BlockApi) CheckApi() {
	_, err := c.Get("")
	if err != nil {
		//c.Available = false
	}
}

func (c *BlockApi) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("doRequest http.Do Error:", err.Error())
		c.Available = false
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("doRequest ioutil.ReadAll Error:", err.Error())
		c.Available = false
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}

	c.Available = true
	return body, nil
}

func (c *BlockApi) Get(api string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", baseURL+api, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *BlockApi) Post(api string, body io.Reader) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", baseURL+api, body)
	if err != nil {
		return nil, err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *BlockApi) GetWalletsList() ([]string, error) {
	data, err := c.Get("/wallets/list")
	if err != nil {
		return nil, err
	}
	//fmt.Println("data: ", data, err)

	var result WalletListResponse
	err = mapstructure.Decode(data, &result)
	//fmt.Println("result", result, err)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, err
	}

	return result.ListWallets, nil
}

func (c *BlockApi) GetWalletInfo(address string) (*WalletHashInfo, error) {
	data, err := c.Get("/wallet/" + address)
	if err != nil {
		return nil, err
	}
	//fmt.Println("data: ", data, err)

	var result WalletHashInfo
	err = mapstructure.Decode(data, &result)
	//fmt.Println("result", result, err)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, err
	}

	return &result, nil
}

func (c *BlockApi) PostWalletCreate(request *WalletCreateRequest) (*WalletCreateInfo, error) {
	j, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	data, err := c.Post("/wallet/new", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	//fmt.Println("data: ", data, err)

	var result WalletCreateInfo
	err = mapstructure.Decode(data, &result)
	//fmt.Println("result", result, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
