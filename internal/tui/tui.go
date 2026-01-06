package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/repo"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type TUI struct {
	authMgr  *auth.Manager
	repoMgr  *repo.Manager
	dataPath string
	sess     ssh.Session
	lang     string // "zh" or "en"
}

func New(authMgr *auth.Manager, repoMgr *repo.Manager, dataPath string) *TUI {
	return &TUI{
		authMgr:  authMgr,
		repoMgr:  repoMgr,
		dataPath: dataPath,
		lang:     "en",
	}
}

func (t *TUI) saveData() {
	if err := t.authMgr.SaveToFile(t.dataPath + "/users.json"); err != nil {
		t.writeln(t.msg("保存用户数据失败: ", "Failed to save user data: ") + err.Error())
	}
	if err := t.repoMgr.SaveToFile(t.dataPath + "/repos.json"); err != nil {
		t.writeln(t.msg("保存仓库权限数据失败: ", "Failed to save repo permissions: ") + err.Error())
	}
}

func (t *TUI) Run(sess ssh.Session) {
	t.sess = sess
	t.writeln("")
	t.writeln("╔══════════════════════════════════════╗")
	t.writeln("║     Git Server Management System     ║")
	t.writeln("╚══════════════════════════════════════╝")
	t.writeln("")
	t.showHelp()

	for {
		t.write("\r\nadmin> ")
		line, err := t.readLine()
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]
		args = args[1:]

		switch cmd {
		case "help", "h":
			t.showHelp()
		case "lang":
			t.handleLang(args)
		case "quit", "exit", "q":
			t.writeln("Bye!")
			return
		case "repo":
			t.handleRepo(args)
		case "user":
			t.handleUser(args)
		default:
			t.writeln(t.msg("未知命令: ", "Unknown command: ") + cmd)
		}
	}
}

func (t *TUI) write(s string) {
	io.WriteString(t.sess, s)
}

func (t *TUI) writeln(s string) {
	io.WriteString(t.sess, s+"\r\n")
}

func (t *TUI) msg(zh, en string) string {
	if t.lang == "en" {
		return en
	}
	return zh
}

func (t *TUI) readLine() (string, error) {
	var line []byte
	buf := make([]byte, 1)

	for {
		n, err := t.sess.Read(buf)
		if err != nil || n == 0 {
			return "", err
		}

		ch := buf[0]
		switch ch {
		case '\r', '\n':
			t.write("\r\n")
			return string(line), nil
		case 127, '\b':
			if len(line) > 0 {
				line = line[:len(line)-1]
				t.write("\b \b")
			}
		case 3: // Ctrl+C
			t.write("^C\r\n")
			return "", io.EOF
		case 4: // Ctrl+D
			if len(line) == 0 {
				return "", io.EOF
			}
		default:
			if ch >= 32 && ch < 127 {
				line = append(line, ch)
				t.sess.Write([]byte{ch})
			}
		}
	}
}

func (t *TUI) showHelp() {
	if t.lang == "en" {
		t.writeln(`Commands:
  repo list                      - List all repositories
  repo create <name>             - Create a repository
  repo delete <name>             - Delete a repository
  repo adduser <repo> <user> <r|rw> - Add user to repository (r=read, rw=read-write)
  repo deluser <repo> <user>     - Remove user from repository

  user list                      - List all users
  user create <name>             - Create a user
  user delete <name>             - Delete a user
  user addkey <name> <pubkey>    - Add SSH key to user
  user delkey <name> <fingerprint> - Remove SSH key from user
  user keys <name>               - List user's SSH keys

  lang <zh|en>                   - Switch language
  help                           - Show this help
  quit                           - Exit

Note: "guest" is a built-in user for read-only access. Add guest to a repo
      with "repo adduser <repo> guest r" to allow all authenticated users
      to read that repository.`)
	} else {
		t.writeln(`命令:
  repo list                      - 列出所有仓库
  repo create <name>             - 创建仓库
  repo delete <name>             - 删除仓库
  repo adduser <repo> <user> <r|rw> - 添加用户到仓库 (r=只读, rw=读写)
  repo deluser <repo> <user>     - 从仓库移除用户

  user list                      - 列出所有用户
  user create <name>             - 创建用户
  user delete <name>             - 删除用户
  user addkey <name> <pubkey>    - 为用户添加 SSH 密钥
  user delkey <name> <fingerprint> - 删除用户的 SSH 密钥
  user keys <name>               - 列出用户的 SSH 密钥

  lang <zh|en>                   - 切换语言
  help                           - 显示帮助
  quit                           - 退出

说明: "guest" 是内置的只读访客用户。使用 "repo adduser <repo> guest r"
      可以让所有已认证用户都能读取该仓库。`)
	}
}

func (t *TUI) handleLang(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg("当前语言: "+t.lang, "Current language: "+t.lang))
		return
	}
	switch args[0] {
	case "zh", "en":
		t.lang = args[0]
		t.writeln(t.msg("语言已切换为中文", "Language switched to English"))
	default:
		t.writeln(t.msg("用法: lang <zh|en>", "Usage: lang <zh|en>"))
	}
}

