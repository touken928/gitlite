# GitLite

[中文文档](README.zh-CN.md)

GitLite is a lightweight private Git server written in pure Go.
Administrators connect via SSH directly into a built-in management CLI,
while developers access repositories strictly through Git commands only.

---

## Features

- **Self-contained SSH Server** - No dependency on system sshd or OS users
- **Separation of Concerns** - Admin manages system via CLI, cannot access code; Users access code via Git, cannot access system
- **Repository-level Access Control** - Fine-grained read/write permissions per user
- **Multi-key Support** - Each user can have multiple SSH keys
- **Guest Access** - Built-in guest user allows anonymous read-only access
- **Security First** - Whitelist-only Git commands, path validation, no shell access

---

## Quick Start

### Build

```bash
go build -o gitlite .
```

### Setup Admin Key

```bash
mkdir -p data
ssh-keygen -t ed25519 -f admin_key -N ""
cp admin_key.pub data/admin.pub
```

### Start Server

```bash
# Default configuration
./gitlite

# With environment variables
GITLITE_PORT=2222 GITLITE_DATA=./data ./gitlite
```

**Environment Variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `GITLITE_PORT` | `2222` | SSH listening port |
| `GITLITE_DATA` | `data` | Data directory |

---

## Usage

### Admin Login

```bash
ssh -t -i admin_key -p 2222 localhost
```

### Admin Commands

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

### User Git Operations

Users can use their existing SSH keys (e.g., `~/.ssh/id_ed25519`). Admin just needs to add their public key.

```bash
# User sends their public key to admin
cat ~/.ssh/id_ed25519.pub

# Admin adds the key in CLI
admin> user create alice
admin> user addkey alice ssh-ed25519 AAAAC3... alice@laptop
admin> repo adduser myrepo alice rw
```

```bash
# User can now access the repository
git clone ssh://localhost:2222/myrepo.git

# Or configure in ~/.ssh/config
Host mygit
    HostName localhost
    Port 2222

git clone mygit:myrepo.git
```

---

## Access Control

### Permission Levels

| Permission | Clone/Pull | Push |
|------------|------------|------|
| `r` (read) | ✓ | ✗ |
| `rw` (read-write) | ✓ | ✓ |

### Guest User

`guest` is a built-in virtual user. When added to a repository, it allows **anyone** (even unauthenticated users) to read that repository.

```
# Enable anonymous read access for a repository
admin> repo adduser myrepo guest r
```

- Cannot be created or deleted
- Can only have read permission
- Enables anonymous read access

---

## Architecture

```
+----------------------------------------------------------+
|                      SSH Server                          |
+----------------------------------------------------------|
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

### Request Routing

| Connection Type | Identity | Action |
|-----------------|----------|--------|
| SSH (no command) | Admin key | Enter CLI |
| SSH (no command) | Other | Denied |
| SSH (git command) | Admin key | Denied |
| SSH (git command) | User key | Check permission |
| SSH (git command) | Unknown key | Check guest permission |

---

## Data Storage

```
data/
├── admin.pub      # Admin public key
├── host_key       # Server host key (auto-generated)
├── users.json     # User data (auto-generated)
├── repos.json     # Repository permissions (auto-generated)
└── repos/         # Git repositories
    ├── repo1.git/
    └── repo2.git/
```

**Note**: User and permission data are now persisted to JSON files and will survive restarts.

---

## Security

- **No shell access** - Only Git commands allowed
- **Command whitelist** - Only `git-upload-pack` and `git-receive-pack`
- **Path validation** - Prevents path traversal attacks
- **No port forwarding** - SSH tunneling disabled
- **Key-based auth only** - No password authentication

---

## License

MIT
