package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

func createAccount(remoteURL string, username string, password string) (err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client http.Client = http.Client{Timeout: time.Duration(120) * time.Second, Transport: tr}
	byteData, err := os.ReadFile("../JSON/MFT_Runner_Account.json")
	if err != nil {
		fmt.Println(err)
	}
	jsonData := bytes.NewReader(byteData)
	req, err := http.NewRequest("POST", remoteURL, jsonData)
	if err != nil {
		fmt.Println(err)
		// continue
		// && req.Response.StatusCode != 201
	}
	req.SetBasicAuth(username, password)
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", "application/json")

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	// Check the response
	if res.StatusCode != 200 {
		err = fmt.Errorf("bad status: %s", res.Status)
	} else {
		fmt.Println("Account MFT_Runner Creation/Update Successful : " + res.Status)
	}
	return
}
