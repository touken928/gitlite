# GitLite

[English](README.md)

GitLite 是一个纯 Go 实现的轻量级私有 Git 服务器，
管理员通过 SSH 直接进入内置管理 CLI，
开发者仅能使用 Git 访问仓库，无法接触系统。

---

## 特性

- **自包含 SSH 服务器** - 不依赖系统 sshd 或操作系统用户
- **职责分离** - 管理员通过命令行管理系统，不能访问代码；用户通过 Git 访问代码，不能访问系统
- **仓库级访问控制** - 每个用户可设置细粒度的读写权限
- **多密钥支持** - 每个用户可以拥有多个 SSH 密钥
- **访客访问** - 内置 guest 用户，允许匿名只读访问
- **安全优先** - 仅允许白名单 Git 命令，路径校验，禁止 shell 访问

---

## 快速开始

### 编译

```bash
go build -o gitlite .
```

### 设置管理员密钥

```bash
mkdir -p data
ssh-keygen -t ed25519 -f admin_key -N ""
cp admin_key.pub data/admin.pub
```

### 启动服务

```bash
# 默认配置
./gitlite

# 使用环境变量
GITLITE_PORT=2222 GITLITE_DATA=./data ./gitlite
```

**环境变量：**

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `GITLITE_PORT` | `2222` | SSH 监听端口 |
| `GITLITE_DATA` | `data` | 数据目录 |

---

## 使用方法

### 管理员登录

```bash
ssh -t -i admin_key -p 2222 localhost
```

### 管理命令

```
命令：
  repo list                         - 列出所有仓库
  repo create <name>                - 创建仓库
  repo delete <name>                - 删除仓库
  repo adduser <repo> <user> <r|rw> - 将用户添加到仓库
  repo deluser <repo> <user>        - 从仓库移除用户

  user list                         - 列出所有用户
  user create <name>                - 创建用户
  user delete <name>                - 删除用户
  user addkey <name> <pubkey>       - 为用户添加 SSH 密钥
  user delkey <name> <fingerprint>  - 从用户移除 SSH 密钥
  user keys <name>                  - 列出用户的 SSH 密钥

  lang <zh|en>                      - 切换语言
  help                              - 显示帮助
  quit                              - 退出
```

### 用户 Git 操作

用户可以使用他们现有的 SSH 密钥（如 `~/.ssh/id_ed25519`）。管理员只需添加他们的公钥即可。

```bash
# 用户将公钥发送给管理员
cat ~/.ssh/id_ed25519.pub

# 管理员在命令行中添加密钥
admin> user create alice
admin> user addkey alice ssh-ed25519 AAAAC3... alice@laptop
admin> repo adduser myrepo alice rw
```

```bash
# 用户现在可以访问仓库
git clone ssh://localhost:2222/myrepo.git

# 或在 ~/.ssh/config 中配置
Host mygit
    HostName localhost
    Port 2222

git clone mygit:myrepo.git
```

---

## 访问控制

### 权限级别

| 权限 | 克隆/拉取 | 推送 |
|------|----------|------|
| `r` (只读) | ✓ | ✗ |
| `rw` (读写) | ✓ | ✓ |

### 访客用户

`guest` 是内置的虚拟用户。当添加到仓库后，**任何人**（包括未认证用户）都可以读取该仓库。

```
# 为仓库启用匿名只读访问
admin> repo adduser myrepo guest r
```

- 不能创建或删除
- 只能设置只读权限
- 启用匿名只读访问

---

## 架构

```
+----------------------------------------------------------+
|                      SSH 服务器                          |
+----------------------------------------------------------|
|                                                          |
|  +-------------+              +---------------------+    |
|  |   管理员    |              |      用户           |    |
|  |  (无命令)   |              |   (git 命令)        |    |
|  +------+------+              +----------+----------+    |
|         |                                |               |
|         v                                v               |
|  +-------------+              +---------------------+    |
|  |  管理命令行 |              |   Git 处理器        |    |
|  |  - repo     |              |   - git-upload-pack |    |
|  |  - user     |              |   - git-receive-pack|    |
|  +-------------+              +---------------------+    |
|                                                          |
+----------------------------------------------------------+
```

### 请求路由

| 连接类型 | 身份 | 操作 |
|---------|------|------|
| SSH (无命令) | 管理员密钥 | 进入 CLI |
| SSH (无命令) | 其他 | 拒绝 |
| SSH (git 命令) | 管理员密钥 | 拒绝 |
| SSH (git 命令) | 用户密钥 | 检查权限 |
| SSH (git 命令) | 未知密钥 | 检查访客权限 |

---

## 数据存储

```
data/
├── admin.pub      # 管理员公钥
├── host_key       # 服务器主机密钥（自动生成）
├── users.json     # 用户数据（自动生成）
├── repos.json     # 仓库权限（自动生成）
└── repos/         # Git 仓库
    ├── repo1.git/
    └── repo2.git/
```

**注意**: 用户和权限数据现在持久化到 JSON 文件中，重启后不会丢失。

---

## 安全性

- **禁止 shell 访问** - 仅允许 Git 命令
- **命令白名单** - 仅允许 `git-upload-pack` 和 `git-receive-pack`
- **路径校验** - 防止路径穿越攻击
- **禁止端口转发** - 禁用 SSH 隧道
- **仅密钥认证** - 无密码认证

---

## 许可证

MIT
