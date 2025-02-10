package Core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func LoadCampaign(path string) (*TestConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check in Campaigns directory
		path = filepath.Join("Campaigns", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TestConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing campaign: %v", err)
	}

	// Generate unique test ID
	config.TestID = fmt.Sprintf("test_%d", time.Now().UnixNano())

	return &config, nil
}
