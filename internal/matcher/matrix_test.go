package matcher

import (
	"testing"

	"regex-validator/internal/validator"
)

func TestBuildMatchOutput(t *testing.T) {
	candidates := []validator.RegexCandidate{
		{ID: 0, Pattern: "^https://a\\.com/.*$", Label: "a"},
		{ID: 1, Pattern: "^https://b\\.com$", Label: "b"},
	}
	compiled := validator.CompileAll(candidates)
	out := BuildMatchOutput([]string{"https://a.com/x", "https://none.com"}, compiled)

	if len(out.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(out.Rows))
	}
	if !out.Rows[0].Results[0] {
		t.Fatalf("expected first URL to match first regex")
	}
	if len(out.UnmatchedURLs) != 1 || out.UnmatchedURLs[0] != "https://none.com" {
		t.Fatalf("unexpected unmatched urls: %#v", out.UnmatchedURLs)
	}
}
