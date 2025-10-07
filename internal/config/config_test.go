package config

import (
	"os"
	"testing"
)

func TestFilePathGet(t *testing.T) {
	filePath, err := getConfigFilePath()
	if err != nil {
		t.Errorf("error found: %v", err)
	}

	expectedPath := "/home/madrid/.gatorconfig.json"
	if filePath != expectedPath {
		t.Errorf("path mismatch found -- expected: %v, actual: %v", expectedPath, filePath)
	}
}

func TestReadFilePrint(t *testing.T) {
	// add -v (verbose) to go test .v
	// checked and found that it reads the file properly
	filePath, err := getConfigFilePath()
	if err != nil {
		t.Errorf("error found: %v", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("error found: %v", err)
	}
	stringified := string(data)
	t.Log(stringified)
}
