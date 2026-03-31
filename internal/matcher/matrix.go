package matcher

import "regex-validator/internal/validator"

// MatrixRow contains one URL's pass/fail results against each regex ID.
type MatrixRow struct {
	URL     string `json:"url"`
	Results []bool `json:"results"`
}

// URLSummary contains per-URL match details.
type URLSummary struct {
	URL        string `json:"url"`
	FirstMatch *int   `json:"firstMatchId,omitempty"`
	AllMatches []int  `json:"allMatchIds"`
}

// MatchOutput includes the full matrix and aggregate summaries.
type MatchOutput struct {
	Rows          []MatrixRow  `json:"rows"`
	Summaries     []URLSummary `json:"summaries"`
	UnmatchedURLs []string     `json:"unmatchedUrls"`
}

// BuildMatchOutput computes URL x regex results using only valid regexes.
func BuildMatchOutput(urls []string, compiled []validator.CompiledRegex) MatchOutput {
	validRegexes := make([]validator.CompiledRegex, 0, len(compiled))
	for _, item := range compiled {
		if item.Valid && item.Compiled != nil {
			validRegexes = append(validRegexes, item)
		}
	}

	rows := make([]MatrixRow, 0, len(urls))
	summaries := make([]URLSummary, 0, len(urls))
	unmatched := make([]string, 0)

	for _, url := range urls {
		row := MatrixRow{URL: url, Results: make([]bool, 0, len(validRegexes))}
		summary := URLSummary{URL: url, AllMatches: make([]int, 0)}

		for _, item := range validRegexes {
			ok := item.Compiled.MatchString(url)
			row.Results = append(row.Results, ok)
			if ok {
				summary.AllMatches = append(summary.AllMatches, item.ID)
				if summary.FirstMatch == nil {
					id := item.ID
					summary.FirstMatch = &id
				}
			}
		}

		if len(summary.AllMatches) == 0 {
			unmatched = append(unmatched, url)
		}

		rows = append(rows, row)
		summaries = append(summaries, summary)
	}

	return MatchOutput{
		Rows:          rows,
		Summaries:     summaries,
		UnmatchedURLs: unmatched,
	}
}
