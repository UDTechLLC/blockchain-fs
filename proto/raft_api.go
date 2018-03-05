package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"os"
)

const raftBaseURL = "http://localhost:11001"

type RaftApi struct {
	http *http.Client
}

func NewRaftApi() *RaftApi {
	return &RaftApi{
		http: &http.Client{},
	}
}

func (c *RaftApi) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("doRequest http.Do Error:", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("doRequest ioutil.ReadAll Error:", err.Error())
		return nil, err
	}
	
	if 200 != resp.StatusCode {
		fmt.Println("doRequest StatusCode:", resp.StatusCode)
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}

func (c *RaftApi) Get(key string) (string, error) {
	url := fmt.Sprintf(raftBaseURL+"/key/%s", key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	
	bytes, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	
	var data map[string]string
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}
	return data[key], nil
}

func (c *RaftApi) Set(key, value string) error {
	m := map[string]string{}
	m[key] = value

	out := bytes.NewBuffer(nil)
	if err := json.NewEncoder(out).Encode(&m); err != nil {
	}

	req, err := http.NewRequest("POST", raftBaseURL+"/key", out)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	//var data map[string]interface{}
	//err = json.Unmarshal(bytes, &data)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (c *RaftApi) Delete(key string) error {
	url := fmt.Sprintf(raftBaseURL+"/key/%s", key)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("Delete Error1:", err.Error())
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("Delete http.Do Error:", err.Error())
		return err
	}
	defer resp.Body.Close()
	
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Delete ioutil.ReadAll Error:", err.Error())
		return err
	}
	return nil
}

func (c *RaftApi) Len() (string, error) {
	url := fmt.Sprintf(raftBaseURL + "/len")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	var data map[string]string
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}
	return data["len"], nil
}

func (c *RaftApi) Clear() (string, error) {
	url := fmt.Sprintf(raftBaseURL + "/clear")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	var data map[string]string
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}
	return data["clear"], nil
}

func (c *RaftApi) Dump() (string, error) {
	url := fmt.Sprintf(raftBaseURL + "/dump")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	var data map[string]string
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}
	return data["dump"], nil
}
