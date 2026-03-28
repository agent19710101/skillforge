package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

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

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  skillforge list [--server URL] [--validation FILTER] [--offset N] [--limit N] [--json]")
	fmt.Fprintln(w, "  skillforge search [--server URL] [--json] <query>")
	fmt.Fprintln(w, "  skillforge get [--server URL] [--json] <skill-name>")
}
