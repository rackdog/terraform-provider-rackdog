package provider

import (
	"context"
	"errors"
	"net/http"

	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverResource struct {
	client *Client
	cfg    resolvedConfig
}

func NewServerResource() resource.Resource { return &serverResource{} }

type serverModel struct {
	ID         types.String `tfsdk:"id"`
	PlanID     types.Int64  `tfsdk:"plan_id"`
	LocationID types.Int64  `tfsdk:"location_id"`
	OSID       types.Int64  `tfsdk:"os_id"`
	Raid       types.Int64  `tfsdk:"raid"`
	Hostname   types.String `tfsdk:"hostname"`
	IPAddress  types.String `tfsdk:"ip_address"`
	Status     types.String `tfsdk:"status"`
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Rackdog servers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"plan_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"location_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"os_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"raid": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd := req.ProviderData.(*ProviderData)
	r.client = pd.Client
	r.cfg = pd.Cfg
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Client was nil")
		return
	}

	var plan serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Raid.IsNull() && !plan.Raid.IsUnknown() {
		ok, err := r.client.CheckRaid(ctx, int(plan.Raid.ValueInt64()), int(plan.PlanID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("RAID validation failed", err.Error())
			return
		}
		if !ok {
			resp.Diagnostics.AddError("Invalid RAID for plan",
				"Selected RAID level is not available for the chosen plan. Choose a supported RAID or omit it.")
			return
		}
	}

	in := &CreateServerRequest{
		PlanID:     int(plan.PlanID.ValueInt64()),
		LocationID: int(plan.LocationID.ValueInt64()),
		OSID:       int(plan.OSID.ValueInt64()),
	}
	if !plan.Raid.IsNull() && !plan.Raid.IsUnknown() {
		rv := int(plan.Raid.ValueInt64())
		in.Raid = &rv
	}
	if !plan.Hostname.IsNull() && !plan.Hostname.IsUnknown() {
		h := plan.Hostname.ValueString()
		in.Hostname = &h
	}

	created, err := r.client.CreateServer(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Create failed", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	if created.Hostname != nil {
		plan.Hostname = types.StringValue(*created.Hostname)
	}
	plan.IPAddress = types.StringValue(created.IPAddress)
	plan.Status = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//r.Read(ctx, resource.ReadRequest{State: resp.State}, &resp.ReadResponse)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Client was nil")
		return
	}

	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetServer(ctx, state.ID.ValueString())
	if err != nil {
		var he *HTTPError
		if errors.As(err, &he) && he.Status == http.StatusNotFound {
			if !r.cfg.RecreateOnMissing {
				resp.Diagnostics.AddError(
					"Server deleted outside Terraform",
					"The server no longer exists (404) and provider setting `recreate_on_missing` is false. "+
						"Enable it in the provider or import an existing server by ID.",
				)
				return
			}
			resp.State.RemoveResource(ctx) // lenient: allow plan to recreate
			return
		}
		resp.Diagnostics.AddError("Read failed", err.Error())
		return
	}
	if s == nil {
		resp.Diagnostics.AddError("Server missing", "API returned no error but also no server")
		return
	}

	if !state.Hostname.IsNull() && s.Hostname != nil && state.Hostname.ValueString() != *s.Hostname {
		resp.Diagnostics.AddError(
			"Out-of-band change detected (hostname)",
			fmt.Sprintf("Remote hostname is %q but state expected %q. This likely happened outside Terraform (portal/api). "+
				"Please reconcile: either update your config to match, import the correct resource, or replace this server.",
				*s.Hostname, state.Hostname.ValueString()),
		)
		return
	}

	if s.Plan.ID != 0 && state.PlanID.ValueInt64() != 0 {
		if state.PlanID.ValueInt64() != int64(s.Plan.ID) {
			resp.Diagnostics.AddError(
				"Out-of-band change detected (plan_id)",
				"Remote plan differs from state; reconcile manually and re-run.",
			)
			return
		}
	}

	if s.Location.ID != 0 && state.LocationID.ValueInt64() != 0 {
		if state.LocationID.ValueInt64() != int64(s.Location.ID) {
			resp.Diagnostics.AddError(
				"Out-of-band change detected (location_id)",
				"Remote location differs from state; reconcile manually and re-run.",
			)
			return
		}
	}

	if s.ServerOS != nil && state.OSID.ValueInt64() != 0 {
		if int64(s.ServerOS.ID) != state.OSID.ValueInt64() {
			resp.Diagnostics.AddError(
				"Out-of-band change detected (os_id)",
				"Remote OS differs from state; reconcile manually and re-run.",
			)
			return
		}
	}

	if s.Raid != nil && !state.Raid.IsNull() && !state.Raid.IsUnknown() {
		if int(state.Raid.ValueInt64()) != *s.Raid {
			resp.Diagnostics.AddError(
				"Out-of-band change detected (raid)",
				"Remote RAID differs from state; reconcile manually and re-run.",
			)
			return
		}
	}

	state.IPAddress = types.StringValue(s.IPAddress)
	if s.Hostname != nil {
		state.Hostname = types.StringValue(*s.Hostname)
	}
	if s.PowerStatus != nil {
		state.Status = types.StringValue(*s.PowerStatus)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("No Update implemented", "Rackdog servers cannot be updated.")
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Client was nil")
		return
	}

	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServer(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())
	}
}
