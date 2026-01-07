package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/touken928/gitlite/internal/storage"
)

// RepoManager defines the interface for repository management
type RepoManager interface {
	// Create creates a new bare git repository
	Create(name string) error
	// Delete removes a repository and its data
	Delete(name string) error
	// Get returns a repository by name
	Get(name string) *Repository
	// List returns all repositories
	List() []*Repository
	// AddUser grants a user access to a repository
	AddUser(repoName, userName string, perm Permission) error
	// RemoveUser revokes a user's access to a repository
	RemoveUser(repoName, userName string) error
	// CheckPermission verifies if a user has the required access
	CheckPermission(repoName, userName string, needWrite bool) bool
	// GetRepoPath returns the filesystem path for a repository
	GetRepoPath(name string) string
	// SaveToFile persists repository permissions to a JSON file
	SaveToFile(path string) error
	// LoadFromFile loads repository permissions from a JSON file
	LoadFromFile(path string) error
}

// Manager handles repository management and provides thread-safe operations
type Manager struct {
	mu       sync.RWMutex
	basePath string                 // Base directory for repository storage
	repos    map[string]*Repository // Map of repository name to Repository struct
}

var _ RepoManager = (*Manager)(nil)

// NewManager creates a new Manager instance
func NewManager(basePath string) *Manager {
	return &Manager{
		basePath: basePath,
		repos:    make(map[string]*Repository),
	}
}

// Create initializes a new bare git repository
func (m *Manager) Create(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repos[name]; exists {
		return fmt.Errorf("repository %s already exists", name)
	}

	repoPath := filepath.Join(m.basePath, "repos", name+".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	// Initialize as a bare git repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		os.RemoveAll(repoPath)
		return fmt.Errorf("failed to init repository: %v", err)
	}

	m.repos[name] = &Repository{
		Name:  name,
		Path:  repoPath,
		Users: make(map[string]Permission),
	}
	return nil
}

// Delete removes a repository and its data from disk
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repo, exists := m.repos[name]
	if !exists {
		return fmt.Errorf("repository %s does not exist", name)
	}

	if err := os.RemoveAll(repo.Path); err != nil {
		return err
	}

	delete(m.repos, name)
	return nil
}

// Get returns a repository by name
func (m *Manager) Get(name string) *Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()
	name = strings.TrimSuffix(name, ".git")
	return m.repos[name]
}

// List returns all repositories
func (m *Manager) List() []*Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}
	return repos
}

// AddUser grants a user access to a repository
func (m *Manager) AddUser(repoName, userName string, perm Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repo, exists := m.repos[repoName]
	if !exists {
		return fmt.Errorf("repository %s does not exist", repoName)
	}
	repo.Users[userName] = perm
	return nil
}

// RemoveUser revokes a user's access to a repository
func (m *Manager) RemoveUser(repoName, userName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repo, exists := m.repos[repoName]
	if !exists {
		return fmt.Errorf("repository %s does not exist", repoName)
	}
	delete(repo.Users, userName)
	return nil
}

// CheckPermission verifies if a user has the required access to a repository
func (m *Manager) CheckPermission(repoName, userName string, needWrite bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repoName = strings.TrimSuffix(repoName, ".git")
	repo, exists := m.repos[repoName]
	if !exists {
		return false
	}

	// Guest access allows read-only for unauthenticated users
	if !needWrite {
		guestPerm, hasGuest := repo.Users["guest"]
		if hasGuest && guestPerm >= PermRead {
			return true
		}
	}

	// Unauthenticated users can only access via guest
	if userName == "" {
		return false
	}

	// Check user's permission
	perm, ok := repo.Users[userName]
	if !ok {
		return false
	}

	if needWrite {
		return perm == PermWrite
	}
	return perm >= PermRead
}

// GetRepoPath returns the filesystem path for a repository
func (m *Manager) GetRepoPath(name string) string {
	name = strings.TrimSuffix(name, ".git")
	return filepath.Join(m.basePath, "repos", name+".git")
}

// SaveToFile persists repository permissions to a JSON file
func (m *Manager) SaveToFile(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]storage.RepoPermission, 0, len(m.repos))
	for _, r := range m.repos {
		users := make(map[string]string)
		for u, p := range r.Users {
			if p == PermRead {
				users[u] = "r"
			} else if p == PermWrite {
				users[u] = "rw"
			}
		}
		repos = append(repos, storage.RepoPermission{
			Name:  r.Name,
			Path:  r.Path,
			Users: users,
		})
	}

	return storage.SaveRepoPermissions(path, repos)
}

// LoadFromFile loads repository permissions from a JSON file
func (m *Manager) LoadFromFile(path string) error {
	repoData, err := storage.LoadRepoPermissions(path)
	if err != nil {
		return err
	}
	if repoData == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rd := range repoData {
		// Only restore permissions if the repository exists on disk
		if _, err := os.Stat(rd.Path); os.IsNotExist(err) {
			continue
		}

		users := make(map[string]Permission)
		for u, pStr := range rd.Users {
			var perm Permission
			switch pStr {
			case "r":
				perm = PermRead
			case "rw":
				perm = PermWrite
			default:
				continue
			}
			users[u] = perm
		}

		// Only add if not already exists
		if _, exists := m.repos[rd.Name]; !exists {
			m.repos[rd.Name] = &Repository{
				Name:  rd.Name,
				Path:  rd.Path,
				Users: users,
			}
		}
	}

	return nil
}
