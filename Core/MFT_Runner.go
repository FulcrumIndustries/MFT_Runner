package Core

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Add this struct definition
type FilesizePolicy struct {
	Size    IntFromString `json:"size"`
	Unit    string        `json:"unit"`
	Percent int64         `json:"percent"`
	Count   int           `json:"-"` // Derived field, not stored
}

// Keep only the essential test configuration
type TestConfig struct {
	NumClients              int              `json:"num_clients"`  // Concurrent workers
	NumRequests             int              `json:"num_requests"` // Total files to transfer
	NumRequestsFirstClients int              `json:"num_requests_first_clients,omitempty"`
	RampUp                  string           `json:"rampup"`
	Type                    string           `json:"type"`
	Protocol                string           `json:"protocol"`
	Host                    string           `json:"host"`
	Port                    int              `json:"port"`
	FilesizePolicies        []FilesizePolicy `json:"filesizepolicies"`
	TestID                  string           `json:"testid"`
	RemotePath              string           `json:"remotepath"`
	LocalPath               string           `json:"localpath"`
	WorkerID                int              `json:"worker_id"`
	Timeout                 int              `json:"timeout"`
	Username                string           `json:"username"`
	Password                string           `json:"password"`
}

// ErrorHandler is a function type for handling test errors
type ErrorHandler func(string)

type TestReport struct {
	Config      TestConfig       `json:"config"`
	Summary     TestSummary      `json:"summary"`
	Latencies   []float64        `json:"latencies"`
	Throughputs []float64        `json:"throughputs"`
	Errors      []string         `json:"errors"`
	Timestamp   time.Time        `json:"timestamp"`
	Duration    time.Duration    `json:"duration"`
	TimeSeries  []TimeSeriesData `json:"time_series"`
	mu          sync.Mutex
}

type TestSummary struct {
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	AvgLatencyMs       float64 `json:"avg_latency_ms"`
	MinLatencyMs       float64 `json:"min_latency_ms"`
	MaxLatencyMs       float64 `json:"max_latency_ms"`
	Percentiles        struct {
		P50 float64 `json:"p50"`
		P90 float64 `json:"p90"`
		P95 float64 `json:"p95"`
		P99 float64 `json:"p99"`
	} `json:"percentiles"`
	ThroughputRps float64 `json:"throughput_rps"`
}

type transferResult struct {
	success  bool
	duration time.Duration
	error    string
}

func NewTestReport(config TestConfig) *TestReport {
	return &TestReport{
		Config:      config,
		Timestamp:   time.Now(),
		Latencies:   make([]float64, 0),
		Throughputs: make([]float64, 0),
		Errors:      make([]string, 0),
		TimeSeries:  make([]TimeSeriesData, 0),
	}
}

func (r *TestReport) Finalize() {
	r.Duration = time.Since(r.Timestamp)
	r.Summary.TotalRequests = r.Summary.SuccessfulRequests + r.Summary.FailedRequests

	// Handle all-failed case
	if r.Summary.SuccessfulRequests == 0 && r.Summary.FailedRequests > 0 {
		r.Summary.MinLatencyMs = 0
		r.Summary.MaxLatencyMs = 0
		r.Summary.AvgLatencyMs = 0
	}

	if len(r.Latencies) > 0 {
		sort.Float64s(r.Latencies)

		var total float64
		r.Summary.MinLatencyMs = r.Latencies[0]
		r.Summary.MaxLatencyMs = r.Latencies[len(r.Latencies)-1]

		for _, l := range r.Latencies {
			total += l
		}
		r.Summary.AvgLatencyMs = total / float64(len(r.Latencies))

		// Percentile calculations
		r.Summary.Percentiles.P50 = percentile(r.Latencies, 0.5)
		r.Summary.Percentiles.P90 = percentile(r.Latencies, 0.9)
		r.Summary.Percentiles.P95 = percentile(r.Latencies, 0.95)
		r.Summary.Percentiles.P99 = percentile(r.Latencies, 0.99)
	}

	if r.Duration.Seconds() > 0 && r.Summary.SuccessfulRequests > 0 {
		r.Summary.ThroughputRps = float64(r.Summary.SuccessfulRequests) / r.Duration.Seconds()
	}
}

