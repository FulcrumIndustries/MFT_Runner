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

func SFTPUpload(filePath, remoteName string, config *TestConfig) error {
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

	srcFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	remotePath := filepath.ToSlash(filepath.Join(config.RemotePath, remoteName))

	fmt.Println("Uploading " + remotePath)

	// Create parent directories
	remoteDir := filepath.ToSlash(filepath.Dir(remotePath))
	fmt.Println("Creating remote directory " + remoteDir)
	if err := client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("create remote directory: %w", err)
	}

	fmt.Println("Creating remote file " + remotePath)
	dstFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
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

func SFTPDownload(pattern string, localPath string, config *TestConfig) error {
	// Ensure remote path uses forward slashes
	remotePath := filepath.ToSlash(config.RemotePath)
	fmt.Printf("Searching in remote path: %s\n", remotePath)

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

	// List files in remote directory
	files, err := client.ReadDir(remotePath)
	if err != nil {
		fmt.Printf("Error listing directory: %v\n", err)
		return err
	}

	// Find files matching pattern
	matched := false
	fmt.Println("Files: ", files)
	fmt.Println("Pattern: ", pattern)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), pattern) {
			remoteFilePath := filepath.ToSlash(filepath.Join(remotePath, file.Name()))
			downloadPath := filepath.Join(config.LocalPath, file.Name())
			fmt.Printf("Downloading %s to %s\n", remoteFilePath, downloadPath)
			if err := downloadFile(client, remoteFilePath, downloadPath); err != nil {
				fmt.Printf("Download error: %v\n", err)
				return err
			}
			matched = true
		}
	}

	if !matched {
		return fmt.Errorf("no files found matching pattern: %s", pattern)
	}
	return nil
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
