package validator

import "testing"

func TestDecodeEncodedRegex(t *testing.T) {
	in := "^https:\\/\\/example\\.com\\/.*$"
	got, err := DecodeEncodedRegex(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "^https://example\\.com/.*$"
	if got != want {
		t.Fatalf("decode mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestParseJSONRegexObject(t *testing.T) {
	input := `{"group": ["^https:\\/\\/example\\.com\\/.*$", "("]}`
	items, _, err := ParseJSONRegexObject(input, true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Pattern != "^https://example\\.com/.*$" {
		t.Fatalf("unexpected decoded pattern: %s", items[0].Pattern)
	}
}

func TestParseMultilineRegexes(t *testing.T) {
	input := "^https://a.com$\n\n^https://b.com$\n"
	items, _ := ParseMultilineRegexes(input, false, 5)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != 5 || items[1].ID != 6 {
		t.Fatalf("unexpected IDs: %d %d", items[0].ID, items[1].ID)
	}
}
