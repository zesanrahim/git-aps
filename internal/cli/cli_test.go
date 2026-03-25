package cli

import (
	"os"
	"testing"
)

func withArgs(t *testing.T, args []string) {
	t.Helper()
	orig := os.Args
	os.Args = append([]string{"git-aps"}, args...)
	t.Cleanup(func() { os.Args = orig })
}

func TestParse_Defaults(t *testing.T) {
	withArgs(t, []string{})
	opts := Parse()
	if opts.Mode != "" {
		t.Errorf("expected empty mode, got %q", opts.Mode)
	}
	if opts.NoAI {
		t.Error("expected NoAI=false")
	}
	if opts.Format != "tui" {
		t.Errorf("expected format tui, got %q", opts.Format)
	}
	if opts.Debug {
		t.Error("expected Debug=false")
	}
	if opts.MinSeverity != "" {
		t.Errorf("expected empty min-severity, got %q", opts.MinSeverity)
	}
}

func TestParse_ModeFlag(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--mode", "staged"}, "staged"},
		{[]string{"--mode", "unstaged"}, "unstaged"},
		{[]string{"--mode", "head"}, "head"},
		{[]string{"--mode", "branch"}, "branch"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			withArgs(t, tc.args)
			opts := Parse()
			if opts.Mode != tc.want {
				t.Errorf("want mode %q, got %q", tc.want, opts.Mode)
			}
		})
	}
}

func TestParse_NoAIFlag(t *testing.T) {
	withArgs(t, []string{"--no-ai"})
	opts := Parse()
	if !opts.NoAI {
		t.Error("expected NoAI=true")
	}
}

func TestParse_FormatFlag(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--format", "json"}, "json"},
		{[]string{"--format", "text"}, "text"},
		{[]string{"--format", "tui"}, "tui"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			withArgs(t, tc.args)
			opts := Parse()
			if opts.Format != tc.want {
				t.Errorf("want format %q, got %q", tc.want, opts.Format)
			}
		})
	}
}

func TestParse_MinSeverityFlag(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--min-severity", "high"}, "high"},
		{[]string{"--min-severity", "med"}, "med"},
		{[]string{"--min-severity", "low"}, "low"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			withArgs(t, tc.args)
			opts := Parse()
			if opts.MinSeverity != tc.want {
				t.Errorf("want min-severity %q, got %q", tc.want, opts.MinSeverity)
			}
		})
	}
}

func TestParse_DebugFlag(t *testing.T) {
	withArgs(t, []string{"--debug"})
	opts := Parse()
	if !opts.Debug {
		t.Error("expected Debug=true")
	}
}

func TestParse_HookSubcommand(t *testing.T) {
	withArgs(t, []string{"hook"})
	opts := Parse()
	if opts.Subcommand != "hook" {
		t.Errorf("expected subcommand hook, got %q", opts.Subcommand)
	}
	if opts.HookAction != "" {
		t.Errorf("expected empty hook action, got %q", opts.HookAction)
	}
}

func TestParse_HookInstall(t *testing.T) {
	withArgs(t, []string{"hook", "install"})
	opts := Parse()
	if opts.Subcommand != "hook" {
		t.Errorf("expected subcommand hook, got %q", opts.Subcommand)
	}
	if opts.HookAction != "install" {
		t.Errorf("expected hook action install, got %q", opts.HookAction)
	}
}

func TestParse_HookUninstall(t *testing.T) {
	withArgs(t, []string{"hook", "uninstall"})
	opts := Parse()
	if opts.Subcommand != "hook" {
		t.Errorf("expected subcommand hook, got %q", opts.Subcommand)
	}
	if opts.HookAction != "uninstall" {
		t.Errorf("expected hook action uninstall, got %q", opts.HookAction)
	}
}

func TestParse_MultipleFlags(t *testing.T) {
	withArgs(t, []string{"--mode", "unstaged", "--no-ai", "--format", "json", "--min-severity", "high", "--debug"})
	opts := Parse()
	if opts.Mode != "unstaged" {
		t.Errorf("expected mode unstaged, got %q", opts.Mode)
	}
	if !opts.NoAI {
		t.Error("expected NoAI=true")
	}
	if opts.Format != "json" {
		t.Errorf("expected format json, got %q", opts.Format)
	}
	if opts.MinSeverity != "high" {
		t.Errorf("expected min-severity high, got %q", opts.MinSeverity)
	}
	if !opts.Debug {
		t.Error("expected Debug=true")
	}
}
