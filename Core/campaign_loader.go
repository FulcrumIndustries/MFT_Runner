package Core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Campaign struct {
	Name             string           `json:"name"`
	Protocol         string           `json:"protocol"` // FTP/SFTP/HTTP
	Type             string           `json:"type"`     // Upload/Download
	Host             string           `json:"host"`
	Port             int              `json:"port"`
	RemotePath       string           `json:"remote_path"`
	LocalPath        string           `json:"local_path"`
	Timeout          int              `json:"timeout"` // in seconds
	RampUp           string           `json:"rampup"`
	HoldFor          string           `json:"holdfor"`
	NumClients       int              `json:"num_clients"`
	NumRequests      int              `json:"num_requests"`
	FilesizePolicies []FilesizePolicy `json:"filesizepolicies"`
	Config           TestConfig       `json:"config,omitempty"`
	SourcePattern    string           `json:"source_pattern,omitempty"`
	Username         string           `json:"username"`
	Password         string           `json:"password"`
}

func LoadCampaign(path string) (*TestConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check in Campaigns directory
		path = filepath.Join("Campaigns", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Raw JSON: %s\n", string(data))

	var campaign Campaign // First unmarshal into Campaign
	if err := json.Unmarshal(data, &campaign); err != nil {
		return nil, fmt.Errorf("error parsing campaign: %v", err)
	}

	// Then copy fields to TestConfig
	config := TestConfig{
		Protocol:         campaign.Protocol,
		Type:             campaign.Type,
		NumClients:       campaign.NumClients,
		Host:             campaign.Host,
		Port:             campaign.Port,
		RemotePath:       campaign.RemotePath,
		LocalPath:        campaign.LocalPath,
		Timeout:          campaign.Timeout,
		RampUp:           campaign.RampUp,
		NumRequests:      campaign.NumRequests,
		FilesizePolicies: campaign.FilesizePolicies,
		Username:         campaign.Username,
		Password:         campaign.Password,
		SourcePattern:    campaign.SourcePattern,
		TestID:           fmt.Sprintf("test_%d", time.Now().UnixNano()),
	}

	config.TestID = fmt.Sprintf("test_%d", time.Now().UnixNano())
	return &config, nil
}
