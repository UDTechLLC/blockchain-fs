package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	//"os/signal"
	"testing"
	"time"

	co "bitbucket.org/udt/wizefs/rest/controllers"
)

const (
	serviceAddr string = ":13000"
	baseURL     string = "http://localhost:13000"
)

//var terminate chan os.Signal

// TODO: REST API Service
func startService() {
	h := NewService(serviceAddr)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	log.Println("rest started successfully")

	//terminate = make(chan os.Signal, 1)
	//signal.Notify(terminate, os.Interrupt)
	//<-terminate
	//log.Println("rest exiting")
}

// TODO: REST API Client
type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{},
	}
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
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
	//if 200 != resp.StatusCode {
	//	fmt.Println("doRequest resp.StatusCode:", resp)
	//	return nil, fmt.Errorf("%s", body)
	//}

	return body, nil
}

func (c *Client) Get() {
	// TODO
}

func (c *Client) Post(api string, body io.Reader) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", baseURL+api, body)
	if err != nil {
		fmt.Printf("ErrorP1: %v\n", err)
		return nil, err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		fmt.Printf("ErrorP2: %v\n", err)
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Printf("ErrorP3: %v\n", err)
		return nil, err
	}
	return data, nil
}

func (c *Client) Delete(api string) (map[string]interface{}, error) {
	req, err := http.NewRequest("DELETE", baseURL+api, nil)
	if err != nil {
		fmt.Println("ErrorD1:", err.Error())
		return nil, err
	}
	bytes, err := c.doRequest(req)
	if err != nil {
		fmt.Printf("ErrorD2: %v\n", err)
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Printf("ErrorD3: %v\n", err)
		return nil, err
	}
	return data, nil
}

// TODO: TestFullCircle
func TestFullCircle(t *testing.T) {
	// start REST service
	startService()

	// create REST client
	client := NewClient()

	origin := "RESTTest"

	// CREATE
	t.Logf("Request Create Bucket %s", origin)
	bucketResource := &co.BucketResource{
		Data: co.BucketModel{
			Origin: origin,
		},
	}
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(&bucketResource); err != nil {
		t.Fatalf("Error1: %v", err)
	}
	resp, err := client.Post("/buckets", body)
	if err != nil {
		t.Fatalf("Error2: %v", err)
	}
	t.Logf("Response: %v", resp)

	// MOUNT
	t.Logf("Request Mount Bucket %s", origin)
	resp, err = client.Post("/buckets/"+origin+"/mount", nil)
	if err != nil {
		t.Fatalf("Error2: %v", err)
	}
	t.Logf("Response: %v", resp)

	time.Sleep(1000 * time.Millisecond)

	// UNMOUNT
	t.Logf("Request Unmount Bucket %s", origin)
	resp, err = client.Post("/buckets/"+origin+"/unmount", nil)
	if err != nil {
		t.Fatalf("Error2: %v", err)
	}
	t.Logf("Response: %v", resp)

	// DELETE
	t.Logf("Request Delete Bucket %s", origin)
	resp, err = client.Delete("/buckets/" + origin)
	if err != nil {
		t.Fatalf("Error2: %v", err)
	}
	t.Logf("Response: %v", resp)
}
