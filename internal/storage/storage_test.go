package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Valid test SSH keys generated for testing
const (
	testKeyEd25519_1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAg61800+D56y8U6E5LmrEzPxOcjwx/pxciKJ6eZcKDL test1@test"
	testKeyEd25519_2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIVa4nvFex8tAbrGhRCbOdZnOsPqEihP0WkZtomoAwNd test2@test"
	testKeyRSA      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxzYur0E+TOY+a3Nyq60owYubSZGq1uljXzlg9Y9s5RqIUsQc8CyD45Q6Jk9nmU/p7Kh8hD2D8+zhPmEzr+4C5MUudjK+hFLKWKkAfCyYZ/hitZuP3Eg8WzsrD41YdDqtvhAhWfIL/Kbg6V+zyakkfrpz3LZyXOUpr9mHVYoWo23XixVpGWQCnNXNdLuO0HNQenZdSyfyKuCB6MXDdAXXiSSE1Nk/qTg5GGqZryxHFWRcedpbFcEvOXeKlrVuVIPNCMkACfr3phls5DMiD7wWpI+6oO/+uxPvUJuoZgoxXowxvgkAYgHTleJ7VqMpmA15J4NmqARKLx7+dXmbafCj/ test3@test"
)

func TestSaveAndLoadUsers(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "users.json")

	// Test data using valid SSH keys
	users := []User{
		{
			Name: "alice",
			Keys: []string{
				testKeyEd25519_1,
				testKeyEd25519_2,
			},
		},
		{
			Name: "bob",
			Keys: []string{
				testKeyRSA,
			},
		},
	}

	// Save users
	err = SaveUsers(testPath, users)
	if err != nil {
		t.Fatalf("SaveUsers failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("users.json was not created")
	}

	// Load users
	loaded, err := LoadUsers(testPath)
	if err != nil {
		t.Fatalf("LoadUsers failed: %v", err)
	}

	// Verify loaded data
	if len(loaded) != 2 {
		t.Errorf("Expected 2 users, got %d", len(loaded))
	}

	if loaded[0].Name != "alice" {
		t.Errorf("Expected first user name 'alice', got '%s'", loaded[0].Name)
	}

	if len(loaded[0].Keys) != 2 {
		t.Errorf("Expected 2 keys for alice, got %d", len(loaded[0].Keys))
	}
}

func TestLoadUsersNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "nonexistent.json")

	// Load non-existent file should return nil, nil
	users, err := LoadUsers(testPath)
	if err != nil {
		t.Errorf("LoadUsers should not return error for non-existent file, got: %v", err)
	}
	if users != nil {
		t.Error("LoadUsers should return nil for non-existent file")
	}
}

func TestLoadUsersEmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "empty.json")

	// Create empty file
	os.WriteFile(testPath, []byte{}, 0644)

	// Load empty file should return nil, nil
	users, err := LoadUsers(testPath)
	if err != nil {
		t.Errorf("LoadUsers should not return error for empty file, got: %v", err)
	}
	if users != nil {
		t.Error("LoadUsers should return nil for empty file")
	}
}

func TestLoadUsersInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "invalid.json")

	// Create invalid JSON file
	os.WriteFile(testPath, []byte("{invalid json}"), 0644)

	// Load invalid JSON should return error
	_, err = LoadUsers(testPath)
	if err == nil {
		t.Error("LoadUsers should return error for invalid JSON")
	}
}

func TestSaveUsersReadonlyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create readonly directory
	readonlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readonlyDir, 0555)
	testPath := filepath.Join(readonlyDir, "users.json")

	users := []User{{Name: "test", Keys: []string{}}}

	// Save to readonly directory should fail
	err = SaveUsers(testPath, users)
	if err == nil {
		t.Error("SaveUsers should return error for readonly directory")
	}
}

