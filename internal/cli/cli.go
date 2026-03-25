package cli

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	Mode        string
	NoAI        bool
	MinSeverity string
	Format      string
	Debug       bool
	Subcommand  string
	HookAction  string
}

func Parse() Options {
	var opts Options

	if len(os.Args) > 1 && os.Args[1] == "hook" {
		opts.Subcommand = "hook"
		if len(os.Args) > 2 {
			opts.HookAction = os.Args[2]
		}
		return opts
	}

	fs := flag.NewFlagSet("git-aps", flag.ExitOnError)
	fs.StringVar(&opts.Mode, "mode", "", "diff mode: staged, unstaged, head, branch")
	fs.BoolVar(&opts.NoAI, "no-ai", false, "disable AI analysis")
	fs.StringVar(&opts.MinSeverity, "min-severity", "", "minimum severity: low, med, high")
	fs.StringVar(&opts.Format, "format", "tui", "output format: tui, json, text")
	fs.BoolVar(&opts.Debug, "debug", false, "enable debug output")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: git-aps [flags]\n       git-aps hook install|uninstall\n\nFlags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])
	return opts
}
