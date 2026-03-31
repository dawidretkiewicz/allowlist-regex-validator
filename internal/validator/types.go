package validator

import "regexp"

// RegexCandidate is a raw pattern source entry before compilation.
type RegexCandidate struct {
	ID         int
	SourceMode string
	Group      string
	Index      int
	Raw        string
	Pattern    string
	Label      string
}

// CompiledRegex stores compilation results for a candidate.
type CompiledRegex struct {
	RegexCandidate
	Compiled *regexp.Regexp
	Valid    bool
	Error    string
}

const (
	SourceJSON      = "json"
	SourceMultiline = "multiline"
)
