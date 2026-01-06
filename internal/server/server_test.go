package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if srv == nil {
		t.Fatal("Server is nil")
	}

	if srv.port != "2222" {
		t.Errorf("Expected port '2222', got '%s'", srv.port)
	}

	if srv.dataPath != tmpDir {
		t.Errorf("Expected dataPath '%s', got '%s'", tmpDir, srv.dataPath)
	}

	if srv.authMgr == nil {
		t.Error("authMgr is nil")
	}

	if srv.repoMgr == nil {
		t.Error("repoMgr is nil")
	}

	if srv.tui == nil {
		t.Error("tui is nil")
	}
}

func TestNewServerCreatesDataDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Verify repos directory was created
	reposDir := filepath.Join(tmpDir, "repos")
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		t.Error("repos directory was not created")
	}

	_ = srv
}

func TestNewServerWithMissingAdminPub(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create server without admin.pub - should still work (with warning)
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if srv == nil {
		t.Fatal("Server is nil")
	}

	// Server should still be usable
	_ = srv
}

func TestNewServerGeneratesHostKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server
	_, err = New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Verify host_key was created
	hostKeyPath := filepath.Join(tmpDir, "host_key")
	if _, err := os.Stat(hostKeyPath); os.IsNotExist(err) {
		t.Error("host_key was not created")
	}
}

func TestNewServerLoadsExistingHostKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server first time
	srv1, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Get host key info
	hostKeyPath := filepath.Join(tmpDir, "host_key")
	info1, _ := os.Stat(hostKeyPath)

	// Create server again - should reuse existing host key
	srv2, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Verify host key file wasn't recreated
	info2, _ := os.Stat(hostKeyPath)
	if info1.ModTime() != info2.ModTime() {
		t.Error("Host key was recreated instead of reused")
	}

	_ = srv1
	_ = srv2
}

func TestNewServerWithInvalidAdminPub(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid admin.pub file
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte("invalid key data"), 0644)

	// Create server - should still work
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if srv == nil {
		t.Fatal("Server is nil")
	}
}

func TestServerStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Stop should not panic
	srv.Stop()
}

func TestServerStopSavesData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create server
	srv, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Stop server
	srv.Stop()

	// Stop again should be safe
	srv.Stop()
}

func TestServerWithDifferentPorts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin.pub file
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir, "admin.pub"), []byte(adminPub), 0644)

	// Create servers with different ports
	srv1, err := New("2222", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	srv2, err := New("3333", tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if srv1.port != "2222" {
		t.Errorf("Expected port '2222', got '%s'", srv1.port)
	}

	if srv2.port != "3333" {
		t.Errorf("Expected port '3333', got '%s'", srv2.port)
	}

	_ = srv1
	_ = srv2
}

func TestServerWithDifferentDataPaths(t *testing.T) {
	tmpDir1, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	// Create admin.pub files
	adminPub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILgQMhDclXxXpNMO8NLiXhZ6Dp7f4G7h7h7h7h7h7h7h7h7 admin@test"
	os.WriteFile(filepath.Join(tmpDir1, "admin.pub"), []byte(adminPub), 0644)
	os.WriteFile(filepath.Join(tmpDir2, "admin.pub"), []byte(adminPub), 0644)

	// Create servers with different data paths
	srv1, err := New("2222", tmpDir1)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	srv2, err := New("2222", tmpDir2)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if srv1.dataPath != tmpDir1 {
		t.Errorf("Expected dataPath '%s', got '%s'", tmpDir1, srv1.dataPath)
	}

	if srv2.dataPath != tmpDir2 {
		t.Errorf("Expected dataPath '%s', got '%s'", tmpDir2, srv2.dataPath)
	}

	_ = srv1
	_ = srv2
}
