package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gliderlabs/ssh"
)

// Allowed Git commands for repository operations
var (
	// Whitelist of permitted git commands
	allowedCommands = map[string]bool{
		"git-upload-pack":  true, // Used for clone, fetch, pull
		"git-receive-pack": true, // Used for push
	}

	// Regex pattern for validating repository paths
	repoPathRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+(/[a-zA-Z0-9_\-]+)*\.git$`)
)

// Command represents a parsed git command with its repository path
type Command struct {
	Cmd      string // Git command name (git-upload-pack or git-receive-pack)
	RepoPath string // Repository path (e.g., "myrepo.git")
	IsWrite  bool   // True if this is a write operation (push)
}

// ParseCommand parses a raw SSH git command string into a Command struct
func ParseCommand(rawCmd string) (*Command, error) {
	parts := strings.SplitN(strings.TrimSpace(rawCmd), " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid command format")
	}

	cmd := parts[0]
	if !allowedCommands[cmd] {
		return nil, fmt.Errorf("command not allowed: %s", cmd)
	}

	// Extract and normalize repository path
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")

	if !repoPathRegex.MatchString(repoPath) {
		return nil, fmt.Errorf("invalid repo path: %s", repoPath)
	}

	return &Command{
		Cmd:      cmd,
		RepoPath: repoPath,
		IsWrite:  cmd == "git-receive-pack",
	}, nil
}

// Execute runs a git command for the given repository
func Execute(sess ssh.Session, gitCmd *Command, repoFullPath string) error {
	cmd := exec.Command(gitCmd.Cmd, repoFullPath)
	cmd.Stdin = sess
	cmd.Stdout = sess
	cmd.Stderr = sess.Stderr()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git command failed: %v", err)
	}
	return nil
}
