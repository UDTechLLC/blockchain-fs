package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const baseURL = "http://localhost:4000"

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

type BlockApi struct {
	http *http.Client
}

func NewBlockApi() *BlockApi {
	return &BlockApi{
		http: &http.Client{},
	}
}

func (c *BlockApi) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}
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
