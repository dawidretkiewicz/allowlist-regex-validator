Regex Validator (Go)

A local tool to validate URL regexes with Go's regexp engine (RE2) and test URLs against those regexes.

Features
- JSON object regex input (list-per-key).
- Multiline regex input (one regex per line).
- Multiline URL input.
- Built-in BMW preset allowlists selectable in the UI:
  - (i)EMEA PROD
  - (i)EMEA E2E
  - (i)US PROD
  - (i)US E2E
- Per-mode decoding toggles for encoded regexes.
- Full URL x valid-regex matrix and per-URL summary.
- Web interface and CLI using the same core logic.

Run web app
1. go run ./cmd/server
2. Open http://localhost:8080
3. Optionally pick a preset in the "Preset BMW Allowlist" dropdown and run validation/matching.

Run CLI
- Validate:
  go run ./cmd/regex-validator validate -json-file testdata/sample_regexes.json -decode-json=true
- Match:
  go run ./cmd/regex-validator match -json-file testdata/sample_regexes.json -urls-inline "https://www.fubo.tv/live\\nhttps://example.com" -decode-json=true

Notes
- Regex compilation uses Go RE2 syntax only.
- Invalid regexes are reported and excluded from matching.

API
- List presets:
  GET /api/v1/allowlists
- Use preset in validate/match payload:
  {
    "presetAllowlist": "emea-prod",
    "multilineUrls": "https://example.com",
    "decodeJson": true,
    "decodeMultiline": false
  }
