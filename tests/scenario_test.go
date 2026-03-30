package main

import (
	"os"
	"path/filepath"
	"testing"
)

const testScenarioContent = `# Scenario: Test Identity Chain

## Description
Tests that the parser correctly extracts all sections.

## Prompt
You are setting up a new app called "my-app" with App ID "myapp".
What GCP project names will be created?

## Assertions
` + "```json" + `
{
  "must_contain": ["ent-myapp-dev", "ent-myapp-prd"],
  "must_not_contain": ["ent-my-app-dev", "google_project"],
  "must_match": ["shortname.*myapp", "app_id.*myapp"]
}
` + "```" + `

## Budget
0.05
`

func TestParseScenario(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "01-test.md")
	if err := os.WriteFile(path, []byte(testScenarioContent), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseScenario(path)
	if err != nil {
		t.Fatalf("ParseScenario: %v", err)
	}

	if s.Name != "Test Identity Chain" {
		t.Errorf("Name = %q, want %q", s.Name, "Test Identity Chain")
	}

	if s.Budget != 0.05 {
		t.Errorf("Budget = %f, want 0.05", s.Budget)
	}

	if !containsStr(s.Prompt, "my-app") {
		t.Error("Prompt should contain 'my-app'")
	}

	if len(s.Assertions.MustContain) != 2 {
		t.Errorf("MustContain length = %d, want 2", len(s.Assertions.MustContain))
	}
	if len(s.Assertions.MustNotContain) != 2 {
		t.Errorf("MustNotContain length = %d, want 2", len(s.Assertions.MustNotContain))
	}
	if len(s.Assertions.MustMatch) != 2 {
		t.Errorf("MustMatch length = %d, want 2", len(s.Assertions.MustMatch))
	}
}

func TestParseScenarioMissingPrompt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	content := "# Test\n\n## Assertions\n```json\n{}\n```\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseScenario(path)
	if err == nil {
		t.Error("expected error for missing Prompt section")
	}
}

func TestParseScenarioDefaultBudget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-budget.md")
	content := "# Test\n\n## Prompt\nHello\n\n## Assertions\n```json\n{\"must_contain\": [\"hi\"]}\n```\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseScenario(path)
	if err != nil {
		t.Fatalf("ParseScenario: %v", err)
	}
	if s.Budget != defaultBudget {
		t.Errorf("Budget = %f, want default %f", s.Budget, defaultBudget)
	}
}

func TestEvaluateAssertions_AllPass(t *testing.T) {
	output := `The GCP projects will be ent-myapp-dev, ent-myapp-tst, ent-myapp-prd.
The Helm shortname should be "myapp".
The Terraform app_id = "myapp".`

	assertions := Assertions{
		MustContain:    []string{"ent-myapp-dev", "ent-myapp-prd"},
		MustNotContain: []string{"google_project", "ent-my-app-dev"},
		MustMatch:      []string{"shortname.*myapp", "app_id.*myapp"},
	}

	results := EvaluateAssertions(output, assertions)

	for _, r := range results {
		if !r.Passed {
			t.Errorf("assertion %s %q failed: %s", r.Kind, r.Value, r.Detail)
		}
	}
}

func TestEvaluateAssertions_MustContainFails(t *testing.T) {
	output := "The project is ent-wrong-dev"

	assertions := Assertions{
		MustContain: []string{"ent-myapp-dev"},
	}

	results := EvaluateAssertions(output, assertions)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("expected must_contain to fail")
	}
}

func TestEvaluateAssertions_MustNotContainFails(t *testing.T) {
	output := "Use resource google_project to create..."

	assertions := Assertions{
		MustNotContain: []string{"google_project"},
	}

	results := EvaluateAssertions(output, assertions)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("expected must_not_contain to fail when forbidden string is present")
	}
}

func TestEvaluateAssertions_CaseInsensitive(t *testing.T) {
	output := "Set SHORTNAME: MyApp in the values"

	assertions := Assertions{
		MustContain: []string{"shortname: myapp"},
		MustMatch:   []string{"shortname.*myapp"},
	}

	results := EvaluateAssertions(output, assertions)
	for _, r := range results {
		if !r.Passed {
			t.Errorf("case-insensitive assertion %s %q failed: %s", r.Kind, r.Value, r.Detail)
		}
	}
}

func TestEvaluateAssertions_RegexWithYAMLVariations(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"unquoted", `shortname: jpapi`},
		{"single-quoted", `shortname: 'jpapi'`},
		{"double-quoted", `shortname: "jpapi"`},
		{"prose", `The shortname is jpapi.`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertions := Assertions{
				MustMatch: []string{`shortname.*jpapi`},
			}
			results := EvaluateAssertions(tt.output, assertions)
			if !results[0].Passed {
				t.Errorf("failed to match %q: %s", tt.output, results[0].Detail)
			}
		})
	}
}

func TestScenarioPassed_Strict(t *testing.T) {
	results := []AssertionResult{
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: false},
		{Kind: "must_not_contain", Passed: true},
	}
	if ScenarioPassed(results, true) {
		t.Error("strict mode should fail when any assertion fails")
	}
}

func TestScenarioPassed_Normal_MustNotContainFails(t *testing.T) {
	results := []AssertionResult{
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: true},
		{Kind: "must_not_contain", Passed: false}, // fundamental misunderstanding
	}
	if ScenarioPassed(results, false) {
		t.Error("should fail when must_not_contain fails, even in normal mode")
	}
}

func TestScenarioPassed_Normal_80Percent(t *testing.T) {
	// 4/5 positive assertions = 80% -> pass
	results := []AssertionResult{
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: true},
		{Kind: "must_contain", Passed: false},
		{Kind: "must_not_contain", Passed: true},
	}
	if !ScenarioPassed(results, false) {
		t.Error("80% pass rate should be sufficient in normal mode")
	}

	// 3/5 = 60% -> fail
	results[3].Passed = false
	if ScenarioPassed(results, false) {
		t.Error("60% pass rate should fail in normal mode")
	}
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && findSubstr(s, substr))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
