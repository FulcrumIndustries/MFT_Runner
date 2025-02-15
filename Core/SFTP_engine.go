package Core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
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

func SFTPUpload(localPath, remoteName string, config *TestConfig) error {
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	remotePath := filepath.ToSlash(filepath.Join(config.RemotePath, remoteName))
	if !strings.HasSuffix(config.RemotePath, "/") {
		return fmt.Errorf("remote path must end with '/'")
	}
	remoteDir := filepath.ToSlash(filepath.Dir(remotePath))

	if err := client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", remoteDir, err)
	}

	fmt.Printf("Uploading %s to %s\n", filepath.Base(localPath), remotePath)

	dstFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}

	fmt.Println("Writing file content")
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("write file content: %w", err)
	}

	fmt.Println("Closing remote file")
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("close remote file: %w", err)
	}

	fmt.Println("Done")
	return nil
}

func SFTPDownload(remoteName, localPath string, config *TestConfig) error {
	// Create SFTP client connection
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	// Use direct file path instead of pattern matching
	remotePath := filepath.ToSlash(filepath.Join(config.RemotePath, remoteName))
	fmt.Printf("Downloading %s to %s\n", remotePath, localPath)
	return downloadFile(client, remotePath, localPath)
}

func downloadFile(client *sftp.Client, remotePath, localPath string) error {
	srcFile, err := client.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, srcFile)
	return err
}