func (t *TUI) handleRepo(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg("用法: repo <list|create|delete|adduser|deluser>", "Usage: repo <list|create|delete|adduser|deluser>"))
		return
	}

	switch args[0] {
	case "list":
		repos := t.repoMgr.List()
		if len(repos) == 0 {
			t.writeln(t.msg("  (暂无仓库)", "  (no repositories)"))
			return
		}
		for _, r := range repos {
			users := make([]string, 0)
			for u, p := range r.Users {
				perm := "r"
				if p == repo.PermWrite {
					perm = "rw"
				}
				users = append(users, fmt.Sprintf("%s(%s)", u, perm))
			}
			userStr := ""
			if len(users) > 0 {
				userStr = " [" + strings.Join(users, ", ") + "]"
			}
			t.writeln(fmt.Sprintf("  %s%s", r.Name, userStr))
		}

	case "create":
		if len(args) < 2 {
			t.writeln(t.msg("用法: repo create <name>", "Usage: repo create <name>"))
			return
		}
		if err := t.repoMgr.Create(args[1]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("仓库 "+args[1]+" 创建成功", "Repository "+args[1]+" created"))
		t.saveData()

	case "delete":
		if len(args) < 2 {
			t.writeln(t.msg("用法: repo delete <name>", "Usage: repo delete <name>"))
			return
		}
		if err := t.repoMgr.Delete(args[1]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("仓库 "+args[1]+" 已删除", "Repository "+args[1]+" deleted"))
		t.saveData()

	case "adduser":
		if len(args) < 4 {
			t.writeln(t.msg("用法: repo adduser <repo> <user> <r|rw>", "Usage: repo adduser <repo> <user> <r|rw>"))
			return
		}
		userName := args[2]
		var perm repo.Permission
		switch args[3] {
		case "r":
			perm = repo.PermRead
		case "rw":
			if userName == "guest" {
				t.writeln(t.msg("guest 用户只能设置只读权限", "Guest user can only have read permission"))
				return
			}
			perm = repo.PermWrite
		default:
			t.writeln(t.msg("权限必须是 r 或 rw", "Permission must be r or rw"))
			return
		}
		if userName != "guest" && t.authMgr.GetUser(userName) == nil {
			t.writeln(t.msg("用户不存在", "User not found"))
			return
		}
		if err := t.repoMgr.AddUser(args[1], userName, perm); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("已添加用户", "User added"))
		t.saveData()

	case "deluser":
		if len(args) < 3 {
			t.writeln(t.msg("用法: repo deluser <repo> <user>", "Usage: repo deluser <repo> <user>"))
			return
		}
		if err := t.repoMgr.RemoveUser(args[1], args[2]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("已移除用户", "User removed"))
		t.saveData()

	default:
		t.writeln(t.msg("未知的 repo 子命令", "Unknown repo subcommand"))
	}
}

func (t *TUI) handleUser(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg("用法: user <list|create|delete|addkey|delkey|keys>", "Usage: user <list|create|delete|addkey|delkey|keys>"))
		return
	}

	switch args[0] {
	case "list":
		users := t.authMgr.ListUsers()
		if len(users) == 0 {
			t.writeln(t.msg("  (暂无用户)", "  (no users)"))
			return
		}
		for _, u := range users {
			t.writeln(fmt.Sprintf("  %s (%d %s)", u.Name, len(u.Keys), t.msg("个密钥", "keys")))
		}

	case "create":
		if len(args) < 2 {
			t.writeln(t.msg("用法: user create <name>", "Usage: user create <name>"))
			return
		}
		if args[1] == "guest" {
			t.writeln(t.msg("不能创建名为 guest 的用户", "Cannot create user named guest"))
			return
		}
		if err := t.authMgr.CreateUser(args[1]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("用户 "+args[1]+" 创建成功", "User "+args[1]+" created"))
		t.saveData()

	case "delete":
		if len(args) < 2 {
			t.writeln(t.msg("用法: user delete <name>", "Usage: user delete <name>"))
			return
		}
		if args[1] == "guest" {
			t.writeln(t.msg("不能删除 guest 用户", "Cannot delete guest user"))
			return
		}
		if err := t.authMgr.DeleteUser(args[1]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("用户 "+args[1]+" 已删除", "User "+args[1]+" deleted"))
		t.saveData()

	case "addkey":
		if len(args) < 3 {
			t.writeln(t.msg("用法: user addkey <name> <pubkey>", "Usage: user addkey <name> <pubkey>"))
			return
		}
		keyStr := strings.Join(args[2:], " ")
		pubKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(keyStr))
		if err != nil {
			t.writeln(t.msg("无效的公钥: ", "Invalid public key: ") + err.Error())
			return
		}
		if err := t.authMgr.AddKeyToUser(args[1], pubKey); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("密钥添加成功", "Key added"))
		t.saveData()

	case "delkey":
		if len(args) < 3 {
			t.writeln(t.msg("用法: user delkey <name> <fingerprint>", "Usage: user delkey <name> <fingerprint>"))
			return
		}
		if err := t.authMgr.RemoveKeyFromUser(args[1], args[2]); err != nil {
			t.writeln(t.msg("错误: ", "Error: ") + err.Error())
			return
		}
		t.writeln(t.msg("密钥已删除", "Key removed"))
		t.saveData()

	case "keys":
		if len(args) < 2 {
			t.writeln(t.msg("用法: user keys <name>", "Usage: user keys <name>"))
			return
		}
		user := t.authMgr.GetUser(args[1])
		if user == nil {
			t.writeln(t.msg("用户不存在", "User not found"))
			return
		}
		if len(user.Keys) == 0 {
			t.writeln(t.msg("  (暂无密钥)", "  (no keys)"))
			return
		}
		for _, k := range user.Keys {
			t.writeln("  " + gossh.FingerprintSHA256(k))
		}

	default:
		t.writeln(t.msg("未知的 user 子命令", "Unknown user subcommand"))
	}
}
