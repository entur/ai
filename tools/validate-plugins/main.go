// validate-plugins checks marketplace.json, every plugin.json, hooks.json,
// and SKILL.md frontmatter against the documented Claude Code plugin schema
// (https://code.claude.com/docs/en/plugins-reference).
//
// Run from anywhere in this repository:
//
//	cd tools/validate-plugins && go run .
//
// Exits non-zero on any validation error.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const marketplacePath = ".claude-plugin/marketplace.json"

var (
	skillNameRE        = regexp.MustCompile(`^[a-z0-9-]+$`)
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
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	UserConfig  map[string]userConfig `json:"userConfig"`
	LSPServers  json.RawMessage       `json:"lspServers"`
	Hooks       json.RawMessage       `json:"hooks"`
}

type codexPluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Skills      string `json:"skills"`
}

type hooksConfig struct {
	Hooks map[string][]hookMatcherGroup `json:"hooks"`
}

type hookMatcherGroup struct {
	Matcher string       `json:"matcher"`
	Hooks   []hookAction `json:"hooks"`
}

type hookAction struct {
	Type    string `json:"type"`
	Command string `json:"command"`
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

	mktPath, err := findMarketplacePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(2)
	}

	mkt, err := loadMarketplace(mktPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("Marketplace %q: %d plugins\n", mkt.Name, len(mkt.Plugins))
	repoRoot := filepath.Dir(filepath.Dir(mktPath))

	for _, ref := range mkt.Plugins {
		validatePlugin(repoRoot, ref, r)
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

func findMarketplacePath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	for {
		candidate := filepath.Join(dir, marketplacePath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	return "", fmt.Errorf("could not find %s in current directory or any parent", marketplacePath)
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

func validatePlugin(repoRoot string, ref marketplaceRef, r *result) {
	prefix := "plugin " + ref.Name
	if ref.Source == "" {
		r.errf("%s: missing source", prefix)
		return
	}
	root, ok := resolvePluginPath(repoRoot, ref.Source)
	if !ok {
		r.errf("%s: invalid source path %q (must be local and contained under marketplace dir)", prefix, ref.Source)
		return
	}

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

	validateCodexManifest(root, prefix, ref, r)

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

func validateCodexManifest(root, prefix string, ref marketplaceRef, r *result) {
	path := filepath.Join(root, ".codex-plugin", "plugin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		r.errf("%s: read %s: %v", prefix, path, err)
		return
	}

	var m codexPluginManifest
	if err := json.Unmarshal(data, &m); err != nil {
		r.errf("%s: parse %s: %v", prefix, path, err)
		return
	}

	if m.Name == "" {
		r.errf("%s: %s missing name", prefix, path)
	} else if m.Name != ref.Name {
		r.errf("%s: marketplace name %q != codex plugin name %q", prefix, ref.Name, m.Name)
	}
	if m.Version == "" {
		r.errf("%s: %s missing version", prefix, path)
	}
	if m.Description == "" {
		r.errf("%s: %s missing description", prefix, path)
	}
	if m.Skills == "" {
		r.errf("%s: %s missing skills path", prefix, path)
		return
	}

	skillsPath, ok := resolvePluginPath(root, m.Skills)
	if !ok {
		r.errf("%s: invalid skills path %q (must be local, start with ./, and stay under plugin root)", prefix, m.Skills)
		return
	}
	info, err := os.Stat(skillsPath)
	if err != nil {
		r.errf("%s: skills path %q does not exist: %v", prefix, m.Skills, err)
		return
	}
	if !info.IsDir() {
		r.errf("%s: skills path %q is not a directory", prefix, m.Skills)
	}
}

func validateHooks(root, prefix string, r *result) {
	hooksPath := filepath.Join(root, "hooks", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		r.errf("%s: read %s: %v", prefix, hooksPath, err)
		return
	}
	var cfg hooksConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		r.errf("%s: parse %s: %v", prefix, hooksPath, err)
		return
	}
	if len(cfg.Hooks) == 0 {
		r.errf("%s: %s missing hooks object or contains no hook events", prefix, hooksPath)
		return
	}

	for event, groups := range cfg.Hooks {
		if len(groups) == 0 {
			r.errf("%s: %s event %q has no matcher groups", prefix, hooksPath, event)
			continue
		}
		for i, group := range groups {
			if len(group.Hooks) == 0 {
				r.errf("%s: %s event %q group %d has no hooks", prefix, hooksPath, event, i)
				continue
			}
			for j, hook := range group.Hooks {
				if hook.Type == "" {
					r.errf("%s: %s event %q group %d hook %d missing type", prefix, hooksPath, event, i, j)
					continue
				}
				if hook.Type != "command" {
					continue
				}
				if hook.Command == "" {
					r.errf("%s: %s event %q group %d hook %d missing command", prefix, hooksPath, event, i, j)
					continue
				}
				const pluginRootPrefix = "${CLAUDE_PLUGIN_ROOT}/"
				if tail, ok := strings.CutPrefix(hook.Command, pluginRootPrefix); ok {
					rel := "./" + tail
					scriptPath, ok := resolvePluginPath(root, rel)
					if !ok {
						r.errf("%s: %s event %q group %d hook %d has invalid command path %q", prefix, hooksPath, event, i, j, hook.Command)
						continue
					}
					info, err := os.Stat(scriptPath)
					if err != nil {
						r.errf("%s: %s event %q group %d hook %d command target does not exist: %q", prefix, hooksPath, event, i, j, hook.Command)
						continue
					}
					if info.IsDir() {
						r.errf("%s: %s event %q group %d hook %d command target is a directory: %q", prefix, hooksPath, event, i, j, hook.Command)
					}
				}
			}
		}
	}
}

func resolvePluginPath(baseDir, relPath string) (string, bool) {
	if !strings.HasPrefix(relPath, "./") {
		return "", false
	}
	cleanRel := filepath.Clean(strings.TrimPrefix(relPath, "./"))
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", false
	}
	path := filepath.Join(baseDir, cleanRel)
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return path, true
}

func validateSkills(root, prefix string, r *result) {
	skillsDir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if errors.Is(err, os.ErrNotExist) {
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
	if errors.Is(err, os.ErrNotExist) {
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
	body, _, found := strings.Cut(rest, "\n"+sep)
	if !found {
		return nil, fmt.Errorf("missing closing --- frontmatter delimiter")
	}
	return []byte(body), nil
}
