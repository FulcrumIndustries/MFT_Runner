package main

import (
	"MFT_Runner/Core"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	// Ensure Work directory exists
	if err := os.MkdirAll("Work/testfiles", 0755); err != nil {
		log.Fatal("Error creating work directories:", err)
	}

	// Handle CLI flags
	listCampaigns := flag.Bool("lc", false, "List available campaigns")
	viewCampaign := flag.String("vc", "", "View campaign details")
	help := flag.Bool("h", false, "Show help")
	version := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *version {
		fmt.Printf("MFT Runner v%s\n", Core.Version)
		return
	}

	if *help {
		printHelp()
		return
	}

	if *listCampaigns {
		listAllCampaigns()
		return
	}

	if *viewCampaign != "" {
		viewCampaignDetails(*viewCampaign)
		return
	}

	args := flag.Args()
	if len(args) < 3 {
		log.Fatal("Usage: ./mft-runner <campaign-file> <clients> <requests>")
	}

	config, err := Core.LoadCampaign(args[0])
	if err != nil {
		log.Fatal("Error loading campaign:", err)
	}

	// Override config with CLI parameters
	config.NumClients, _ = strconv.Atoi(args[1])
	totalRequests, _ := strconv.Atoi(args[2])
	config.NumRequests = totalRequests / config.NumClients

	// Calculate requests per client correctly
	if rem := totalRequests % config.NumClients; rem > 0 {
		config.NumRequests++
		// Adjust first 'rem' clients to handle remainder
		config.NumRequestsFirstClients = rem
	}

	// File generation
	if err := Core.CreateTestFiles(*config, totalRequests); err != nil {
		log.Fatal("Error creating test files:", err)
	}

	// Run test
	report, err := Core.RunMFTTest(config, func(msg string) {
		log.Printf("Error: %s", msg)
	})
	if err != nil {
		log.Fatal("Test execution failed:", err)
	}
	report.Finalize()

	// Write final report
	campaignName := filepath.Base(args[0])
	campaignName = campaignName[:len(campaignName)-len(filepath.Ext(campaignName))]
	reportPath := fmt.Sprintf("TestReports/%s_%s.json",
		campaignName,
		time.Now().Format("20060102_150405"))

	// Ensure reports directory exists
	os.MkdirAll("TestReports", 0755)

	if err := report.WriteToFile(reportPath); err != nil {
		log.Fatal("Failed to write report:", err)
	}
	log.Printf("\n=== Test Report Generated ===")
	log.Printf("Location: %s", reportPath)
	log.Printf("To visualize results:")
	log.Printf("1. Launch MFT Runner UI")
	log.Printf("2. Import this report file")
	log.Printf("3. View interactive performance charts")

	log.Printf("Initializing test parameters:")
	log.Printf("Concurrent Clients: %d", config.NumClients)
	log.Printf("Total Files to Transfer: %d", totalRequests)
	log.Printf("Files per Client: %d", config.NumRequests)
	if rem := totalRequests % config.NumClients; rem > 0 {
		log.Printf("First %d clients will transfer %d files", rem, config.NumRequests)
	}
}

func printHelp() {
	fmt.Println(`MFT Runner - Managed File Transfer Load Testing Tool
Version: ` + Core.Version + `

Basic Usage:
  Run a campaign:
    ./mft-runner <campaign-file> <clients> <requests>
  
  Example:
    ./mft-runner Campaigns/UPLOAD_FTP_1KB.json 10 100
    → 10 concurrent clients sending 100 files total (10 files/client)

Campaign Management:
  -lc        List available campaigns in ./Campaigns/
  -vc <name> View details of specified campaign
  
  Create new campaigns:
    1. Create JSON file in Campaigns/ directory
    2. Use existing campaigns as templates
    3. Supported protocols: FTP, SFTP, HTTP

Other Options:
  -h         Show this help message

Campaign File Format:
{
  "protocol": "FTP",
  "host": "ftp.example.com",
  "port": 21,
  "username": "testuser",
  "password": "testpass",
  "type": "Upload",
  "remote_path": "/uploads",
  "local_path": "./testfiles",
  "timeout": 30,
  "filesizepolicies": [
    {"size": 1, "unit": "MB", "percent": 50},
    {"size": 5, "unit": "MB", "percent": 50}
  ]
}`)
}

func viewCampaignDetails(name string) {
	path := filepath.Join("Campaigns", name)
	config, err := Core.LoadCampaign(path)
	if err != nil {
		log.Fatal("Error loading campaign:", err)
	}

	fmt.Printf("Campaign: %s\n", name)
	fmt.Printf("Protocol: %s\n", config.Protocol)
	fmt.Printf("Host: %s:%d\n", config.Host, config.Port)
	fmt.Printf("Type: %s\n", config.Type)
	fmt.Printf("Ramp Up: %s\n", config.RampUp)
	fmt.Printf("Remote Path: %s\n", config.RemotePath)
	fmt.Printf("Local Path: %s\n", config.LocalPath)
	fmt.Printf("Username: %s\n", config.Username)
	fmt.Printf("Password: %s\n", config.Password)
	fmt.Printf("Timeout: %d\n", config.Timeout)
	fmt.Printf("Num Clients: %d\n", config.NumClients)
	fmt.Printf("Num Requests: %d\n", config.NumRequests)
	fmt.Printf("Filesize Policies: ")
	for _, p := range config.FilesizePolicies {
		fmt.Printf("%d%s (%d%%) ", p.Size, p.Unit, p.Percent)
	}
	fmt.Println()
}

func listAllCampaigns() {
	files, err := filepath.Glob("Campaigns/*.json")
	if err != nil {
		log.Fatal("Error finding campaigns:", err)
	}

	fmt.Println("Available campaigns:")
	for _, f := range files {
		fmt.Println("  ", filepath.Base(f))
	}
}
