package i18n

// Messages holds all localized message strings for the admin TUI
type Messages struct {
	// Common messages
	SaveUserDataFailed   string
	SaveRepoDataFailed   string
	UnknownCommand       string
	Error                string

	// Help command messages
	HelpRepoList         string
	HelpRepoCreate       string
	HelpRepoDelete       string
	HelpRepoAddUser      string
	HelpRepoDelUser      string
	HelpUserList         string
	HelpUserCreate       string
	HelpUserDelete       string
	HelpUserAddKey       string
	HelpUserDelKey       string
	HelpUserKeys         string
	HelpLang             string
	HelpHelp             string
	HelpQuit             string
	HelpNote             string

	// Language switching messages
	CurrentLang          string
	LangSwitched         string
	LangUsage            string

	// Repository management messages
	RepoUsage            string
	NoRepositories       string
	RepoCreateUsage      string
	RepoCreated          string
	RepoDeleteUsage      string
	RepoDeleted          string
	RepoAddUserUsage     string
	RepoDelUserUsage     string
	GuestReadOnly        string
	PermissionInvalid    string
	UserNotFound         string
	UserAdded            string
	UserRemoved          string
	UnknownRepoCommand   string

	// User management messages
	UserUsage            string
	NoUsers              string
	UserCreateUsage      string
	CannotCreateGuest    string
	UserCreated          string
	UserDeleteUsage      string
	CannotDeleteGuest    string
	UserDeleted          string
	UserAddKeyUsage      string
	InvalidPublicKey     string
	KeyAdded             string
	UserDelKeyUsage      string
	KeyRemoved           string
	UserKeysUsage        string
	NoKeys               string
	UnknownUserCommand   string

	// Miscellaneous
	KeysCount            string
}

// Translations maps language codes to their message sets
type Translations map[string]Messages

// GetMessages returns the message set for the specified language, defaulting to English
func GetMessages(lang string) Messages {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["en"]
}

