package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPlansDataSource_Schema(t *testing.T) {
	ds := NewPlansDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	if _, ok := resp.Schema.Attributes["location"]; !ok {
		t.Error("expected 'location' attribute in schema")
	}

	if _, ok := resp.Schema.Attributes["plans"]; !ok {
		t.Error("expected 'plans' attribute in schema")
	}
}

func TestPlansDataSource_ClientIntegration(t *testing.T) {
	// Create test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id":   10,
					"name": "Standard Plan",
					"cpu": map[string]any{
						"name":     "Intel Xeon E5",
						"cores":    16,
						"speedGhz": 2.4,
					},
					"locations": []map[string]any{
						{
							"id":           1,
							"name":         "New York",
							"keyword":      "NY",
							"monthlyPrice": 149,
						},
					},
					"ram":       32,
					"storageGb": 200,
				},
			},
		})
	}))
	defer srv.Close()

	// Test through client directly
	client := NewClient(srv.URL, "test-key")
	plans, err := client.ListPlans(context.Background(), "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}

	if plans[0].Name != "Standard Plan" {
		t.Errorf("expected plan name 'Standard Plan', got %s", plans[0].Name)
	}
}

func TestPlansDataSource_Metadata(t *testing.T) {
	ds := NewPlansDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "rackdog",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	if resp.TypeName != "rackdog_plans" {
		t.Errorf("expected TypeName 'rackdog_plans', got %s", resp.TypeName)
	}
}
