package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const systemPrompt = `You are being evaluated on your ability to correctly apply Entur platform conventions.
You MUST read the documentation files in this repository before answering.
Start by reading AGENTS.md to understand the documentation structure, then read the
specific guides referenced for your task. Pay close attention to:
- The difference between metadata.id and metadata.name
- GCP project naming patterns for each manifest kind
- Language-specific conventions (health paths, Docker images, metrics paths)
Do not make assumptions. If the documentation specifies a pattern, use it exactly.`

var (
	activeSysPrompt string
	activeTools     string
)

// claudeResponse is the JSON structure returned by claude --output-format json.
type claudeResponse struct {
	Type       string  `json:"type"`
	Result     string  `json:"result"`
	IsError    bool    `json:"is_error"`
	CostUSD    float64 `json:"total_cost_usd"`
	DurationMS int64   `json:"duration_ms"`
}

func main() {
	scenarioFilter := flag.String("scenario", "", "glob pattern to filter scenarios (e.g. \"01-*\")")
	model := flag.String("model", "haiku", "Claude model to use (haiku, sonnet, opus)")
	totalBudget := flag.Float64("budget", 1.00, "total budget cap in USD")
	junitPath := flag.String("junit", "", "write JUnit XML report to this file")
	verbose := flag.Bool("verbose", false, "print full Claude responses for failed scenarios")
	dryRun := flag.Bool("dry-run", false, "parse scenarios and print commands without executing")
	strict := flag.Bool("strict", false, "require 100% assertion pass rate")
	noRetry := flag.Bool("no-retry", false, "disable retry on failure")
	scenarioDir := flag.String("dir", "", "scenario directory (default: ./scenarios relative to binary)")
	sysPromptOverride := flag.String("system-prompt", "", "override default system prompt ('none' to omit)")
	allowedToolsFlag := flag.String("allowed-tools", "", "override allowed tools ('none' to omit, default: Read,Grep,Glob)")
	flag.Parse()

	// Apply system prompt and tools configuration
	activeSysPrompt = systemPrompt
	if *sysPromptOverride != "" {
		if *sysPromptOverride == "none" {
			activeSysPrompt = ""
		} else {
			activeSysPrompt = *sysPromptOverride
		}
	}
	activeTools = "Read,Grep,Glob"
	if *allowedToolsFlag != "" {
		if *allowedToolsFlag == "none" {
			activeTools = ""
		} else {
			activeTools = *allowedToolsFlag
		}
	}

	dir := *scenarioDir
	if dir == "" {
		// Find scenarios relative to the current working directory or binary location
		candidates := []string{
			"tests/scenarios",
			"scenarios",
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				dir = c
				break
			}
		}
		if dir == "" {
			fmt.Fprintln(os.Stderr, "error: could not find scenarios directory. Use --dir to specify.")
			os.Exit(1)
		}
	}

	// Find the repo root (for CWD when running claude)
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		fmt.Fprintln(os.Stderr, "error: could not find repository root (no AGENTS.md found)")
		os.Exit(1)
	}

	// Check claude is available
	if !*dryRun {
		if _, err := exec.LookPath("claude"); err != nil {
			fmt.Fprintln(os.Stderr, "error: claude CLI not found. Install with: npm install -g @anthropic-ai/claude-code")
			os.Exit(1)
		}
	}

	// Discover and parse scenarios
	scenarios, err := discoverScenarios(dir, *scenarioFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(scenarios) == 0 {
		fmt.Fprintln(os.Stderr, "no scenarios found")
		os.Exit(1)
	}

	fmt.Printf("=== Entur AI Documentation Comprehension Tests ===\n")
	fmt.Printf("Model: %s | Budget cap: $%.2f | Scenarios: %d\n\n", *model, *totalBudget, len(scenarios))

	if *dryRun {
		for i, s := range scenarios {
			cmd := buildCommand(s, *model, repoRoot)
			fmt.Printf("[%d/%d] %s\n", i+1, len(scenarios), s.Name)
			fmt.Printf("  Budget: $%.2f\n", s.Budget)
			fmt.Printf("  Assertions: %d must_contain, %d must_not_contain, %d must_match\n",
				len(s.Assertions.MustContain), len(s.Assertions.MustNotContain), len(s.Assertions.MustMatch))
			fmt.Printf("  Command: %s\n\n", strings.Join(cmd, " "))
		}
		fmt.Println("Dry run complete. No API calls made.")
		return
	}

	var results []ScenarioResult
	var totalCost float64

	for i, s := range scenarios {
		if totalCost >= *totalBudget {
			fmt.Printf("\n[!] Budget cap reached ($%.2f). Skipping remaining scenarios.\n", *totalBudget)
			break
		}

		fmt.Printf("[%d/%d] %s ", i+1, len(scenarios), s.Name)

		result := runScenario(s, *model, *strict, repoRoot)
		totalCost += result.CostUSD

		// Retry once if failed and retries enabled
		if !result.Passed && !*noRetry && result.Error == "" {
			fmt.Print("-> retrying ")
			retry := runScenario(s, *model, *strict, repoRoot)
			totalCost += retry.CostUSD
			if retry.Passed {
				retry.Flaky = true
				result = retry
			}
		}

		results = append(results, result)

		// Print result
		totalAssertions := len(result.AssertionResults)
		passedAssertions := 0
		for _, a := range result.AssertionResults {
			if a.Passed {
				passedAssertions++
			}
		}

		status := "PASS"
		if result.Error != "" {
			status = "ERROR"
		} else if !result.Passed {
			status = "FAIL"
		} else if result.Flaky {
			status = "FLAKY"
		}

		fmt.Printf("... %s (%d/%d assertions, $%.3f, %.1fs)\n",
			status, passedAssertions, totalAssertions, result.CostUSD,
			float64(result.DurationMS)/1000)

		// Print failures
		if !result.Passed || *verbose {
			for _, a := range result.AssertionResults {
				if !a.Passed {
					fmt.Printf("      FAIL %s: %s\n", a.Kind, a.Detail)
				}
			}
			if result.Error != "" {
				fmt.Printf("      ERROR: %s\n", result.Error)
			}
		}
		if !result.Passed && *verbose && result.RawOutput != "" {
			fmt.Printf("      --- Raw output ---\n%s\n      --- End ---\n", truncate(result.RawOutput, 2000))
		}
	}

	// Summary
	passed, flaky, failed, errored := 0, 0, 0, 0
	for _, r := range results {
		switch {
		case r.Error != "":
			errored++
		case !r.Passed:
			failed++
		case r.Flaky:
			flaky++
		default:
			passed++
		}
	}

	fmt.Printf("\nResults: %d passed", passed)
	if flaky > 0 {
		fmt.Printf(", %d flaky", flaky)
	}
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	if errored > 0 {
		fmt.Printf(", %d errors", errored)
	}
	fmt.Printf(" (total: $%.3f)\n", totalCost)

	// JUnit XML output
	if *junitPath != "" {
		if err := writeJUnit(results, *junitPath, totalCost); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JUnit report: %v\n", err)
		} else {
			fmt.Printf("JUnit report written to %s\n", *junitPath)
		}
	}

	if failed > 0 || errored > 0 {
		os.Exit(1)
	}
}

