package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeGitDir(t *testing.T) (root string, hooksDir string, hookFilePath string) {
	t.Helper()
	root = t.TempDir()
	hooksDir = filepath.Join(root, ".git", "hooks")
	hookFilePath = filepath.Join(hooksDir, "pre-commit")
	return root, hooksDir, hookFilePath
}

func overrideHookPath(t *testing.T, path string) {
	t.Helper()
	orig := overriddenHookPath
	overriddenHookPath = path
	t.Cleanup(func() { overriddenHookPath = orig })
}

func TestInstall_CreatesHookFile(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	overrideHookPath(t, hookFile)

	if err := Install(); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	data, err := os.ReadFile(hookFile)
	if err != nil {
		t.Fatalf("reading hook file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, marker) {
		t.Errorf("hook file does not contain marker %q:\n%s", marker, content)
	}
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	if err := os.WriteFile(hookFile, []byte(hookScript), 0755); err != nil {
		t.Fatalf("writing existing hook: %v", err)
	}
	overrideHookPath(t, hookFile)

	err := Install()
	if err != nil {
		t.Fatalf("Install() should succeed when already installed, got: %v", err)
	}
}

func TestInstall_ExistingHookNotOurs(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	if err := os.WriteFile(hookFile, []byte("#!/bin/sh\nsome_other_hook"), 0755); err != nil {
		t.Fatalf("writing existing hook: %v", err)
	}
	overrideHookPath(t, hookFile)

	err := Install()
	if err == nil {
		t.Error("expected error when existing hook is not ours")
	}
}

func TestUninstall_RemovesHook(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	if err := os.WriteFile(hookFile, []byte(hookScript), 0755); err != nil {
		t.Fatalf("writing hook: %v", err)
	}
	overrideHookPath(t, hookFile)

	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if _, err := os.Stat(hookFile); !os.IsNotExist(err) {
		t.Error("expected hook file to be removed")
	}
}

func TestUninstall_NoHookFound(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	overrideHookPath(t, hookFile)

	err := Uninstall()
	if err != nil {
		t.Fatalf("Uninstall() should succeed when no hook, got: %v", err)
	}
}

func TestUninstall_RefusesNonOurHook(t *testing.T) {
	t.Parallel()
	_, hooksDir, hookFile := makeGitDir(t)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	if err := os.WriteFile(hookFile, []byte("#!/bin/sh\nsome_other_hook"), 0755); err != nil {
		t.Fatalf("writing hook: %v", err)
	}
	overrideHookPath(t, hookFile)

	err := Uninstall()
	if err == nil {
		t.Error("expected error when hook was not installed by us")
	}
}
