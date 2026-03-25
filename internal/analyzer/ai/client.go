package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type Config struct {
	Model      string
	BaseURL    string
	APIKey     string
	Debug      bool
	MaxRetries int
}

func DefaultConfig() Config {
	apiKey := os.Getenv("AI_API_KEY")
	return Config{
		Model:      "gemini-2.5-flash",
		BaseURL:    "https://generativelanguage.googleapis.com/v1beta/openai",
		APIKey:     apiKey,
		MaxRetries: 3,
	}
}

type AIAnalyzer struct {
	client *openai.Client
	config Config
}

func New(cfg Config) *AIAnalyzer {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	ocfg := openai.DefaultConfig(cfg.APIKey)
	ocfg.BaseURL = cfg.BaseURL
	return &AIAnalyzer{
		client: openai.NewClientWithConfig(ocfg),
		config: cfg,
	}
}

func (a *AIAnalyzer) Analyze(diffs []git.FileDiff) ([]analyzer.Finding, error) {
	if a.config.APIKey == "" {
		return nil, nil
	}

	var allFindings []analyzer.Finding
	for _, diff := range diffs {
		if diff.Deleted {
			continue
		}
		findings, err := a.analyzeFile(diff)
		if err != nil {
			return allFindings, fmt.Errorf("ai analyze %s: %w", diff.Path, err)
		}
		allFindings = append(allFindings, findings...)
	}
	return allFindings, nil
}

func (a *AIAnalyzer) analyzeFile(diff git.FileDiff) ([]analyzer.Finding, error) {
	diffText := formatDiffForPrompt(diff)
	fileContent := readFullFile(diff.Path)
	prompt := buildPrompt(diff.Path, diffText, fileContent)

	if a.config.Debug {
		fmt.Fprintf(os.Stderr, "\n=== DIFF TEXT for %s ===\n%s\n=== END DIFF ===\n", diff.Path, diffText)
		fmt.Fprintf(os.Stderr, "File content length: %d bytes\n", len(fileContent))
	}

	var resp openai.ChatCompletionResponse
	var err error
	for attempt := 0; attempt < a.config.MaxRetries; attempt++ {
		resp, err = a.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: a.config.Model,
				Messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
					{Role: openai.ChatMessageRoleUser, Content: prompt},
				},
				Temperature: 0.3,
			},
		)
		if err == nil {
			break
		}
		if !isRetryableError(err) {
			return nil, err
		}
		if attempt < a.config.MaxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	raw := resp.Choices[0].Message.Content
	if a.config.Debug {
		fmt.Fprintf(os.Stderr, "\n=== AI RAW RESPONSE for %s ===\n%s\n=== END ===\n", diff.Path, raw)
	}

	findings, err := parseResponse(diff.Path, raw)
	if err != nil {
		return nil, err
	}

	findings = validateFindings(findings, diff)

	fileLines := strings.Split(fileContent, "\n")
	for i := range findings {
		f := &findings[i]
		if f.Line < 1 || f.Line > len(fileLines) {
			continue
		}

		endLine := f.EndLine
		if endLine < f.Line {
			if f.FixCode != "" {
				endLine = f.Line + strings.Count(f.FixCode, "\n")
			} else {
				endLine = f.Line
			}
		}
		if endLine > len(fileLines) {
			endLine = len(fileLines)
		}
		f.EndLine = endLine
		f.OriginalCode = strings.Join(fileLines[f.Line-1:endLine], "\n")
	}

	return findings, nil
}

func validateFindings(findings []analyzer.Finding, diff git.FileDiff) []analyzer.Finding {
	addedLines := make(map[int]bool)
	maxLine := 0
	for _, hunk := range diff.Hunks {
		for _, line := range hunk.Lines {
			if line.Type == git.LineAdded {
				addedLines[line.NewNum] = true
			}
			if line.NewNum > maxLine {
				maxLine = line.NewNum
			}
		}
	}

	var valid []analyzer.Finding
	for _, f := range findings {
		if f.Line < 1 {
			continue
		}
		if f.Description == "" || f.Rule == "" {
			continue
		}
		if maxLine > 0 && f.Line > maxLine {
			continue
		}
		valid = append(valid, f)
	}
	return valid
}

func isRetryableError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "500") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "504") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "temporarily")
}

func readFullFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func formatDiffForPrompt(diff git.FileDiff) string {
	var sb string
	for _, hunk := range diff.Hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case git.LineAdded:
				sb += fmt.Sprintf("+%d: %s\n", line.NewNum, line.Content)
			case git.LineRemoved:
				sb += fmt.Sprintf("-%d: %s\n", line.OldNum, line.Content)
			default:
				sb += fmt.Sprintf(" %d: %s\n", line.NewNum, line.Content)
			}
		}
	}
	return sb
}
