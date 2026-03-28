package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/agent19710101/skillforge/internal/catalog"
	"github.com/agent19710101/skillforge/internal/client"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	ctx := context.Background()
	switch args[0] {
	case "list":
		if err := runList(ctx, args[1:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "search":
		if err := runSearch(ctx, args[1:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "get":
		if err := runGet(ctx, args[1:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "draft":
		if err := runDraft(ctx, args[1:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runList(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	validation := fs.String("validation", "", "validation filter: all, valid, invalid")
	offset := fs.Int("offset", 0, "pagination offset")
	limit := fs.Int("limit", 0, "page size")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("list does not accept positional arguments")
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	resp, err := c.ListSkills(ctx, client.ListOptions{Validation: *validation, Offset: *offset, Limit: *limit})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, resp)
	}
	writeSkillTable(stdout, resp.Skills)
	fmt.Fprintf(stdout, "\n%d skill(s) shown (%d total)\n", len(resp.Skills), resp.Total)
	return nil
}

func runSearch(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: skillforge search [--server URL] [--json] <query>")
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	resp, err := c.SearchSkills(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, resp)
	}
	fmt.Fprintf(stdout, "Query: %s\n\n", resp.Query)
	writeSkillTable(stdout, resp.Skills)
	fmt.Fprintf(stdout, "\n%d match(es)\n", resp.Total)
	return nil
}

func runGet(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: skillforge get [--server URL] [--json] <skill-name>")
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	skill, err := c.GetSkill(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, skill)
	}
	writeSkillDetail(stdout, skill)
	return nil
}

func runDraft(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: skillforge draft <create|status|submit> [...]")
	}

	switch args[0] {
	case "create":
		return runDraftCreate(ctx, args[1:], stdout, stderr)
	case "status":
		return runDraftStatus(ctx, args[1:], stdout, stderr)
	case "submit":
		return runDraftSubmit(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown draft command %q", args[0])
	}
}

func runDraftCreate(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("draft create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	operation := fs.String("operation", "", "draft operation: create, update, delete")
	skillName := fs.String("skill", "", "skill name")
	content := fs.String("content", "", "draft content")
	filePath := fs.String("file", "", "path to draft content file or - for stdin")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: skillforge draft create [--server URL] [--json] --operation <create|update|delete> --skill <name> [--content TEXT | --file PATH]")
	}

	req, err := buildDraftCreateRequest(*operation, *skillName, *content, *filePath)
	if err != nil {
		return err
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	resp, err := c.CreateDraft(ctx, req)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, resp)
	}
	writeDraftDetail(stdout, resp)
	return nil
}

func runDraftStatus(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("draft status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: skillforge draft status [--server URL] [--json] <draft-id>")
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	resp, err := c.GetDraft(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, resp)
	}
	writeDraftDetail(stdout, resp)
	return nil
}

func runDraftSubmit(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("draft submit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	serverURL := fs.String("server", defaultServerURL, "Skillforge API base URL")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: skillforge draft submit [--server URL] [--json] <draft-id>")
	}

	c, err := client.New(*serverURL)
	if err != nil {
		return err
	}
	resp, err := c.SubmitDraft(ctx, fs.Arg(0))
	if err != nil {
		var apiErr *client.APIError
		if !*jsonOutput && errors.As(err, &apiErr) {
			writeDraftSubmissionError(stdout, apiErr)
		}
		return err
	}
	if *jsonOutput {
		return writeJSON(stdout, resp)
	}
	writeDraftSubmissionResult(stdout, resp)
	return nil
}

func buildDraftCreateRequest(operation, skillName, inlineContent, filePath string) (client.DraftCreateRequest, error) {
	req := client.DraftCreateRequest{
		Operation: strings.TrimSpace(operation),
		SkillName: strings.TrimSpace(skillName),
	}
	switch {
	case strings.TrimSpace(filePath) != "" && strings.TrimSpace(inlineContent) != "":
		return client.DraftCreateRequest{}, fmt.Errorf("only one of --content or --file may be set")
	case strings.TrimSpace(filePath) != "":
		content, err := readDraftContent(filePath)
		if err != nil {
			return client.DraftCreateRequest{}, err
		}
		req.Content = content
	default:
		req.Content = inlineContent
	}
	return req, nil
}

func readDraftContent(path string) (string, error) {
	var (
		data []byte
		err  error
	)
	if strings.TrimSpace(path) == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", fmt.Errorf("read draft content: %w", err)
	}
	return string(data), nil
}

func writeJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func writeSkillTable(w io.Writer, skills []catalog.SkillRecord) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tVALID\tDESCRIPTION")
	for _, skill := range skills {
		validity := "invalid"
		if skill.Valid {
			validity = "valid"
		}
		description := strings.TrimSpace(skill.Description)
		if description == "" {
			description = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", skill.Name, validity, description)
	}
	_ = tw.Flush()
}

func writeSkillDetail(w io.Writer, skill catalog.SkillRecord) {
	fmt.Fprintf(w, "Name: %s\n", skill.Name)
	fmt.Fprintf(w, "Path: %s\n", skill.Path)
	fmt.Fprintf(w, "Valid: %t\n", skill.Valid)
	if skill.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", skill.Description)
	}
	if skill.License != "" {
		fmt.Fprintf(w, "License: %s\n", skill.License)
	}
	if len(skill.Tags) > 0 {
		tags := append([]string(nil), skill.Tags...)
		sort.Strings(tags)
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(tags, ", "))
	}
	if len(skill.Compatibility) > 0 {
		fmt.Fprintf(w, "Compatibility: %s\n", strings.Join(skill.Compatibility, ", "))
	}
	if len(skill.AllowedTools) > 0 {
		fmt.Fprintf(w, "Allowed tools: %s\n", strings.Join(skill.AllowedTools, ", "))
	}
	if len(skill.Findings) > 0 {
		fmt.Fprintln(w, "Findings:")
		for _, finding := range skill.Findings {
			if finding.Path != "" {
				fmt.Fprintf(w, "- [%s] %s (%s)\n", finding.Code, finding.Message, finding.Path)
				continue
			}
			fmt.Fprintf(w, "- [%s] %s\n", finding.Code, finding.Message)
		}
	}
	if body := strings.TrimSpace(skill.Body); body != "" {
		fmt.Fprintf(w, "\n%s\n", body)
	}
}

