package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gliderlabs/ssh"
)

var (
	// 白名单 Git 命令
	allowedCommands = map[string]bool{
		"git-upload-pack":  true, // clone, fetch, pull
		"git-receive-pack": true, // push
	}

	// 仓库路径校验
	repoPathRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+(/[a-zA-Z0-9_\-]+)*\.git$`)
)

type Command struct {
	Cmd      string
	RepoPath string
	IsWrite  bool
}

func ParseCommand(rawCmd string) (*Command, error) {
	parts := strings.SplitN(strings.TrimSpace(rawCmd), " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的命令格式")
	}

	cmd := parts[0]
	if !allowedCommands[cmd] {
		return nil, fmt.Errorf("不允许的命令: %s", cmd)
	}

	// 解析仓库路径，移除引号
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")

	if !repoPathRegex.MatchString(repoPath) {
		return nil, fmt.Errorf("无效的仓库路径: %s", repoPath)
	}

	return &Command{
		Cmd:      cmd,
		RepoPath: repoPath,
		IsWrite:  cmd == "git-receive-pack",
	}, nil
}

func Execute(sess ssh.Session, gitCmd *Command, repoFullPath string) error {
	cmd := exec.Command(gitCmd.Cmd, repoFullPath)
	cmd.Stdin = sess
	cmd.Stdout = sess
	cmd.Stderr = sess.Stderr()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git 命令执行失败: %v", err)
	}
	return nil
}
