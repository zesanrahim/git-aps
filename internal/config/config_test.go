package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".git-aps.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	return path
}

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	t.Setenv("GIT_APS_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yaml"))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if cfg.Diff.Mode != "staged" {
		t.Errorf("expected default mode staged, got %q", cfg.Diff.Mode)
	}
	if cfg.AI.Model != "gemini-2.5-flash" {
		t.Errorf("expected default AI model, got %q", cfg.AI.Model)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	yaml := `
diff:
  mode: unstaged
ai:
  enabled: false
  model: gpt-4
output:
  min_severity: high
`
	path := writeConfigFile(t, yaml)
	t.Setenv("GIT_APS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Diff.Mode != "unstaged" {
		t.Errorf("expected mode unstaged, got %q", cfg.Diff.Mode)
	}
	if cfg.AI.Enabled {
		t.Error("expected AI disabled")
	}
	if cfg.AI.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %q", cfg.AI.Model)
	}
	if cfg.Output.MinSeverity != "high" {
		t.Errorf("expected min_severity high, got %q", cfg.Output.MinSeverity)
	}
}

func TestLoad_PartialConfig_PreservesDefaults(t *testing.T) {
	yaml := `
diff:
  mode: head
`
	path := writeConfigFile(t, yaml)
	t.Setenv("GIT_APS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Diff.Mode != "head" {
		t.Errorf("expected mode head, got %q", cfg.Diff.Mode)
	}
	if cfg.AI.Model != "gemini-2.5-flash" {
		t.Errorf("expected default AI model preserved, got %q", cfg.AI.Model)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeConfigFile(t, "diff: [invalid yaml }{")
	t.Setenv("GIT_APS_CONFIG", path)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_RuleConfigs(t *testing.T) {
	yaml := `
rules:
  magic_numbers:
    enabled: true
    threshold: 10
  deep_nesting:
    enabled: false
    max_depth: 5
  long_functions:
    enabled: true
    max_lines: 100
  many_params:
    enabled: true
    max_params: 8
`
	path := writeConfigFile(t, yaml)
	t.Setenv("GIT_APS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mn := cfg.Rules["magic_numbers"]
	if mn.Threshold != 10 {
		t.Errorf("expected threshold 10, got %d", mn.Threshold)
	}
	if !mn.Enabled {
		t.Error("expected magic_numbers enabled")
	}

	dn := cfg.Rules["deep_nesting"]
	if dn.Enabled {
		t.Error("expected deep_nesting disabled")
	}
	if dn.MaxDepth != 5 {
		t.Errorf("expected max_depth 5, got %d", dn.MaxDepth)
	}

	lf := cfg.Rules["long_functions"]
	if lf.MaxLines != 100 {
		t.Errorf("expected max_lines 100, got %d", lf.MaxLines)
	}

	mp := cfg.Rules["many_params"]
	if mp.MaxParams != 8 {
		t.Errorf("expected max_params 8, got %d", mp.MaxParams)
	}
}

func TestLoad_CustomRules(t *testing.T) {
	yaml := `
custom_rules:
  - name: no_panic
    pattern: 'panic\('
    severity: high
    description: Avoid panics
    suggestion: Return an error
  - name: no_print
    pattern: 'fmt\.Print'
    severity: low
    description: Remove debug prints
    suggestion: Use a logger
`
	path := writeConfigFile(t, yaml)
	t.Setenv("GIT_APS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.CustomRules) != 2 {
		t.Fatalf("expected 2 custom rules, got %d", len(cfg.CustomRules))
	}
	if cfg.CustomRules[0].Name != "no_panic" {
		t.Errorf("expected first rule no_panic, got %q", cfg.CustomRules[0].Name)
	}
	if cfg.CustomRules[0].Severity != "high" {
		t.Errorf("expected severity high, got %q", cfg.CustomRules[0].Severity)
	}
	if cfg.CustomRules[1].Name != "no_print" {
		t.Errorf("expected second rule no_print, got %q", cfg.CustomRules[1].Name)
	}
}

func TestLoad_EnvVarOverridesPath(t *testing.T) {
	yaml := `
diff:
  mode: branch
`
	path := writeConfigFile(t, yaml)
	t.Setenv("GIT_APS_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Diff.Mode != "branch" {
		t.Errorf("expected mode branch from env override, got %q", cfg.Diff.Mode)
	}
}

func TestDefaults(t *testing.T) {
	t.Parallel()
	cfg := defaults()
	if cfg.Diff.Mode != "staged" {
		t.Errorf("expected staged, got %q", cfg.Diff.Mode)
	}
	requiredRules := []string{"magic_numbers", "deep_nesting", "long_functions", "many_params", "todo_comments", "error_ignored"}
	for _, name := range requiredRules {
		if _, ok := cfg.Rules[name]; !ok {
			t.Errorf("expected default rule %q", name)
		}
	}
	if !cfg.Rules["magic_numbers"].Enabled {
		t.Error("expected magic_numbers enabled by default")
	}
	if cfg.Rules["magic_numbers"].Threshold != 2 {
		t.Errorf("expected magic_numbers threshold 2, got %d", cfg.Rules["magic_numbers"].Threshold)
	}
	if cfg.Rules["deep_nesting"].MaxDepth != 3 {
		t.Errorf("expected deep_nesting max_depth 3, got %d", cfg.Rules["deep_nesting"].MaxDepth)
	}
	if cfg.Rules["long_functions"].MaxLines != 50 {
		t.Errorf("expected long_functions max_lines 50, got %d", cfg.Rules["long_functions"].MaxLines)
	}
	if cfg.Rules["many_params"].MaxParams != 5 {
		t.Errorf("expected many_params max_params 5, got %d", cfg.Rules["many_params"].MaxParams)
	}
	if cfg.AI.Model != "gemini-2.5-flash" {
		t.Errorf("expected default AI model, got %q", cfg.AI.Model)
	}
}
