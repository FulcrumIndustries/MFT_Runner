package main

import (
	"fmt"
	"io"
	"io/fs"
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
