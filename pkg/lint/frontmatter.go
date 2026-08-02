package lint

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillFrontmatter defines the expected YAML frontmatter structure for SKILL.md files.
type SkillFrontmatter struct {
	// Name is the unique kebab-case or descriptive name of the skill.
	Name string `yaml:"name"`
	// Description explains the skill's purpose and usage triggers.
	Description string `yaml:"description"`
}

// ValidateSkillFrontmatter inspects all SKILL.md files inside skills/*/ and verifies frontmatter metadata.
func ValidateSkillFrontmatter(repoRoot string) ([]LintIssue, error) {
	var issues []LintIssue

	skillsDir := filepath.Join(repoRoot, "skills")
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return nil, nil
	}

	err = filepath.Walk(skillsDir, func(path string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fileInfo.IsDir() {
			return nil
		}

		if filepath.Base(path) == "SKILL.md" {
			relPath, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				relPath = path
			}

			if issue := validateSingleSkillFrontmatter(path, relPath); issue != nil {
				issues = append(issues, *issue)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking skills directory for SKILL.md: %w", err)
	}

	return issues, nil
}

func validateSingleSkillFrontmatter(fullPath, displayPath string) *LintIssue {
	file, err := os.Open(fullPath)
	if err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  fmt.Sprintf("unable to open file: %v", err),
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  fmt.Sprintf("error reading file: %v", err),
		}
	}

	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  "missing YAML frontmatter starting delimiter ('---')",
		}
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  "unclosed YAML frontmatter (missing closing '---' delimiter)",
		}
	}

	frontmatterText := strings.Join(lines[1:endIdx], "\n")
	var meta SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterText), &meta); err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  fmt.Sprintf("invalid YAML in frontmatter: %v", err),
		}
	}

	if strings.TrimSpace(meta.Name) == "" {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  "frontmatter is missing required 'name' field",
		}
	}

	if strings.TrimSpace(meta.Description) == "" {
		return &LintIssue{
			Path:     displayPath,
			Category: "Skill Metadata",
			Message:  "frontmatter is missing required 'description' field",
		}
	}

	return nil
}
