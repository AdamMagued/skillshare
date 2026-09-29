package audit

import (
	"path/filepath"
	"strings"
)

// ScanContent scans raw content for security issues and returns findings.
// filename is used for reporting (e.g. "SKILL.md").
func ScanContent(content []byte, filename string) []Finding {
	return ScanContentWithRules(content, filename, nil)
}

// ScanContentWithRules scans content using the given rules.
// If rules is nil, the default global rules are used.
func ScanContentWithRules(content []byte, filename string, activeRules []rule) []Finding {
	if activeRules == nil {
		var err error
		activeRules, err = Rules()
		if err != nil {
			return nil
		}
	}

	var findings []Finding
	text := string(content)
	lineNum := 0
	for start := 0; start <= len(text); {
		lineNum++
		end := strings.IndexByte(text[start:], '\n')
		var line string
		if end == -1 {
			line = text[start:]
			start = len(text) + 1
		} else {
			line = text[start : start+end]
			start = start + end + 1
		}
		lineLower := ""
		lineLowerReady := false
		for _, r := range activeRules {
			if !rulePrefilterAllows(r, line, &lineLower, &lineLowerReady) {
				continue
			}
			if r.Regex.MatchString(line) {
				if r.Exclude != nil && r.Exclude.MatchString(line) {
					continue
				}
				findings = append(findings, Finding{
					Severity:   r.Severity,
					Pattern:    r.Pattern,
					Message:    r.Message,
					File:       filename,
					Line:       lineNum,
					Snippet:    strings.TrimSpace(line),
					RuleID:     r.ID,
					Analyzer:   AnalyzerStatic,
					Category:   categoryForPattern(r.Pattern),
					Confidence: 0.95,
				})
			}
		}
	}

	return findings
}

// ScanMarkdownContentWithRules scans markdown content and suppresses selected
// non-critical patterns when they appear in educational example context.
func ScanMarkdownContentWithRules(content []byte, filename string, activeRules []rule) []Finding {
	if activeRules == nil {
		var err error
		activeRules, err = Rules()
		if err != nil {
			return nil
		}
	}

	var findings []Finding
	text := string(content)
	inCodeFence := false
	fenceMarker := ""
	tutorialPath := isLikelyTutorialPath(filename)
	lineNum := 0

	for start := 0; start <= len(text); {
		lineNum++
		end := strings.IndexByte(text[start:], '\n')
		var line string
		if end == -1 {
			line = text[start:]
			start = len(text) + 1
		} else {
			line = text[start : start+end]
			start = start + end + 1
		}

		if marker, ok := detectFenceMarker(line); ok {
			if !inCodeFence {
				inCodeFence = true
				fenceMarker = marker
			} else if marker == fenceMarker {
				inCodeFence = false
				fenceMarker = ""
			}
			continue
		}

		lineLower := ""
		lineLowerReady := false
		for _, r := range activeRules {
			if !rulePrefilterAllows(r, line, &lineLower, &lineLowerReady) {
				continue
			}
			if !r.Regex.MatchString(line) {
				continue
			}
			if r.Exclude != nil && r.Exclude.MatchString(line) {
				continue
			}
			if shouldSuppressTutorialExample(r.Pattern, line, inCodeFence, tutorialPath) {
				continue
			}
			findings = append(findings, Finding{
				Severity:   r.Severity,
				Pattern:    r.Pattern,
				Message:    r.Message,
				File:       filename,
				Line:       lineNum,
				Snippet:    strings.TrimSpace(line),
				RuleID:     r.ID,
				Analyzer:   AnalyzerStatic,
				Category:   categoryForPattern(r.Pattern),
				Confidence: 0.95,
			})
		}
	}

	return findings
}

var tutorialSuppressedPatterns = map[string]bool{
	"dynamic-code-exec":    true,
	"shell-execution":      true,
	"destructive-commands": true,
	"suspicious-fetch":     true,

	"insecure-http":      true,
	"escape-obfuscation": true,
	"hidden-unicode":     true,
	"fetch-with-pipe":    true,
	"untrusted-install":  true,
}

func shouldSuppressTutorialExample(pattern, line string, inCodeFence, tutorialPath bool) bool {
	if !tutorialSuppressedPatterns[pattern] {
		return false
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if inCodeFence || tutorialPath {
		return true
	}
	return mdTutorialMarkerRe.MatchString(trimmed)
}

func isLikelyTutorialPath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	parts := strings.Split(lower, "/")
	for _, p := range parts {
		switch p {
		case "reference", "references", "resource", "resources", "template", "templates", "example", "examples":
			return true
		}
	}
	return false
}
