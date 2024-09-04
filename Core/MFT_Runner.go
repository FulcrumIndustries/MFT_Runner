package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Campaign struct {
	SourceProtocol   string           `json:"sourceprotocol"`
	TargetProtocol   string           `json:"targetprotocol"`
	Url              string           `json:"url"`
	User             string           `json:"user"`
	Password         string           `json:"password"`
	Type             string           `json:"type"`
	Timeout          int              `json:"timeout"`
	FilesizePolicies []FilesizePolicy `json:"filesizepolicies"`
}
type FilesizePolicy struct {
	Size    int64  `json:"size"`
	Unit    string `json:"unit"`
	Percent int64  `json:"percent"`
}

func main() {
	fmt.Printf(">>>>> MFT_Runner Program Started\n")
	fmt.Printf("Analyzing Arguments [%v] : %v\n", len(os.Args), os.Args[1:])
	if len(os.Args) == 2 && os.Args[1] == "init" {
		fmt.Printf("Setting up SecureTransport : %v\n", os.Args[1])
		err := createAccount("https://10.128.144.168:8444/api/v2.0/accountSetup", "admin", "admin")
		if err != nil {
			fmt.Printf("Something went wrong: %v\n", err)
		}
		os.Exit(1)
	} else if len(os.Args) != 4 {
		fmt.Printf("Bad arguments : %v\n", os.Args)
		fmt.Printf("Expecting : [%v <campaign> <nb_client> <nb_requests>]", os.Args[0])
		os.Exit(1)
	}
	campaign_arg := os.Args[1]
	nbclient_arg := os.Args[2]
	nbrequests_arg := os.Args[3]
	nbclients, e := strconv.Atoi(nbclient_arg)
	if e != nil {
		fmt.Printf("Bad argument #1 : %v", nbclients)
		os.Exit(1)
	}
	nbrequests, e := strconv.Atoi(nbrequests_arg)
	if e != nil {
		fmt.Printf("Bad argument #2 : %v", nbrequests)
		os.Exit(1)
	}
	// Load JSON Campaign
	campaign := loadCampaign(campaign_arg)
	fmt.Printf("Using user/password: %v/%v \n", campaign.User, campaign.Password)
	fmt.Printf("URL: %v\n", campaign.Url)
	// Make Work Directory
	startf := time.Now()
	dirname := fileNameWithoutExtTrimSuffix(filepath.Base(campaign_arg)) + startf.Format("20060102150405")
	makeDir(dirname)

	// SFTP parameters
	var ipaddr string = ""
	var port uint = 0
	var targetdir string = ""
	var dial string = ""
	// In case of SFTP source protocol, we prepare a few more variables
	if campaign.SourceProtocol == "SFTP" {
		url_array := strings.Split(campaign.Url, ":")
		if len(url_array) < 2 {
			fmt.Printf("Bad URL in configuration file : %v", campaign.Url)
			os.Exit(1)
		}
		ipaddr = url_array[0]
		fmt.Println("Target Server : " + ipaddr)
		sftp_port_dir_str := url_array[1]
		sftp_port_dir_str_array := strings.Split(sftp_port_dir_str, "/")
		if len(sftp_port_dir_str_array) < 2 {
			fmt.Printf("Bad URL in configuration file : %v", campaign.Url)
			os.Exit(1)
		}
		sftp_port_str := sftp_port_dir_str_array[0]
		fmt.Println("Target Port : " + sftp_port_str)
		targetdir = sftp_port_dir_str_array[1]
		fmt.Println("Target Dir : " + targetdir)
		portu, err := strconv.ParseUint(sftp_port_str, 10, 32)
		if err != nil {
			fmt.Printf("Bad Port number in configuration file : %v", campaign.Url)
			os.Exit(1)
		}
		port = uint(portu)
	} else if campaign.SourceProtocol == "FTP" {
		url_array := strings.Split(campaign.Url, "/")
		if len(url_array) < 2 {
			fmt.Printf("Bad URL in configuration file : %v", campaign.Url)
			os.Exit(1)
		}
		dial = url_array[0]
		fmt.Println("Target FTP Dial : " + dial)
		targetdir = url_array[1]
		fmt.Println("Target Dir : " + targetdir)
	}
	// Make Work Files : only 1 per size
	// Also prepares the size distribution arrays
	fmt.Printf("Generating Work Files in ../Work/%v\n", dirname)
	var fileDistributionWithName []string
	for i := 0; i < len(campaign.FilesizePolicies); i++ {
		sSize := strconv.FormatInt(campaign.FilesizePolicies[i].Size, 10)
		var int64CalculatedSize = int64(1024)
		fmt.Printf(" #%d - %s%s Files [%d]\n", i+1, sSize, campaign.FilesizePolicies[i].Unit, int(campaign.FilesizePolicies[i].Percent)*nbrequests/100+1)
		if campaign.FilesizePolicies[i].Unit == "K" {
			int64CalculatedSize = int64CalculatedSize * campaign.FilesizePolicies[i].Size
		} else if campaign.FilesizePolicies[i].Unit == "M" {
			int64CalculatedSize = int64CalculatedSize * 1024 * campaign.FilesizePolicies[i].Size
		} else if campaign.FilesizePolicies[i].Unit == "G" {
			int64CalculatedSize = int64CalculatedSize * 1024 * 1024 * campaign.FilesizePolicies[i].Size
		}
		for j := 0; j < int(campaign.FilesizePolicies[i].Percent)*nbrequests/100+1; j++ {
			fileDistributionWithName = append(fileDistributionWithName, "../Work/"+dirname+"/test_"+sSize+campaign.FilesizePolicies[i].Unit) //+"_"+strconv.Itoa(j)
		}
		makeFile("test_"+sSize+campaign.FilesizePolicies[i].Unit, dirname, int64CalculatedSize)
		// If reqtype=GET we setup the Download files
		if campaign.Type == "GET" {
			if campaign.SourceProtocol == "HTTP" {
				multipartCall("", 300, campaign.Url, campaign.User, campaign.Password, "POST", "../Work/"+dirname+"/test_"+sSize+campaign.FilesizePolicies[i].Unit)
			} else if campaign.SourceProtocol == "SFTP" {
				SSH_Operation("", campaign.User, ipaddr, port, campaign.Password, "../Work/"+dirname+"/test_"+sSize+campaign.FilesizePolicies[i].Unit, targetdir, 300, "POST")
			} else if campaign.SourceProtocol == "FTP" {
				FTP_Operation("", campaign.User, dial, campaign.Password, "../Work/"+dirname+"/test_"+sSize+campaign.FilesizePolicies[i].Unit, targetdir, 300, "POST")
			}
			fmt.Printf(" >> File of Size %v uploaded in MFT_Runner/Download\n", int64CalculatedSize)
		}
	}

	// Shuffle fileDistributionWithName array
	for i := range fileDistributionWithName {
		j := rand.IntN(i + 1)
		fileDistributionWithName[i], fileDistributionWithName[j] = fileDistributionWithName[j], fileDistributionWithName[i]
	}
	fmt.Println(len(fileDistributionWithName))
	// for j := 0; j < len(fileDistributionWithName); j++ {
	// 	fmt.Println(fileDistributionWithName[j])
	// }
	// sPercent := strconv.FormatInt(campaign.FilesizePolicies[i].Percent, 10)
	// Start the Concurrent Clients test
	start := time.Now()
	fmt.Println("===========================================================================")
	fmt.Println("> Running Campaign " + campaign_arg + " with " + nbclient_arg + " concurrent Clients")
	var numJobs = nbrequests
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	for w := 1; w <= nbclients; w++ {
		go worker(w, campaign.Url, campaign.Timeout, campaign.User, campaign.Password, campaign.Type, jobs, results, fileDistributionWithName, campaign.SourceProtocol, ipaddr, port, targetdir, dial)
	}
	// Feed jobs into workers
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
	elapsed := time.Since(start)
	fmt.Printf("MFT_Runner took %s to send %s requests from %s clients.", elapsed, nbrequests_arg, nbclient_arg)
}
func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
func worker(id int, urlstr string, maxtimeout int, username string, password string, reqtype string, jobs <-chan int, results chan<- int, fileDistributionWithName []string, sourceprotocol string, ipaddr string, port uint, targetdir string, dial string) {
	for j := range jobs {
		fmt.Println(sourceprotocol, "worker", id, "started  job", j)
		if sourceprotocol == "HTTP" {
			multipartCall(strconv.Itoa(j), maxtimeout, urlstr, username, password, reqtype, fileDistributionWithName[j-1])
		} else if sourceprotocol == "SFTP" {
			// time.Sleep(time.Second * 1)
			SSH_Operation(strconv.Itoa(j), username, ipaddr, port, password, fileDistributionWithName[j-1], targetdir, maxtimeout, reqtype)
		} else if sourceprotocol == "FTP" {
			// time.Sleep(time.Second * 1)
			FTP_Operation(strconv.Itoa(j), username, dial, password, fileDistributionWithName[j-1], targetdir, maxtimeout, reqtype)
		}
		fmt.Println(sourceprotocol, "worker", id, "finished job", j)
		// time.Sleep(time.Second)
		results <- j * 2
	}
}

// func pickFile(fileDistributionWithName []string) (file string) {
// 	rng := rand.IntN(len(fileDistributionWithName))
// 	file = fileDistributionWithName[rng]
// 	return
// }

func loadCampaign(campaignfile string) (campaign Campaign) {
	fmt.Println("Loading Campaign " + campaignfile + "...")
	jsonFile, err := os.Open(campaignfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &campaign)
	fmt.Println("Campaign Loaded Successfully.")
	// fmt.Println("> User/Pwd: " + campaign.User + "/" + campaign.Password)
	// fmt.Println("> Source Protocol: " + campaign.SourceProtocol)
	// fmt.Println("> Target Protocol: " + campaign.TargetProtocol)
	// fmt.Println("> File Size Policies:")
	// for i := 0; i < len(campaign.FilesizePolicies); i++ {
	// 	s := strconv.FormatInt(campaign.FilesizePolicies[i].Percent, 10)
	// 	fmt.Println("   - " + campaign.FilesizePolicies[i].Size + " : " + s + "%")
	// }
	return
}
