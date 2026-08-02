package mcp

import (
	"testing"
)

func TestMaskSecrets(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-1234567890123456789012345", "sk-***"},
		{"ghp_1234567890123456789012345", "ghp_***"},
		{"eyJ1MjM0NTY3ODkwMTIzNDU2Nzg5MDE", "eyJ***"},
		{"plain_string", "plain_string"},
	}

	for _, tt := range tests {
		got := MaskSecrets(tt.input)
		if got != tt.expected {
			t.Errorf("MaskSecrets(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestValidateConfig_SecretMasking(t *testing.T) {
	configData := []byte(`{
		"mcpServers": {
			"api_server": {
				"command": "node",
				"env": {
					"API_KEY": "sk-123456789012345678901234567890"
				}
			}
		}
	}`)

	issues, err := ValidateConfig(configData)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}

	if len(issues) == 0 {
		t.Fatalf("Expected validation issues, got 0")
	}

	foundSecretIssue := false
	for _, issue := range issues {
		if issue.Type == IssuePossibleSecret {
			foundSecretIssue = true
			if issue.Value == "sk-123456789012345678901234567890" {
				t.Errorf("ValidationIssue.Value contains raw secret: %s", issue.Value)
			}
			if issue.Value != "sk-***" {
				t.Errorf("Expected masked value 'sk-***', got %q", issue.Value)
			}
		}
	}

	if !foundSecretIssue {
		t.Errorf("Expected IssuePossibleSecret to be reported")
	}
}