func writeDraftDetail(w io.Writer, draft client.DraftResponse) {
	fmt.Fprintf(w, "Draft: %s\n", draft.ID)
	fmt.Fprintf(w, "Operation: %s\n", draft.Operation)
	fmt.Fprintf(w, "Skill: %s\n", draft.SkillName)
	fmt.Fprintf(w, "Branch: %s\n", draft.BranchName)
	if createdAt := formatTimestamp(draft.CreatedAt); createdAt != "" {
		fmt.Fprintf(w, "Created: %s\n", createdAt)
	}
	writeDraftValidation(w, draft.Validation)
	writeDraftSubmissionStatus(w, draft.Submission)
}

func writeDraftSubmissionResult(w io.Writer, result client.DraftSubmissionResponse) {
	fmt.Fprintf(w, "Draft: %s\n", result.ID)
	fmt.Fprintf(w, "Submitted: yes\n")
	fmt.Fprintf(w, "Operation: %s\n", result.Operation)
	fmt.Fprintf(w, "Skill: %s\n", result.SkillName)
	fmt.Fprintf(w, "Branch: %s\n", result.BranchName)
	fmt.Fprintf(w, "Base branch: %s\n", result.BaseBranch)
	if commitHash := strings.TrimSpace(result.CommitHash); commitHash != "" {
		fmt.Fprintf(w, "Commit: %s\n", commitHash)
	}
	if result.PullRequest != nil {
		if result.PullRequest.Number != 0 {
			fmt.Fprintf(w, "Pull request: #%d\n", result.PullRequest.Number)
		}
		if result.PullRequest.URL != "" {
			fmt.Fprintf(w, "Pull request URL: %s\n", result.PullRequest.URL)
		}
	}
	writeDraftValidation(w, result.Validation)
}

func writeDraftSubmissionError(w io.Writer, apiErr *client.APIError) {
	switch apiErr.Code {
	case "submission_unavailable":
		fmt.Fprintln(w, "Submission unavailable")
		if apiErr.Message != "" {
			fmt.Fprintf(w, "Reason: %s\n", apiErr.Message)
		}
		if apiErr.Submission != nil {
			writeDraftSubmissionStatus(w, *apiErr.Submission)
		}
	case "draft_invalid":
		fmt.Fprintln(w, "Submission blocked")
		if apiErr.Message != "" {
			fmt.Fprintf(w, "Reason: %s\n", apiErr.Message)
		}
		if apiErr.Validation != nil {
			writeDraftValidation(w, *apiErr.Validation)
		}
		if apiErr.Submission != nil {
			writeDraftSubmissionStatus(w, *apiErr.Submission)
		}
	}
}

func writeDraftValidation(w io.Writer, validation client.DraftValidation) {
	fmt.Fprintf(w, "Valid: %t\n", validation.Valid)
	if len(validation.Findings) == 0 {
		return
	}
	fmt.Fprintln(w, "Findings:")
	for _, finding := range validation.Findings {
		if finding.Path != "" {
			fmt.Fprintf(w, "- [%s] %s (%s)\n", finding.Code, finding.Message, finding.Path)
			continue
		}
		fmt.Fprintf(w, "- [%s] %s\n", finding.Code, finding.Message)
	}
}

func writeDraftSubmissionStatus(w io.Writer, status client.DraftSubmissionStatus) {
	fmt.Fprintf(w, "Submission enabled: %t\n", status.Enabled)
	if status.BaseBranch != "" {
		fmt.Fprintf(w, "Submission base branch: %s\n", status.BaseBranch)
	}
	if status.Reason != "" {
		fmt.Fprintf(w, "Submission reason: %s\n", status.Reason)
	}
}

func formatTimestamp(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return ts.Format(time.RFC3339)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  skillforge list [--server URL] [--validation FILTER] [--offset N] [--limit N] [--json]")
	fmt.Fprintln(w, "  skillforge search [--server URL] [--json] <query>")
	fmt.Fprintln(w, "  skillforge get [--server URL] [--json] <skill-name>")
	fmt.Fprintln(w, "  skillforge draft create [--server URL] [--json] --operation <create|update|delete> --skill <name> [--content TEXT | --file PATH]")
	fmt.Fprintln(w, "  skillforge draft status [--server URL] [--json] <draft-id>")
	fmt.Fprintln(w, "  skillforge draft submit [--server URL] [--json] <draft-id>")
}
