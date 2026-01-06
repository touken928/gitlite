# GitLite

GitLite is a lightweight private Git server written in pure Go.
Administrators connect via SSH directly into a built-in management CLI,
while developers access repositories strictly through Git commands only.

GitLite 是一个纯 Go 实现的轻量级私有 Git 服务器，
管理员通过 SSH 直接进入内置管理 CLI，
开发者仅能使用 Git 访问仓库，无法接触系统。


---

## Features / 特性

- **Self-contained SSH Server** - No dependency on system sshd or OS users
- **Separation of Concerns** - Admin manages system via CLI, cannot access code; Users access code via Git, cannot access system
- **Repository-level Access Control** - Fine-grained read/write permissions per user
- **Multi-key Support** - Each user can have multiple SSH keys
- **Guest Access** - Built-in guest user allows anonymous read-only access
- **Security First** - Whitelist-only Git commands, path validation, no shell access

---

- **自包含 SSH 服务器** - 不依赖系统 sshd 或操作系统用户
- **职责分离** - 管理员通过命令行管理系统，不能访问代码；用户通过 Git 访问代码，不能访问系统
- **仓库级访问控制** - 每个用户可设置细粒度的读写权限
- **多密钥支持** - 每个用户可以拥有多个 SSH 密钥
- **访客访问** - 内置 guest 用户，允许匿名只读访问
- **安全优先** - 仅允许白名单 Git 命令，路径校验，禁止 shell 访问

---

## Quick Start / 快速开始

### Build / 编译

```bash
go build -o gitlite .
```

### Setup Admin Key / 设置管理员密钥

```bash
mkdir -p data
ssh-keygen -t ed25519 -f admin_key -N ""
cp admin_key.pub data/admin.pub
```

### Start Server / 启动服务

```bash
# Default configuration / 默认配置
./gitlite

# With environment variables / 使用环境变量
GITLITE_PORT=2222 GITLITE_DATA=./data ./gitlite
```

**Environment Variables / 环境变量:**

| Variable | Default | Description |
|----------|---------|-------------|
| `GITLITE_PORT` | `2222` | SSH listening port / SSH 监听端口 |
| `GITLITE_DATA` | `data` | Data directory / 数据目录 |

---

## Usage / 使用方法

### Admin Login / 管理员登录

```bash
ssh -t -i admin_key -p 2222 localhost
```

### Admin Commands / 管理命令

```
Commands:
  repo list                         - List all repositories
  repo create <name>                - Create a repository
  repo delete <name>                - Delete a repository
  repo adduser <repo> <user> <r|rw> - Add user to repository
  repo deluser <repo> <user>        - Remove user from repository

  user list                         - List all users
  user create <name>                - Create a user
  user delete <name>                - Delete a user
  user addkey <name> <pubkey>       - Add SSH key to user
  user delkey <name> <fingerprint>  - Remove SSH key from user
  user keys <name>                  - List user's SSH keys

  lang <zh|en>                      - Switch language
  help                              - Show help
  quit                              - Exit
```

### User Git Operations / 用户 Git 操作

Users can use their existing SSH keys (e.g., `~/.ssh/id_ed25519`). Admin just needs to add their public key.

用户可以使用他们现有的 SSH 密钥（如 `~/.ssh/id_ed25519`）。管理员只需添加他们的公钥即可。

```bash
# User sends their public key to admin
# 用户将公钥发送给管理员
cat ~/.ssh/id_ed25519.pub

# Admin adds the key in CLI
# 管理员在命令行中添加密钥
admin> user create alice
admin> user addkey alice ssh-ed25519 AAAAC3... alice@laptop
admin> repo adduser myrepo alice rw
```

```bash
# User can now access the repository
# 用户现在可以访问仓库
git clone ssh://localhost:2222/myrepo.git

# Or configure in ~/.ssh/config
# 或在 ~/.ssh/config 中配置
Host mygit
    HostName localhost
    Port 2222

git clone mygit:myrepo.git
```

---

## Access Control / 访问控制

### Permission Levels / 权限级别

| Permission | Clone/Pull | Push |
|------------|------------|------|
| `r` (read) | ✓ | ✗ |
| `rw` (read-write) | ✓ | ✓ |

### Guest User / 访客用户

`guest` is a built-in virtual user. When added to a repository, it allows **anyone** (even unauthenticated users) to read that repository.

`guest` 是内置的虚拟用户。当添加到仓库后，**任何人**（包括未认证用户）都可以读取该仓库。

```
# Enable anonymous read access for a repository
# 为仓库启用匿名只读访问
admin> repo adduser myrepo guest r
```

- Cannot be created or deleted / 不能创建或删除
- Can only have read permission / 只能设置只读权限
- Enables anonymous read access / 启用匿名只读访问

---

## Architecture / 架构

```
+----------------------------------------------------------+
|                      SSH Server                          |
+----------------------------------------------------------+
|                                                          |
|  +-------------+              +---------------------+    |
|  |   Admin     |              |      Users          |    |
|  |  (no cmd)   |              |   (git command)     |    |
|  +------+------+              +----------+----------+    |
|         |                                |               |
|         v                                v               |
|  +-------------+              +---------------------+    |
|  |  Admin CLI  |              |   Git Handler       |    |
|  |  - repo     |              |   - git-upload-pack |    |
|  |  - user     |              |   - git-receive-pack|    |
|  +-------------+              +---------------------+    |
|                                                          |
+----------------------------------------------------------+
```

### Request Routing / 请求路由

| Connection Type | Identity | Action |
|-----------------|----------|--------|
| SSH (no command) | Admin key | Enter CLI |
| SSH (no command) | Other | Denied |
| SSH (git command) | Admin key | Denied |
| SSH (git command) | User key | Check permission |
| SSH (git command) | Unknown key | Check guest permission |

---

## Data Storage / 数据存储

```
data/
├── admin.pub      # Admin public key / 管理员公钥
├── host_key       # Server host key (auto-generated) / 服务器主机密钥（自动生成）
├── users.json     # User data (auto-generated) / 用户数据（自动生成）
├── repos.json     # Repository permissions (auto-generated) / 仓库权限（自动生成）
└── repos/         # Git repositories / Git 仓库
    ├── repo1.git/
    └── repo2.git/
```

**Note**: User and permission data are now persisted to JSON files and will survive restarts.

**注意**: 用户和权限数据现在持久化到 JSON 文件中，重启后不会丢失。

---

## Security / 安全性

- **No shell access** - Only Git commands allowed / 禁止 shell 访问，仅允许 Git 命令
- **Command whitelist** - Only `git-upload-pack` and `git-receive-pack` / 命令白名单
- **Path validation** - Prevents path traversal attacks / 路径校验，防止路径穿越攻击
- **No port forwarding** - SSH tunneling disabled / 禁止端口转发
- **Key-based auth only** - No password authentication / 仅密钥认证，无密码认证

---

## License

MIT
