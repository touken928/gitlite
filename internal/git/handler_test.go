package git

import (
	"testing"
)

func TestParseCommandGitUploadPack(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/myproject.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.Cmd != "git-upload-pack" {
		t.Errorf("Expected Cmd 'git-upload-pack', got '%s'", cmd.Cmd)
	}

	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}

	if cmd.IsWrite {
		t.Error("IsWrite should be false for git-upload-pack")
	}
}

func TestParseCommandGitReceivePack(t *testing.T) {
	cmd, err := ParseCommand("git-receive-pack '/myproject.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.Cmd != "git-receive-pack" {
		t.Errorf("Expected Cmd 'git-receive-pack', got '%s'", cmd.Cmd)
	}

	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}

	if !cmd.IsWrite {
		t.Error("IsWrite should be true for git-receive-pack")
	}
}

func TestParseCommandWithPath(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/group/project.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "group/project.git" {
		t.Errorf("Expected RepoPath 'group/project.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandInvalidCommand(t *testing.T) {
	_, err := ParseCommand("ls -la")
	if err == nil {
		t.Error("ParseCommand should return error for invalid command")
	}

	if err.Error() != "不允许的命令: ls" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestParseCommandInvalidFormat(t *testing.T) {
	_, err := ParseCommand("git-upload-pack")
	if err == nil {
		t.Error("ParseCommand should return error for invalid format")
	}

	if err.Error() != "无效的命令格式" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestParseCommandInvalidRepoPath(t *testing.T) {
	_, err := ParseCommand("git-upload-pack '/../../../etc/passwd'")
	if err == nil {
		t.Error("ParseCommand should return error for path traversal attempt")
	}

	if err.Error() != "无效的仓库路径: ../../../etc/passwd" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestParseCommandEmptyRepo(t *testing.T) {
	_, err := ParseCommand("git-upload-pack '/.git'")
	if err == nil {
		t.Error("ParseCommand should return error for invalid repo name")
	}
}

func TestParseCommandWithSingleQuotes(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/myproject.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	// Should handle single quotes
	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandWithDoubleQuotes(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack \"myproject.git\"")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	// Should handle double quotes (stripped)
	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandNoLeadingSlash(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack 'myproject.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandWithUnderscores(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/my_project_name.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "my_project_name.git" {
		t.Errorf("Expected RepoPath 'my_project_name.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandWithHyphens(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/my-repo-name.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "my-repo-name.git" {
		t.Errorf("Expected RepoPath 'my-repo-name.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandWithNumbers(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/project123.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "project123.git" {
		t.Errorf("Expected RepoPath 'project123.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandNestedPath(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/org/team/project.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "org/team/project.git" {
		t.Errorf("Expected RepoPath 'org/team/project.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandDeepNestedPath(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/a/b/c/d/e/project.git'")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "a/b/c/d/e/project.git" {
		t.Errorf("Expected RepoPath 'a/b/c/d/e/project.git', got '%s'", cmd.RepoPath)
	}
}

func TestParseCommandWithTrailingWhitespace(t *testing.T) {
	cmd, err := ParseCommand("git-upload-pack '/myproject.git'   ")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}

	if cmd.RepoPath != "myproject.git" {
		t.Errorf("Expected RepoPath 'myproject.git', got '%s'", cmd.RepoPath)
	}
}

func TestAllowedCommands(t *testing.T) {
	// Verify allowed commands
	if !allowedCommands["git-upload-pack"] {
		t.Error("git-upload-pack should be allowed")
	}
	if !allowedCommands["git-receive-pack"] {
		t.Error("git-receive-pack should be allowed")
	}

	// Verify other commands are not allowed
	if allowedCommands["git-shell"] {
		t.Error("git-shell should not be allowed")
	}
	if allowedCommands["git"] {
		t.Error("git should not be allowed")
	}
}

func TestRepoPathRegexValid(t *testing.T) {
	validPaths := []string{
		"myproject.git",
		"my_project.git",
		"my-repo.git",
		"project123.git",
		"a.git",
		"org/team/project.git",
		"a/b/c/d/e/project.git",
	}

	for _, path := range validPaths {
		if !repoPathRegex.MatchString(path) {
			t.Errorf("Path '%s' should be valid", path)
		}
	}
}

func TestRepoPathRegexInvalid(t *testing.T) {
	invalidPaths := []string{
		"../../../etc/passwd",
		"/etc/passwd.git",
		"my project.git",
		"myproject.",
		".git",
		"my project/with spaces.git",
		"../parent.git",
		"/abs/path.git",
		"~user.git",
		"$VAR.git",
	}

	for _, path := range invalidPaths {
		if repoPathRegex.MatchString(path) {
			t.Errorf("Path '%s' should be invalid", path)
		}
	}
}
