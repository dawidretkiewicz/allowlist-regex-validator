package validator

import "strings"

// DecodeEncodedRegex converts JSON-escaped content like \/ and \\ into literal values.
func DecodeEncodedRegex(input string) (string, error) {
	decoded := strings.NewReplacer(
		`\/`, `/`,
		`\\`, `\`,
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
	).Replace(input)
	return decoded, nil
}
