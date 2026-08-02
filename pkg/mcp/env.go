package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/korakotlee/koharness/pkg/credentials"
)

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)

// EnvOptions configures options for environment token expansion and external credential injection.
type EnvOptions struct {
	// HomeDir specifies an explicit home directory path override for tilde (~) expansion.
	HomeDir string
	// EnvMap provides explicit key-value environment variable overrides.
	EnvMap map[string]string
	// Resolvers specifies external credential resolvers to expand secret reference URIs (e.g. pass://).
	Resolvers []credentials.CredentialResolver
	// WarningHandler receives non-fatal diagnostic warning messages during secret resolution.
	WarningHandler func(warning string)
}

// ExpandJSONTokens parses JSON bytes and replaces all string values containing ${VAR} or ~
// with resolved environment variable values and home directory paths.
func ExpandJSONTokens(data []byte, opts ...EnvOptions) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var opt EnvOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.HomeDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			opt.HomeDir = h
		}
	}

	var val interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return nil, err
	}

	expandedVal := expandValue(val, opt)
	return json.MarshalIndent(expandedVal, "", "  ")
}

// ExpandString performs token and tilde expansion on a single input string.
func ExpandString(s string, opt EnvOptions) string {
	if opt.HomeDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			opt.HomeDir = h
		}
	}

	// Default HOME in env map to opt.HomeDir if specified
	if opt.HomeDir != "" {
		if opt.EnvMap == nil {
			opt.EnvMap = make(map[string]string)
		}
		if _, ok := opt.EnvMap["HOME"]; !ok {
			opt.EnvMap["HOME"] = opt.HomeDir
		}
	}

	// Expand ~ at start of string or ~/
	if opt.HomeDir != "" {
		if s == "~" {
			s = opt.HomeDir
		} else if strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
			s = filepath.Join(opt.HomeDir, s[2:])
		}
	}

	// Expand ${VAR}
	s = envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		if opt.EnvMap != nil {
			if val, ok := opt.EnvMap[varName]; ok {
				return val
			}
		}
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Retain token if unresolvable
	})

	// Resolve secret references via registered credential resolvers
	for _, resolver := range opt.Resolvers {
		if resolver != nil && resolver.CanResolve(s) {
			resolvedSecret, err := resolver.Resolve(context.Background(), s)
			if err == nil {
				return resolvedSecret
			}
			if opt.WarningHandler != nil {
				opt.WarningHandler(fmt.Sprintf("failed to resolve secret reference %q via %s: %v", s, resolver.ProviderName(), err))
			}
			return s
		}
	}

	return s
}

func expandValue(val interface{}, opt EnvOptions) interface{} {
	switch v := val.(type) {
	case string:
		return ExpandString(v, opt)
	case map[string]interface{}:
		res := make(map[string]interface{}, len(v))
		for k, child := range v {
			res[k] = expandValue(child, opt)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, child := range v {
			res[i] = expandValue(child, opt)
		}
		return res
	default:
		return val
	}
}
