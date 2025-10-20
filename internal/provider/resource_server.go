package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverResource struct{ client *Client }

func NewServerResource() resource.Resource { return &serverResource{} }

// Terraform model: what users set / what we store
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
			"id":          schema.StringAttribute{Computed: true},
			"plan_id":     schema.Int64Attribute{Required: true},
			"location_id": schema.Int64Attribute{Required: true},
			"os_id":       schema.Int64Attribute{Required: true},
			"raid":        schema.Int64Attribute{Optional: true},         
			"hostname":    schema.StringAttribute{Optional: true},        
			"ip_address":  schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
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

	// Seed state
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

	// Hydrate full details
	//r.Read(ctx, resource.ReadRequest{State: resp.State}, &resp.ReadResponse)
	return
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
		// If your client returns structured status codes, handle 404 → RemoveResource
		var httpErr interface{ StatusCode() int }
		if errors.As(err, &httpErr) && httpErr.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read failed", err.Error())
		return
	}
	if s == nil {
		resp.Diagnostics.AddError("Server missing", "API returned no error but also no server")
		return
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

