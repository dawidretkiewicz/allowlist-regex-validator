package validator

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ParseJSONRegexObject parses a JSON object where each key maps to a string array.
func ParseJSONRegexObject(input string, decode bool, startID int) ([]RegexCandidate, int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, startID, nil
	}

	var payload map[string][]string
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, startID, fmt.Errorf("invalid JSON object: %w", err)
	}

	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]RegexCandidate, 0)
	nextID := startID
	for _, group := range keys {
		list := payload[group]
		for i, raw := range list {
			pattern := raw
			if decode {
				decoded, err := DecodeEncodedRegex(raw)
				if err == nil {
					pattern = decoded
				}
			}
			out = append(out, RegexCandidate{
				ID:         nextID,
				SourceMode: SourceJSON,
				Group:      group,
				Index:      i,
				Raw:        raw,
				Pattern:    pattern,
				Label:      fmt.Sprintf("json:%s[%d]", group, i),
			})
			nextID++
		}
	}

	return out, nextID, nil
}

// ParseMultilineRegexes parses one pattern per non-empty line.
func ParseMultilineRegexes(input string, decode bool, startID int) ([]RegexCandidate, int) {
	lines := strings.Split(input, "\n")
	out := make([]RegexCandidate, 0)
	nextID := startID
	lineIndex := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		pattern := trimmed
		if decode {
			decoded, err := DecodeEncodedRegex(trimmed)
			if err == nil {
				pattern = decoded
			}
		}

		out = append(out, RegexCandidate{
			ID:         nextID,
			SourceMode: SourceMultiline,
			Group:      SourceMultiline,
			Index:      lineIndex,
			Raw:        trimmed,
			Pattern:    pattern,
			Label:      fmt.Sprintf("multiline:%d", lineIndex),
		})
		nextID++
		lineIndex++
	}

	return out, nextID
}

// ParseMultilineURLs parses one URL per non-empty line.
func ParseMultilineURLs(input string) []string {
	lines := strings.Split(input, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
