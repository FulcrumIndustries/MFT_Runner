package Core

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

func FTP_Operation(id string, user string, dial string, password string, sourcefile string, targetdir string, maxtimeout int, reqtype string) {
	fmt.Println("Logging in " + user + "/" + password + " @ " + dial)
	shortfilename := filepath.Base(sourcefile) + id
	c, err := ftp.Dial(dial, ftp.DialWithTimeout(time.Duration(maxtimeout)*time.Second))
	if err != nil {
		fmt.Println(err)
	}
	err = c.Login(user, password)
	if err != nil {
		fmt.Println(err)
	}
	if reqtype == "POST" {
		local, err := os.Open(sourcefile)
		if err != nil {
			return
		}
		defer local.Close()
		err = c.Stor(targetdir+"/"+shortfilename, local)
		if err != nil {
			fmt.Println(err)
		}
	} else if reqtype == "GET" {
		r, err := c.Retr(targetdir + "/" + filepath.Base(sourcefile))
		if err != nil {
			fmt.Println(err)
		}
		defer r.Close()
		buf, err := io.ReadAll(r)
		if err != nil {
			fmt.Println(err)
		}
		err2 := os.WriteFile(sourcefile+id, buf, fs.ModeAppend)
		if err2 != nil {
			fmt.Println(err)
		}
	}
	if err := c.Quit(); err != nil {
		fmt.Println(err)
	}
}

func FTPUpload(filePath, remoteName string, config *TestConfig, workerID, transferID int) error {
	log.Printf("Worker %d Transfer %d: Starting FTP upload", workerID, transferID)
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		log.Printf("Worker %d Transfer %d: Connection failed - %v", workerID, transferID, err)
		return err
	}
	defer conn.Quit()

	if err := conn.Login(config.Username, config.Password); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := filepath.Join(config.RemotePath, remoteName)
	return conn.Stor(remotePath, file)
}

func FTPDownload(filePath, remoteName string, config *TestConfig, workerID int) error {
	log.Printf("Worker %d: Initiating FTP download to %s:%d (Timeout: %ds)",
		workerID, config.Host, config.Port, config.Timeout)

	client, err := ftp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port),
		ftp.DialWithTimeout(time.Duration(config.Timeout)*time.Second),
		ftp.DialWithDebugOutput(os.Stdout))

	if err != nil {
		log.Printf("Worker %d: Connection failed - %v", workerID, err)
		return err
	}

	defer client.Quit()

	if err := client.Login(config.Username, config.Password); err != nil {
		log.Printf("Worker %d: Login failed - %v", workerID, err)
		return err
	}

	r, err := client.Retr(config.RemotePath)
	if err != nil {
		log.Printf("Worker %d: File retrieval failed - %v", workerID, err)
		return err
	}
	defer r.Close()

	file, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	_, err = io.Copy(file, r)
	return err
}
