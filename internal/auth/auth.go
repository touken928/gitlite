package auth

import (
	"fmt"
	"sync"

	"github.com/touken928/gitlite/internal/storage"

	"golang.org/x/crypto/ssh"
)

// AuthManager defines the interface for user authentication management
type AuthManager interface {
	// Authenticate validates an SSH public key and returns the user and their type
	Authenticate(key ssh.PublicKey) (*User, UserType)
	// CreateUser creates a new user with the given name
	CreateUser(name string) error
	// DeleteUser removes a user by name
	DeleteUser(name string) error
	// GetUser returns a user by name, or nil if not found
	GetUser(name string) *User
	// ListUsers returns all users in the system
	ListUsers() []*User
	// AddKeyToUser adds an SSH key to a user
	AddKeyToUser(userName string, key ssh.PublicKey) error
	// RemoveKeyFromUser removes an SSH key from a user by fingerprint
	RemoveKeyFromUser(userName, fingerprint string) error
	// SaveToFile persists user data to a JSON file
	SaveToFile(path string) error
	// LoadFromFile loads user data from a JSON file
	LoadFromFile(path string) error
}

// Manager handles user authentication and provides thread-safe operations
type Manager struct {
	mu       sync.RWMutex
	adminKey ssh.PublicKey      // SSH public key for admin authentication
	users    map[string]*User   // Map of username to User struct
}

var _ AuthManager = (*Manager)(nil)

// NewManager creates a new Manager instance
func NewManager() *Manager {
	return &Manager{
		users: make(map[string]*User),
	}
}

// SetAdminKey sets the administrator's SSH public key
func (m *Manager) SetAdminKey(key ssh.PublicKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adminKey = key
}

// Authenticate validates an SSH public key and returns the corresponding user and type
func (m *Manager) Authenticate(key ssh.PublicKey) (*User, UserType) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if the key matches the admin key
	if m.adminKey != nil && ssh.FingerprintSHA256(key) == ssh.FingerprintSHA256(m.adminKey) {
		return &User{Name: "admin"}, UserTypeAdmin
	}

	// Check if the key belongs to any registered user
	for _, user := range m.users {
		if user.HasKey(key) {
			return user, UserTypeNormal
		}
	}

	return nil, UserTypeUnknown
}

// CreateUser creates a new user with the given name
func (m *Manager) CreateUser(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "admin" {
		return fmt.Errorf("cannot create user named 'admin'")
	}
	if _, exists := m.users[name]; exists {
		return fmt.Errorf("user %s already exists", name)
	}

	m.users[name] = &User{Name: name, Keys: []ssh.PublicKey{}}
	return nil
}

// DeleteUser removes a user by name
func (m *Manager) DeleteUser(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[name]; !exists {
		return fmt.Errorf("user %s does not exist", name)
	}
	delete(m.users, name)
	return nil
}

// GetUser returns a user by name, or nil if not found
func (m *Manager) GetUser(name string) *User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[name]
}

// ListUsers returns all users in the system
func (m *Manager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users
}

// AddKeyToUser adds an SSH public key to a user
func (m *Manager) AddKeyToUser(userName string, key ssh.PublicKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userName]
	if !exists {
		return fmt.Errorf("user %s does not exist", userName)
	}
	return user.AddKey(key)
}

// RemoveKeyFromUser removes an SSH public key from a user by fingerprint
func (m *Manager) RemoveKeyFromUser(userName, fingerprint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userName]
	if !exists {
		return fmt.Errorf("user %s does not exist", userName)
	}
	if !user.RemoveKey(fingerprint) {
		return fmt.Errorf("key not found")
	}
	return nil
}

// SaveToFile persists user data to a JSON file
func (m *Manager) SaveToFile(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]storage.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, storage.User{
			Name: u.Name,
			Keys: storage.SaveSSHKeys(u.Keys),
		})
	}

	return storage.SaveUsers(path, users)
}

// LoadFromFile loads user data from a JSON file
func (m *Manager) LoadFromFile(path string) error {
	userData, err := storage.LoadUsers(path)
	if err != nil {
		return err
	}
	if userData == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ud := range userData {
		user := &User{
			Name: ud.Name,
			Keys: storage.LoadSSHKeys(ud.Keys),
		}
		m.users[ud.Name] = user
	}

	return nil
}
