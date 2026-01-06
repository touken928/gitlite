package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/touken928/gitlite/internal/auth"
	"github.com/touken928/gitlite/internal/git"
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

	// 确保数据目录存在
	if err := os.MkdirAll(filepath.Join(dataPath, "repos"), 0755); err != nil {
		return nil, err
	}

	// 加载或生成主机密钥
	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return nil, err
	}

	// 加载管理员公钥
	if err := s.loadAdminKey(); err != nil {
		log.Printf("警告: 未找到管理员公钥，请创建 %s/admin.pub", dataPath)
	}

	// 加载持久化的用户数据
	if err := s.authMgr.LoadFromFile(filepath.Join(dataPath, "users.json")); err != nil {
		log.Printf("警告: 加载用户数据失败: %v", err)
	}

	// 加载持久化的仓库权限数据
	if err := s.repoMgr.LoadFromFile(filepath.Join(dataPath, "repos.json")); err != nil {
		log.Printf("警告: 加载仓库权限数据失败: %v", err)
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
	// 保存用户数据
	if err := s.authMgr.SaveToFile(filepath.Join(s.dataPath, "users.json")); err != nil {
		log.Printf("保存用户数据失败: %v", err)
	}

	// 保存仓库权限数据
	if err := s.repoMgr.SaveToFile(filepath.Join(s.dataPath, "repos.json")); err != nil {
		log.Printf("保存仓库权限数据失败: %v", err)
	}

	s.sshSrv.Close()
}

func (s *Server) loadOrGenerateHostKey() (ssh.Signer, error) {
	keyPath := filepath.Join(s.dataPath, "host_key")

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.dataPath, 0755); err != nil {
			return nil, err
		}
		log.Printf("生成新的主机密钥: %s", keyPath)
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
	log.Printf("已加载管理员公钥")
	return nil
}

func (s *Server) handlePublicKey(ctx ssh.Context, key ssh.PublicKey) bool {
	user, userType := s.authMgr.Authenticate(key)

	ctx.SetValue("user", user)
	ctx.SetValue("userType", userType)

	// 管理员和已知用户允许连接，未知用户也允许（可能是 guest 访问）
	return true
}

func (s *Server) handleSession(sess ssh.Session) {
	user, _ := sess.Context().Value("user").(*auth.User)
	userType, _ := sess.Context().Value("userType").(auth.UserType)

	rawCmd := sess.RawCommand()

	// 无命令 = 管理员 TUI 登录
	if rawCmd == "" {
		if userType != auth.UserTypeAdmin {
			io.WriteString(sess, "拒绝访问: 仅管理员可登录\r\n")
			sess.Exit(1)
			return
		}
		s.tui.Run(sess)
		return
	}

	// 有命令 = Git 操作
	if userType == auth.UserTypeAdmin {
		io.WriteString(sess, "拒绝访问: 管理员不能执行 Git 操作\r\n")
		sess.Exit(1)
		return
	}

	gitCmd, err := git.ParseCommand(rawCmd)
	if err != nil {
		io.WriteString(sess, fmt.Sprintf("错误: %v\r\n", err))
		sess.Exit(1)
		return
	}

	userName := ""
	if user != nil {
		userName = user.Name
	}

	if !s.repoMgr.CheckPermission(gitCmd.RepoPath, userName, gitCmd.IsWrite) {
		io.WriteString(sess, "拒绝访问: 权限不足\r\n")
		sess.Exit(1)
		return
	}

	repoFullPath := s.repoMgr.GetRepoPath(gitCmd.RepoPath)
	if _, err := os.Stat(repoFullPath); os.IsNotExist(err) {
		io.WriteString(sess, "错误: 仓库不存在\r\n")
		sess.Exit(1)
		return
	}

	if err := git.Execute(sess, gitCmd, repoFullPath); err != nil {
		log.Printf("Git 执行错误: %v", err)
		sess.Exit(1)
		return
	}
}
