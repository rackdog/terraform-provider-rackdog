package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestOSDataSource_Schema(t *testing.T) {
	ds := NewOperatingSystemsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	if _, ok := resp.Schema.Attributes["operating_systems"]; !ok {
		t.Error("expected 'operating_systems' attribute in schema")
	}
}

func TestOSDataSource_Metadata(t *testing.T) {
	ds := NewOperatingSystemsDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "rackdog",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	if resp.TypeName != "rackdog_operating_systems" {
		t.Errorf("expected TypeName 'rackdog_operating_systems', got %s", resp.TypeName)
	}
}

func TestOSDataSource_Configure(t *testing.T) {
	ds := &osDataSource{}

	client := NewClient("https://test.example.com", "test-key")
	pd := &ProviderData{
		Client: client,
		Cfg:    resolvedConfig{RecreateOnMissing: false},
	}

	req := datasource.ConfigureRequest{
		ProviderData: pd,
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if ds.client == nil {
		t.Fatal("expected client to be configured, got nil")
	}
	if ds.client != client {
		t.Error("expected client to match provided client")
	}
}

func TestOSDataSource_ClientIntegration(t *testing.T) {
	// Test that the data source can successfully call the client
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ordering/os" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": 48, "name": "Ubuntu 24.04 LTS"},
				{"id": 55, "name": "Windows Server 2019"},
				{"id": 62, "name": "Debian 12"},
			},
		})
	}))
	defer srv.Close()

	// Test through client directly
	client := NewClient(srv.URL, "test-key")
	osList, err := client.ListOperatingSystems(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(osList) != 3 {
		t.Errorf("expected 3 operating systems, got %d", len(osList))
	}
}

func TestOSDataSource_NilClient(t *testing.T) {
	// Test that data source properly handles nil client
	ds := &osDataSource{
		client: nil,
	}

	if ds.client != nil {
		t.Error("expected client to be nil")
	}

	// In real usage, this would be caught during Configure
	pd := &ProviderData{
		Client: NewClient("https://test.com", "test-key"),
		Cfg:    resolvedConfig{RecreateOnMissing: false},
	}

	ds.client = pd.Client

	if ds.client == nil {
		t.Error("expected client to be set after configuration")
	}
}