func percentile(sortedData []float64, p float64) float64 {
	index := p * float64(len(sortedData)-1)
	if index == float64(int(index)) {
		return sortedData[int(index)]
	}
	i := int(index)
	f := index - float64(i)
	return sortedData[i]*(1-f) + sortedData[i+1]*f
}

// Log formatting constants
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	logPrefix   = "[MFT] "
)

func RunMFTTest(config *TestConfig, onError ErrorHandler) (*TestReport, error) {
	fmt.Printf("\n%s%s=== STARTING TEST: %s ===%s\n", colorCyan, logPrefix, config.TestID, colorReset)
	defer fmt.Printf("\n%s%s=== TEST COMPLETED ===%s\n", colorCyan, logPrefix, colorReset)

	// Use configured values instead of raw arguments
	numClients := config.NumClients
	numRequests := config.NumRequests

	fmt.Printf("%s%s%-18s: %s%s:%d%s\n", colorReset, logPrefix, "Protocol", colorCyan, config.Host, config.Port, colorReset)
	fmt.Printf("%s%s%-18s: %s%d workers%s\n", colorReset, logPrefix, "Concurrency", colorCyan, config.NumClients, colorReset)
	fmt.Printf("%s%s%-18s: %s%d transfers%s\n", colorReset, logPrefix, "Total Transfers", colorCyan, config.NumClients*config.NumRequests, colorReset)
	fmt.Printf("%s%s%-18s: %s%.2f KB avg%s\n", colorReset, logPrefix, "File Size", colorCyan, averageFileSize(config), colorReset)

	var wg sync.WaitGroup
	results := make(chan transferResult, numClients*numRequests)

	log.Printf("Creating %d test clients", numClients)
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d starting...", workerID)

			// Create worker-specific config copy
			workerConfig := *config
			workerConfig.WorkerID = workerID

			for j := 0; j < numRequests; j++ {
				transferNum := j + 1
				result := executeTransfer(workerConfig, transferNum, onError)
				results <- result

				if j%10 == 0 {
					log.Printf("Worker %d completed %d/%d transfers", workerID, transferNum, numRequests)
				}
			}
			log.Printf("Worker %d finished all transfers", workerID)
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(results)
		log.Printf("All clients completed their transfers")
	}()

	// Create report with initial counts
	report := &TestReport{
		Config:      *config,
		Timestamp:   time.Now(),
		Latencies:   make([]float64, 0),
		Throughputs: make([]float64, 0),
		Errors:      make([]string, 0),
		TimeSeries:  make([]TimeSeriesData, 0),
		Summary: TestSummary{
			TotalRequests: numClients * numRequests,
		},
	}

	// Process results...
	for result := range results {
		if result.success {
			report.Latencies = append(report.Latencies, result.duration.Seconds()*1000)
			report.Summary.SuccessfulRequests++
		} else {
			if result.error != "" {
				report.Errors = append(report.Errors, result.error)
			}
			report.Summary.FailedRequests++
		}
	}

	// Calculate percentages
	var successPercent, failPercent float64
	if report.Summary.TotalRequests > 0 {
		successPercent = float64(report.Summary.SuccessfulRequests) / float64(report.Summary.TotalRequests) * 100
		failPercent = float64(report.Summary.FailedRequests) / float64(report.Summary.TotalRequests) * 100
	}

	// Log summary after all results are processed
	fmt.Printf("\n%s%s=== TEST SUMMARY ===%s", colorCyan, logPrefix, colorReset)
	fmt.Printf("\n%s%s%-20s: %s%d%s", colorReset, logPrefix, "Total Transfers", colorCyan, report.Summary.TotalRequests, colorReset)
	fmt.Printf("\n%s%s%-20s: %s%d (%.1f%%)%s", colorReset, logPrefix, "Successful", colorGreen, report.Summary.SuccessfulRequests, successPercent, colorReset)
	fmt.Printf("\n%s%s%-20s: %s%d (%.1f%%)%s", colorReset, logPrefix, "Failed", colorRed, report.Summary.FailedRequests, failPercent, colorReset)
	fmt.Printf("\n%s%s%-20s: %s%.2f req/s%s", colorReset, logPrefix, "Throughput", colorCyan, report.Summary.ThroughputRps, colorReset)
	fmt.Printf("\n%s%s%-20s: %s%.2fms%s", colorReset, logPrefix, "Avg Latency", colorCyan, report.Summary.AvgLatencyMs, colorReset)

	defer os.RemoveAll(fmt.Sprintf("testfiles/%s", config.TestID)) // Cleanup files

	// Return report for writing in main
	return report, nil
}

