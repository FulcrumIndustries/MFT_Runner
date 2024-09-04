package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

func SSH_Operation(id string, user string, ipaddr string, port uint, password string, sourcefile string, targetdir string, maxtimeout int, reqtype string) {
	fmt.Println("Logging in " + user + "/" + password + " @ " + ipaddr)
	shortfilename := filepath.Base(sourcefile) + id

	client, err := goph.NewConn(&goph.Config{
		User:     user,
		Addr:     ipaddr,
		Port:     port,
		Auth:     goph.Password(password),
		Timeout:  time.Duration(float64(maxtimeout) * float64(time.Second)),
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err == nil {
		// Defer closing the network connection.
		defer client.Close()
		fmt.Println("Connection OK")
		if reqtype == "POST" {
			fmt.Println("Sending " + targetdir + "/" + shortfilename + "...")
			err = client.Upload(sourcefile, targetdir+"/"+shortfilename)
			if err != nil {
				// log.Fatal(err)
				fmt.Println(err)
			}
		} else if reqtype == "GET" {
			fmt.Println("Downloading " + targetdir + "/" + filepath.Base(sourcefile) + "... to " + sourcefile + id)
			err = client.Download(targetdir+"/"+filepath.Base(sourcefile), sourcefile+id)
			if err != nil {
				// log.Fatal(err)
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("ID#" + id + " Error while handling " + sourcefile + " >> " + targetdir + "/" + shortfilename)
		// log.Fatal(err)
		fmt.Println(err)
	}

	// sftp, err := client.NewSftp()
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	fmt.Println("Doing the SFTP")
	// }
	// file, err := sftp.Create(targetdir + "/" + shortfilename)
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	fmt.Println("Writing file")
	// }
	// file.Write([]byte(`Hello world`))
	// file.Close()
	// Execute your command.
	// err := client.Upload("/path/to/local/file", "/path/to/remote/file")
	// out, err := client.Run("ls /Upload/")
	// err := client.Download("/path/to/remote/file", "/path/to/local/file")
	// out, err := client.Run("bash -c 'printenv'")
}
