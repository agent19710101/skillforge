package catalog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var canonicalSkillNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Finding struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type SkillRecord struct {
	Name          string         `json:"name,omitempty"`
	Path          string         `json:"path"`
	Description   string         `json:"description,omitempty"`
	License       string         `json:"license,omitempty"`
	Compatibility []string       `json:"compatibility,omitempty"`
	AllowedTools  []string       `json:"allowedTools,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Body          string         `json:"body,omitempty"`
	Valid         bool           `json:"valid"`
	Findings      []Finding      `json:"findings,omitempty"`
}

type ScanResult struct {
	Root       string        `json:"root"`
	ScannedAt  time.Time     `json:"scannedAt"`
	Skills     []SkillRecord `json:"skills"`
	ValidCount int           `json:"validCount"`
	ErrorCount int           `json:"errorCount"`
}

type Scanner struct {
	Root string
}

func (s Scanner) Scan() (ScanResult, error) {
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("resolve root: %w", err)
	}

	skillsRoot := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ScanResult{Root: root, ScannedAt: time.Now().UTC(), Skills: []SkillRecord{}}, nil
		}
		return ScanResult{}, fmt.Errorf("read skills root: %w", err)
	}

	result := ScanResult{Root: root, ScannedAt: time.Now().UTC()}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record := SkillRecord{Path: filepath.ToSlash(filepath.Join("skills", entry.Name()))}
		skillDir := filepath.Join(skillsRoot, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		if !canonicalSkillNameRE.MatchString(entry.Name()) {
			record.Findings = append(record.Findings, Finding{
				Code:    "invalid_directory_name",
				Message: fmt.Sprintf("skill directory %q must match %s", entry.Name(), canonicalSkillNameRE.String()),
				Path:    filepath.ToSlash(filepath.Join("skills", entry.Name())),
			})
		}

		content, readErr := os.ReadFile(skillFile)
		if readErr != nil {
			if errors.Is(readErr, fs.ErrNotExist) {
				record.Name = entry.Name()
				record.Findings = append(record.Findings, Finding{
					Code:    "missing_skill_md",
					Message: "skill directory is missing SKILL.md",
					Path:    filepath.ToSlash(filepath.Join("skills", entry.Name(), "SKILL.md")),
				})
				record.Valid = false
				result.Skills = append(result.Skills, record)
				continue
			}
			return ScanResult{}, fmt.Errorf("read %s: %w", skillFile, readErr)
		}

		record.Path = filepath.ToSlash(filepath.Join("skills", entry.Name(), "SKILL.md"))
		parsed, parseErr := parseSkillMarkdown(string(content))
		if parseErr != nil {
			record.Name = entry.Name()
			record.Findings = append(record.Findings, Finding{
				Code:    "invalid_frontmatter",
				Message: parseErr.Error(),
				Path:    record.Path,
			})
			record.Valid = false
			result.Skills = append(result.Skills, record)
			continue
		}

		record.Name = parsed.Name
		record.Description = parsed.Description
		record.License = parsed.License
		record.Compatibility = parsed.Compatibility
		record.AllowedTools = parsed.AllowedTools
		record.Metadata = parsed.Metadata
		record.Tags = parsed.Tags
		record.Body = parsed.Body

		if parsed.Name == "" {
			record.Findings = append(record.Findings, Finding{
				Code:    "missing_name",
				Message: "frontmatter must define name",
				Path:    record.Path,
			})
		}
		if parsed.Name != "" && parsed.Name != entry.Name() {
			record.Findings = append(record.Findings, Finding{
				Code:    "name_directory_mismatch",
				Message: fmt.Sprintf("frontmatter name %q does not match directory %q", parsed.Name, entry.Name()),
				Path:    record.Path,
			})
		}
		record.Valid = len(record.Findings) == 0
		result.Skills = append(result.Skills, record)
	}

	sort.Slice(result.Skills, func(i, j int) bool {
		return result.Skills[i].Path < result.Skills[j].Path
	})
	for _, skill := range result.Skills {
		if skill.Valid {
			result.ValidCount++
		} else {
			result.ErrorCount += len(skill.Findings)
		}
	}
	return result, nil
}

type parsedSkill struct {
	Name          string
	Description   string
	License       string
	Compatibility []string
	AllowedTools  []string
	Tags          []string
	Metadata      map[string]any
	Body          string
}

func parseSkillMarkdown(content string) (parsedSkill, error) {
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return parsedSkill{}, err
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &raw); err != nil {
		return parsedSkill{}, fmt.Errorf("parse YAML frontmatter: %w", err)
	}

	metadata := mapStringAny(raw["metadata"])
	tags := uniqueSortedStrings(append(stringSlice(raw["tags"]), stringSlice(metadata["tags"])...))

	return parsedSkill{
		Name:          stringValue(raw["name"]),
		Description:   stringValue(raw["description"]),
		License:       stringValue(raw["license"]),
		Compatibility: uniqueSortedStrings(stringSlice(raw["compatibility"])),
		AllowedTools:  uniqueSortedStrings(stringSlice(raw["allowed-tools"])),
		Tags:          tags,
		Metadata:      metadata,
		Body:          strings.TrimSpace(body),
	}, nil
}

func splitFrontmatter(content string) (string, string, error) {
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", "", errors.New("missing YAML frontmatter header")
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	parts := strings.SplitN(normalized, "\n---\n", 2)
	if len(parts) != 2 {
		return "", "", errors.New("missing YAML frontmatter terminator")
	}
	return strings.TrimPrefix(parts[0], "---\n"), parts[1], nil
}

func stringValue(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func stringSlice(v any) []string {
	switch typed := v.(type) {
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if s := stringValue(item); s != "" {
				items = append(items, s)
			}
		}
		return items
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				items = append(items, trimmed)
			}
		}
		return items
	case string:
		if trimmed := strings.TrimSpace(typed); trimmed != "" {
			return []string{trimmed}
		}
	}
	return nil
}

func mapStringAny(v any) map[string]any {
	if v == nil {
		return nil
	}
	if typed, ok := v.(map[string]any); ok {
		return typed
	}
	if typed, ok := v.(map[any]any); ok {
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			keyString, ok := key.(string)
			if !ok {
				continue
			}
			out[keyString] = value
		}
		return out
	}
	return nil
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	slices.Sort(result)
	return result
}
