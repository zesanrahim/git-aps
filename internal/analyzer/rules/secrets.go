package rules

import (
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type SecretsRule struct{}

func (r *SecretsRule) Name() string { return "secrets" }

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|secret|api_?key|auth_?token|access_?token|private_?key)\s*[:=]\s*["'][^"']{8,}["']`),
	regexp.MustCompile(`(?:AKIA|ASIA)[A-Z0-9]{16}`),
	regexp.MustCompile(`(?i)(token|key|secret)\s*[:=]\s*["'][a-zA-Z0-9+/=_\-]{20,}["']`),
	regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA )?PRIVATE KEY-----`),
	regexp.MustCompile(`(?:ghp_|glpat-|github_pat_)[a-zA-Z0-9_]{20,}`),
}

var secretExclusions = []string{
	"os.Getenv",
	"os.LookupEnv",
	"config.Get",
	"config.",
	"env.",
	"viper.",
}

func (r *SecretsRule) Check(file git.FileDiff) []analyzer.Finding {
	if strings.HasSuffix(file.Path, "_test.go") || strings.HasSuffix(file.Path, "_test.py") {
		return nil
	}
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			lang := DetectLanguage(file.Path)
			if IsComment(line.Content, lang) {
				continue
			}
			excluded := false
			for _, ex := range secretExclusions {
				if strings.Contains(line.Content, ex) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
			for _, pat := range secretPatterns {
				if pat.MatchString(line.Content) {
					findings = append(findings, analyzer.Finding{
						File:        file.Path,
						Line:        line.NewNum,
						Severity:    analyzer.SeverityHigh,
						Rule:        r.Name(),
						Description: "Possible hardcoded secret detected — use environment variables or a secret manager",
						Suggestion:  "Move the secret to an environment variable or config file excluded from version control",
					})
					break
				}
			}
		}
	}
	return findings
}
