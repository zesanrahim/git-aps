package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/config"
	"github.com/zesanrahim/git-aps/internal/git"
)

func TestNewEngine_DefaultsWhenEmptyConfig(t *testing.T) {
	t.Parallel()
	engine := NewEngine(nil)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(engine.rules) == 0 {
		t.Error("expected default rules to be registered")
	}
}

func TestNewEngine_AllDefaultRulesRegistered(t *testing.T) {
	t.Parallel()
	engine := NewEngine(nil)
	names := make(map[string]bool)
	for _, r := range engine.rules {
		names[r.Name()] = true
	}
	required := []string{"magic_numbers", "deep_nesting", "long_functions", "many_params", "todo_comments", "error_ignored"}
	for _, n := range required {
		if !names[n] {
			t.Errorf("expected rule %q to be registered", n)
		}
	}
}

func TestNewEngine_DisableRule(t *testing.T) {
	t.Parallel()
	cfg := map[string]config.RuleConfig{
		"magic_numbers": {Enabled: false},
	}
	engine := NewEngine(cfg)
	for _, r := range engine.rules {
		if r.Name() == "magic_numbers" {
			t.Error("expected magic_numbers to be disabled")
		}
	}
}

func TestNewEngine_CustomThresholds(t *testing.T) {
	t.Parallel()
	cfg := map[string]config.RuleConfig{
		"magic_numbers":  {Enabled: true, Threshold: 100},
		"deep_nesting":   {Enabled: true, MaxDepth: 10},
		"long_functions": {Enabled: true, MaxLines: 200},
		"many_params":    {Enabled: true, MaxParams: 10},
	}
	engine := NewEngine(cfg)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	diff := []git.FileDiff{
		makeFileDiff("foo.go",
			makeAddedLine("x := 50", 1),
		),
	}
	findings, err := engine.Analyze(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range findings {
		if f.Rule == "magic_numbers" {
			t.Error("expected no magic_numbers finding with threshold 100 for value 50")
		}
	}
}

func TestNewEngine_CustomRules(t *testing.T) {
	t.Parallel()
	customCfg := []config.CustomRuleConfig{
		{
			Name:        "no_panic",
			Pattern:     `panic\(`,
			Severity:    "high",
			Description: "avoid panic",
			Suggestion:  "use errors",
		},
	}
	engine := NewEngine(nil, customCfg)
	names := make(map[string]bool)
	for _, r := range engine.rules {
		names[r.Name()] = true
	}
	if !names["no_panic"] {
		t.Error("expected custom rule no_panic to be registered")
	}
}

func TestNewEngine_SkipsInvalidCustomRulePattern(t *testing.T) {
	t.Parallel()
	customCfg := []config.CustomRuleConfig{
		{
			Name:    "bad_pattern",
			Pattern: `[invalid`,
		},
	}
	engine := NewEngine(nil, customCfg)
	for _, r := range engine.rules {
		if r.Name() == "bad_pattern" {
			t.Error("expected invalid pattern rule to be skipped")
		}
	}
}

func TestNewEngine_SkipsCustomRuleWithEmptyName(t *testing.T) {
	t.Parallel()
	customCfg := []config.CustomRuleConfig{
		{
			Name:    "",
			Pattern: `panic\(`,
		},
	}
	count := len(NewEngine(nil).rules)
	engine := NewEngine(nil, customCfg)
	if len(engine.rules) != count {
		t.Error("expected rule with empty name to be skipped")
	}
}

func TestNewEngine_SkipsCustomRuleWithEmptyPattern(t *testing.T) {
	t.Parallel()
	customCfg := []config.CustomRuleConfig{
		{
			Name:    "empty_pattern",
			Pattern: "",
		},
	}
	count := len(NewEngine(nil).rules)
	engine := NewEngine(nil, customCfg)
	if len(engine.rules) != count {
		t.Error("expected rule with empty pattern to be skipped")
	}
}

func TestEngine_Analyze_ReturnsFindings(t *testing.T) {
	t.Parallel()
	cfg := map[string]config.RuleConfig{
		"todo_comments":  {Enabled: true},
		"magic_numbers":  {Enabled: false},
		"deep_nesting":   {Enabled: false},
		"long_functions": {Enabled: false},
		"many_params":    {Enabled: false},
		"error_ignored":  {Enabled: false},
	}
	engine := NewEngine(cfg)
	diffs := []git.FileDiff{
		makeFileDiff("foo.go",
			makeAddedLine("// TODO: fix this", 5),
		),
	}
	findings, err := engine.Analyze(diffs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("expected at least one finding for TODO comment")
	}
	if findings[0].Rule != "todo_comments" {
		t.Errorf("expected rule todo_comments, got %q", findings[0].Rule)
	}
}

func TestEngine_Analyze_EmptyDiff(t *testing.T) {
	t.Parallel()
	engine := NewEngine(nil)
	findings, err := engine.Analyze(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings for empty diff, got %d", len(findings))
	}
}
