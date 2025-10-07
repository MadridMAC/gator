package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const configFileName = "/.gatorconfig.json"

type Config struct {
	Db_url            string
	Current_user_name string
}

// tested to provide proper file path
func getConfigFilePath() (string, error) {

	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error: $home environment variable not set or found")
	}
	configPath := filepath.Join(homePath, configFileName)
	return configPath, nil
}

func Read() Config {
	// get config file path
	filePath, err := getConfigFilePath()
	if err != nil {
		log.Fatalf("error getting config file path: %v", err)

	}

	// read JSON data
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error: error reading file from path: %v", filePath)
	}

	// unmarshal JSON data
	var configStruct Config
	if err := json.Unmarshal(data, &configStruct); err != nil {
		log.Fatal("error: error unmarshaling data")
	}

	// return config
	return configStruct
}

func write(cfg Config) error {
	// marshal config struct into JSON
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// get config file path
	filePath, err := getConfigFilePath()
	if err != nil {
		log.Fatalf("error getting config file path: %v", err)

	}

	// write config to file
	os.WriteFile(filePath, jsonData, 0644)
	return nil
}

func (c *Config) SetUser(username string) {
	c.Current_user_name = username
	write(*c)
}
