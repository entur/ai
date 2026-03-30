package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Scenario represents a parsed test scenario from a markdown file.
type Scenario struct {
	Name        string
	Description string
	Prompt      string
	Assertions  Assertions
	Budget      float64
	FilePath    string
}

// Assertions defines the expected outputs for a scenario.
type Assertions struct {
	MustContain    []string `json:"must_contain"`
	MustNotContain []string `json:"must_not_contain"`
	MustMatch      []string `json:"must_match"`
}

// AssertionResult records whether a single assertion passed.
type AssertionResult struct {
	Kind    string // "must_contain", "must_not_contain", "must_match"
	Value   string // the expected string or regex pattern
	Passed  bool
	Detail  string // explanation
}

// ScenarioResult records the outcome of running a scenario.
type ScenarioResult struct {
	Scenario         Scenario
	Passed           bool
	Flaky            bool // passed on retry
	AssertionResults []AssertionResult
	RawOutput        string
	CostUSD          float64
	DurationMS       int64
	Error            string
}

const defaultBudget = 0.08

// ParseScenario reads a markdown scenario file and extracts its components.
func ParseScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	s := Scenario{
		FilePath: path,
		Budget:   defaultBudget,
	}

	// Extract name from first H1
	if name := extractSection(content, "# "); name != "" {
		s.Name = strings.TrimPrefix(name, "Scenario: ")
		s.Name = strings.TrimSpace(s.Name)
	} else {
		// Derive from filename
		base := path[strings.LastIndex(path, "/")+1:]
		s.Name = strings.TrimSuffix(base, ".md")
	}

	s.Description = extractH2Section(content, "Description")
	s.Prompt = extractH2Section(content, "Prompt")
	if s.Prompt == "" {
		return s, fmt.Errorf("scenario %s: missing ## Prompt section", path)
	}

	// Extract assertions from JSON code block inside ## Assertions section
	assertionsRaw := extractH2Section(content, "Assertions")
	if assertionsRaw == "" {
		return s, fmt.Errorf("scenario %s: missing ## Assertions section", path)
	}

	jsonBlock := extractJSONBlock(assertionsRaw)
	if jsonBlock == "" {
		return s, fmt.Errorf("scenario %s: no JSON code block found in ## Assertions", path)
	}

	if err := json.Unmarshal([]byte(jsonBlock), &s.Assertions); err != nil {
		return s, fmt.Errorf("scenario %s: invalid assertions JSON: %w", path, err)
	}

	// Extract budget
	budgetStr := extractH2Section(content, "Budget")
	if budgetStr != "" {
		if b, err := strconv.ParseFloat(strings.TrimSpace(budgetStr), 64); err == nil {
			s.Budget = b
		}
	}

	return s, nil
}

// EvaluateAssertions checks the AI output against the scenario's assertions.
func EvaluateAssertions(output string, assertions Assertions) []AssertionResult {
	var results []AssertionResult
	outputLower := strings.ToLower(output)

	for _, expected := range assertions.MustContain {
		found := strings.Contains(outputLower, strings.ToLower(expected))
		detail := ""
		if found {
			idx := strings.Index(outputLower, strings.ToLower(expected))
			detail = fmt.Sprintf("found at position %d", idx)
		} else {
			detail = fmt.Sprintf("%q not found in output", expected)
		}
		results = append(results, AssertionResult{
			Kind:   "must_contain",
			Value:  expected,
			Passed: found,
			Detail: detail,
		})
	}

	for _, forbidden := range assertions.MustNotContain {
		found := strings.Contains(outputLower, strings.ToLower(forbidden))
		detail := ""
		if found {
			idx := strings.Index(outputLower, strings.ToLower(forbidden))
			detail = fmt.Sprintf("forbidden string %q found at position %d", forbidden, idx)
		} else {
			detail = fmt.Sprintf("%q correctly absent", forbidden)
		}
		results = append(results, AssertionResult{
			Kind:   "must_not_contain",
			Value:  forbidden,
			Passed: !found,
			Detail: detail,
		})
	}

	for _, pattern := range assertions.MustMatch {
		// Prepend (?i) for case-insensitive matching
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			results = append(results, AssertionResult{
				Kind:   "must_match",
				Value:  pattern,
				Passed: false,
				Detail: fmt.Sprintf("invalid regex: %v", err),
			})
			continue
		}

		match := re.FindString(output)
		passed := match != ""
		detail := ""
		if passed {
			detail = fmt.Sprintf("matched: %q", match)
		} else {
			detail = fmt.Sprintf("pattern /%s/ not matched", pattern)
		}
		results = append(results, AssertionResult{
			Kind:   "must_match",
			Value:  pattern,
			Passed: passed,
			Detail: detail,
		})
	}

	return results
}

// ScenarioPassed determines if a scenario passed based on its assertion results.
// In strict mode, all assertions must pass.
// In normal mode, all must_not_contain must pass and >= 80% of other assertions must pass.
func ScenarioPassed(results []AssertionResult, strict bool) bool {
	if len(results) == 0 {
		return true
	}

	if strict {
		for _, r := range results {
			if !r.Passed {
				return false
			}
		}
		return true
	}

	// All must_not_contain must pass
	positiveTotal := 0
	positivePassed := 0
	for _, r := range results {
		if r.Kind == "must_not_contain" {
			if !r.Passed {
				return false
			}
		} else {
			positiveTotal++
			if r.Passed {
				positivePassed++
			}
		}
	}

	if positiveTotal == 0 {
		return true
	}

	return float64(positivePassed)/float64(positiveTotal) >= 0.8
}

// --- Markdown parsing helpers ---

// extractSection extracts text following a markdown heading prefix on the first matching line.
func extractSection(content, prefix string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// extractH2Section extracts the content under an H2 heading, stopping at the next H2 or end.
func extractH2Section(content, heading string) string {
	marker := "## " + heading
	lines := strings.Split(content, "\n")

	start := -1
	for i, line := range lines {
		if strings.EqualFold(strings.TrimSpace(line), marker) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}

	var sb strings.Builder
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

// extractJSONBlock extracts the first JSON fenced code block from text.
func extractJSONBlock(text string) string {
	// Match ```json ... ``` or ``` ... ```
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n(.*?)\\n```")
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}
