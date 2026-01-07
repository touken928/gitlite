package server

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/git"
	"github.com/touken928/gitlite/internal/logging"

	"github.com/gliderlabs/ssh"
)

// handlePublicKey validates SSH public keys for authentication
func (s *Server) handlePublicKey(ctx ssh.Context, key ssh.PublicKey) bool {
	user, userType := s.authMgr.Authenticate(key)

	ctx.SetValue("user", user)
	ctx.SetValue("userType", userType)

	// Allow admin, registered users, and guest access
	return true
}

// handleSession processes incoming SSH sessions for git or admin operations
func (s *Server) handleSession(sess ssh.Session) {
	user, _ := sess.Context().Value("user").(*auth.User)
	userType, _ := sess.Context().Value("userType").(auth.UserType)

	rawCmd := sess.RawCommand()

	// Empty command triggers admin TUI login
	if rawCmd == "" {
		if userType != auth.UserTypeAdmin {
			io.WriteString(sess, "Access denied: admin only\r\n")
			sess.Exit(1)
			return
		}
		s.tui.Run(sess)
		return
	}

	// Admin users cannot perform git operations
	if userType == auth.UserTypeAdmin {
		io.WriteString(sess, "Access denied: admins cannot perform Git operations\r\n")
		sess.Exit(1)
		return
	}

	// Parse and execute git command
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

	// Check user permissions
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
		logging.Get().Error("Git execution error", zap.Error(err))
		sess.Exit(1)
		return
	}
}