func discoverScenarios(dir, filter string) ([]Scenario, error) {
	pattern := "*.md"
	if filter != "" {
		pattern = filter
		if !strings.HasSuffix(pattern, ".md") {
			pattern += ".md"
		}
	}

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("globbing scenarios: %w", err)
	}

	sort.Strings(matches)

	var scenarios []Scenario
	for _, path := range matches {
		s, err := ParseScenario(path)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}

	return scenarios, nil
}

func buildCommand(s Scenario, model, repoRoot string) []string {
	cmd := []string{
		"claude",
		"-p", s.Prompt,
		"--output-format", "json",
		"--model", model,
		"--max-budget-usd", fmt.Sprintf("%.2f", s.Budget),
	}
	if activeTools != "" {
		cmd = append(cmd, "--allowedTools", activeTools)
	}
	if activeSysPrompt != "" {
		cmd = append(cmd, "--append-system-prompt", activeSysPrompt)
	}
	return cmd
}

func runScenario(s Scenario, model string, strict bool, repoRoot string) ScenarioResult {
	args := buildCommand(s, model, repoRoot)

	start := time.Now()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	duration := time.Since(start)

	result := ScenarioResult{
		Scenario:   s,
		DurationMS: duration.Milliseconds(),
	}

	if err != nil {
		// Try to extract stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Error = fmt.Sprintf("claude exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		} else {
			result.Error = err.Error()
		}
		// Still try to parse output in case there's a partial response
		if len(output) == 0 {
			return result
		}
	}

	// Parse JSON response
	var resp claudeResponse
	if jsonErr := json.Unmarshal(output, &resp); jsonErr != nil {
		result.Error = fmt.Sprintf("failed to parse claude response: %v (raw: %s)", jsonErr, truncate(string(output), 200))
		return result
	}

	result.RawOutput = resp.Result
	result.CostUSD = resp.CostUSD
	if resp.DurationMS > 0 {
		result.DurationMS = resp.DurationMS
	}

	if resp.IsError {
		result.Error = fmt.Sprintf("claude returned error: %s", truncate(resp.Result, 500))
		return result
	}

	// Evaluate assertions
	result.AssertionResults = EvaluateAssertions(resp.Result, s.Assertions)
	result.Passed = ScenarioPassed(result.AssertionResults, strict)

	return result
}

