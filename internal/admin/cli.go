package admin

import (
	"fmt"
	"io"
	"strings"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/i18n"
	"github.com/touken928/gitlite/internal/repo"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// TUI provides an interactive command-line interface for server administration
type TUI struct {
	authMgr  auth.AuthManager      // User authentication manager interface
	repoMgr  repo.RepoManager      // Repository manager interface
	dataPath string                // Base directory for data storage
	sess     ssh.Session           // SSH session for I/O
	msg      i18n.Messages         // Localized messages
}

// New creates a new admin TUI instance
func New(authMgr auth.AuthManager, repoMgr repo.RepoManager, dataPath string) *TUI {
	return &TUI{
		authMgr:  authMgr,
		repoMgr:  repoMgr,
		dataPath: dataPath,
		msg:      i18n.GetMessages("en"),
	}
}

// setLang switches the display language
func (t *TUI) setLang(lang string) {
	t.msg = i18n.GetMessages(lang)
}

// saveData persists user and repository data to disk
func (t *TUI) saveData() {
	if err := t.authMgr.SaveToFile(t.dataPath + "/users.json"); err != nil {
		t.writeln(t.msg.SaveUserDataFailed + err.Error())
	}
	if err := t.repoMgr.SaveToFile(t.dataPath + "/repos.json"); err != nil {
		t.writeln(t.msg.SaveRepoDataFailed + err.Error())
	}
}

// Run starts the admin TUI main loop
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
			t.writeln(t.msg.UnknownCommand + cmd)
		}
	}
}

// write sends a string to the SSH session
func (t *TUI) write(s string) {
	io.WriteString(t.sess, s)
}

// writeln sends a string with newline to the SSH session
func (t *TUI) writeln(s string) {
	io.WriteString(t.sess, s+"\r\n")
}

// readLine reads a line from the SSH session with basic line editing
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
		case 127, '\b': // Backspace
			if len(line) > 0 {
				line = line[:len(line)-1]
				t.write("\b \b")
			}
		case 3: // Ctrl+C - interrupt
			t.write("^C\r\n")
			return "", io.EOF
		case 4: // Ctrl+D - EOF
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

// showHelp displays available commands
func (t *TUI) showHelp() {
	help := t.msg.HelpRepoList + "\n" +
		t.msg.HelpRepoCreate + "\n" +
		t.msg.HelpRepoDelete + "\n" +
		t.msg.HelpRepoAddUser + "\n" +
		t.msg.HelpRepoDelUser + "\n" +
		t.msg.HelpUserList + "\n" +
		t.msg.HelpUserCreate + "\n" +
		t.msg.HelpUserDelete + "\n" +
		t.msg.HelpUserAddKey + "\n" +
		t.msg.HelpUserDelKey + "\n" +
		t.msg.HelpUserKeys + "\n" +
		t.msg.HelpLang + "\n" +
		t.msg.HelpHelp + "\n" +
		t.msg.HelpQuit + "\n\n" +
		t.msg.HelpNote
	t.writeln(help)
}

// handleLang processes the language switching command
func (t *TUI) handleLang(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg.CurrentLang + t.msg.CurrentLang[:len(t.msg.CurrentLang)-1])
		return
	}
	switch args[0] {
	case "zh":
		t.setLang("zh")
		t.writeln(t.msg.LangSwitched)
	case "en":
		t.setLang("en")
		t.writeln(t.msg.LangSwitched)
	default:
		t.writeln(t.msg.LangUsage)
	}
}

