package harness

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestInjectOrUpdateMemorySection_NewFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	targetFile := "/home/user/.gemini/AGENTS.md"
	repoMemoryPath := "/home/user/.koharness/repo/memory/AGENTS.md"

	err := InjectOrUpdateMemorySection(fs, targetFile, repoMemoryPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := afero.ReadFile(fs, targetFile)
	if err != nil {
		t.Fatalf("failed reading target file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, MemoryStartMarker) || !strings.Contains(content, MemoryEndMarker) {
		t.Errorf("expected content to contain memory markers, got:\n%s", content)
	}
	if !strings.Contains(content, repoMemoryPath) {
		t.Errorf("expected content to contain repo memory path %s", repoMemoryPath)
	}
}

func TestInjectOrUpdateMemorySection_UpdateExistingBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	targetFile := "/home/user/.gemini/AGENTS.md"
	repoMemoryPath := "/home/user/.koharness/repo/memory/AGENTS.md"

	initialContent := `# User Customs
Do not use em-dashes.

` + MemoryStartMarker + `
Old Memory Content
` + MemoryEndMarker + `

# End Rules
`
	_ = fs.MkdirAll(filepath.Dir(targetFile), 0755)
	_ = afero.WriteFile(fs, targetFile, []byte(initialContent), 0644)

	err := InjectOrUpdateMemorySection(fs, targetFile, repoMemoryPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := afero.ReadFile(fs, targetFile)
	content := string(data)

	if !strings.Contains(content, "# User Customs") || !strings.Contains(content, "# End Rules") {
		t.Errorf("user custom content was preserved improperly:\n%s", content)
	}
	if strings.Contains(content, "Old Memory Content") {
		t.Errorf("old memory content was not replaced")
	}
	if !strings.Contains(content, repoMemoryPath) {
		t.Errorf("expected content to include updated path %s", repoMemoryPath)
	}
}
