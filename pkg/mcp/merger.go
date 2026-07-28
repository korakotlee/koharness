// Package mcp provides engines for deep JSON configuration merging, dynamic
// environment variable expansion, and secret validation for Model Context Protocol tools.
package mcp

import (
	"encoding/json"
	"fmt"
)

// MergeJSON performs a deep merge of overrideJSON over baseJSON.
// Properties defined in overrideJSON take precedence over properties in baseJSON.
// Both baseJSON and overrideJSON must be valid JSON objects. If overrideJSON is empty
// or contains empty bytes, baseJSON is returned intact.
func MergeJSON(baseJSON, overrideJSON []byte) ([]byte, error) {
	if len(baseJSON) == 0 {
		return overrideJSON, nil
	}
	if len(overrideJSON) == 0 {
		return baseJSON, nil
	}

	var baseMap map[string]interface{}
	if err := json.Unmarshal(baseJSON, &baseMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base JSON: %w", err)
	}

	var overrideMap map[string]interface{}
	if err := json.Unmarshal(overrideJSON, &overrideMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal override JSON: %w", err)
	}

	mergedMap := MergeMaps(baseMap, overrideMap)
	mergedBytes, err := json.MarshalIndent(mergedMap, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged JSON: %w", err)
	}

	return mergedBytes, nil
}

// MergeMaps recursively merges two string-keyed maps, giving priority to override.
// If both base and override contain a map for the same key, those maps are merged recursively.
// Otherwise, the value in override replaces the value in base.
func MergeMaps(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range base {
		result[k] = v
	}

	for k, overrideVal := range override {
		baseVal, exists := result[k]
		if !exists {
			result[k] = overrideVal
			continue
		}

		baseMap, baseIsMap := baseVal.(map[string]interface{})
		overrideMap, overrideIsMap := overrideVal.(map[string]interface{})

		if baseIsMap && overrideIsMap {
			result[k] = MergeMaps(baseMap, overrideMap)
		} else {
			result[k] = overrideVal
		}
	}

	return result
}
