package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// RepoPermission represents repository permissions for persistence
type RepoPermission struct {
	Name  string            `json:"name"`
	Path  string            `json:"path"`
	Users map[string]string `json:"users"` // username -> permission string ("r" or "rw")
}

// LoadRepoPermissions loads repository permissions from a JSON file
func LoadRepoPermissions(path string) ([]RepoPermission, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No data file yet
		}
		return nil, fmt.Errorf("failed to read repo permission data: %v", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var repos []RepoPermission
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repo permission data: %v", err)
	}

	return repos, nil
}

// SaveRepoPermissions saves repository permissions to a JSON file
func SaveRepoPermissions(path string, repos []RepoPermission) error {
	jsonData, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize repo permission data: %v", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save repo permission data: %v", err)
	}

	return nil
}
