package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProvider_Metadata(t *testing.T) {
	p := New("test-version")()
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "rackdog" {
		t.Errorf("expected TypeName 'rackdog', got %s", resp.TypeName)
	}
	if resp.Version != "test-version" {
		t.Errorf("expected Version 'test-version', got %s", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	p := New("dev")()
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	// required attributes
	attrs := []string{"endpoint", "api_key", "recreate_on_missing"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected '%s' attribute in schema", attr)
		}
	}
}

func TestProvider_Configure_WithConfig(t *testing.T) {
	client := NewClient("https://test.rackdog.com", "test-api-key-123")

	if client == nil {
		t.Fatal("expected Client to be created, got nil")
	}

	if client.base != "https://test.rackdog.com" {
		t.Errorf("expected base URL 'https://test.rackdog.com', got '%s'", client.base)
	}

	if client.apiKey != "test-api-key-123" {
		t.Error("expected API key to be set")
	}
}

func TestProvider_Configure_WithEnv(t *testing.T) {
	os.Setenv("RACKDOG_API_KEY", "env-api-key")
	os.Setenv("RACKDOG_ENDPOINT", "https://env.rackdog.com")
	defer func() {
		os.Unsetenv("RACKDOG_API_KEY")
		os.Unsetenv("RACKDOG_ENDPOINT")
	}()

	apiKey := os.Getenv("RACKDOG_API_KEY")
	endpoint := os.Getenv("RACKDOG_ENDPOINT")

	if apiKey != "env-api-key" {
		t.Errorf("expected RACKDOG_API_KEY 'env-api-key', got '%s'", apiKey)
	}

	if endpoint != "https://env.rackdog.com" {
		t.Errorf("expected RACKDOG_ENDPOINT 'https://env.rackdog.com', got '%s'", endpoint)
	}

	client := NewClient(endpoint, apiKey)
	if client == nil {
		t.Fatal("expected Client to be created, got nil")
	}
}

func TestProvider_Configure_MissingAPIKey(t *testing.T) {
	p := New("dev")()
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	if _, ok := schemaResp.Schema.Attributes["api_key"]; !ok {
		t.Error("expected 'api_key' attribute in schema")
	}
}

func TestProvider_Resources(t *testing.T) {
	p := New("dev")()

	resources := p.Resources(context.Background())

	if len(resources) == 0 {
		t.Fatal("expected at least one resource, got none")
	}

	foundServer := false
	for _, res := range resources {
		r := res()
		if _, ok := r.(*serverResource); ok {
			foundServer = true
			break
		}
	}

	if !foundServer {
		t.Error("expected server resource to be registered")
	}
}

func TestProvider_DataSources(t *testing.T) {
	p := New("dev")()

	dataSources := p.DataSources(context.Background())

	if len(dataSources) < 2 {
		t.Fatalf("expected at least 2 data sources, got %d", len(dataSources))
	}

	foundPlans := false
	foundOS := false

	for _, ds := range dataSources {
		d := ds()
		switch d.(type) {
		case *plansDataSource:
			foundPlans = true
		case *osDataSource:
			foundOS = true
		}
	}

	if !foundPlans {
		t.Error("expected plans data source to be registered")
	}
	if !foundOS {
		t.Error("expected operating systems data source to be registered")
	}
}
