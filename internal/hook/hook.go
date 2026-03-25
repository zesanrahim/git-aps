package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const marker = "# git-aps hook"

var overriddenHookPath string

const hookScript = `#!/bin/sh
# git-aps hook — auto-installed, do not edit
git-aps --format text --mode staged --min-severity high
`

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("finding git root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func hookPath() (string, error) {
	if overriddenHookPath != "" {
		return overriddenHookPath, nil
	}
	root, err := gitRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".git", "hooks", "pre-commit"), nil
}

func Install() error {
	path, err := hookPath()
	if err != nil {
		return err
	}

	if data, err := os.ReadFile(path); err == nil {
		content := string(data)
		if strings.Contains(content, marker) {
			fmt.Println("git-aps hook is already installed.")
			return nil
		}
		return fmt.Errorf("a pre-commit hook already exists at %s — remove it first or add git-aps manually", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("writing hook: %w", err)
	}

	fmt.Println("Installed pre-commit hook.")
	fmt.Println("Make sure the git-aps binary is on your PATH.")
	fmt.Println("Build with: go build -o $(go env GOPATH)/bin/git-aps ./cmd/git-aps")
	return nil
}

func Uninstall() error {
	path, err := hookPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No pre-commit hook found.")
			return nil
		}
		return fmt.Errorf("reading hook: %w", err)
	}

	if !strings.Contains(string(data), marker) {
		return fmt.Errorf("pre-commit hook at %s was not installed by git-aps — refusing to remove", path)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing hook: %w", err)
	}

	fmt.Println("Uninstalled pre-commit hook.")
	return nil
}
