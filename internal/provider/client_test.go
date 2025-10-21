package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
	"time"
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

func TestListPlans(t *testing.T) {
	tests := []struct {
		name     string
		location string
		wantPath string
	}{
		{
			name:     "without location filter",
			location: "",
			wantPath: "/v1/ordering/plans",
		},
		{
			name:     "with location filter",
			location: "NY",
			wantPath: "/v1/ordering/plans?location=NY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != tt.wantPath {
					t.Fatalf("unexpected path: got %s, want %s", r.URL.String(), tt.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"success": true,
					"data": []map[string]any{
						{
							"id":   10,
							"name": "Test Plan",
							"cpu": map[string]any{
								"name":     "Intel Xeon",
								"cores":    8,
								"speedGhz": 3.5,
							},
							"locations": []map[string]any{
								{
									"id":           1,
									"name":         "New York",
									"keyword":      "NY",
									"monthlyPrice": 99,
								},
							},
							"ram":       16,
							"storageGb": 500,
						},
					},
				})
			}))
			defer srv.Close()

			c := NewClient(srv.URL, "k123")
			plans, err := c.ListPlans(context.Background(), tt.location)
			if err != nil {
				t.Fatalf("ListPlans error: %v", err)
			}
			if len(plans) != 1 {
				t.Fatalf("expected 1 plan, got %d", len(plans))
			}
			if plans[0].ID != 10 {
				t.Fatalf("expected plan ID 10, got %d", plans[0].ID)
			}
		})
	}
}

func TestCreateServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ordering/allocate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.PlanID != 10 || req.LocationID != 1 || req.OSID != 62 {
			t.Fatalf("unexpected request data: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"id":          "server-123",
				"hostname":    "test-server",
				"ipAddress":   "192.168.1.100",
				"powerStatus": "ON",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k123")
	hostname := "test-server"
	server, err := c.CreateServer(context.Background(), &CreateServerRequest{
		PlanID:     10,
		LocationID: 1,
		OSID:       62,
		Hostname:   &hostname,
	})
	if err != nil {
		t.Fatalf("CreateServer error: %v", err)
	}
	if server.ID != "server-123" {
		t.Fatalf("expected server ID 'server-123', got %s", server.ID)
	}
}

func TestGetServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/servers/server-123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		hostname := "test-server"
		status := "ON"
		price := "$99.99"
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"id":       "server-123",
				"hostname": hostname,
				"plan": map[string]any{
					"id":      10,
					"name":    "Test Plan",
					"ram":     16,
					"storage": 500,
					"cpuName": "Intel Xeon",
					"cores":   8,
				},
				"location": map[string]any{
					"id":      1,
					"name":    "New York",
					"keyword": "NY",
					"country": "USA",
				},
				"serverOS": map[string]any{
					"id":   62,
					"name": "Ubuntu 24.04",
				},
				"ipAddress":         "192.168.1.100",
				"devicePowerStatus": status,
				"monthlyPrice":      price,
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k123")
	server, err := c.GetServer(context.Background(), "server-123")
	if err != nil {
		t.Fatalf("GetServer error: %v", err)
	}
	if server.ID != "server-123" {
		t.Fatalf("expected server ID 'server-123', got %s", server.ID)
	}
	if server.Plan.ID != 10 {
		t.Fatalf("expected plan ID 10, got %d", server.Plan.ID)
	}
}

func TestDeleteServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/servers/server-123/destroy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k123")
	err := c.DeleteServer(context.Background(), "server-123")
	if err != nil {
		t.Fatalf("DeleteServer error: %v", err)
	}
}

func TestCheckRaid(t *testing.T) {
	tests := []struct {
		name       string
		raid       int
		planID     int
		wantPath   string
		apiSuccess bool
		wantErr    bool
	}{
		{
			name:       "valid raid configuration",
			raid:       1,
			planID:     10,
			wantPath:   "/v1/ordering/plans/10/raid/1/check",
			apiSuccess: true,
			wantErr:    false,
		},
		{
			name:       "invalid raid configuration",
			raid:       5,
			planID:     10,
			wantPath:   "/v1/ordering/plans/10/raid/5/check",
			apiSuccess: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Fatalf("unexpected path: got %s, want %s", r.URL.Path, tt.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"success": tt.apiSuccess,
					"message": "RAID check result",
				})
			}))
			defer srv.Close()

			c := NewClient(srv.URL, "k123")
			valid, err := c.CheckRaid(context.Background(), tt.raid, tt.planID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !valid {
				t.Fatal("expected valid=true")
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "message": "Server not found"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k123")
	_, err := c.GetServer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected status 404, got %d", httpErr.Status)
	}
}