func TestLoadAndSaveSSHKeys(t *testing.T) {
	// Load keys
	keys := LoadSSHKeys([]string{testKeyEd25519_1, testKeyEd25519_2, testKeyRSA})
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify keys are valid ssh.PublicKey
	for i, key := range keys {
		if key == nil {
			t.Errorf("Key %d is nil", i)
		}
		// Verify we can get the fingerprint
		fp := ssh.FingerprintSHA256(key)
		if len(fp) == 0 {
			t.Errorf("Key %d has empty fingerprint", i)
		}
	}

	// Save keys back to strings
	savedStrs := SaveSSHKeys(keys)
	if len(savedStrs) != 3 {
		t.Errorf("Expected 3 saved key strings, got %d", len(savedStrs))
	}

	// Verify saved strings contain expected key type
	foundEd25519 := false
	foundRsa := false
	for _, s := range savedStrs {
		if len(s) > 0 {
			if strings.Contains(s, "ed25519") {
				foundEd25519 = true
			}
			if strings.Contains(s, "rsa") {
				foundRsa = true
			}
		}
	}

	if !foundEd25519 {
		t.Error("Expected to find ed25519 key in saved strings")
	}
	if !foundRsa {
		t.Error("Expected to find rsa key in saved strings")
	}
}

func TestLoadSSHKeysInvalidKeys(t *testing.T) {
	// Mix of valid and invalid keys
	keyStrs := []string{
		testKeyEd25519_1,
		"invalid key data",
		testKeyRSA,
		"",
		"another invalid key",
	}

	// Load keys - should skip invalid ones
	keys := LoadSSHKeys(keyStrs)
	if len(keys) != 2 {
		t.Errorf("Expected 2 valid keys, got %d", len(keys))
	}
}

func TestLoadSSHKeysEmpty(t *testing.T) {
	keys := LoadSSHKeys([]string{})
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for empty input, got %d", len(keys))
	}
}

func TestSaveSSHKeysEmpty(t *testing.T) {
	var keys []ssh.PublicKey
	savedStrs := SaveSSHKeys(keys)
	if len(savedStrs) != 0 {
		t.Errorf("Expected 0 saved strings for empty input, got %d", len(savedStrs))
	}
}

func TestSaveAndLoadRepoPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "repos.json")

	// Test data
	repos := []RepoPermission{
		{
			Name: "myproject",
			Path: "/data/repos/myproject.git",
			Users: map[string]string{
				"alice": "rw",
				"bob":   "r",
				"guest": "r",
			},
		},
		{
			Name: "another-repo",
			Path: "/data/repos/another-repo.git",
			Users: map[string]string{
				"alice": "rw",
			},
		},
	}

	// Save repos
	err = SaveRepoPermissions(testPath, repos)
	if err != nil {
		t.Fatalf("SaveRepoPermissions failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("repos.json was not created")
	}

	// Load repos
	loaded, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Fatalf("LoadRepoPermissions failed: %v", err)
	}

	// Verify loaded data
	if len(loaded) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(loaded))
	}

	if loaded[0].Name != "myproject" {
		t.Errorf("Expected first repo name 'myproject', got '%s'", loaded[0].Name)
	}

	if loaded[0].Users["alice"] != "rw" {
		t.Errorf("Expected alice permission 'rw', got '%s'", loaded[0].Users["alice"])
	}

	if loaded[0].Users["bob"] != "r" {
		t.Errorf("Expected bob permission 'r', got '%s'", loaded[0].Users["bob"])
	}

	if loaded[0].Users["guest"] != "r" {
		t.Errorf("Expected guest permission 'r', got '%s'", loaded[0].Users["guest"])
	}
}

func TestLoadRepoPermissionsNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "nonexistent.json")

	// Load non-existent file should return nil, nil
	repos, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Errorf("LoadRepoPermissions should not return error for non-existent file, got: %v", err)
	}
	if repos != nil {
		t.Error("LoadRepoPermissions should return nil for non-existent file")
	}
}

func TestLoadRepoPermissionsEmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "empty.json")

	// Create empty file
	os.WriteFile(testPath, []byte{}, 0644)

	// Load empty file should return nil, nil
	repos, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Errorf("LoadRepoPermissions should not return error for empty file, got: %v", err)
	}
	if repos != nil {
		t.Error("LoadRepoPermissions should return nil for empty file")
	}
}

func TestLoadRepoPermissionsInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "invalid.json")

	// Create invalid JSON file
	os.WriteFile(testPath, []byte("{invalid json}"), 0644)

	// Load invalid JSON should return error
	_, err = LoadRepoPermissions(testPath)
	if err == nil {
		t.Error("LoadRepoPermissions should return error for invalid JSON")
	}
}

func TestSaveRepoPermissionsReadonlyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create readonly directory
	readonlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readonlyDir, 0555)
	testPath := filepath.Join(readonlyDir, "repos.json")

	repos := []RepoPermission{{Name: "test", Path: "/test", Users: map[string]string{}}}

	// Save to readonly directory should fail
	err = SaveRepoPermissions(testPath, repos)
	if err == nil {
		t.Error("SaveRepoPermissions should return error for readonly directory")
	}
}

func TestRepoPermissionEmptyUsers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "repos.json")

	// Test with empty users map
	repos := []RepoPermission{
		{
			Name:  "empty-repo",
			Path:  "/data/repos/empty-repo.git",
			Users: map[string]string{},
		},
	}

	err = SaveRepoPermissions(testPath, repos)
	if err != nil {
		t.Fatalf("SaveRepoPermissions failed: %v", err)
	}

	loaded, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Fatalf("LoadRepoPermissions failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(loaded))
	}

	if len(loaded[0].Users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(loaded[0].Users))
	}
}

func TestUserRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "users_roundtrip.json")

	// Create users with SSH keys
	originalUsers := []User{
		{
			Name: "alice",
			Keys: []string{testKeyEd25519_1},
		},
	}

	// Save
	err = SaveUsers(testPath, originalUsers)
	if err != nil {
		t.Fatalf("SaveUsers failed: %v", err)
	}

	// Load
	loadedUsers, err := LoadUsers(testPath)
	if err != nil {
		t.Fatalf("LoadUsers failed: %v", err)
	}

	// Verify
	if len(loadedUsers) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(loadedUsers))
	}

	if loadedUsers[0].Name != "alice" {
		t.Errorf("Expected user name 'alice', got '%s'", loadedUsers[0].Name)
	}

	if len(loadedUsers[0].Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(loadedUsers[0].Keys))
	}

	// Verify key content (might have different newline handling)
	if !strings.Contains(loadedUsers[0].Keys[0], "AAAAC3NzaC1lZDI1NTE5") {
		t.Errorf("Key mismatch, expected key to contain fingerprint")
	}
}

func TestRepoPermissionRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "repos_roundtrip.json")

	// Create repos
	originalRepos := []RepoPermission{
		{
			Name: "test-repo",
			Path: "/var/data/repos/test-repo.git",
			Users: map[string]string{
				"admin": "rw",
				"user1": "r",
				"user2": "rw",
			},
		},
	}

	// Save
	err = SaveRepoPermissions(testPath, originalRepos)
	if err != nil {
		t.Fatalf("SaveRepoPermissions failed: %v", err)
	}

	// Load
	loadedRepos, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Fatalf("LoadRepoPermissions failed: %v", err)
	}

	// Verify
	if len(loadedRepos) != 1 {
		t.Fatalf("Expected 1 repo, got %d", len(loadedRepos))
	}

	if loadedRepos[0].Name != "test-repo" {
		t.Errorf("Expected repo name 'test-repo', got '%s'", loadedRepos[0].Name)
	}

	if loadedRepos[0].Path != "/var/data/repos/test-repo.git" {
		t.Errorf("Expected path '/var/data/repos/test-repo.git', got '%s'", loadedRepos[0].Path)
	}

	if loadedRepos[0].Users["admin"] != "rw" {
		t.Errorf("Expected admin permission 'rw', got '%s'", loadedRepos[0].Users["admin"])
	}

	if loadedRepos[0].Users["user1"] != "r" {
		t.Errorf("Expected user1 permission 'r', got '%s'", loadedRepos[0].Users["user1"])
	}

	if loadedRepos[0].Users["user2"] != "rw" {
		t.Errorf("Expected user2 permission 'rw', got '%s'", loadedRepos[0].Users["user2"])
	}
}

func TestMultipleUsersRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "multiple_users.json")

	// Create multiple users with multiple keys
	originalUsers := []User{
		{
			Name: "alice",
			Keys: []string{testKeyEd25519_1, testKeyEd25519_2},
		},
		{
			Name: "bob",
			Keys: []string{testKeyRSA},
		},
		{
			Name: "charlie",
			Keys: []string{},
		},
	}

	// Save
	err = SaveUsers(testPath, originalUsers)
	if err != nil {
		t.Fatalf("SaveUsers failed: %v", err)
	}

	// Load
	loadedUsers, err := LoadUsers(testPath)
	if err != nil {
		t.Fatalf("LoadUsers failed: %v", err)
	}

	// Verify
	if len(loadedUsers) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(loadedUsers))
	}

	// Check alice has 2 keys
	var alice *User
	for i := range loadedUsers {
		if loadedUsers[i].Name == "alice" {
			alice = &loadedUsers[i]
			break
		}
	}
	if alice == nil {
		t.Fatal("alice not found in loaded users")
	}
	if len(alice.Keys) != 2 {
		t.Errorf("Expected alice to have 2 keys, got %d", len(alice.Keys))
	}

	// Check charlie has no keys
	var charlie *User
	for i := range loadedUsers {
		if loadedUsers[i].Name == "charlie" {
			charlie = &loadedUsers[i]
			break
		}
	}
	if charlie == nil {
		t.Fatal("charlie not found in loaded users")
	}
	if len(charlie.Keys) != 0 {
		t.Errorf("Expected charlie to have 0 keys, got %d", len(charlie.Keys))
	}
}

func TestMultipleReposRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "multiple_repos.json")

	// Create multiple repos with various permissions
	originalRepos := []RepoPermission{
		{
			Name: "project-a",
			Path: "/data/repos/project-a.git",
			Users: map[string]string{
				"alice": "rw",
				"bob":   "r",
			},
		},
		{
			Name: "project-b",
			Path: "/data/repos/project-b.git",
			Users: map[string]string{
				"alice": "rw",
				"bob":   "rw",
				"guest": "r",
			},
		},
		{
			Name: "private-repo",
			Path: "/data/repos/private-repo.git",
			Users: map[string]string{
				"alice": "rw",
			},
		},
	}

	// Save
	err = SaveRepoPermissions(testPath, originalRepos)
	if err != nil {
		t.Fatalf("SaveRepoPermissions failed: %v", err)
	}

	// Load
	loadedRepos, err := LoadRepoPermissions(testPath)
	if err != nil {
		t.Fatalf("LoadRepoPermissions failed: %v", err)
	}

	// Verify
	if len(loadedRepos) != 3 {
		t.Fatalf("Expected 3 repos, got %d", len(loadedRepos))
	}

	// Check project-b has guest access
	var projectB *RepoPermission
	for i := range loadedRepos {
		if loadedRepos[i].Name == "project-b" {
			projectB = &loadedRepos[i]
			break
		}
	}
	if projectB == nil {
		t.Fatal("project-b not found in loaded repos")
	}
	if projectB.Users["guest"] != "r" {
		t.Errorf("Expected guest permission 'r' for project-b, got '%s'", projectB.Users["guest"])
	}
}
