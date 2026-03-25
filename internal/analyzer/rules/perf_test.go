package rules

import "testing"

func TestStringConcatRule_Name(t *testing.T) {
	t.Parallel()
	rule := &StringConcatRule{}
	if rule.Name() != "perf_string_concat" {
		t.Errorf("expected name perf_string_concat, got %q", rule.Name())
	}
}

func TestStringConcatRule_DetectsConcatInLoop(t *testing.T) {
	t.Parallel()
	rule := &StringConcatRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, s := range items {", 10),
		makeAddedLine("\tresult += s", 11),
		makeAddedLine("}", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 11 {
		t.Errorf("expected line 11, got %d", findings[0].Line)
	}
	if findings[0].Rule != "perf_string_concat" {
		t.Errorf("expected rule perf_string_concat, got %q", findings[0].Rule)
	}
}

func TestStringConcatRule_NoConcatOutsideLoop(t *testing.T) {
	t.Parallel()
	rule := &StringConcatRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("result += extra", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for concat outside loop, got %d", len(findings))
	}
}

func TestStringConcatRule_SkipsCommentInsideLoop(t *testing.T) {
	t.Parallel()
	rule := &StringConcatRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, s := range items {", 10),
		makeAddedLine("\t// result += s", 11),
		makeAddedLine("}", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for comment inside loop, got %d", len(findings))
	}
}

func TestRegexLoopRule_Name(t *testing.T) {
	t.Parallel()
	rule := &RegexLoopRule{}
	if rule.Name() != "perf_regex_loop" {
		t.Errorf("expected name perf_regex_loop, got %q", rule.Name())
	}
}

func TestRegexLoopRule_DetectsRegexpInLoop(t *testing.T) {
	t.Parallel()
	rule := &RegexLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, s := range items {", 10),
		makeAddedLine("\tre := regexp.MustCompile(`\\d+`)", 11),
		makeAddedLine("}", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 11 {
		t.Errorf("expected line 11, got %d", findings[0].Line)
	}
	if findings[0].Rule != "perf_regex_loop" {
		t.Errorf("expected rule perf_regex_loop, got %q", findings[0].Rule)
	}
}

func TestRegexLoopRule_NoRegexpAtPackageLevel(t *testing.T) {
	t.Parallel()
	rule := &RegexLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine(`var re = regexp.MustCompile(`+"`"+`\d+`+"`"+`)`, 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for regexp at package level, got %d", len(findings))
	}
}

func TestRegexLoopRule_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &RegexLoopRule{}
	diff := makeFileDiff("main.py",
		makeAddedLine("for s in items:", 10),
		makeAddedLine("\tre = regexp.MustCompile(r'\\d+')", 11),
		makeAddedLine("", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-Go file, got %d", len(findings))
	}
}

func TestRegexLoopRule_DetectsCompileVariant(t *testing.T) {
	t.Parallel()
	rule := &RegexLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, s := range items {", 10),
		makeAddedLine("\tre, _ := regexp.Compile(`\\d+`)", 11),
		makeAddedLine("}", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for regexp.Compile in loop, got %d", len(findings))
	}
}
