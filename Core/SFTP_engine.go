package Core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

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