// translations contains all supported language message sets
var translations = Translations{
	"en": {
		// Common
		SaveUserDataFailed:   "Failed to save user data: ",
		SaveRepoDataFailed:   "Failed to save repo permissions: ",
		UnknownCommand:       "Unknown command: ",
		Error:                "Error: ",

		// Help
		HelpRepoList:         "repo list                      - List all repositories",
		HelpRepoCreate:       "repo create <name>             - Create a repository",
		HelpRepoDelete:       "repo delete <name>             - Delete a repository",
		HelpRepoAddUser:      "repo adduser <repo> <user> <r|rw> - Add user to repository (r=read, rw=read-write)",
		HelpRepoDelUser:      "repo deluser <repo> <user>     - Remove user from repository",
		HelpUserList:         "user list                      - List all users",
		HelpUserCreate:       "user create <name>             - Create a user",
		HelpUserDelete:       "user delete <name>             - Delete a user",
		HelpUserAddKey:       "user addkey <name> <pubkey>    - Add SSH key to user",
		HelpUserDelKey:       "user delkey <name> <fingerprint> - Remove SSH key from user",
		HelpUserKeys:         "user keys <name>               - List user's SSH keys",
		HelpLang:             "lang <zh|en>                   - Switch language",
		HelpHelp:             "help                           - Show this help",
		HelpQuit:             "quit                           - Exit",
		HelpNote:             "Note: \"guest\" is a built-in user for read-only access. Add guest to a repo\n      with \"repo adduser <repo> guest r\" to allow all authenticated users\n      to read that repository.",

		// Lang
		CurrentLang:          "Current language: ",
		LangSwitched:         "Language switched to English",
		LangUsage:            "Usage: lang <zh|en>",

		// Repo
		RepoUsage:            "Usage: repo <list|create|delete|adduser|deluser>",
		NoRepositories:       "  (no repositories)",
		RepoCreateUsage:      "Usage: repo create <name>",
		RepoCreated:          "Repository %s created",
		RepoDeleteUsage:      "Usage: repo delete <name>",
		RepoDeleted:          "Repository %s deleted",
		RepoAddUserUsage:     "Usage: repo adduser <repo> <user> <r|rw>",
		RepoDelUserUsage:     "Usage: repo deluser <repo> <user>",
		GuestReadOnly:        "Guest user can only have read permission",
		PermissionInvalid:    "Permission must be r or rw",
		UserNotFound:         "User not found",
		UserAdded:            "User added",
		UserRemoved:          "User removed",
		UnknownRepoCommand:   "Unknown repo subcommand",

		// User
		UserUsage:            "Usage: user <list|create|delete|addkey|delkey|keys>",
		NoUsers:              "  (no users)",
		UserCreateUsage:      "Usage: user create <name>",
		CannotCreateGuest:    "Cannot create user named guest",
		UserCreated:          "User %s created",
		UserDeleteUsage:      "Usage: user delete <name>",
		CannotDeleteGuest:    "Cannot delete guest user",
		UserDeleted:          "User %s deleted",
		UserAddKeyUsage:      "Usage: user addkey <name> <pubkey>",
		InvalidPublicKey:     "Invalid public key: ",
		KeyAdded:             "Key added",
		UserDelKeyUsage:      "Usage: user delkey <name> <fingerprint>",
		KeyRemoved:           "Key removed",
		UserKeysUsage:        "Usage: user keys <name>",
		NoKeys:               "  (no keys)",
		UnknownUserCommand:   "Unknown user subcommand",

		// Misc
		KeysCount:            "keys",
	},
	"zh": {
		// Common
		SaveUserDataFailed:   "保存用户数据失败: ",
		SaveRepoDataFailed:   "保存仓库权限数据失败: ",
		UnknownCommand:       "未知命令: ",
		Error:                "错误: ",

		// Help
		HelpRepoList:         "repo list                      - 列出所有仓库",
		HelpRepoCreate:       "repo create <name>             - 创建仓库",
		HelpRepoDelete:       "repo delete <name>             - 删除仓库",
		HelpRepoAddUser:      "repo adduser <repo> <user> <r|rw> - 添加用户到仓库 (r=只读, rw=读写)",
		HelpRepoDelUser:      "repo deluser <repo> <user>     - 从仓库移除用户",
		HelpUserList:         "user list                      - 列出所有用户",
		HelpUserCreate:       "user create <name>             - 创建用户",
		HelpUserDelete:       "user delete <name>             - 删除用户",
		HelpUserAddKey:       "user addkey <name> <pubkey>    - 为用户添加 SSH 密钥",
		HelpUserDelKey:       "user delkey <name> <fingerprint> - 删除用户的 SSH 密钥",
		HelpUserKeys:         "user keys <name>               - 列出用户的 SSH 密钥",
		HelpLang:             "lang <zh|en>                   - 切换语言",
		HelpHelp:             "help                           - 显示帮助",
		HelpQuit:             "quit                           - 退出",
		HelpNote:             "说明: \"guest\" 是内置的只读访客用户。使用 \"repo adduser <repo> guest r\"\n      可以让所有已认证用户都能读取该仓库。",

		// Lang
		CurrentLang:          "当前语言: ",
		LangSwitched:         "语言已切换为中文",
		LangUsage:            "用法: lang <zh|en>",

		// Repo
		RepoUsage:            "用法: repo <list|create|delete|adduser|deluser>",
		NoRepositories:       "  (暂无仓库)",
		RepoCreateUsage:      "用法: repo create <name>",
		RepoCreated:          "仓库 %s 创建成功",
		RepoDeleteUsage:      "用法: repo delete <name>",
		RepoDeleted:          "仓库 %s 已删除",
		RepoAddUserUsage:     "用法: repo adduser <repo> <user> <r|rw>",
		RepoDelUserUsage:     "用法: repo deluser <repo> <user>",
		GuestReadOnly:        "guest 用户只能设置只读权限",
		PermissionInvalid:    "权限必须是 r 或 rw",
		UserNotFound:         "用户不存在",
		UserAdded:            "已添加用户",
		UserRemoved:          "已移除用户",
		UnknownRepoCommand:   "未知的 repo 子命令",

		// User
		UserUsage:            "用法: user <list|create|delete|addkey|delkey|keys>",
		NoUsers:              "  (暂无用户)",
		UserCreateUsage:      "用法: user create <name>",
		CannotCreateGuest:    "不能创建名为 guest 的用户",
		UserCreated:          "用户 %s 创建成功",
		UserDeleteUsage:      "用法: user delete <name>",
		CannotDeleteGuest:    "不能删除 guest 用户",
		UserDeleted:          "用户 %s 已删除",
		UserAddKeyUsage:      "用法: user addkey <name> <pubkey>",
		InvalidPublicKey:     "无效的公钥: ",
		KeyAdded:             "密钥添加成功",
		UserDelKeyUsage:      "用法: user delkey <name> <fingerprint>",
		KeyRemoved:           "密钥已删除",
		UserKeysUsage:        "用法: user keys <name>",
		NoKeys:               "  (暂无密钥)",
		UnknownUserCommand:   "未知的 user 子命令",

		// Misc
		KeysCount:            "个密钥",
	},
}
