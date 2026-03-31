package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"regex-validator/internal/matcher"
	"regex-validator/internal/validator"
)

// Handler holds shared request processing behavior.
type Handler struct{}

type allowlistSource struct {
	Label string
	URL   string
}

var presetAllowlists = map[string]allowlistSource{
	"emea-prod": {
		Label: "(i)EMEA PROD",
		URL:   "https://onboard-config-emea.bmwgroup.com/allowlist.json",
	},
	"emea-e2e": {
		Label: "(i)EMEA E2E",
		URL:   "https://onboard-config-e2e-emea.bmwgroup.com/allowlist.json",
	},
	"us-prod": {
		Label: "(i)US PROD",
		URL:   "https://onboard-config-us.bmwgroup.com/allowlist.json",
	},
	"us-e2e": {
		Label: "(i)US E2E",
		URL:   "https://onboard-config-e2e-us.bmwgroup.com/allowlist.json",
	},
}

// NewHandler creates a new HTTP handler set.
func NewHandler() *Handler {
	return &Handler{}
}

// Register wires API routes into mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/allowlists", h.handleListAllowlists)
	mux.HandleFunc("/api/v1/validate", h.handleValidate)
	mux.HandleFunc("/api/v1/match", h.handleMatch)
}

func (h *Handler) handleListAllowlists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	keys := make([]string, 0, len(presetAllowlists))
	for key := range presetAllowlists {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]PresetAllowlist, 0, len(keys))
	for _, key := range keys {
		source := presetAllowlists[key]
		items = append(items, PresetAllowlist{
			Key:   key,
			Label: source.Label,
			URL:   source.URL,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	resp, err := parseAndProcess(r, false)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	resp, err := parseAndProcess(r, true)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseAndProcess(r *http.Request, includeMatch bool) (ProcessResponse, error) {
	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ProcessResponse{}, err
	}

	presetJSON := ""
	if req.PresetAllowlist != "" {
		fetched, err := fetchPresetAllowlistJSON(req.PresetAllowlist)
		if err != nil {
			return ProcessResponse{}, err
		}
		presetJSON = fetched
	}

	jsonCandidates, nextID, err := validator.ParseJSONRegexObject(presetJSON, req.DecodeJSON, 0)
	if err != nil {
		return ProcessResponse{}, err
	}

	manualJSONCandidates, nextID, err := validator.ParseJSONRegexObject(req.JSONRegexObject, req.DecodeJSON, nextID)
	if err != nil {
		return ProcessResponse{}, err
	}
	jsonCandidates = append(jsonCandidates, manualJSONCandidates...)

	lineCandidates, _ := validator.ParseMultilineRegexes(req.MultilineRegexes, req.DecodeMultiline, nextID)
	allCandidates := append(jsonCandidates, lineCandidates...)
	compiled := validator.CompileAll(allCandidates)

	regexes := make([]RegexResult, 0, len(compiled))
	validIDs := make([]int, 0)
	validOrder := make([]int, 0)
	validCount := 0
	for _, item := range compiled {
		regexes = append(regexes, RegexResult{
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

	resp := ProcessResponse{
		Regexes: regexes,
		Summary: ValidationSummary{
			TotalCount:   len(compiled),
			ValidCount:   validCount,
			InvalidCount: len(compiled) - validCount,
		},
		ValidIDs:   validIDs,
		ValidOrder: validOrder,
	}

	if includeMatch {
		urls := validator.ParseMultilineURLs(req.MultilineURLs)
		m := matcher.BuildMatchOutput(urls, compiled)
		resp.Match = &m
	}

	return resp, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func fetchPresetAllowlistJSON(key string) (string, error) {
	source, ok := presetAllowlists[key]
	if !ok {
		return "", fmt.Errorf("unknown preset allowlist key: %s", key)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(source.URL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch preset allowlist %s: %w", source.Label, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("preset allowlist %s returned status %d", source.Label, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read preset allowlist %s: %w", source.Label, err)
	}

	return string(body), nil
}
