package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("/test/path")
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.repos == nil {
		t.Error("repos map is nil")
	}
	if len(mgr.repos) != 0 {
		t.Errorf("Expected empty repos map, got %d repos", len(mgr.repos))
	}
	if mgr.basePath != "/test/path" {
		t.Errorf("Expected basePath '/test/path', got '%s'", mgr.basePath)
	}
}

func TestCreateRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Create a repo
	err = mgr.Create("myproject")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify repo exists
	repo := mgr.Get("myproject")
	if repo == nil {
		t.Fatal("Get returned nil for myproject")
	}
	if repo.Name != "myproject" {
		t.Errorf("Expected repo name 'myproject', got '%s'", repo.Name)
	}

	// Verify repo path
	expectedPath := filepath.Join(tmpDir, "repos", "myproject.git")
	if repo.Path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, repo.Path)
	}

	// Verify repo directory was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Repo directory was not created")
	}

	// Try to create duplicate repo
	err = mgr.Create("myproject")
	if err == nil {
		t.Error("Create should return error for duplicate repo")
	}
}

func TestDeleteRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Create a repo
	mgr.Create("myproject")

	// Delete repo
	err = mgr.Delete("myproject")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify repo is gone
	repo := mgr.Get("myproject")
	if repo != nil {
		t.Error("Get should return nil after delete")
	}

	// Verify repo directory was removed
	repoPath := filepath.Join(tmpDir, "repos", "myproject.git")
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("Repo directory was not removed")
	}

	// Try to delete non-existent repo
	err = mgr.Delete("nonexistent")
	if err == nil {
		t.Error("Delete should return error for non-existent repo")
	}
}

func TestListRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// List empty repos
	repos := mgr.List()
	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}

	// Create repos
	mgr.Create("project1")
	mgr.Create("project2")
	mgr.Create("project3")

	// List repos
	repos = mgr.List()
	if len(repos) != 3 {
		t.Errorf("Expected 3 repos, got %d", len(repos))
	}
}

func TestGetRepoWithSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.Create("myproject")

	// Get with .git suffix
	repo := mgr.Get("myproject.git")
	if repo == nil {
		t.Fatal("Get returned nil for myproject.git")
	}
	if repo.Name != "myproject" {
		t.Errorf("Expected repo name 'myproject', got '%s'", repo.Name)
	}
}

func TestGetRepoNonExistent(t *testing.T) {
	mgr := NewManager("/test")
	repo := mgr.Get("nonexistent")
	if repo != nil {
		t.Error("Get should return nil for non-existent repo")
	}
}

func TestAddUserToRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.Create("myproject")

	// Add user with read permission
	err = mgr.AddUser("myproject", "alice", PermRead)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	// Add user with write permission
	err = mgr.AddUser("myproject", "bob", PermWrite)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	// Verify permissions
	repo := mgr.Get("myproject")
	if repo.Users["alice"] != PermRead {
		t.Errorf("Expected alice to have PermRead, got %v", repo.Users["alice"])
	}
	if repo.Users["bob"] != PermWrite {
		t.Errorf("Expected bob to have PermWrite, got %v", repo.Users["bob"])
	}

	// Try to add user to non-existent repo
	err = mgr.AddUser("nonexistent", "alice", PermRead)
	if err == nil {
		t.Error("AddUser should return error for non-existent repo")
	}
}

func TestRemoveUserFromRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.Create("myproject")
	mgr.AddUser("myproject", "alice", PermRead)

	// Remove user
	err = mgr.RemoveUser("myproject", "alice")
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}

	// Verify user is removed
	repo := mgr.Get("myproject")
	if _, exists := repo.Users["alice"]; exists {
		t.Error("alice should be removed from repo")
	}

	// Try to remove non-existent user (silently succeeds in implementation)
	err = mgr.RemoveUser("myproject", "nonexistent")
	if err != nil {
		t.Errorf("RemoveUser should not return error for non-existent user: %v", err)
	}

	// Try to remove user from non-existent repo
	err = mgr.RemoveUser("nonexistent", "alice")
	if err == nil {
		t.Error("RemoveUser should return error for non-existent repo")
	}
}

