package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type osDataSource struct{ client *Client }

func NewOperatingSystemsDataSource() datasource.DataSource { return &osDataSource{} }

type osModel struct {
	OperatingSystems []osItem `tfsdk:"operating_systems"`
}

type osItem struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *osDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operating_systems"
}

func (d *osDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all available operating systems from Rackdog /ordering/os.",
		Attributes: map[string]schema.Attribute{
			"operating_systems": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.Int64Attribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *osDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *osDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Client was nil")
		return
	}

	osList, err := d.client.ListOperatingSystems(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list operating systems", err.Error())
		return
	}

	state := osModel{OperatingSystems: make([]osItem, 0, len(osList))}
	for _, o := range osList {
		state.OperatingSystems = append(state.OperatingSystems, osItem{
			ID:   types.Int64Value(int64(o.ID)),
			Name: types.StringValue(o.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

