package api

import "regex-validator/internal/matcher"

// ProcessRequest is shared by both /validate and /match.
type ProcessRequest struct {
	JSONRegexObject  string `json:"jsonRegexObject"`
	MultilineRegexes string `json:"multilineRegexes"`
	MultilineURLs    string `json:"multilineUrls"`
	PresetAllowlist  string `json:"presetAllowlist"`
	DecodeJSON       bool   `json:"decodeJson"`
	DecodeMultiline  bool   `json:"decodeMultiline"`
}

// PresetAllowlist describes a predefined allowlist source.
type PresetAllowlist struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// RegexResult is the compiled representation returned to clients.
type RegexResult struct {
	ID         int    `json:"id"`
	SourceMode string `json:"sourceMode"`
	Group      string `json:"group"`
	Index      int    `json:"index"`
	Label      string `json:"label"`
	Raw        string `json:"raw"`
	Pattern    string `json:"pattern"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}

// ValidationSummary reports compile outcomes.
type ValidationSummary struct {
	TotalCount   int `json:"totalCount"`
	ValidCount   int `json:"validCount"`
	InvalidCount int `json:"invalidCount"`
}

// ProcessResponse returns validation and optional match results.
type ProcessResponse struct {
	Regexes    []RegexResult        `json:"regexes"`
	Summary    ValidationSummary    `json:"summary"`
	ValidIDs   []int                `json:"validIds"`
	Match      *matcher.MatchOutput `json:"match,omitempty"`
	ValidOrder []int                `json:"validOrder"`
}

// ErrorResponse is an API error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}