func loadTestConfig(path string) (*TestConfig, error) {
	log.Printf("Loading campaign file: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config TestConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config JSON: %w", err)
	}
	return &config, nil
}

// Helper function to select file size based on distribution
func selectFileSize(policies []FilesizePolicy) *FilesizePolicy {
	// Generate random number between 0 and 100
	r := rand.Float64() * 100

	// Accumulate percentages and select appropriate policy
	var accum float64
	for i := range policies {
		accum += float64(policies[i].Percent)
		if r <= accum {
			return &policies[i]
		}
	}

	// Fallback to last policy if rounding issues
	return &policies[len(policies)-1]
}

// formatSize converts filesize policy to human-readable string
func formatSize(policy *FilesizePolicy) string {
	unit := strings.ToUpper(policy.Unit)
	switch unit {
	case "K":
		unit = "KB"
	case "M":
		unit = "MB"
	case "G":
		unit = "GB"
	}
	return fmt.Sprintf("%d%s", policy.Size, unit)
}

func executeTransfer(config TestConfig, transferID int, onError ErrorHandler) transferResult {
	workerID := config.WorkerID
	policy := selectFileSize(config.FilesizePolicies)

	fmt.Printf("%s%sWorker %d - Starting transfer %d (%s)%s\n",
		colorReset, logPrefix, workerID, transferID, formatSize(policy), colorReset)

	start := time.Now()

	if policy.Count < 1 {
		log.Printf("Worker %d Transfer %d: No files available for policy", workerID, transferID)
		return transferResult{success: false, duration: 0, error: ""}
	}

	// Get random file from manifest
	fileList := getFileList(config.TestID)
	if len(fileList) == 0 {
		log.Printf("Worker %d Transfer %d: No files available", workerID, transferID)
		return transferResult{success: false, duration: 0, error: ""}
	}
	selectedFile := fileList[rand.Intn(len(fileList))]

	// Get existing test file
	filePath := filepath.Join("Work", "testfiles", config.TestID, selectedFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Worker %d Transfer %d: Test file missing - %v", workerID, transferID, err)
		onError(fmt.Sprintf("Test file missing: %s", filePath))
		return transferResult{
			success:  false,
			duration: time.Since(start),
			error:    fmt.Sprintf("file_not_found: %s", filePath),
		}
	}

	var remoteName string
	if config.Type == "Upload" {
		// Generate unique name only for uploads
		remoteName = fmt.Sprintf("%s_%d_%d_%d.dat",
			strings.TrimSuffix(selectedFile, filepath.Ext(selectedFile)),
			time.Now().UnixNano(),
			workerID,
			transferID,
		)
	} else {
		// Use original filename for downloads
		remoteName = selectedFile
	}

	// Create a channel to signal completion
	done := make(chan bool)
	var transferErr error

	// Execute transfer in goroutine
	go func() {
		fmt.Printf("%s%sWorker %d - Transfer %d: %s %s to %s%s\n",
			colorReset, logPrefix, workerID, transferID, config.Type,
			formatSize(policy), config.RemotePath, colorReset)
		switch config.Protocol {
		case "FTP":
			if config.Type == "Upload" {
				transferErr = FTPUpload(filePath, remoteName, &config, workerID, transferID)
			} else {
				transferErr = FTPDownload(filePath, remoteName, &config, workerID)
			}
		case "SFTP":
			if config.Type == "Upload" {
				transferErr = SFTPUpload(filePath, remoteName, &config)
			} else {
				transferErr = SFTPDownload(remoteName, filePath, &config)
			}
		case "HTTP":
			if config.Type == "Upload" {
				transferErr = HTTPUpload(filePath, &config)
			} else {
				transferErr = HTTPDownload(config.RemotePath, &config)
			}
		default:
			log.Printf("Unsupported protocol: %s", config.Protocol)
			transferErr = fmt.Errorf("unsupported protocol: %s", config.Protocol)
		}
		done <- true
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		duration := time.Since(start)
		if transferErr != nil {
			fmt.Printf("%s%sWorker %d - Failed after %s | %s | Error: %s%s\n",
				colorRed, logPrefix, workerID, duration.Round(time.Millisecond), selectedFile, transferErr.Error(), colorReset)
			onError(fmt.Sprintf("Transfer failed: %v", transferErr))
			return transferResult{success: false, duration: duration, error: transferErr.Error()}
		}
		fmt.Printf("%s%sWorker %d - Completed in %s | %s%s\n",
			colorGreen, logPrefix, workerID, duration.Round(time.Millisecond), selectedFile, colorReset)
		return transferResult{success: true, duration: duration}
	case <-time.After(time.Duration(config.Timeout) * time.Second):
		log.Printf("Transfer %s timed out after %ds", selectedFile, config.Timeout)
		onError(fmt.Sprintf("Transfer %s timed out after %ds", selectedFile, config.Timeout))
		return transferResult{
			success:  false,
			duration: time.Duration(config.Timeout) * time.Second,
			error:    fmt.Sprintf("timeout_%ds: %s", config.Timeout, selectedFile),
		}
	}
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}