func findRepoRoot() string {
	// Walk up from CWD looking for AGENTS.md
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- JUnit XML ---

type junitTestSuites struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Name     string           `xml:"name,attr"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Time     string           `xml:"time,attr"`
	Suites   []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Error     *junitError   `xml:"error,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitError struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

func writeJUnit(results []ScenarioResult, path string, totalCost float64) error {
	failures := 0
	errors := 0
	var cases []junitTestCase

	for _, r := range results {
		tc := junitTestCase{
			Name:      r.Scenario.Name,
			ClassName: "scenarios",
			Time:      fmt.Sprintf("%.1f", float64(r.DurationMS)/1000),
		}

		if r.Error != "" {
			errors++
			tc.Error = &junitError{Message: r.Error}
		} else if !r.Passed {
			failures++
			var failDetails []string
			for _, a := range r.AssertionResults {
				if !a.Passed {
					failDetails = append(failDetails, fmt.Sprintf("%s: %s", a.Kind, a.Detail))
				}
			}
			passedCount := 0
			for _, a := range r.AssertionResults {
				if a.Passed {
					passedCount++
				}
			}
			tc.Failure = &junitFailure{
				Message: fmt.Sprintf("%d/%d assertions passed", passedCount, len(r.AssertionResults)),
				Body:    strings.Join(failDetails, "\n"),
			}
		} else if r.Flaky {
			tc.SystemOut = "Flaky: passed on retry"
		}

		cases = append(cases, tc)
	}

	suites := junitTestSuites{
		Name:     "entur-ai-doc-tests",
		Tests:    len(results),
		Failures: failures,
		Errors:   errors,
		Time:     fmt.Sprintf("%.1f", totalCost),
		Suites: []junitTestSuite{{
			Name:     "scenarios",
			Tests:    len(results),
			Failures: failures,
			Errors:   errors,
			Cases:    cases,
		}},
	}

	data, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return err
	}

	header := []byte(xml.Header)
	return os.WriteFile(path, append(header, data...), 0644)
}
