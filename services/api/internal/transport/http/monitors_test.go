package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestMonitorsCRUD(t *testing.T) {
	srv := newTestServer(t)

	body := `{"name":"api","type":"http","target":"https://example.com","intervalSeconds":30,"timeoutSeconds":5,"enabled":true}`
	resp, err := http.Post(srv.URL+"/v1/monitors", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	var created monitorResponse
	mustDecode(t, resp, &created)
	if created.ID == "" || created.ExpectedStatus != 200 {
		t.Fatalf("unexpected create response: %+v", created)
	}

	resp, err = http.Get(srv.URL + "/v1/monitors/" + created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = http.Get(srv.URL + "/v1/monitors")
	if err != nil {
		t.Fatal(err)
	}
	var list []monitorResponse
	mustDecode(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}

	upd := `{"name":"api2","type":"http","target":"https://example.com","intervalSeconds":60,"timeoutSeconds":5,"enabled":false}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/v1/monitors/"+created.ID, bytes.NewBufferString(upd))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var updated monitorResponse
	mustDecode(t, resp, &updated)
	if updated.Name != "api2" || updated.Enabled || updated.IntervalSeconds != 60 {
		t.Fatalf("update response: %+v", updated)
	}

	req, _ = http.NewRequest(http.MethodDelete, srv.URL+"/v1/monitors/"+created.ID, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = http.Get(srv.URL + "/v1/monitors/" + created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete = %d, want 404", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestMonitorsErrors(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"invalid spec", http.MethodPost, "/v1/monitors", `{"name":"","type":"http","target":"https://x","intervalSeconds":30,"timeoutSeconds":5}`, http.StatusBadRequest},
		{"malformed json", http.MethodPost, "/v1/monitors", `{`, http.StatusBadRequest},
		{"bad id", http.MethodGet, "/v1/monitors/not-a-uuid", "", http.StatusBadRequest},
		{"missing", http.MethodGet, "/v1/monitors/" + uuid.NewString(), "", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, srv.URL+tc.path, bytes.NewBufferString(tc.body))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

func mustDecode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
