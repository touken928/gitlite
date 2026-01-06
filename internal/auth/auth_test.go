package auth

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func generateSSHKey(t *testing.T, comment string) ssh.PublicKey {
	// Create temp directory for key generation
	tmpDir, err := os.MkdirTemp("", "keygen_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyFile := filepath.Join(tmpDir, "test_key")
	pubFile := keyFile + ".pub"

	// Generate SSH key
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyFile, "-N", "", "-C", comment)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate SSH key: %v", err)
	}

	// Read public key
	pubKeyData, err := os.ReadFile(pubFile)
	if err != nil {
		t.Fatalf("Failed to read public key: %v", err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		t.Fatalf("Failed to parse SSH key: %v", err)
	}

	return pubKey
}

func generateRSAKey(t *testing.T, comment string) ssh.PublicKey {
	tmpDir, err := os.MkdirTemp("", "keygen_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyFile := filepath.Join(tmpDir, "test_key")
	pubFile := keyFile + ".pub"

	// Generate RSA key
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", keyFile, "-N", "", "-C", comment)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to generate SSH key: %v", err)
	}

	pubKeyData, err := os.ReadFile(pubFile)
	if err != nil {
		t.Fatalf("Failed to read public key: %v", err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		t.Fatalf("Failed to parse SSH key: %v", err)
	}

	return pubKey
}

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.users == nil {
		t.Error("users map is nil")
	}
	if len(mgr.users) != 0 {
		t.Errorf("Expected empty users map, got %d users", len(mgr.users))
	}
}

func TestCreateUser(t *testing.T) {
	mgr := NewManager()

	// Create a user
	err := mgr.CreateUser("alice")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify user exists
	user := mgr.GetUser("alice")
	if user == nil {
		t.Fatal("GetUser returned nil for alice")
	}
	if user.Name != "alice" {
		t.Errorf("Expected user name 'alice', got '%s'", user.Name)
	}

	// Try to create duplicate user
	err = mgr.CreateUser("alice")
	if err == nil {
		t.Error("CreateUser should return error for duplicate user")
	}

	// Try to create admin user
	err = mgr.CreateUser("admin")
	if err == nil {
		t.Error("CreateUser should return error for admin user")
	}
}

func TestDeleteUser(t *testing.T) {
	mgr := NewManager()

	// Create a user
	mgr.CreateUser("alice")

	// Delete user
	err := mgr.DeleteUser("alice")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user is gone
	user := mgr.GetUser("alice")
	if user != nil {
		t.Error("GetUser should return nil after delete")
	}

	// Try to delete non-existent user
	err = mgr.DeleteUser("bob")
	if err == nil {
		t.Error("DeleteUser should return error for non-existent user")
	}
}

func TestListUsers(t *testing.T) {
	mgr := NewManager()

	// List empty users
	users := mgr.ListUsers()
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}

	// Create users
	mgr.CreateUser("alice")
	mgr.CreateUser("bob")
	mgr.CreateUser("charlie")

	// List users
	users = mgr.ListUsers()
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestAddKeyToUser(t *testing.T) {
	mgr := NewManager()
	mgr.CreateUser("alice")

	key := generateSSHKey(t, "alice@test")

	// Add key
	err := mgr.AddKeyToUser("alice", key)
	if err != nil {
		t.Fatalf("AddKeyToUser failed: %v", err)
	}

	// Verify key was added
	user := mgr.GetUser("alice")
	if len(user.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(user.Keys))
	}

	// Try to add duplicate key
	err = mgr.AddKeyToUser("alice", key)
	if err == nil {
		t.Error("AddKeyToUser should return error for duplicate key")
	}

	// Try to add key to non-existent user
	err = mgr.AddKeyToUser("bob", key)
	if err == nil {
		t.Error("AddKeyToUser should return error for non-existent user")
	}
}

func TestRemoveKeyFromUser(t *testing.T) {
	mgr := NewManager()
	mgr.CreateUser("alice")

	key1 := generateSSHKey(t, "alice@test")
	key2 := generateSSHKey(t, "alice2@test")
	mgr.AddKeyToUser("alice", key1)
	mgr.AddKeyToUser("alice", key2)

	// Get fingerprint of key1
	fp1 := ssh.FingerprintSHA256(key1)

	// Remove key1
	err := mgr.RemoveKeyFromUser("alice", fp1)
	if err != nil {
		t.Fatalf("RemoveKeyFromUser failed: %v", err)
	}

	// Verify key was removed
	user := mgr.GetUser("alice")
	if len(user.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(user.Keys))
	}

	// Try to remove non-existent key
	err = mgr.RemoveKeyFromUser("alice", "nonexistent")
	if err == nil {
		t.Error("RemoveKeyFromUser should return error for non-existent key")
	}
}

func TestUserHasKey(t *testing.T) {
	mgr := NewManager()
	mgr.CreateUser("alice")

	key := generateSSHKey(t, "alice@test")

	// User should not have key initially
	user := mgr.GetUser("alice")
	if user.HasKey(key) {
		t.Error("User should not have key before adding")
	}

	// Add key
	mgr.AddKeyToUser("alice", key)

	// User should have key now
	user = mgr.GetUser("alice")
	if !user.HasKey(key) {
		t.Error("User should have key after adding")
	}
}

func TestUserAddKeyDuplicate(t *testing.T) {
	user := &User{Name: "alice", Keys: []ssh.PublicKey{}}
	key := generateSSHKey(t, "alice@test")

	// Add key first time
	err := user.AddKey(key)
	if err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}

	// Try to add duplicate
	err = user.AddKey(key)
	if err == nil {
		t.Error("AddKey should return error for duplicate key")
	}
}

