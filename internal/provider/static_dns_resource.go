package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.joshhogle.dev/terraform-provider-ubiquiti-udm/internal/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &staticDNSResource{}
	_ resource.ResourceWithConfigure   = &staticDNSResource{}
	_ resource.ResourceWithImportState = &staticDNSResource{}
)

// NewStaticDNSResource is a helper function to simplify the provider implementation.
func NewStaticDNSResource() resource.Resource {
	return &staticDNSResource{}
}

// staticDNSResource is the resource implementation.
type staticDNSResource struct {
	client *api.Client
}

type staticDNSEntryResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Key        types.String `tfsdk:"key"`
	Port       types.Int32  `tfsdk:"port"`
	Priority   types.Int32  `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	TTL        types.Int32  `tfsdk:"ttl"`
	Value      types.String `tfsdk:"value"`
	Weight     types.Int32  `tfsdk:"weight"`
}

func (r *staticDNSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *staticDNSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_dns"
}

// Schema defines the schema for the resource.
func (r *staticDNSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"key": schema.StringAttribute{
				Required: true,
			},
			"port": schema.Int32Attribute{
				Computed: true,
				Optional: true,
			},
			"priority": schema.Int32Attribute{
				Computed: true,
				Optional: true,
			},
			"record_type": schema.StringAttribute{
				Required: true,
			},
			"ttl": schema.Int32Attribute{
				Computed: true,
				Optional: true,
			},
			"value": schema.StringAttribute{
				Required: true,
			},
			"weight": schema.Int32Attribute{
				Computed: true,
				Optional: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *staticDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan staticDNSEntryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	entry := api.StaticDNSEntry{
		Key:        plan.Key.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		Value:      plan.Value.ValueString(),
	}
	if !plan.Enabled.IsNull() {
		entry.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.Port.IsNull() {
		entry.Port = int(plan.Port.ValueInt32())
	}
	if !plan.Priority.IsNull() {
		entry.Priority = int(plan.Priority.ValueInt32())
	}
	if !plan.TTL.IsNull() {
		entry.TTL = int(plan.TTL.ValueInt32())
	}
	if !plan.Weight.IsNull() {
		entry.Weight = int(plan.Weight.ValueInt32())
	}

	// create the entry
	createdEntry, err := r.client.CreateStaticDNSEntry(ctx, entry)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Create Static DNS Entry",
			fmt.Sprintf("Failed to create static DNS entry using the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	plan.ID = types.StringValue(createdEntry.ID)
	plan.Enabled = types.BoolValue(createdEntry.Enabled)
	plan.Key = types.StringValue(createdEntry.Key)
	plan.Port = types.Int32Value(int32(createdEntry.Port))
	plan.Priority = types.Int32Value(int32(createdEntry.Priority))
	plan.RecordType = types.StringValue(createdEntry.RecordType)
	plan.TTL = types.Int32Value(int32(createdEntry.TTL))
	plan.Value = types.StringValue(createdEntry.Value)
	plan.Weight = types.Int32Value(int32(createdEntry.Weight))

	// save state with the populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *staticDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// retrieve the current state
	var state staticDNSEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh value from the API
	entry, err := r.client.GetStaticDNSEntry(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Retrieve Static DNS Entry",
			fmt.Sprintf("Failed to retrieve the static DNS entry with the ID '%s': %s",
				state.ID.ValueString(), err.Error()),
		)
		return
	}

	// update the state
	state.Enabled = types.BoolValue(entry.Enabled)
	state.Key = types.StringValue(entry.Key)
	state.Port = types.Int32Value(int32(entry.Port))
	state.Priority = types.Int32Value(int32(entry.Priority))
	state.RecordType = types.StringValue(entry.RecordType)
	state.TTL = types.Int32Value(int32(entry.TTL))
	state.Value = types.StringValue(entry.Value)
	state.Weight = types.Int32Value(int32(entry.Weight))

	// set the refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *staticDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan staticDNSEntryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	entry := api.StaticDNSEntry{
		Key:        plan.Key.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		Value:      plan.Value.ValueString(),
	}
	if !plan.Enabled.IsNull() {
		entry.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.Port.IsNull() {
		entry.Port = int(plan.Port.ValueInt32())
	}
	if !plan.Priority.IsNull() {
		entry.Priority = int(plan.Priority.ValueInt32())
	}
	if !plan.TTL.IsNull() {
		entry.TTL = int(plan.TTL.ValueInt32())
	}
	if !plan.Weight.IsNull() {
		entry.Weight = int(plan.Weight.ValueInt32())
	}

	// create the entry
	updatedEntry, err := r.client.UpdateStaticDNSEntry(ctx, plan.ID.ValueString(), entry)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Update Static DNS Entry",
			fmt.Sprintf("Failed to update static DNS entry using the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	plan.ID = types.StringValue(updatedEntry.ID)
	plan.Enabled = types.BoolValue(updatedEntry.Enabled)
	plan.Key = types.StringValue(updatedEntry.Key)
	plan.Port = types.Int32Value(int32(updatedEntry.Port))
	plan.Priority = types.Int32Value(int32(updatedEntry.Priority))
	plan.RecordType = types.StringValue(updatedEntry.RecordType)
	plan.TTL = types.Int32Value(int32(updatedEntry.TTL))
	plan.Value = types.StringValue(updatedEntry.Value)
	plan.Weight = types.Int32Value(int32(updatedEntry.Weight))

	// save state with the populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *staticDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve the current state
	var state staticDNSEntryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// delete the entry
	if err := r.client.DeleteStaticDNSEntry(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Update Static DNS Entry",
			fmt.Sprintf("Failed to update static DNS entry using the UDM API:\n\t%s", err.Error()),
		)
		return
	}
}

func (r *staticDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
