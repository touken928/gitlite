package auth

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// UserType represents the type of user in the system
type UserType int

const (
	UserTypeUnknown UserType = iota // Unauthenticated or unknown user
	UserTypeAdmin                   // Administrator with full access
	UserTypeNormal                  // Regular authenticated user
)

// User represents a user with their associated SSH public keys
type User struct {
	Name string             // Unique username
	Keys []ssh.PublicKey    // SSH public keys for authentication
}

// AddKey adds a new SSH public key to the user, returns error if key already exists
func (u *User) AddKey(key ssh.PublicKey) error {
	fingerprint := ssh.FingerprintSHA256(key)
	for _, k := range u.Keys {
		if ssh.FingerprintSHA256(k) == fingerprint {
			return fmt.Errorf("key already exists")
		}
	}
	u.Keys = append(u.Keys, key)
	return nil
}

// RemoveKey removes an SSH public key by its fingerprint, returns true if found and removed
func (u *User) RemoveKey(fingerprint string) bool {
	for i, k := range u.Keys {
		if ssh.FingerprintSHA256(k) == fingerprint {
			u.Keys = append(u.Keys[:i], u.Keys[i+1:]...)
			return true
		}
	}
	return false
}

// HasKey returns true if the user has the given SSH public key
func (u *User) HasKey(key ssh.PublicKey) bool {
	fingerprint := ssh.FingerprintSHA256(key)
	for _, k := range u.Keys {
		if ssh.FingerprintSHA256(k) == fingerprint {
			return true
		}
	}
	return false
}
