package auth

import (
	"fmt"
	"sync"

	"github.com/touken928/gitlite/internal/storage"

	"golang.org/x/crypto/ssh"
)

type UserType int

const (
	UserTypeUnknown UserType = iota
	UserTypeAdmin
	UserTypeNormal
)

type User struct {
	Name string
	Keys []ssh.PublicKey
}

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

func (u *User) RemoveKey(fingerprint string) bool {
	for i, k := range u.Keys {
		if ssh.FingerprintSHA256(k) == fingerprint {
			u.Keys = append(u.Keys[:i], u.Keys[i+1:]...)
			return true
		}
	}
	return false
}

func (u *User) HasKey(key ssh.PublicKey) bool {
	fingerprint := ssh.FingerprintSHA256(key)
	for _, k := range u.Keys {
		if ssh.FingerprintSHA256(k) == fingerprint {
			return true
		}
	}
	return false
}

type Manager struct {
	mu       sync.RWMutex
	adminKey ssh.PublicKey
	users    map[string]*User // name -> User
}

func NewManager() *Manager {
	return &Manager{
		users: make(map[string]*User),
	}
}

func (m *Manager) SetAdminKey(key ssh.PublicKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adminKey = key
}

func (m *Manager) Authenticate(key ssh.PublicKey) (*User, UserType) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 检查是否为管理员
	if m.adminKey != nil && ssh.FingerprintSHA256(key) == ssh.FingerprintSHA256(m.adminKey) {
		return &User{Name: "admin"}, UserTypeAdmin
	}

	// 检查是否为普通用户
	for _, user := range m.users {
		if user.HasKey(key) {
			return user, UserTypeNormal
		}
	}

	return nil, UserTypeUnknown
}

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

func (m *Manager) DeleteUser(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[name]; !exists {
		return fmt.Errorf("user %s does not exist", name)
	}
	delete(m.users, name)
	return nil
}

func (m *Manager) GetUser(name string) *User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[name]
}

func (m *Manager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users
}

func (m *Manager) AddKeyToUser(userName string, key ssh.PublicKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userName]
	if !exists {
		return fmt.Errorf("user %s does not exist", userName)
	}
	return user.AddKey(key)
}

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
