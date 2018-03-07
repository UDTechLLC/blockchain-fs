package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const raftBaseURL = "http://localhost:"

var raftPorts = []string{"11001", "11002", "11003"}

type RaftApi struct {
	Available  bool
	LeaderPort string
	http       *http.Client
}

func NewRaftApi() *RaftApi {
	raftApi := &RaftApi{
		Available: true,
		http:      &http.Client{},
	}

	raftApi.CheckApi()
	return raftApi
}

func (c *RaftApi) CheckApi() {
	// choosing the Leader
	for _, port := range raftPorts {
		c.LeaderPort = port
		//fmt.Println("port:", port)
		isLeader, err := c.Check(port)
		if err != nil {
			//c.Available = false
			//fmt.Println("error:", err)
			break
		}
		//fmt.Println("isLeader:", isLeader)
		if isLeader {

			break
		}
	}

	//fmt.Println("Leader:", c.LeaderPort)
	//c.Available = true
}

func (c *RaftApi) Check(port string) (bool, error) {
	data, err := c.Get("/check", false)
	if err != nil {
		return false, err
	}

	check, err := strconv.ParseBool(data["check"])
	return check, nil
}

func (c *RaftApi) doRequest(req *http.Request, checkStatus bool) ([]byte, error) {
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

	// FIXME
	if checkStatus {
		if 200 != resp.StatusCode {
			fmt.Println("doRequest StatusCode:", resp.StatusCode)
			return nil, fmt.Errorf("%s", body)
		}
	}

	c.Available = true
	return body, nil
}

func (c *RaftApi) Get(api string, checkStatus bool) (map[string]string, error) {
	//url := fmt.Sprintf(raftBaseURL+"/%s", api)
	url := bytes.NewBufferString(raftBaseURL)
	url.WriteString(c.LeaderPort)
	url.WriteString(api)

	//fmt.Println("url:", url.String())

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	bytes, err := c.doRequest(req, checkStatus)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *RaftApi) GetKey(key string) (string, error) {
	api := bytes.NewBufferString("/key/")
	api.WriteString(key)
	data, err := c.Get(api.String(), true)
	if err != nil {
		return "", err
	}

	return data[key], nil
}

func (c *RaftApi) SetKey(key, value string) error {
	m := map[string]string{}
	m[key] = value

	out := bytes.NewBuffer(nil)
	if err := json.NewEncoder(out).Encode(&m); err != nil {
		return err
	}

	url := bytes.NewBufferString(raftBaseURL)
	url.WriteString(c.LeaderPort)
	url.WriteString("/key")

	req, err := http.NewRequest("POST", url.String(), out)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, true)
	if err != nil {
		return err
	}

	return nil
}

func (c *RaftApi) DeleteKey(key string) error {
	url := bytes.NewBufferString(raftBaseURL)
	url.WriteString(c.LeaderPort)
	url.WriteString("/key/")
	url.WriteString(key)

	req, err := http.NewRequest("DELETE", url.String(), nil)
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

/*
func (c *RaftApi) Len() (string, error) {
	data, err := c.Get("/len", false)
	if err != nil {
		return "", err
	}

	return data["len"], nil
}

func (c *RaftApi) Clear() (string, error) {
	data, err := c.Get("/clear", false)
	if err != nil {
		return "", err
	}

	return data["clear"], nil
}

func (c *RaftApi) Dump() (string, error) {
	data, err := c.Get("/dump", false)
	if err != nil {
		return "", err
	}

	return data["dump"], nil
}
*/
