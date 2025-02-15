package Core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Campaign struct {
	Name             string           `json:"Name"`
	Protocol         string           `json:"Protocol"` // FTP/SFTP/HTTP
	Type             string           `json:"Type"`     // Upload/Download
	Host             string           `json:"Host"`
	Port             int              `json:"Port"`
	RemotePath       string           `json:"RemotePath"`
	LocalPath        string           `json:"LocalPath"`
	Timeout          int              `json:"Timeout"` // in seconds
	RampUp           string           `json:"RampUp"`
	HoldFor          string           `json:"HoldFor"`
	NumClients       int              `json:"NumClients"`
	NumRequests      int              `json:"NumRequests"`
	FilesizePolicies []FilesizePolicy `json:"FilesizePolicies"`
	Config           TestConfig       `json:"Config,omitempty"`
	SourcePattern    string           `json:"SourcePattern,omitempty"`
	Username         string           `json:"Username"`
	Password         string           `json:"Password"`
	UploadTestID     string           `json:"UploadTestID"`
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

	// fmt.Printf("Raw JSON: %s\n", string(data))

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
		UploadTestID:     campaign.UploadTestID,
		TestID:           fmt.Sprintf("test_%d", time.Now().UnixNano()),
	}

	if config.Type == "DOWNLOAD" && config.UploadTestID == "" {
		return nil, fmt.Errorf("download campaigns require 'upload_test_id' field in campaign file")
	}

	if config.Type == "UPLOAD" && !strings.HasSuffix(config.RemotePath, "/") {
		return nil, fmt.Errorf("upload remote path must end with '/'")
	}

	fmt.Printf("Config: %+v\n", config)

	return &config, nil
}
