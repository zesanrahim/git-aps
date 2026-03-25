package rules

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/config"
	"github.com/zesanrahim/git-aps/internal/git"
)

type Rule interface {
	Name() string
	Check(file git.FileDiff) []analyzer.Finding
}

type Engine struct {
	rules []Rule
}

func NewEngine(rulesCfg map[string]config.RuleConfig, customCfg ...[]config.CustomRuleConfig) *Engine {
	var rules []Rule

	if rc, ok := rulesCfg["magic_numbers"]; !ok || rc.Enabled {
		threshold := 2
		if ok && rc.Threshold > 0 {
			threshold = rc.Threshold
		}
		rules = append(rules, &MagicNumberRule{Threshold: threshold})
	}

	if rc, ok := rulesCfg["deep_nesting"]; !ok || rc.Enabled {
		maxDepth := 3
		if ok && rc.MaxDepth > 0 {
			maxDepth = rc.MaxDepth
		}
		rules = append(rules, &DeepNestingRule{MaxDepth: maxDepth})
	}

	if rc, ok := rulesCfg["long_functions"]; !ok || rc.Enabled {
		maxLines := 50
		if ok && rc.MaxLines > 0 {
			maxLines = rc.MaxLines
		}
		rules = append(rules, &LongFunctionRule{MaxLines: maxLines})
	}

	if rc, ok := rulesCfg["many_params"]; !ok || rc.Enabled {
		maxParams := 5
		if ok && rc.MaxParams > 0 {
			maxParams = rc.MaxParams
		}
		rules = append(rules, &ManyParamsRule{MaxParams: maxParams})
	}

	if rc, ok := rulesCfg["todo_comments"]; !ok || rc.Enabled {
		rules = append(rules, &TodoCommentRule{})
	}

	if rc, ok := rulesCfg["error_ignored"]; !ok || rc.Enabled {
		rules = append(rules, &ErrorIgnoredRule{})
	}

	if rc, ok := rulesCfg["error_wrap"]; !ok || rc.Enabled {
		rules = append(rules, &ErrorWrapRule{})
	}

	if rc, ok := rulesCfg["defer_loop"]; !ok || rc.Enabled {
		rules = append(rules, &DeferLoopRule{})
	}

	if rc, ok := rulesCfg["type_assert"]; !ok || rc.Enabled {
		rules = append(rules, &TypeAssertRule{})
	}

	if rc, ok := rulesCfg["goroutine_ctx"]; !ok || rc.Enabled {
		rules = append(rules, &GoroutineCtxRule{})
	}

	if rc, ok := rulesCfg["secrets"]; !ok || rc.Enabled {
		rules = append(rules, &SecretsRule{})
	}

	if rc, ok := rulesCfg["sql_injection"]; !ok || rc.Enabled {
		rules = append(rules, &SQLInjectionRule{})
	}

	if rc, ok := rulesCfg["cognitive_complexity"]; !ok || rc.Enabled {
		threshold := 15
		if ok && rc.MaxComplexity > 0 {
			threshold = rc.MaxComplexity
		}
		rules = append(rules, &CognitiveComplexityRule{Threshold: threshold})
	}

	if rc, ok := rulesCfg["perf_string_concat"]; !ok || rc.Enabled {
		rules = append(rules, &StringConcatRule{})
	}

	if rc, ok := rulesCfg["perf_regex_loop"]; !ok || rc.Enabled {
		rules = append(rules, &RegexLoopRule{})
	}

	if len(customCfg) > 0 {
		for _, cr := range customCfg[0] {
			if cr.Name == "" || cr.Pattern == "" {
				continue
			}
			re, err := regexp.Compile(cr.Pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "custom rule %q: invalid pattern: %v\n", cr.Name, err)
				continue
			}
			rules = append(rules, &CustomRule{
				RuleName:    cr.Name,
				Pattern:     re,
				Sev:         analyzer.ParseSeverity(cr.Severity),
				Desc:        cr.Description,
				Suggestion_: cr.Suggestion,
			})
		}
	}

	return &Engine{rules: rules}
}

func (e *Engine) Analyze(diffs []git.FileDiff) ([]analyzer.Finding, error) {
	var findings []analyzer.Finding
	for _, diff := range diffs {
		for _, rule := range e.rules {
			findings = append(findings, rule.Check(diff)...)
		}
	}

	fileCache := make(map[string][]string)
	for i := range findings {
		f := &findings[i]
		if f.OriginalCode != "" || f.Line < 1 {
			continue
		}
		lines, ok := fileCache[f.File]
		if !ok {
			data, err := os.ReadFile(f.File)
			if err != nil {
				continue
			}
			lines = strings.Split(string(data), "\n")
			fileCache[f.File] = lines
		}
		if f.Line > len(lines) {
			continue
		}
		end := f.EndLine
		if end < f.Line {
			end = f.Line
		}
		if end > len(lines) {
			end = len(lines)
		}
		f.EndLine = end
		f.OriginalCode = strings.Join(lines[f.Line-1:end], "\n")
	}

	return findings, nil
}