// handleRepo processes repository management commands
func (t *TUI) handleRepo(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg.RepoUsage)
		return
	}

	switch args[0] {
	case "list":
		repos := t.repoMgr.List()
		if len(repos) == 0 {
			t.writeln(t.msg.NoRepositories)
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
			t.writeln(t.msg.RepoCreateUsage)
			return
		}
		if err := t.repoMgr.Create(args[1]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(fmt.Sprintf(t.msg.RepoCreated, args[1]))
		t.saveData()

	case "delete":
		if len(args) < 2 {
			t.writeln(t.msg.RepoDeleteUsage)
			return
		}
		if err := t.repoMgr.Delete(args[1]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(fmt.Sprintf(t.msg.RepoDeleted, args[1]))
		t.saveData()

	case "adduser":
		if len(args) < 4 {
			t.writeln(t.msg.RepoAddUserUsage)
			return
		}
		userName := args[2]
		var perm repo.Permission
		switch args[3] {
		case "r":
			perm = repo.PermRead
		case "rw":
			if userName == "guest" {
				t.writeln(t.msg.GuestReadOnly)
				return
			}
			perm = repo.PermWrite
		default:
			t.writeln(t.msg.PermissionInvalid)
			return
		}
		if userName != "guest" && t.authMgr.GetUser(userName) == nil {
			t.writeln(t.msg.UserNotFound)
			return
		}
		if err := t.repoMgr.AddUser(args[1], userName, perm); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(t.msg.UserAdded)
		t.saveData()

	case "deluser":
		if len(args) < 3 {
			t.writeln(t.msg.RepoDelUserUsage)
			return
		}
		if err := t.repoMgr.RemoveUser(args[1], args[2]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(t.msg.UserRemoved)
		t.saveData()

	default:
		t.writeln(t.msg.UnknownRepoCommand)
	}
}

// handleUser processes user management commands
func (t *TUI) handleUser(args []string) {
	if len(args) == 0 {
		t.writeln(t.msg.UserUsage)
		return
	}

	switch args[0] {
	case "list":
		users := t.authMgr.ListUsers()
		if len(users) == 0 {
			t.writeln(t.msg.NoUsers)
			return
		}
		for _, u := range users {
			t.writeln(fmt.Sprintf("  %s (%d %s)", u.Name, len(u.Keys), t.msg.KeysCount))
		}

	case "create":
		if len(args) < 2 {
			t.writeln(t.msg.UserCreateUsage)
			return
		}
		if args[1] == "guest" {
			t.writeln(t.msg.CannotCreateGuest)
			return
		}
		if err := t.authMgr.CreateUser(args[1]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(fmt.Sprintf(t.msg.UserCreated, args[1]))
		t.saveData()

	case "delete":
		if len(args) < 2 {
			t.writeln(t.msg.UserDeleteUsage)
			return
		}
		if args[1] == "guest" {
			t.writeln(t.msg.CannotDeleteGuest)
			return
		}
		if err := t.authMgr.DeleteUser(args[1]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(fmt.Sprintf(t.msg.UserDeleted, args[1]))
		t.saveData()

	case "addkey":
		if len(args) < 3 {
			t.writeln(t.msg.UserAddKeyUsage)
			return
		}
		keyStr := strings.Join(args[2:], " ")
		pubKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(keyStr))
		if err != nil {
			t.writeln(t.msg.InvalidPublicKey + err.Error())
			return
		}
		if err := t.authMgr.AddKeyToUser(args[1], pubKey); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(t.msg.KeyAdded)
		t.saveData()

	case "delkey":
		if len(args) < 3 {
			t.writeln(t.msg.UserDelKeyUsage)
			return
		}
		if err := t.authMgr.RemoveKeyFromUser(args[1], args[2]); err != nil {
			t.writeln(t.msg.Error + err.Error())
			return
		}
		t.writeln(t.msg.KeyRemoved)
		t.saveData()

	case "keys":
		if len(args) < 2 {
			t.writeln(t.msg.UserKeysUsage)
			return
		}
		user := t.authMgr.GetUser(args[1])
		if user == nil {
			t.writeln(t.msg.UserNotFound)
			return
		}
		if len(user.Keys) == 0 {
			t.writeln(t.msg.NoKeys)
			return
		}
		for _, k := range user.Keys {
			t.writeln("  " + gossh.FingerprintSHA256(k))
		}

	default:
		t.writeln(t.msg.UnknownUserCommand)
	}
}
