package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type plansDataSource struct{ client *Client }

func NewPlansDataSource() datasource.DataSource { return &plansDataSource{} }

type plansModel struct {
	Location types.String `tfsdk:"location"`
	Plans    []planItem   `tfsdk:"plans"`
}

type planItem struct {
	ID       types.Int64 `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	RAMGB    types.Int64  `tfsdk:"ram"`
	Storage  types.String `tfsdk:"storage"`
	CPUName  types.String `tfsdk:"cpu_name"`
	Cores    types.Int64  `tfsdk:"cores"`
	Price    types.Float64 `tfsdk:"price_monthly"`
}

func (d *plansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plans"
}

func (d *plansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves available hardware plans, optionally filtered by location keyword.",
		Attributes: map[string]schema.Attribute{
			"location": schema.StringAttribute{Optional: true},
			"plans": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"ram":           schema.Int64Attribute{Computed: true},
						"storage":       schema.StringAttribute{Computed: true},
						"cpu_name":      schema.StringAttribute{Computed: true},
						"cores":         schema.Int64Attribute{Computed: true},
						"price_monthly": schema.Float64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *plansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *plansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Client was nil")
		return
	}

	var config plansModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := ""
	if !config.Location.IsNull() && !config.Location.IsUnknown() {
		loc = config.Location.ValueString()
	}

	plans, err := d.client.ListPlans(ctx, loc)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list plans", err.Error())
		return
	}

	state := plansModel{Location: config.Location, Plans: make([]planItem, 0, len(plans))}
	for _, p := range plans {
		state.Plans = append(state.Plans, planItem{
			ID:           types.Int64Value(int64(p.ID)),
			Name:         types.StringValue(p.Name),
			RAMGB:        types.Int64Value(int64(p.RAMGB)),
			Storage:      types.StringValue(p.Storage),
			CPUName:      types.StringValue(p.CPU.Name),
			Cores:        types.Int64Value(int64(p.CPU.Cores)),
			Price:        types.Float64Value(p.Price.Monthly),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

