package tui

import (
	"testing"

	"gitlite/internal/auth"
	"gitlite/internal/repo"
)

func TestNewTUI(t *testing.T) {
	authMgr := auth.NewManager()
	repoMgr := repo.NewManager("/tmp")
	tui := New(authMgr, repoMgr, "/tmp/data")

	if tui == nil {
		t.Fatal("NewTUI returned nil")
	}
	if tui.authMgr != authMgr {
		t.Error("authMgr not set correctly")
	}
	if tui.repoMgr != repoMgr {
		t.Error("repoMgr not set correctly")
	}
	if tui.dataPath != "/tmp/data" {
		t.Error("dataPath not set correctly")
	}
	if tui.lang != "en" {
		t.Error("default lang should be en")
	}
}

func TestMsgChinese(t *testing.T) {
	authMgr := auth.NewManager()
	repoMgr := repo.NewManager("/tmp")
	tui := New(authMgr, repoMgr, "/tmp")
	tui.lang = "zh"

	result := tui.msg("中文消息", "English message")
	if result != "中文消息" {
		t.Errorf("Expected Chinese message, got '%s'", result)
	}
}

func TestMsgEnglish(t *testing.T) {
	authMgr := auth.NewManager()
	repoMgr := repo.NewManager("/tmp")
	tui := New(authMgr, repoMgr, "/tmp")
	tui.lang = "en"

	result := tui.msg("中文消息", "English message")
	if result != "English message" {
		t.Errorf("Expected English message, got '%s'", result)
	}
}
