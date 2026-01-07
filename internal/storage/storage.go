package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// User represents a user with SSH keys for JSON persistence
type User struct {
	Name string   `json:"name"` // Unique username
	Keys []string `json:"keys"` // SSH public key strings in authorized_keys format
}

// LoadUsers loads user data from a JSON file
func LoadUsers(path string) ([]User, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No data file exists yet
		}
		return nil, fmt.Errorf("failed to read user data: %v", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %v", err)
	}

	return users, nil
}

// SaveUsers persists user data to a JSON file
func SaveUsers(path string, users []User) error {
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize user data: %v", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save user data: %v", err)
	}

	return nil
}

// LoadSSHKeys converts SSH public key strings to ssh.PublicKey objects
func LoadSSHKeys(keyStrs []string) []ssh.PublicKey {
	keys := make([]ssh.PublicKey, 0, len(keyStrs))
	for _, keyStr := range keyStrs {
		pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyStr))
		if err != nil {
			continue // Skip invalid keys
		}
		keys = append(keys, pubKey)
	}
	return keys
}

// SaveSSHKeys converts ssh.PublicKey objects to strings for persistence
func SaveSSHKeys(keys []ssh.PublicKey) []string {
	keyStrs := make([]string, len(keys))
	for i, k := range keys {
		keyStrs[i] = string(ssh.MarshalAuthorizedKey(k))
	}
	return keyStrs
}

// RepoPermission represents repository permissions for JSON persistence
type RepoPermission struct {
	Name  string            `json:"name"`            // Repository name
	Path  string            `json:"path"`            // Filesystem path to repository
	Users map[string]string `json:"users"`           // Username to permission mapping ("r" or "rw")
}

// LoadRepoPermissions loads repository permission data from a JSON file
func LoadRepoPermissions(path string) ([]RepoPermission, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No data file exists yet
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

// SaveRepoPermissions persists repository permission data to a JSON file
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
