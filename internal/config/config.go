package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Diff        DiffConfig            `yaml:"diff"`
	Rules       map[string]RuleConfig `yaml:"rules"`
	CustomRules []CustomRuleConfig    `yaml:"custom_rules"`
	AI          AIConfig              `yaml:"ai"`
	Output      OutputConfig          `yaml:"output"`
}

type DiffConfig struct {
	Mode string `yaml:"mode"`
}

type RuleConfig struct {
	Enabled   bool `yaml:"enabled"`
	Threshold int  `yaml:"threshold,omitempty"`
	MaxDepth  int  `yaml:"max_depth,omitempty"`
	MaxLines  int  `yaml:"max_lines,omitempty"`
	MaxParams int  `yaml:"max_params,omitempty"`
}

type AIConfig struct {
	Enabled bool   `yaml:"enabled"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
}

type OutputConfig struct {
	MinSeverity string `yaml:"min_severity"`
}

type CustomRuleConfig struct {
	Name        string `yaml:"name"`
	Pattern     string `yaml:"pattern"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
	Suggestion  string `yaml:"suggestion"`
}

func Load() (Config, error) {
	path := os.Getenv("GIT_APS_CONFIG")
	if path == "" {
		path = ".git-aps.yaml"
	}

	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func defaults() Config {
	return Config{
		Diff: DiffConfig{Mode: "staged"},
		Rules: map[string]RuleConfig{
			"magic_numbers": {Enabled: true, Threshold: 2},
			"deep_nesting":  {Enabled: true, MaxDepth: 3},
			"long_functions": {Enabled: true, MaxLines: 50},
			"many_params":   {Enabled: true, MaxParams: 5},
			"todo_comments": {Enabled: true},
			"error_ignored": {Enabled: true},
		},
		AI: AIConfig{
			Enabled: true,
			Model:   "gemini-2.5-flash",
			BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
		},
	}
}
