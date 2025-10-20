package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type rackdogProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &rackdogProvider{version: version}
	}
}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey 	types.String `tfsdk:"api_key"`
	RecreateOnMissing 	types.Bool `tfsdk:"recreate_on_missing"`
}

type resolvedConfig struct {
    RecreateOnMissing bool
}

type ProviderData struct {
    Client *Client
    Cfg    resolvedConfig
}

func (p *rackdogProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rackdog"
	resp.Version = p.version
}

func (p *rackdogProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for Rackdog infrastructure.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "Rackdog API base URL.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for Rackdog.",
				Optional:    true,
				Sensitive:   true,
			},
			"recreate_on_missing": schema.BoolAttribute{
				Optional: true,
				Description: "If true, resources missing on Read (404) will be removed from state so Terraform can recreate them.",
			},
		},
	}
}

func (p *rackdogProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config providerModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() { return }

    endpoint := getString(config.Endpoint, "RACKDOG_ENDPOINT", "https://metal.rackdog.com")
    key := getString(config.APIKey, "RACKDOG_API_KEY", "")
    if key == "" {
        resp.Diagnostics.AddError("Missing API Key", "No api_key or RACKDOG_API_KEY found.")
        return
    }

    recreate := false
    if !config.RecreateOnMissing.IsNull() && !config.RecreateOnMissing.IsUnknown() {
        recreate = config.RecreateOnMissing.ValueBool()
    } else if v := os.Getenv("RACKDOG_RECREATE_ON_MISSING"); v != "" {
        recreate = strings.EqualFold(v, "1") || strings.EqualFold(v, "true")
    }

    client := NewClient(endpoint, key)
    pd := &ProviderData{
        Client: client,
        Cfg:    resolvedConfig{RecreateOnMissing: recreate},
    }

    // Make available to resources and data sources
    resp.DataSourceData = pd
    resp.ResourceData   = pd

    tflog.Info(ctx, "Rackdog provider configured", map[string]any{
        "endpoint":            endpoint,
        "recreate_on_missing": recreate,
    })
}

func (p *rackdogProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
	}
}

func (p *rackdogProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPlansDataSource,
		NewOperatingSystemsDataSource,
	}
}

func getString(v types.String, env, def string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	if val := os.Getenv(env); val != "" {
		return val
	}
	return def
}
