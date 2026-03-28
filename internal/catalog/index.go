package catalog

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ListOptions struct {
	Validation string
	Offset     int
	Limit      int
}

type IndexStatus struct {
	Root       string    `json:"root"`
	Revision   string    `json:"revision,omitempty"`
	ScannedAt  time.Time `json:"scannedAt"`
	SkillCount int       `json:"skillCount"`
	ValidCount int       `json:"validCount"`
	ErrorCount int       `json:"errorCount"`
}

type Index struct {
	result   ScanResult
	revision string
	byName   map[string]SkillRecord
}

func BuildIndex(root string) (*Index, error) {
	result, err := Scanner{Root: root}.Scan()
	if err != nil {
		return nil, err
	}
	return NewIndex(result, gitRevision(result.Root)), nil
}

func NewIndex(result ScanResult, revision string) *Index {
	byName := make(map[string]SkillRecord, len(result.Skills))
	for _, skill := range result.Skills {
		if skill.Name == "" {
			continue
		}
		byName[skill.Name] = skill
	}
	return &Index{result: result, revision: revision, byName: byName}
}

func (i *Index) List(opts ListOptions) []SkillRecord {
	filtered := make([]SkillRecord, 0, len(i.result.Skills))
	for _, skill := range i.result.Skills {
		if !matchesValidation(skill, opts.Validation) {
			continue
		}
		filtered = append(filtered, skill)
	}

	if opts.Offset < 0 {
		opts.Offset = 0
	}
	if opts.Offset >= len(filtered) {
		return []SkillRecord{}
	}
	end := len(filtered)
	if opts.Limit > 0 && opts.Offset+opts.Limit < end {
		end = opts.Offset + opts.Limit
	}
	return filtered[opts.Offset:end]
}

func (i *Index) Total(opts ListOptions) int {
	total := 0
	for _, skill := range i.result.Skills {
		if matchesValidation(skill, opts.Validation) {
			total++
		}
	}
	return total
}

func (i *Index) Get(name string) (SkillRecord, bool) {
	skill, ok := i.byName[name]
	return skill, ok
}

func (i *Index) Status() IndexStatus {
	return IndexStatus{
		Root:       i.result.Root,
		Revision:   i.revision,
		ScannedAt:  i.result.ScannedAt,
		SkillCount: len(i.result.Skills),
		ValidCount: i.result.ValidCount,
		ErrorCount: i.result.ErrorCount,
	}
}

func matchesValidation(skill SkillRecord, filter string) bool {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "", "all":
		return true
	case "valid":
		return skill.Valid
	case "invalid":
		return !skill.Valid
	default:
		return true
	}
}

func gitRevision(root string) string {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ParseListOptions(validation, offset, limit string) (ListOptions, error) {
	opts := ListOptions{Validation: validation}
	if offset != "" {
		if _, err := fmt.Sscanf(offset, "%d", &opts.Offset); err != nil || opts.Offset < 0 {
			return ListOptions{}, fmt.Errorf("invalid offset %q", offset)
		}
	}
	if limit != "" {
		if _, err := fmt.Sscanf(limit, "%d", &opts.Limit); err != nil || opts.Limit < 0 {
			return ListOptions{}, fmt.Errorf("invalid limit %q", limit)
		}
	}
	return opts, nil
}
