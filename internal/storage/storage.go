package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// User represents a user with SSH keys for persistence
type User struct {
	Name string   `json:"name"`
	Keys []string `json:"keys"` // SSH public key strings
}

// LoadUsers loads users from a JSON file
func LoadUsers(path string) ([]User, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No data file yet
		}
		return nil, fmt.Errorf("读取用户数据失败: %v", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("解析用户数据失败: %v", err)
	}

	return users, nil
}

// SaveUsers saves users to a JSON file
func SaveUsers(path string, users []User) error {
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化用户数据失败: %v", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("保存用户数据失败: %v", err)
	}

	return nil
}

// LoadSSHKeys loads SSH public keys from key strings
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

// SaveSSHKeys saves SSH public keys to strings
func SaveSSHKeys(keys []ssh.PublicKey) []string {
	keyStrs := make([]string, len(keys))
	for i, k := range keys {
		keyStrs[i] = string(ssh.MarshalAuthorizedKey(k))
	}
	return keyStrs
}
