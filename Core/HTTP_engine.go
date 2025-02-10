package Core

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func multipartCall(id string, maxtimeout int, remoteURL string, username string, password string, reqtype string, filename string) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client http.Client = http.Client{Timeout: time.Duration(maxtimeout) * time.Second, Transport: tr}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"file":  mustOpen(filename), // lets assume its this file
		"other": strings.NewReader("MFT_Runner"),
	}
	if reqtype == "POST" {
		fmt.Println("Sending " + filename + "...")
		err := Upload(id, client, username, password, remoteURL, values)
		if err != nil {
			fmt.Println(err)
		}
	} else if reqtype == "GET" {
		shortfilename := filepath.Base(filename)
		targetfile := filename + id
		fmt.Println("Downloading " + remoteURL + "%2F" + shortfilename + "...")
		err := Download(id, client, targetfile, shortfilename, username, password, remoteURL, values)
		if err != nil {
			// panic(err)
			fmt.Println(err)
		}
	}
}

func Upload(id string, client http.Client, username string, password string, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()+id); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	// w.Close()
	w.Close()
	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		fmt.Println(err)
		// && req.Response.StatusCode != 201
	}
	req.SetBasicAuth(username, password)
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	// Check the response
	if res.StatusCode != 201 && res.StatusCode != 200 {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func Download(id string, client http.Client, targetfile string, downloadfilename string, username string, password string, url string, values map[string]io.Reader) (err error) {
	// Get the data
	req, err := http.NewRequest("GET", url+"%2F"+downloadfilename, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.SetBasicAuth(username, password)
	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}
	fmt.Printf("Download result : %v\n", res.Status)
	// Check the response
	if res.StatusCode != 201 && res.StatusCode != 200 {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	// return
	f, _ := os.OpenFile(targetfile, os.O_CREATE|os.O_WRONLY, 0644)
	io.Copy(io.Writer(f), res.Body)
	// fmt.Printf("body: %v\n", res.Body)
	defer res.Body.Close()
	return nil
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

type Post struct {
	Userid string `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func HTTPUpload(filePath string, config *TestConfig) error {
	url := fmt.Sprintf("http://%s:%d%s", config.Host, config.Port, config.RemotePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err

	}
	defer file.Close()

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Post(url, "application/octet-stream", file)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}
	return nil
}

func HTTPDownload(filePath string, config *TestConfig) error {
	url := fmt.Sprintf("http://%s:%d%s", config.Host, config.Port, config.RemotePath)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	file, err := os.CreateTemp("", "http-download-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	_, err = io.Copy(file, resp.Body)
	return err
}
