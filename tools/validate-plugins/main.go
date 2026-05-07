// validate-plugins checks marketplace.json, every plugin.json, hooks.json,
// and SKILL.md frontmatter against the documented Claude Code plugin schema
// (https://code.claude.com/docs/en/plugins-reference).
//
// Run from the repo root:
//
//	go run ./tools/validate-plugins
//
// Exits non-zero on any validation error.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const marketplacePath = ".claude-plugin/marketplace.json"

var (
	skillNameRE = regexp.MustCompile(`^[a-z0-9-]+$`)
	allowedConfigTypes = map[string]bool{
		"string": true, "number": true, "boolean": true, "directory": true, "file": true,
	}
)

type marketplace struct {
	Name    string           `json:"name"`
	Plugins []marketplaceRef `json:"plugins"`
}

type marketplaceRef struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type pluginManifest struct {
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	UserConfig  map[string]userConfig   `json:"userConfig"`
	LSPServers  json.RawMessage         `json:"lspServers"`
	Hooks       json.RawMessage         `json:"hooks"`
}

type userConfig struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Required    *bool  `json:"required"`
	Default     any    `json:"default"`
}

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type commandFrontmatter struct {
	Description  string `yaml:"description"`
	ArgumentHint string `yaml:"argument-hint"`
}

type result struct {
	errors []string
}

func (r *result) errf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func main() {
	r := &result{}

	mkt, err := loadMarketplace(marketplacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("Marketplace %q: %d plugins\n", mkt.Name, len(mkt.Plugins))

	for _, ref := range mkt.Plugins {
		validatePlugin(ref, r)
	}

	if len(r.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nFAIL: %d issue(s)\n", len(r.errors))
		for _, e := range r.errors {
			fmt.Fprintln(os.Stderr, " - "+e)
		}
		os.Exit(1)
	}
	fmt.Println("OK")
}

func loadMarketplace(path string) (*marketplace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m marketplace
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(m.Plugins) == 0 {
		return nil, fmt.Errorf("%s: no plugins listed", path)
	}
	return &m, nil
}

func validatePlugin(ref marketplaceRef, r *result) {
	prefix := "plugin " + ref.Name
	if ref.Source == "" {
		r.errf("%s: missing source", prefix)
		return
	}
	root := strings.TrimPrefix(ref.Source, "./")

	manifestPath := filepath.Join(root, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		r.errf("%s: read %s: %v", prefix, manifestPath, err)
		return
	}
	var m pluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		r.errf("%s: parse %s: %v", prefix, manifestPath, err)
		return
	}

	if m.Name == "" {
		r.errf("%s: %s missing name", prefix, manifestPath)
	} else if m.Name != ref.Name {
		r.errf("%s: marketplace name %q != plugin.json name %q", prefix, ref.Name, m.Name)
	}
	if m.Version == "" {
		r.errf("%s: missing version", prefix)
	}
	if m.Description == "" {
		r.errf("%s: missing description", prefix)
	}

	for key, cfg := range m.UserConfig {
		ctx := fmt.Sprintf("%s userConfig.%s", prefix, key)
		if cfg.Type == "" {
			r.errf("%s: missing type", ctx)
		} else if !allowedConfigTypes[cfg.Type] {
			r.errf("%s: type %q not in allowed set (string, number, boolean, directory, file)", ctx, cfg.Type)
		}
		if cfg.Title == "" {
			r.errf("%s: missing title", ctx)
		}
		if cfg.Description == "" {
			r.errf("%s: missing description", ctx)
		}
	}

	validateHooks(root, prefix, r)
	validateSkills(root, prefix, r)
	validateCommands(root, prefix, r)
}

func validateHooks(root, prefix string, r *result) {
	hooksPath := filepath.Join(root, "hooks", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		r.errf("%s: read %s: %v", prefix, hooksPath, err)
		return
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		r.errf("%s: parse %s: %v", prefix, hooksPath, err)
	}
}

func validateSkills(root, prefix string, r *result) {
	skillsDir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		r.errf("%s: read %s: %v", prefix, skillsDir, err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		validateSkill(skillMD, prefix+" skill "+e.Name(), r)
	}
}

func validateSkill(path, prefix string, r *result) {
	data, err := os.ReadFile(path)
	if err != nil {
		r.errf("%s: read %s: %v", prefix, path, err)
		return
	}
	fm, err := extractFrontmatter(data)
	if err != nil {
		r.errf("%s: %v", prefix, err)
		return
	}
	var sf skillFrontmatter
	if err := yaml.Unmarshal(fm, &sf); err != nil {
		r.errf("%s: parse frontmatter: %v", prefix, err)
		return
	}
	if sf.Description == "" {
		r.errf("%s: SKILL.md missing description (recommended for model-invocation triggering)", prefix)
	}
	if sf.Name != "" {
		if !skillNameRE.MatchString(sf.Name) {
			r.errf("%s: name %q must match [a-z0-9-]+", prefix, sf.Name)
		}
		if len(sf.Name) > 64 {
			r.errf("%s: name %q exceeds 64 chars", prefix, sf.Name)
		}
	}
	combined := len(sf.Description) + len(sf.Name)
	if combined > 1536 {
		r.errf("%s: combined name+description=%d exceeds documented 1536-char skill listing cap", prefix, combined)
	}
}

func validateCommands(root, prefix string, r *result) {
	commandsDir := filepath.Join(root, "commands")
	entries, err := os.ReadDir(commandsDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		r.errf("%s: read %s: %v", prefix, commandsDir, err)
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(commandsDir, e.Name())
		ctx := prefix + " command " + strings.TrimSuffix(e.Name(), ".md")
		validateCommand(path, ctx, r)
	}
}

func validateCommand(path, prefix string, r *result) {
	data, err := os.ReadFile(path)
	if err != nil {
		r.errf("%s: read %s: %v", prefix, path, err)
		return
	}
	fm, err := extractFrontmatter(data)
	if err != nil {
		r.errf("%s: %v", prefix, err)
		return
	}
	var cf commandFrontmatter
	if err := yaml.Unmarshal(fm, &cf); err != nil {
		r.errf("%s: parse frontmatter: %v", prefix, err)
		return
	}
	if cf.Description == "" {
		r.errf("%s: missing description (required for /help and slash-command discovery)", prefix)
	}
}

func extractFrontmatter(data []byte) ([]byte, error) {
	const sep = "---"
	s := string(data)
	if !strings.HasPrefix(s, sep) {
		return nil, fmt.Errorf("missing leading --- frontmatter delimiter")
	}
	rest := s[len(sep):]
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\n"+sep)
	if end < 0 {
		return nil, fmt.Errorf("missing closing --- frontmatter delimiter")
	}
	return []byte(rest[:end]), nil
}
