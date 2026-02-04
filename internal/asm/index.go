package asm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/linker"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

const defaultIndexFilename = "skills-index.md"

type indexRow struct {
	Name        string
	Title       string
	Description string
	Directory   string
}

type skillDoc struct {
	Title       string
	Description string
}

func Index(outputPath string) (IndexReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return IndexReport{}, err
	}

	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		outputPath = filepath.Join(state.Root, defaultIndexFilename)
	} else if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(state.Root, outputPath)
	}
	outputPath = filepath.Clean(outputPath)

	skills := make([]manifest.Skill, len(state.Config.Skills))
	copy(skills, state.Config.Skills)
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	rows := make([]indexRow, 0, len(skills))
	warnings := []string{}
	for _, skill := range skills {
		safeName, err := linker.SafeNamePath(skill.Name)
		if err != nil {
			return IndexReport{}, err
		}
		dirDisplay := filepath.ToSlash(filepath.Join("skills", safeName))
		if !strings.HasSuffix(dirDisplay, "/") {
			dirDisplay += "/"
		}

		doc, warning, err := loadSkillDoc(state, skill, safeName)
		if err != nil {
			return IndexReport{}, err
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
		if doc.Title == "" {
			doc.Title = skill.Name
		}

		rows = append(rows, indexRow{
			Name:        skill.Name,
			Title:       doc.Title,
			Description: doc.Description,
			Directory:   dirDisplay,
		})
	}

	content := renderIndexMarkdown(rows)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return IndexReport{}, err
	}
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return IndexReport{}, err
	}

	return IndexReport{OutputPath: outputPath, SkillCount: len(rows), Warnings: warnings}, nil
}

func loadSkillDoc(state manifest.State, skill manifest.Skill, safeName string) (skillDoc, string, error) {
	candidates := skillDocCandidates(state, skill, safeName)
	for _, candidate := range candidates {
		doc, err := readSkillDoc(candidate)
		if err == nil {
			return doc, "", nil
		}
		if !os.IsNotExist(err) {
			return skillDoc{}, "", err
		}
	}

	if len(candidates) == 0 {
		return skillDoc{}, "", nil
	}
	return skillDoc{}, fmt.Sprintf("skill %s missing SKILL.md; run asm install", skill.Name), nil
}

func skillDocCandidates(state manifest.State, skill manifest.Skill, safeName string) []string {
	seen := map[string]struct{}{}
	add := func(path string) []string {
		if path == "" {
			return nil
		}
		path = filepath.Clean(path)
		if _, ok := seen[path]; ok {
			return nil
		}
		seen[path] = struct{}{}
		return []string{path}
	}

	candidates := []string{}
	installed := filepath.Join(state.Paths.SkillsDir, safeName, "SKILL.md")
	candidates = append(candidates, add(installed)...)

	subdir := filepath.FromSlash(skill.Subdir)
	pathForBase := func(base string) string {
		if base == "" {
			return ""
		}
		if subdir != "" {
			base = filepath.Join(base, subdir)
		}
		return filepath.Join(base, "SKILL.md")
	}

	if skill.Version == "" {
		candidates = append(candidates, add(pathForBase(skill.Origin))...)
		return candidates
	}

	replacePath := ""
	if state.Config.Replace != nil {
		replacePath = state.Config.Replace[skill.Origin]
	}
	if replacePath != "" {
		candidates = append(candidates, add(pathForBase(replacePath))...)
	}
	storePath := gitstore.RepoPath(state.Paths.StoreDir, skill.Origin)
	candidates = append(candidates, add(pathForBase(storePath))...)
	return candidates
}

func readSkillDoc(path string) (skillDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return skillDoc{}, err
	}
	return parseSkillDoc(data), nil
}

func parseSkillDoc(data []byte) skillDoc {
	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	frontMatter, startIndex := parseFrontMatter(lines)
	if startIndex > len(lines) {
		startIndex = len(lines)
	}

	title := frontMatter.Title
	description := frontMatter.Description

	if title == "" {
		title = extractTitle(lines[startIndex:])
	}
	if description == "" {
		description = extractDescription(lines[startIndex:], title)
	}

	return skillDoc{
		Title:       title,
		Description: description,
	}
}

func parseFrontMatter(lines []string) (skillDoc, int) {
	if len(lines) == 0 {
		return skillDoc{}, 0
	}
	if strings.TrimSpace(lines[0]) != "---" {
		return skillDoc{}, 0
	}

	front := skillDoc{}
	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if line == "---" {
			return front, index + 1
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")
		switch key {
		case "title":
			if front.Title == "" {
				front.Title = value
			}
		case "description":
			if front.Description == "" {
				front.Description = value
			}
		}
	}

	return skillDoc{}, 0
}

func extractTitle(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			value := strings.TrimLeft(trimmed, "#")
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func extractDescription(lines []string, title string) string {
	startIndex := 0
	if title != "" {
		for index, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				value := strings.TrimLeft(trimmed, "#")
				value = strings.TrimSpace(value)
				if value != "" {
					startIndex = index + 1
					break
				}
			}
		}
	}

	paragraph := []string{}
	for _, line := range lines[startIndex:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(paragraph) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") && len(paragraph) == 0 {
			continue
		}
		paragraph = append(paragraph, trimmed)
	}

	return collapseWhitespace(strings.Join(paragraph, " "))
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func renderIndexMarkdown(rows []indexRow) string {
	var builder strings.Builder
	builder.WriteString("# Skills Index\n\n")
	builder.WriteString("| Name | Title | Description | Directory |\n")
	builder.WriteString("| --- | --- | --- | --- |\n")
	for _, row := range rows {
		builder.WriteString("| ")
		builder.WriteString(escapeTableCell(row.Name))
		builder.WriteString(" | ")
		builder.WriteString(escapeTableCell(row.Title))
		builder.WriteString(" | ")
		builder.WriteString(escapeTableCell(row.Description))
		builder.WriteString(" | ")
		builder.WriteString(escapeTableCell(row.Directory))
		builder.WriteString(" |\n")
	}

	content := builder.String()
	if content == "" || content[len(content)-1] != '\n' {
		content += "\n"
	}
	return content
}

func escapeTableCell(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.TrimSpace(value)
	return value
}
