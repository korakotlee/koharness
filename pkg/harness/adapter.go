// Package harness provides domain models, adapter interfaces, and a detection
// engine for discovering and interfacing with local AI client harnesses such as
// Google Antigravity, Claude Code, and OpenAI Codex.
package harness

// HarnessID uniquely identifies a supported AI client harness.
type HarnessID string

const (
	// HarnessAntigravity identifies the Google Antigravity harness environment.
	HarnessAntigravity HarnessID = "antigravity"
	// HarnessClaude identifies the Claude Code harness environment.
	HarnessClaude HarnessID = "claude"
	// HarnessCodex identifies the OpenAI Codex harness environment.
	HarnessCodex HarnessID = "codex"
)

// CapabilityType defines the classification of a harness feature payload.
type CapabilityType string

const (
	// CapabilitySkill represents a modular skill extension.
	CapabilitySkill CapabilityType = "skill"
	// CapabilityWorkflow represents an automated multi-step workflow.
	CapabilityWorkflow CapabilityType = "workflow"
	// CapabilityPrompt represents a reusable prompt template.
	CapabilityPrompt CapabilityType = "prompt"
	// CapabilityMCP represents a Model Context Protocol server configuration.
	CapabilityMCP CapabilityType = "mcp"
)

// Capability represents an installable or linkable harness capability asset.
type Capability struct {
	// Type specifies the category of the capability payload.
	Type CapabilityType
	// Name specifies the unique identifier or title of the capability.
	Name string
	// SourcePath specifies the absolute filesystem path to the capability payload.
	SourcePath string
}

// HarnessPaths contains the resolved filesystem locations for a specific harness.
type HarnessPaths struct {
	// ConfigDir is the primary configuration root directory.
	ConfigDir string
	// ConfigFile is the primary JSON or YAML configuration file, if applicable.
	ConfigFile string
	// SkillsDir is the directory where custom skills are installed.
	SkillsDir string
	// WorkflowsDir is the directory where custom workflows are installed.
	WorkflowsDir string
	// MCPDir is the directory or file location for MCP server definitions.
	MCPDir string
}

// HarnessStatus details the local installation footprint for a client harness.
type HarnessStatus struct {
	// ID is the unique identifier of the target harness.
	ID HarnessID
	// Name is the human-readable display name of the target harness.
	Name string
	// Installed indicates whether required configuration footprints exist on disk.
	Installed bool
	// PathsFound lists configuration paths that were successfully resolved.
	PathsFound []string
	// PathsMissing lists standard configuration paths that were absent on disk.
	PathsMissing []string
}

// HarnessAdapter defines the unified operational interface implemented by all
// supported client harness integration backends.
type HarnessAdapter interface {
	// ID returns the unique HarnessID for this adapter.
	ID() HarnessID
	// Name returns the human-readable title of the target client harness.
	Name() string
	// IsInstalled checks if the target harness is present on the workstation.
	IsInstalled() bool
	// GetConfigPaths returns the resolved absolute filesystem locations.
	GetConfigPaths() HarnessPaths
	// GetStatus compiles a comprehensive status snapshot for the target harness.
	GetStatus() HarnessStatus
	// LinkCapability establishes linkage for the provided capability asset into the target harness.
	LinkCapability(cap Capability) error
}