func TestUserRemoveKey(t *testing.T) {
	user := &User{Name: "alice", Keys: []ssh.PublicKey{}}
	key := generateSSHKey(t, "alice@test")
	user.AddKey(key)

	fp := ssh.FingerprintSHA256(key)

	// Remove key
	removed := user.RemoveKey(fp)
	if !removed {
		t.Error("RemoveKey should return true for existing key")
	}

	if len(user.Keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(user.Keys))
	}

	// Try to remove non-existent key
	removed = user.RemoveKey(fp)
	if removed {
		t.Error("RemoveKey should return false for non-existent key")
	}
}

func TestAuthenticate(t *testing.T) {
	mgr := NewManager()

	// Generate keys for users
	aliceKey := generateSSHKey(t, "alice@test")
	bobKey := generateSSHKey(t, "bob@test")

	// Create users with keys
	mgr.CreateUser("alice")
	mgr.CreateUser("bob")
	mgr.AddKeyToUser("alice", aliceKey)
	mgr.AddKeyToUser("bob", bobKey)

	// Set admin key
	adminKey := generateRSAKey(t, "admin@test")
	mgr.SetAdminKey(adminKey)

	// Test admin authentication
	user, userType := mgr.Authenticate(adminKey)
	if userType != UserTypeAdmin {
		t.Errorf("Expected UserTypeAdmin, got %v", userType)
	}
	if user == nil || user.Name != "admin" {
		t.Error("Expected admin user")
	}

	// Test normal user authentication
	user, userType = mgr.Authenticate(aliceKey)
	if userType != UserTypeNormal {
		t.Errorf("Expected UserTypeNormal, got %v", userType)
	}
	if user == nil || user.Name != "alice" {
		t.Error("Expected alice user")
	}

	// Test unknown key
	unknownKey := generateSSHKey(t, "unknown@test")
	user, userType = mgr.Authenticate(unknownKey)
	if userType != UserTypeUnknown {
		t.Errorf("Expected UserTypeUnknown, got %v", userType)
	}
	if user != nil {
		t.Error("Expected nil user for unknown key")
	}
}

func TestAuthenticateUnknownKey(t *testing.T) {
	mgr := NewManager()

	// Create user
	mgr.CreateUser("alice")

	// Generate a key that's not registered
	unknownKey := generateSSHKey(t, "unknown@test")
	user, userType := mgr.Authenticate(unknownKey)
	if userType != UserTypeUnknown {
		t.Errorf("Expected UserTypeUnknown, got %v", userType)
	}
	if user != nil {
		t.Error("Expected nil user for unknown key")
	}
}

func TestSetAdminKey(t *testing.T) {
	mgr := NewManager()
	key := generateSSHKey(t, "admin@test")

	// Initially no admin key set
	unknownKey := generateSSHKey(t, "unknown@test")
	_, userType := mgr.Authenticate(unknownKey)
	if userType != UserTypeUnknown {
		t.Error("Should not authenticate before setting admin key")
	}

	// Set admin key
	mgr.SetAdminKey(key)

	// Should authenticate now
	_, userType = mgr.Authenticate(key)
	if userType != UserTypeAdmin {
		t.Errorf("Expected UserTypeAdmin after setting admin key, got %v", userType)
	}
}

func TestManagerSaveAndLoadUsers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "users.json")

	// Create manager with users
	mgr := NewManager()
	mgr.CreateUser("alice")
	mgr.CreateUser("bob")
	mgr.AddKeyToUser("alice", generateSSHKey(t, "alice@test"))
	mgr.AddKeyToUser("alice", generateSSHKey(t, "alice2@test"))
	mgr.AddKeyToUser("bob", generateRSAKey(t, "bob@test"))

	// Save
	err = mgr.SaveToFile(testPath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Create new manager and load
	newMgr := NewManager()
	err = newMgr.LoadFromFile(testPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify loaded data
	alice := newMgr.GetUser("alice")
	if alice == nil {
		t.Fatal("alice not found after loading")
	}
	if len(alice.Keys) != 2 {
		t.Errorf("Expected alice to have 2 keys, got %d", len(alice.Keys))
	}

	bob := newMgr.GetUser("bob")
	if bob == nil {
		t.Fatal("bob not found after loading")
	}
	if len(bob.Keys) != 1 {
		t.Errorf("Expected bob to have 1 key, got %d", len(bob.Keys))
	}
}

func TestManagerLoadNonExistent(t *testing.T) {
	mgr := NewManager()
	err := mgr.LoadFromFile("/nonexistent/path/users.json")
	if err != nil {
		t.Errorf("LoadFromFile should not return error for non-existent file, got: %v", err)
	}
}

func TestGetUserNonExistent(t *testing.T) {
	mgr := NewManager()
	user := mgr.GetUser("nonexistent")
	if user != nil {
		t.Error("GetUser should return nil for non-existent user")
	}
}

func TestConcurrentAccess(t *testing.T) {
	mgr := NewManager()
	done := make(chan bool)

	// Concurrent user creation
	for i := 0; i < 10; i++ {
		go func(id int) {
			mgr.CreateUser("user" + string(rune('a'+id)))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all users were created
	users := mgr.ListUsers()
	if len(users) != 10 {
		t.Errorf("Expected 10 users, got %d", len(users))
	}
}