// IntFromString handles both string and number JSON values
type IntFromString int64

func (i *IntFromString) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*i = IntFromString(value)
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid number string: %s", value)
		}
		*i = IntFromString(parsed)
	default:
		return fmt.Errorf("invalid type for size: %T", value)
	}
	return nil
}

func (r *TestReport) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}

type TimeSeriesData struct {
	Timestamp     time.Time `json:"timestamp"`
	Requests      int       `json:"requests"`
	Errors        int       `json:"errors"`
	AvgLatencyMs  float64   `json:"avg_latency_ms"`
	ThroughputRPS float64   `json:"throughput_rps"`
}

func (r *TestReport) AddTimeSeriesSample() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Calculate current window stats
	success := len(r.Latencies)
	failed := len(r.Errors)
	total := success + failed

	var avgLatency float64
	if success > 0 {
		var sum float64
		for _, l := range r.Latencies {
			sum += l
		}
		avgLatency = sum / float64(success)
	}

	r.TimeSeries = append(r.TimeSeries, TimeSeriesData{
		Timestamp:     time.Now(),
		Requests:      total,
		Errors:        failed,
		AvgLatencyMs:  avgLatency,
		ThroughputRPS: float64(total) / time.Since(r.Timestamp).Seconds(),
	})
}

// averageFileSize calculates average file size in KB
func averageFileSize(config *TestConfig) float64 {
	var totalKB float64
	for _, p := range config.FilesizePolicies {
		sizeKB := float64(p.Size)
		switch strings.ToUpper(p.Unit) {
		case "M":
			sizeKB *= 1024
		case "G":
			sizeKB *= 1024 * 1024
		case "K":
		default: // Assume bytes
			sizeKB /= 1024
		}
		totalKB += sizeKB * float64(p.Percent) / 100
	}
	return totalKB
}
