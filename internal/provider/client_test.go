package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListOperatingSystems(t *testing.T) {
	// Fake API server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ordering/os" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-rd-key"); got != "k123" {
			t.Fatalf("missing/incorrect x-rd-key header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 62, "name": "Ubuntu 24.04"},
				{"id": 63, "name": "Debian 12"},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k123")
	got, err := c.ListOperatingSystems(context.Background())
	if err != nil {
		t.Fatalf("ListOperatingSystems error: %v", err)
	}
	if len(got) != 2 || got[0].ID != 62 {
		t.Fatalf("unexpected os list: %+v", got)
	}
}

