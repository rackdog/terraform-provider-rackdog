package provider

import (
	"context"
	"os"

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
		},
	}
}

func (p *rackdogProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := getString(config.Endpoint, "RACKDOG_ENDPOINT", "https://metal.rackdog.com")
	key := getString(config.APIKey, "RACKDOG_API_KEY", "")

	if key == "" {
		resp.Diagnostics.AddError("Missing API Key", "No api_key or RACKDOG_KEY found.")
		return
	}

	client := NewClient(endpoint, key)
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Rackdog provider configured", map[string]any{
		"endpoint": endpoint,
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
