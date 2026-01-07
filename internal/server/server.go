package server

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/config"
	"github.com/touken928/gitlite/internal/logging"
	"github.com/touken928/gitlite/internal/repo"

	"github.com/touken928/gitlite/internal/admin"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// Server represents the SSH git server instance
type Server struct {
	port     string              // SSH listening port
	dataPath string              // Base directory for data storage
	sshSrv   *ssh.Server         // SSH server instance
	authMgr  *auth.Manager       // User authentication manager
	repoMgr  *repo.Manager       // Repository manager
	tui      *admin.TUI          // Admin TUI instance
}

// New creates a new server instance with the given configuration
func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		port:     cfg.Port,
		dataPath: cfg.DataPath,
		authMgr:  auth.NewManager(),
		repoMgr:  repo.NewManager(cfg.DataPath),
	}
	s.tui = admin.New(s.authMgr, s.repoMgr, cfg.DataPath)

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Join(cfg.DataPath, "repos"), 0755); err != nil {
		return nil, err
	}

	// Load or generate host key for SSH
	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return nil, err
	}

	// Load admin public key if available
	if err := s.loadAdminKey(); err != nil {
		logging.Get().Warn("Admin public key not found, please create "+cfg.DataPath+"/admin.pub")
	}

	// Load persisted user data
	if err := s.authMgr.LoadFromFile(filepath.Join(cfg.DataPath, "users.json")); err != nil {
		logging.Get().Warn("Failed to load user data", zap.Error(err))
	}

	// Load persisted repository permission data
	if err := s.repoMgr.LoadFromFile(filepath.Join(cfg.DataPath, "repos.json")); err != nil {
		logging.Get().Warn("Failed to load repo permission data", zap.Error(err))
	}

	s.sshSrv = &ssh.Server{
		Addr:             ":" + cfg.Port,
		Handler:          s.handleSession,
		PublicKeyHandler: s.handlePublicKey,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			userType, _ := ctx.Value("userType").(auth.UserType)
			return userType == auth.UserTypeAdmin
		},
	}
	s.sshSrv.AddHostKey(hostKey)

	return s, nil
}

// Start begins listening for SSH connections
func (s *Server) Start() error {
	return s.sshSrv.ListenAndServe()
}

// Stop gracefully shuts down the server and persists data
func (s *Server) Stop() {
	// Persist user data
	if err := s.authMgr.SaveToFile(filepath.Join(s.dataPath, "users.json")); err != nil {
		logging.Get().Error("Failed to save user data", zap.Error(err))
	}

	// Persist repository permission data
	if err := s.repoMgr.SaveToFile(filepath.Join(s.dataPath, "repos.json")); err != nil {
		logging.Get().Error("Failed to save repo permission data", zap.Error(err))
	}

	s.sshSrv.Close()
}

// loadOrGenerateHostKey loads an existing host key or generates a new one
func (s *Server) loadOrGenerateHostKey() (ssh.Signer, error) {
	keyPath := filepath.Join(s.dataPath, "host_key")

	// Generate new key if it doesn't exist
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.dataPath, 0755); err != nil {
			return nil, err
		}
		logging.Get().Info("Generating new host key", zap.String("path", keyPath))
		return generateHostKey(keyPath)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(keyData)
}

// loadAdminKey loads the administrator's public key from file
func (s *Server) loadAdminKey() error {
	keyPath := filepath.Join(s.dataPath, "admin.pub")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	pubKey, _, _, _, err := gossh.ParseAuthorizedKey(keyData)
	if err != nil {
		return err
	}

	s.authMgr.SetAdminKey(pubKey)
	logging.Get().Info("Admin public key loaded")
	return nil
}
