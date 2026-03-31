package validator

import "regexp"

// CompileAll compiles all candidates and preserves positional metadata.
func CompileAll(candidates []RegexCandidate) []CompiledRegex {
	out := make([]CompiledRegex, 0, len(candidates))
	for _, candidate := range candidates {
		re, err := regexp.Compile(candidate.Pattern)
		result := CompiledRegex{
			RegexCandidate: candidate,
			Compiled:       re,
			Valid:          err == nil,
		}
		if err != nil {
			result.Error = err.Error()
		}
		out = append(out, result)
	}
	return out
}
