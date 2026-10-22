package contract_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

func do(t *testing.T, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

// nullPaths returns the path of every JSON null in v, with array indexes
// collapsed to [] so one rule covers every row of a list.
func nullPaths(v any, path string) []string {
	var out []string
	switch x := v.(type) {
	case nil:
		out = append(out, path)
	case map[string]any:
		for k, child := range x {
			out = append(out, nullPaths(child, path+"."+k)...)
		}
	case []any:
		for _, child := range x {
			out = append(out, nullPaths(child, path+"[]")...)
		}
	}
	return out
}

// assertNoUnexpectedNulls fails for every null in body not covered by
// allowed. allowed holds paths like "$.data[].phone"; "$.data.*" covers every
// direct field of $.data.
func assertNoUnexpectedNulls(t *testing.T, body []byte, allowed []string) {
	t.Helper()
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decode response: %v\n%s", err, body)
	}
	seen := map[string]bool{}
	var unexpected []string
	for _, p := range nullPaths(v, "$") {
		if !seen[p] && !isAllowed(p, allowed) {
			unexpected = append(unexpected, p)
		}
		seen[p] = true
	}
	sort.Strings(unexpected)
	if len(unexpected) > 0 {
		t.Errorf("unexpected null (a nil slice encodes as null and breaks clients):\n  %s", strings.Join(unexpected, "\n  "))
	}
}

func isAllowed(path string, allowed []string) bool {
	for _, a := range allowed {
		if a == path {
			return true
		}
		if prefix, ok := strings.CutSuffix(a, ".*"); ok {
			rest, found := strings.CutPrefix(path, prefix+".")
			if found && !strings.ContainsAny(rest, ".[") {
				return true
			}
		}
	}
	return false
}

func decodeData[T any](t *testing.T, body []byte) T {
	t.Helper()
	var env struct {
		Data T `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode response: %v\n%s", err, body)
	}
	return env.Data
}

func mustStatus(t *testing.T, want, got int, body []byte) {
	t.Helper()
	if got != want {
		t.Fatalf("status = %d, want %d: %s", got, want, body)
	}
}

func uniqueSuffix() string { return fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000) }
