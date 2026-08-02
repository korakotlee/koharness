package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/korakotlee/koharness/pkg/credentials"
	"github.com/korakotlee/koharness/pkg/mcp"
	"github.com/spf13/afero"
)

func TestEndToEndMCPMergingAndExpansion(t *testing.T) {
	fs := afero.NewMemMapFs()

	baseJSON := []byte(`{
		"mcpServers": {
			"shared-tool": {
				"command": "shared-tool-bin",
				"args": ["--mode", "shared"]
			},
			"local-tool": {
				"command": "default-local-bin"
			}
		}
	}`)

	overrideJSON := []byte(`{
		"mcpServers": {
			"local-tool": {
				"command": "${HOME}/.local/bin/local-tool-custom",
				"env": {
					"API_KEY": "${TEST_API_KEY}"
				}
			}
		}
	}`)

	// 1. Deep merge
	mergedBytes, err := mcp.MergeJSON(baseJSON, overrideJSON)
	if err != nil {
		t.Fatalf("failed merging MCP JSON: %v", err)
	}

	// 2. Expand tokens
	opts := mcp.EnvOptions{
		HomeDir: "/home/testuser",
		EnvMap: map[string]string{
			"TEST_API_KEY": "secret-key-12345",
		},
	}

	finalBytes, err := mcp.ExpandJSONTokens(mergedBytes, opts)
	if err != nil {
		t.Fatalf("failed expanding tokens in merged JSON: %v", err)
	}

	// 3. Write to active client target path
	targetPath := "/home/testuser/.gemini/config/mcp_config.json"
	if err := afero.WriteFile(fs, targetPath, finalBytes, 0644); err != nil {
		t.Fatalf("failed writing active client config target: %v", err)
	}

	// 4. Verify rendered target file contents
	readBytes, err := afero.ReadFile(fs, targetPath)
	if err != nil {
		t.Fatalf("failed reading back target file: %v", err)
	}

	var rendered map[string]interface{}
	if err := json.Unmarshal(readBytes, &rendered); err != nil {
		t.Fatalf("failed unmarshaling rendered target file: %v", err)
	}

	servers := rendered["mcpServers"].(map[string]interface{})
	sharedTool := servers["shared-tool"].(map[string]interface{})
	localTool := servers["local-tool"].(map[string]interface{})

	if sharedTool["command"] != "shared-tool-bin" {
		t.Errorf("expected shared-tool command to be preserved, got %v", sharedTool["command"])
	}

	if localTool["command"] != "/home/testuser/.local/bin/local-tool-custom" {
		t.Errorf("expected local-tool command to be expanded, got %v", localTool["command"])
	}

	localEnv := localTool["env"].(map[string]interface{})
	if localEnv["API_KEY"] != "secret-key-12345" {
		t.Errorf("expected expanded API_KEY, got %v", localEnv["API_KEY"])
	}
}

type mockResolver struct {
	resolved map[string]string
}

func (m *mockResolver) ProviderName() string {
	return "Mock Provider"
}

func (m *mockResolver) CanResolve(uri string) bool {
	return uri == "pass://Vault/Anthropic/api_key" || uri == "pass://Vault/Missing/key"
}

func (m *mockResolver) Resolve(ctx context.Context, uri string) (string, error) {
	if val, ok := m.resolved[uri]; ok {
		return val, nil
	}
	return "", errors.New("key not found in mock provider")
}

func TestEndToEndMCPCredentialInjectionAndFallback(t *testing.T) {
	templateJSON := []byte(`{
		"mcpServers": {
			"anthropic": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-anthropic"],
				"env": {
					"ANTHROPIC_API_KEY": "pass://Vault/Anthropic/api_key",
					"FALLBACK_KEY": "pass://Vault/Missing/key"
				}
			}
		}
	}`)

	mock := &mockResolver{
		resolved: map[string]string{
			"pass://Vault/Anthropic/api_key": "proton-pass-secret-token-9999",
		},
	}

	var warnings []string
	opts := mcp.EnvOptions{
		HomeDir:   "/home/testuser",
		Resolvers: []credentials.CredentialResolver{mock},
		WarningHandler: func(warning string) {
			warnings = append(warnings, warning)
		},
	}

	resultBytes, err := mcp.ExpandJSONTokens(templateJSON, opts)
	if err != nil {
		t.Fatalf("failed expanding credential tokens: %v", err)
	}

	var rendered map[string]interface{}
	if err := json.Unmarshal(resultBytes, &rendered); err != nil {
		t.Fatalf("failed unmarshaling expanded JSON: %v", err)
	}

	servers := rendered["mcpServers"].(map[string]interface{})
	anthropic := servers["anthropic"].(map[string]interface{})
	envMap := anthropic["env"].(map[string]interface{})

	if envMap["ANTHROPIC_API_KEY"] != "proton-pass-secret-token-9999" {
		t.Errorf("expected ANTHROPIC_API_KEY to be resolved, got %v", envMap["ANTHROPIC_API_KEY"])
	}

	if envMap["FALLBACK_KEY"] != "pass://Vault/Missing/key" {
		t.Errorf("expected missing key to retain original URI fallback, got %v", envMap["FALLBACK_KEY"])
	}

	if len(warnings) != 1 {
		t.Errorf("expected 1 resolution warning, got %d (%v)", len(warnings), warnings)
	}
}
