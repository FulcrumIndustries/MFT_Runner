package Core

import (
	"os"
	"sync"
	"time"
)

// Now only contains logging-specific enhancements
type MFTHandler struct {
	report   *TestReport
	mu       sync.Mutex
	logFile  *os.File
	interval time.Duration
}

func NewMFTHandler(config TestConfig) *MFTHandler {
	return &MFTHandler{
		report:   NewTestReport(config),
		interval: 5 * time.Second,
	}
}

// Wrapper methods using core TestReport
func (h *MFTHandler) RecordLatency(latency time.Duration, success bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.report.Latencies = append(h.report.Latencies, latency.Seconds()*1000)
}

func (h *MFTHandler) WriteLog(path string) error {
	h.report.Finalize()
	return h.report.WriteToFile(path)
}
