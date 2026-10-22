package contract_test

import (
	"net/http"
	"strings"
	"testing"
)

// Fields that are legitimately null for a freshly registered organization. Any
// other null is treated as a bug.
var allowedNulls = map[string][]string{}

func register(t *testing.T) string {
	t.Helper()
	status, body := do(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"name":              "Contract Owner",
		"email":             "owner" + uniqueSuffix() + "@contract.test",
		"password":          "contract-pass",
		"organization_name": "Contract Org",
	})
	mustStatus(t, http.StatusCreated, status, body)
	return decodeData[struct {
		AccessToken string `json:"access_token"`
	}](t, body).AccessToken
}

// Every list a client renders must be [] rather than null when there is
// nothing in it, or the client crashes calling .map/.forEach on it.
func TestEmptyAccount_GetRoutesReturnNoNullLists(t *testing.T) {
	token := register(t)
	checked := 0
	for _, r := range api.Router().Routes() {
		if r.Method != http.MethodGet || len(r.Parameters) > 0 || !strings.HasPrefix(r.Path, "/api/") {
			continue
		}
		checked++
		t.Run(r.Path, func(t *testing.T) {
			status, body := do(t, http.MethodGet, r.Path, token, nil)
			mustStatus(t, http.StatusOK, status, body)
			assertNoUnexpectedNulls(t, body, allowedNulls[r.Path])
		})
	}
	if checked == 0 {
		t.Fatal("no GET routes found")
	}
}
