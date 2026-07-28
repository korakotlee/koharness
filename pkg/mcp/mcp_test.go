package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/korakotlee/koharness/pkg/mcp"
)

func TestMergeJSON(t *testing.T) {
	baseJSON := []byte(`{
		"mcpServers": {
			"server-a": {
				"command": "server-a-cmd",
				"args": ["--port", "8080"]
			},
			"server-b": {
				"command": "server-b-cmd"
			}
		}
	}`)

	overrideJSON := []byte(`{
		"mcpServers": {
			"server-a": {
				"command": "/custom/path/server-a-cmd"
			},
			"server-c": {
				"command": "server-c-cmd"
			}
		}
	}`)

	mergedBytes, err := mcp.MergeJSON(baseJSON, overrideJSON)
	if err != nil {
		t.Fatalf("unexpected MergeJSON error: %v", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(mergedBytes, &resultMap); err != nil {
		t.Fatalf("failed unmarshaling merged JSON: %v", err)
	}

	servers := resultMap["mcpServers"].(map[string]interface{})
	serverA := servers["server-a"].(map[string]interface{})
	serverB := servers["server-b"].(map[string]interface{})
	serverC := servers["server-c"].(map[string]interface{})

	if serverA["command"] != "/custom/path/server-a-cmd" {
		t.Errorf("expected server-a command override, got %v", serverA["command"])
	}
	if serverB["command"] != "server-b-cmd" {
		t.Errorf("expected server-b command preserved, got %v", serverB["command"])
	}
	if serverC["command"] != "server-c-cmd" {
		t.Errorf("expected server-c added, got %v", serverC["command"])
	}
}

func TestExpandJSONTokens(t *testing.T) {
	inputJSON := []byte(`{
		"mcpServers": {
			"my-server": {
				"command": "${HOME}/bin/server",
				"env": {
					"DATABASE_URL": "${DB_URL}",
					"STATIC_PATH": "~/data"
				}
			}
		}
	}`)

	opts := mcp.EnvOptions{
		HomeDir: "/Users/testuser",
		EnvMap: map[string]string{
			"DB_URL": "postgres://localhost:5432/mydb",
		},
	}

	expandedBytes, err := mcp.ExpandJSONTokens(inputJSON, opts)
	if err != nil {
		t.Fatalf("unexpected ExpandJSONTokens error: %v", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(expandedBytes, &res); err != nil {
		t.Fatalf("failed unmarshaling expanded JSON: %v", err)
	}

	servers := res["mcpServers"].(map[string]interface{})
	myServer := servers["my-server"].(map[string]interface{})
	cmd := myServer["command"].(string)
	env := myServer["env"].(map[string]interface{})

	if cmd != "/Users/testuser/bin/server" {
		t.Errorf("expected expanded command, got %s", cmd)
	}
	if env["DATABASE_URL"] != "postgres://localhost:5432/mydb" {
		t.Errorf("expected expanded DB_URL, got %v", env["DATABASE_URL"])
	}
	if env["STATIC_PATH"] != "/Users/testuser/data" {
		t.Errorf("expected expanded tilde path, got %v", env["STATIC_PATH"])
	}
}

func TestValidateConfig(t *testing.T) {
	inputJSON := []byte(`{
		"mcpServers": {
			"bad-server": {
				"command": "/Users/korakot/.local/bin/my-mcp",
				"env": {
					"api_key": "sk-proj-1234567890123456789012345"
				}
			}
		}
	}`)

	issues, err := mcp.ValidateConfig(inputJSON)
	if err != nil {
		t.Fatalf("unexpected ValidateConfig error: %v", err)
	}

	if len(issues) < 2 {
		t.Fatalf("expected at least 2 issues, got %d", len(issues))
	}

	hasNonPortable := false
	hasSecret := false
	for _, issue := range issues {
		if issue.Type == mcp.IssueNonPortablePath {
			hasNonPortable = true
		}
		if issue.Type == mcp.IssuePossibleSecret {
			hasSecret = true
		}
	}

	if !hasNonPortable {
		t.Errorf("expected IssueNonPortablePath warning")
	}
	if !hasSecret {
		t.Errorf("expected IssuePossibleSecret warning")
	}
}
