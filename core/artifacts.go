package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func writeContentList(path string, contentList []MinerUContent) error {
	data, err := json.MarshalIndent(contentList, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func writeArtifacts(resultDir string, result *Result, csvBytes []byte) error {
	jsonPath := filepath.Join(resultDir, "result.json")
	result.Artifacts.JSONPath = jsonPath

	if len(csvBytes) > 0 {
		csvPath := filepath.Join(resultDir, "result.csv")
		result.Artifacts.CSVPath = csvPath
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
		return err
	}

	if result.Artifacts.CSVPath != "" {
		if err := os.WriteFile(result.Artifacts.CSVPath, csvBytes, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func newTaskID(t time.Time) (string, error) {
	var suffix [4]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", t.Format("20060102-150405"), hex.EncodeToString(suffix[:])), nil
}
