package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"regex-validator/internal/api"
	"regex-validator/internal/matcher"
	"regex-validator/internal/validator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		if err := runValidate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "match":
		if err := runMatch(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  regex-validator validate [flags]")
	fmt.Println("  regex-validator match [flags]")
	fmt.Println("")
	fmt.Println("Common flags:")
	fmt.Println("  -json-file string         JSON object file with regex lists")
	fmt.Println("  -json-inline string       JSON object inline")
	fmt.Println("  -regex-file string        Multiline regex file")
	fmt.Println("  -regex-inline string      Multiline regexes inline with \\n")
	fmt.Println("  -decode-json              Decode escaped regexes from JSON input (default true)")
	fmt.Println("  -decode-multiline         Decode escaped regexes from multiline input")
	fmt.Println("  -output string            text|json (default text)")
	fmt.Println("")
	fmt.Println("Match-only flags:")
	fmt.Println("  -urls-file string         Multiline URL file")
	fmt.Println("  -urls-inline string       Multiline URLs inline with \\n")
}

func runValidate(args []string) error {
	cfg, err := parseCommonFlags("validate", args)
	if err != nil {
		return err
	}

	resp, err := process(cfg.req, false)
	if err != nil {
		return err
	}

	return writeOutput(cfg.output, resp)
}

func runMatch(args []string) error {
	fs := flag.NewFlagSet("match", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	jsonFile := fs.String("json-file", "", "")
	jsonInline := fs.String("json-inline", "", "")
	regexFile := fs.String("regex-file", "", "")
	regexInline := fs.String("regex-inline", "", "")
	urlsFile := fs.String("urls-file", "", "")
	urlsInline := fs.String("urls-inline", "", "")
	decodeJSON := fs.Bool("decode-json", true, "")
	decodeMultiline := fs.Bool("decode-multiline", false, "")
	output := fs.String("output", "text", "")

	if err := fs.Parse(args); err != nil {
		return err
	}

	req, err := makeRequest(*jsonFile, *jsonInline, *regexFile, *regexInline, *urlsFile, *urlsInline, *decodeJSON, *decodeMultiline)
	if err != nil {
		return err
	}

	resp, err := process(req, true)
	if err != nil {
		return err
	}

	return writeOutput(*output, resp)
}

type commonConfig struct {
	req    api.ProcessRequest
	output string
}

func parseCommonFlags(name string, args []string) (commonConfig, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	jsonFile := fs.String("json-file", "", "")
	jsonInline := fs.String("json-inline", "", "")
	regexFile := fs.String("regex-file", "", "")
	regexInline := fs.String("regex-inline", "", "")
	decodeJSON := fs.Bool("decode-json", true, "")
	decodeMultiline := fs.Bool("decode-multiline", false, "")
	output := fs.String("output", "text", "")

	if err := fs.Parse(args); err != nil {
		return commonConfig{}, err
	}

	req, err := makeRequest(*jsonFile, *jsonInline, *regexFile, *regexInline, "", "", *decodeJSON, *decodeMultiline)
	if err != nil {
		return commonConfig{}, err
	}

	return commonConfig{req: req, output: *output}, nil
}

func makeRequest(
	jsonFile, jsonInline, regexFile, regexInline, urlsFile, urlsInline string,
	decodeJSON, decodeMultiline bool,
) (api.ProcessRequest, error) {
	jsonBlob, err := mergeFileAndInline(jsonFile, jsonInline)
	if err != nil {
		return api.ProcessRequest{}, err
	}

	regexBlob, err := mergeFileAndInline(regexFile, regexInline)
	if err != nil {
		return api.ProcessRequest{}, err
	}

	urlsBlob, err := mergeFileAndInline(urlsFile, urlsInline)
	if err != nil {
		return api.ProcessRequest{}, err
	}

	return api.ProcessRequest{
		JSONRegexObject:  jsonBlob,
		MultilineRegexes: regexBlob,
		MultilineURLs:    urlsBlob,
		DecodeJSON:       decodeJSON,
		DecodeMultiline:  decodeMultiline,
	}, nil
}

func mergeFileAndInline(filePath, inline string) (string, error) {
	parts := make([]string, 0, 3)
	if filePath != "" {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		parts = append(parts, string(b))
	}
	if inline != "" {
		parts = append(parts, strings.ReplaceAll(inline, "\\n", "\n"))
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, "\n"), nil
}

func process(req api.ProcessRequest, includeMatch bool) (api.ProcessResponse, error) {
	jsonCandidates, nextID, err := validator.ParseJSONRegexObject(req.JSONRegexObject, req.DecodeJSON, 0)
	if err != nil {
		return api.ProcessResponse{}, err
	}
	lineCandidates, _ := validator.ParseMultilineRegexes(req.MultilineRegexes, req.DecodeMultiline, nextID)
	allCandidates := append(jsonCandidates, lineCandidates...)
	compiled := validator.CompileAll(allCandidates)

	regexes := make([]api.RegexResult, 0, len(compiled))
	validIDs := make([]int, 0)
	validOrder := make([]int, 0)
	validCount := 0
	for _, item := range compiled {
		regexes = append(regexes, api.RegexResult{
			ID:         item.ID,
			SourceMode: item.SourceMode,
			Group:      item.Group,
			Index:      item.Index,
			Label:      item.Label,
			Raw:        item.Raw,
			Pattern:    item.Pattern,
			Valid:      item.Valid,
			Error:      item.Error,
		})
		if item.Valid {
			validCount++
			validIDs = append(validIDs, item.ID)
			validOrder = append(validOrder, item.ID)
		}
	}

	resp := api.ProcessResponse{
		Regexes: regexes,
		Summary: api.ValidationSummary{
			TotalCount:   len(compiled),
			ValidCount:   validCount,
			InvalidCount: len(compiled) - validCount,
		},
		ValidIDs:   validIDs,
		ValidOrder: validOrder,
	}

	if includeMatch {
		urls := validator.ParseMultilineURLs(req.MultilineURLs)
		match := matcher.BuildMatchOutput(urls, compiled)
		resp.Match = &match
	}

	return resp, nil
}

func writeOutput(format string, resp api.ProcessResponse) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	case "text":
		fmt.Printf("Regexes: total=%d valid=%d invalid=%d\n", resp.Summary.TotalCount, resp.Summary.ValidCount, resp.Summary.InvalidCount)
		for _, rx := range resp.Regexes {
			status := "ok"
			if !rx.Valid {
				status = "invalid: " + rx.Error
			}
			fmt.Printf("- id=%d label=%s status=%s\n", rx.ID, rx.Label, status)
		}

		if resp.Match != nil {
			fmt.Println("")
			fmt.Println("Match summaries:")
			for _, s := range resp.Match.Summaries {
				if s.FirstMatch == nil {
					fmt.Printf("- %s -> no match\n", s.URL)
					continue
				}
				fmt.Printf("- %s -> first=%d all=%v\n", s.URL, *s.FirstMatch, s.AllMatches)
			}
		}
		return nil
	default:
		return errors.New("unknown output format, use text or json")
	}
}

func prettyJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes.TrimSpace(b))
}
