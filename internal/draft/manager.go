package draft

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agent19710101/skillforge/internal/catalog"
)

type IDGenerator func() string

type Manager struct {
	RepoRoot     string
	WorkspaceDir string
	NewID        IDGenerator
}

type Workspace struct {
	RepoRoot   string    `json:"repoRoot"`
	Root       string    `json:"root"`
	BranchName string    `json:"branchName"`
	Operation  string    `json:"operation"`
	SkillName  string    `json:"skillName"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ValidationResult struct {
	Valid    bool               `json:"valid"`
	Scan     catalog.ScanResult `json:"scan"`
	Findings []catalog.Finding  `json:"findings,omitempty"`
}

func (m Manager) CreateWorkspace(operation, skillName string) (*Workspace, error) {
	repoRoot, err := filepath.Abs(m.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repo root: %w", err)
	}
	if err := validateOperation(operation); err != nil {
		return nil, err
	}
	if !catalog.IsCanonicalSkillName(skillName) {
		return nil, fmt.Errorf("invalid skill name %q", skillName)
	}

	id := m.nextID()
	branchName := fmt.Sprintf("skillforge/%s/%s/%s", operation, skillName, id)
	workspaceRoot, err := m.makeWorkspaceRoot(operation, skillName, id)
	if err != nil {
		return nil, err
	}
	if err := copyRepoTree(repoRoot, workspaceRoot); err != nil {
		return nil, fmt.Errorf("copy repository into workspace: %w", err)
	}

	return &Workspace{
		RepoRoot:   repoRoot,
		Root:       workspaceRoot,
		BranchName: branchName,
		Operation:  operation,
		SkillName:  skillName,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func (m Manager) nextID() string {
	if m.NewID != nil {
		if id := strings.TrimSpace(m.NewID()); id != "" {
			return id
		}
	}
	return fmt.Sprintf("%x", time.Now().UTC().UnixNano())
}

func (m Manager) makeWorkspaceRoot(operation, skillName, id string) (string, error) {
	prefix := fmt.Sprintf("skillforge-%s-%s-%s-", operation, skillName, id)
	if strings.TrimSpace(m.WorkspaceDir) == "" {
		return os.MkdirTemp("", prefix)
	}
	if err := os.MkdirAll(m.WorkspaceDir, 0o755); err != nil {
		return "", fmt.Errorf("create workspace dir: %w", err)
	}
	return os.MkdirTemp(m.WorkspaceDir, prefix)
}

func (w *Workspace) CreateSkill(content string) error {
	path := w.skillPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("skill %q already exists", w.SkillName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat skill path: %w", err)
	}
	return w.writeSkillFile(path, content)
}

func (w *Workspace) UpdateSkill(content string) error {
	path := w.skillPath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q does not exist", w.SkillName)
		}
		return fmt.Errorf("stat skill path: %w", err)
	}
	return w.writeSkillFile(path, content)
}

func (w *Workspace) DeleteSkill() error {
	dir := filepath.Dir(w.skillPath())
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q does not exist", w.SkillName)
		}
		return fmt.Errorf("stat skill dir: %w", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove skill dir: %w", err)
	}
	return nil
}

func (w *Workspace) Validate() (ValidationResult, error) {
	scan, err := catalog.Scanner{Root: w.Root}.Scan()
	if err != nil {
		return ValidationResult{}, fmt.Errorf("scan workspace: %w", err)
	}
	findings := make([]catalog.Finding, 0)
	for _, skill := range scan.Skills {
		if skill.Valid {
			continue
		}
		findings = append(findings, skill.Findings...)
	}
	return ValidationResult{
		Valid:    scan.ErrorCount == 0,
		Scan:     scan,
		Findings: findings,
	}, nil
}

func (w *Workspace) skillPath() string {
	return filepath.Join(w.Root, "skills", w.SkillName, "SKILL.md")
}

func (w *Workspace) writeSkillFile(path, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("skill content must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write skill file: %w", err)
	}
	return nil
}

func validateOperation(operation string) error {
	switch strings.TrimSpace(operation) {
	case "create", "update", "delete":
		return nil
	default:
		return fmt.Errorf("unsupported draft operation %q", operation)
	}
}

func copyRepoTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == ".git" {
			return filepath.SkipDir
		}
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
