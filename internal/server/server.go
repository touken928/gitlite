package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/git"
	"github.com/touken928/gitlite/internal/logger"
	"github.com/touken928/gitlite/internal/repo"
	"github.com/touken928/gitlite/internal/tui"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	port     string
	dataPath string
	sshSrv   *ssh.Server
	authMgr  *auth.Manager
	repoMgr  *repo.Manager
	tui      *tui.TUI
}

func New(port, dataPath string) (*Server, error) {
	s := &Server{
		port:     port,
		dataPath: dataPath,
		authMgr:  auth.NewManager(),
		repoMgr:  repo.NewManager(dataPath),
	}
	s.tui = tui.New(s.authMgr, s.repoMgr, dataPath)

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Join(dataPath, "repos"), 0755); err != nil {
		return nil, err
	}

	// Load or generate host key
	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return nil, err
	}

	// Load admin public key
	if err := s.loadAdminKey(); err != nil {
		logger.Get().Warn("Admin public key not found, please create "+dataPath+"/admin.pub")
	}

	// Load persisted user data
	if err := s.authMgr.LoadFromFile(filepath.Join(dataPath, "users.json")); err != nil {
		logger.Get().Warn("Failed to load user data", zap.Error(err))
	}

	// Load persisted repo permission data
	if err := s.repoMgr.LoadFromFile(filepath.Join(dataPath, "repos.json")); err != nil {
		logger.Get().Warn("Failed to load repo permission data", zap.Error(err))
	}

	s.sshSrv = &ssh.Server{
		Addr:             ":" + port,
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

func (s *Server) Start() error {
	return s.sshSrv.ListenAndServe()
}

func (s *Server) Stop() {
	// Save user data
	if err := s.authMgr.SaveToFile(filepath.Join(s.dataPath, "users.json")); err != nil {
		logger.Get().Error("Failed to save user data", zap.Error(err))
	}

	// Save repo permission data
	if err := s.repoMgr.SaveToFile(filepath.Join(s.dataPath, "repos.json")); err != nil {
		logger.Get().Error("Failed to save repo permission data", zap.Error(err))
	}

	s.sshSrv.Close()
}

func (s *Server) loadOrGenerateHostKey() (ssh.Signer, error) {
	keyPath := filepath.Join(s.dataPath, "host_key")

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.dataPath, 0755); err != nil {
			return nil, err
		}
		logger.Get().Info("Generating new host key", zap.String("path", keyPath))
		return generateHostKey(keyPath)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(keyData)
}

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
	logger.Get().Info("Admin public key loaded")
	return nil
}

func (s *Server) handlePublicKey(ctx ssh.Context, key ssh.PublicKey) bool {
	user, userType := s.authMgr.Authenticate(key)

	ctx.SetValue("user", user)
	ctx.SetValue("userType", userType)

	// Admin and known users are allowed to connect, unknown users are also allowed (guest access)
	return true
}

func (s *Server) handleSession(sess ssh.Session) {
	user, _ := sess.Context().Value("user").(*auth.User)
	userType, _ := sess.Context().Value("userType").(auth.UserType)

	rawCmd := sess.RawCommand()

	// No command = admin TUI login
	if rawCmd == "" {
		if userType != auth.UserTypeAdmin {
			io.WriteString(sess, "Access denied: admin only\r\n")
			sess.Exit(1)
			return
		}
		s.tui.Run(sess)
		return
	}

	// Has command = Git operation
	if userType == auth.UserTypeAdmin {
		io.WriteString(sess, "Access denied: admins cannot perform Git operations\r\n")
		sess.Exit(1)
		return
	}

	gitCmd, err := git.ParseCommand(rawCmd)
	if err != nil {
		io.WriteString(sess, fmt.Sprintf("Error: %v\r\n", err))
		sess.Exit(1)
		return
	}

	userName := ""
	if user != nil {
		userName = user.Name
	}

	if !s.repoMgr.CheckPermission(gitCmd.RepoPath, userName, gitCmd.IsWrite) {
		io.WriteString(sess, "Access denied: insufficient permissions\r\n")
		sess.Exit(1)
		return
	}

	repoFullPath := s.repoMgr.GetRepoPath(gitCmd.RepoPath)
	if _, err := os.Stat(repoFullPath); os.IsNotExist(err) {
		io.WriteString(sess, "Error: repository does not exist\r\n")
		sess.Exit(1)
		return
	}

	if err := git.Execute(sess, gitCmd, repoFullPath); err != nil {
		logger.Get().Error("Git execution error", zap.Error(err))
		sess.Exit(1)
		return
	}
}
