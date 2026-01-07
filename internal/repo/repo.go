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

type Permission int

const (
	PermNone  Permission = 0
	PermRead  Permission = 1
	PermWrite Permission = 2
)

type Repository struct {
	Name     string
	Path     string
	Users    map[string]Permission // username -> permission
}

type Manager struct {
	mu       sync.RWMutex
	basePath string
	repos    map[string]*Repository
}

func NewManager(basePath string) *Manager {
	return &Manager{
		basePath: basePath,
		repos:    make(map[string]*Repository),
	}
}

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

func (m *Manager) Get(name string) *Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()
	name = strings.TrimSuffix(name, ".git")
	return m.repos[name]
}

func (m *Manager) List() []*Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}
	return repos
}

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

func (m *Manager) CheckPermission(repoName, userName string, needWrite bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repoName = strings.TrimSuffix(repoName, ".git")
	repo, exists := m.repos[repoName]
	if !exists {
		return false
	}

	// 检查 guest 权限（允许任何人只读）
	if !needWrite {
		guestPerm, hasGuest := repo.Users["guest"]
		if hasGuest && guestPerm >= PermRead {
			return true
		}
	}

	// 未认证用户只能通过 guest 访问
	if userName == "" {
		return false
	}

	// 检查用户自身权限
	perm, ok := repo.Users[userName]
	if !ok {
		return false
	}

	if needWrite {
		return perm == PermWrite
	}
	return perm >= PermRead
}

func (m *Manager) GetRepoPath(name string) string {
	name = strings.TrimSuffix(name, ".git")
	return filepath.Join(m.basePath, "repos", name+".git")
}

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
		// Only restore permissions if the repository exists
		if _, err := os.Stat(rd.Path); os.IsNotExist(err) {
			continue // Skip non-existent repos
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