func TestClientAPIKeyHeader(t *testing.T) {
	apiKey := "test-key-xyz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("x-rd-key")
		if got != apiKey {
			t.Fatalf("expected x-rd-key=%s, got %s", apiKey, got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    []any{},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, apiKey)
	_, err := c.ListOperatingSystems(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

///////////
/// integration tests that run against actual api 
///////////
type planWire struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RAM         int    `json:"ram"`
	StorageGb   int    `json:"storageGb"`
	CPU         struct {
		Name     string  `json:"name"`
		Cores    int     `json:"cores"`
		SpeedGhz float64 `json:"speedGhz"`
	} `json:"cpu"`
	Price struct {
		Monthly float64 `json:"monthly"`
	} `json:"price"`
	RaidOptions []any `json:"raidOptions"`
	Locations   []struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Keyword      string  `json:"keyword"`
		MonthlyPrice float64 `json:"monthlyPrice"`
	} `json:"locations"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func TestListPlans_Integration_WireSchema(t *testing.T) {
	base := os.Getenv("RACKDOG_API_URL") 
	key := os.Getenv("RACKDOG_API_KEY") 
	location := os.Getenv("RACKDOG_LOCATION") 

	if base == "" || key == "" {
		t.Skip("set RACKDOG_API_URL and RACKDOG_API_KEY to run integration tests")
	}

	url := base + "/v1/ordering/plans"
	if location != "" {
		url += "?location=" + location
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	// Try both common auth styles (whichever your API uses will work).
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("X-API-Key", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body struct {
			Error   any `json:"error"`
			Message any `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("non-200 status: %d; body: %+v", resp.StatusCode, body)
	}

	// Some APIs wrap data { success, data: [...] }, so handle both shapes.
	var (
		raw any
		errDec = json.NewDecoder(resp.Body).Decode(&raw)
	)
	if errDec != nil {
		t.Fatalf("decode: %v", errDec)
	}

	var plans []planWire

	switch v := raw.(type) {
	case map[string]any:
		// expect wrapper with "data"
		d, ok := v["data"]
		if !ok {
			t.Fatalf("response missing 'data' field")
		}
		b, _ := json.Marshal(d)
		if err := json.Unmarshal(b, &plans); err != nil {
			t.Fatalf("decode data[]: %v", err)
		}
	case []any:
		// bare array
		b, _ := json.Marshal(v)
		if err := json.Unmarshal(b, &plans); err != nil {
			t.Fatalf("decode []: %v", err)
		}
	default:
		t.Fatalf("unexpected top-level JSON type %T", v)
	}

	if len(plans) == 0 {
		t.Fatalf("expected at least one plan")
	}

	// Validate the first plan’s shape (and parse timestamps).
	p := plans[0]

	if p.ID <= 0 {
		t.Fatalf("invalid id: %d", p.ID)
	}
	if p.Name == "" {
		t.Fatalf("missing name")
	}
	if p.CPU.Name == "" || p.CPU.Cores <= 0 || p.CPU.SpeedGhz <= 0 {
		t.Fatalf("invalid cpu: %+v", p.CPU)
	}
	if p.Price.Monthly < 0 {
		t.Fatalf("invalid price.monthly: %v", p.Price.Monthly)
	}
	if p.RAM < 0 || p.StorageGb < 0 {
		t.Fatalf("invalid ram/storage: ram=%d storageGb=%d", p.RAM, p.StorageGb)
	}
	if len(p.Locations) == 0 {
		t.Fatalf("expected at least one location")
	}
	loc := p.Locations[0]
	if loc.ID <= 0 || loc.Name == "" || loc.Keyword == "" || loc.MonthlyPrice < 0 {
		t.Fatalf("invalid location: %+v", loc)
	}

	// createdAt/updatedAt should be RFC3339-like strings per your example.
	// Try a couple common layouts.
	parse := func(s string) time.Time {
		if s == "" {
			return time.Time{}
		}
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05", // your example format (no zone)
		} {
			if ts, err := time.Parse(layout, s); err == nil {
				return ts
			}
		}
		return time.Time{}
	}
	created := parse(p.CreatedAt)
	updated := parse(p.UpdatedAt)
	if created.IsZero() || updated.IsZero() {
		t.Fatalf("timestamps not parseable: createdAt=%q updatedAt=%q", p.CreatedAt, p.UpdatedAt)
	}
}