func TestCheckPermission(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.Create("myproject")
	mgr.AddUser("myproject", "alice", PermRead)
	mgr.AddUser("myproject", "bob", PermWrite)
	mgr.AddUser("myproject", "guest", PermRead)

	// Alice has read permission
	if !mgr.CheckPermission("myproject", "alice", false) {
		t.Error("alice should have read permission")
	}
	if mgr.CheckPermission("myproject", "alice", true) {
		t.Error("alice should not have write permission")
	}

	// Bob has write permission
	if !mgr.CheckPermission("myproject", "bob", false) {
		t.Error("bob should have read permission")
	}
	if !mgr.CheckPermission("myproject", "bob", true) {
		t.Error("bob should have write permission")
	}

	// Guest can read (anonymous)
	if !mgr.CheckPermission("myproject", "", false) {
		t.Error("guest should allow anonymous read")
	}
	if mgr.CheckPermission("myproject", "", true) {
		t.Error("guest should not allow anonymous write")
	}

	// Non-existent repo
	if mgr.CheckPermission("nonexistent", "alice", false) {
		t.Error("Should not have permission for non-existent repo")
	}

	// User without permission (but guest is enabled, so they can read)
	// Note: when guest is added, any authenticated user can read
	if !mgr.CheckPermission("myproject", "charlie", false) {
		t.Error("charlie should have read permission when guest is enabled")
	}
	// But cannot write
	if mgr.CheckPermission("myproject", "charlie", true) {
		t.Error("charlie should not have write permission")
	}
}

func TestCheckPermissionNoGuest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	mgr.Create("myproject")
	mgr.AddUser("myproject", "bob", PermWrite)
	mgr.AddUser("myproject", "alice", PermRead)

	// Bob can write
	if !mgr.CheckPermission("myproject", "bob", true) {
		t.Error("bob should have write permission")
	}

	// Alice can read but not write
	if !mgr.CheckPermission("myproject", "alice", false) {
		t.Error("alice should have read permission")
	}
	if mgr.CheckPermission("myproject", "alice", true) {
		t.Error("alice should not have write permission")
	}

	// Unauthenticated user cannot access without guest
	if mgr.CheckPermission("myproject", "", false) {
		t.Error("unauthenticated user should not have read access without guest")
	}

	// User without any permission
	if mgr.CheckPermission("myproject", "charlie", false) {
		t.Error("charlie should not have permission")
	}
}

func TestGetRepoPath(t *testing.T) {
	tmpDir := "/test/data"
	mgr := NewManager(tmpDir)

	path := mgr.GetRepoPath("myproject")
	expected := filepath.Join(tmpDir, "repos", "myproject.git")
	if path != expected {
		t.Errorf("Expected '%s', got '%s'", expected, path)
	}

	// With .git suffix
	path = mgr.GetRepoPath("myproject.git")
	if path != expected {
		t.Errorf("Expected '%s', got '%s'", expected, path)
	}
}

func TestRepoSaveAndLoadPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create repo manager with permissions
	mgr := NewManager(tmpDir)
	mgr.Create("project1")
	mgr.Create("project2")
	mgr.AddUser("project1", "alice", PermRead)
	mgr.AddUser("project1", "bob", PermWrite)
	mgr.AddUser("project1", "guest", PermRead)

	testPath := filepath.Join(tmpDir, "repos.json")

	// Save
	err = mgr.SaveToFile(testPath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Create new manager and load
	newMgr := NewManager(tmpDir)
	err = newMgr.LoadFromFile(testPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify permissions were loaded (for existing repos)
	repo := newMgr.Get("project1")
	if repo == nil {
		t.Fatal("project1 not found after loading")
	}
	if repo.Users["alice"] != PermRead {
		t.Errorf("Expected alice to have PermRead, got %v", repo.Users["alice"])
	}
	if repo.Users["bob"] != PermWrite {
		t.Errorf("Expected bob to have PermWrite, got %v", repo.Users["bob"])
	}
}

func TestRepoLoadNonExistent(t *testing.T) {
	mgr := NewManager("/test")
	err := mgr.LoadFromFile("/nonexistent/path/repos.json")
	if err != nil {
		t.Errorf("LoadFromFile should not return error for non-existent file, got: %v", err)
	}
}

func TestRepoLoadSkipsNonExistentRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "repos.json")

	// Create a permission file with non-existent repo
	perms := `[{"name":"deleted-repo","path":"/nonexistent/path/deleted.git","users":{"alice":"rw"}}]`
	os.WriteFile(testPath, []byte(perms), 0644)

	mgr := NewManager(tmpDir)
	err = mgr.LoadFromFile(testPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Repo should not be added since path doesn't exist
	repo := mgr.Get("deleted-repo")
	if repo != nil {
		t.Error("Non-existent repo should not be loaded")
	}
}

func TestConcurrentRepoOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	done := make(chan bool)

	// Concurrent repo creation
	for i := 0; i < 10; i++ {
		go func(id int) {
			mgr.Create("project" + string(rune('a'+id)))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all repos were created
	repos := mgr.List()
	if len(repos) != 10 {
		t.Errorf("Expected 10 repos, got %d", len(repos))
	}
}
