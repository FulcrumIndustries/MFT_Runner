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

type FTPConnPool struct {
	pool   chan *ftp.ServerConn
	config *TestConfig
}

func NewFTPConnPool(config *TestConfig, max int) *FTPConnPool {
	return &FTPConnPool{
		pool:   make(chan *ftp.ServerConn, max),
		config: config,
	}
}

func (p *FTPConnPool) Get() (*ftp.ServerConn, error) {
	select {
	case conn := <-p.pool:
		return conn, nil
	default:
		return p.createConnection()
	}
}

func (p *FTPConnPool) Put(conn *ftp.ServerConn) {
	select {
	case p.pool <- conn:
	default:
		conn.Quit()
	}
}

func (p *FTPConnPool) createConnection() (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", p.config.Host, p.config.Port),
		ftp.DialWithTimeout(time.Duration(p.config.Timeout)*time.Second))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	if err := conn.Login(p.config.Username, p.config.Password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return conn, nil
}

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
	start := time.Now()
	log.Printf("Worker %d Transfer %d: Starting FTP upload", workerID, transferID)

	pool := NewFTPConnPool(config, 5)
	conn, err := pool.Get()
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer pool.Put(conn)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file open error: %w", err)
	}
	defer file.Close()

	// Start transfer timer after connection is established
	transferStart := time.Now()
	remotePath := filepath.Join(config.RemotePath, remoteName)
	err = conn.Stor(remotePath, file)
	transferDuration := time.Since(transferStart)

	log.Printf("Worker %d Transfer %d: Transfer duration %s",
		workerID, transferID, transferDuration.Round(time.Millisecond))

	if err != nil {
		return fmt.Errorf("transfer error: %w", err)
	}

	totalDuration := time.Since(start)
	if totalDuration > time.Duration(config.Timeout)*time.Second {
		log.Printf("Worker %d Transfer %d: Total operation time exceeded timeout (%s)",
			workerID, transferID, totalDuration)
	}

	return nil
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
