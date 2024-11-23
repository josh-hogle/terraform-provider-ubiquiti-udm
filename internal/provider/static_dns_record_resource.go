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
	_ resource.Resource                = &staticDNSRecordResource{}
	_ resource.ResourceWithConfigure   = &staticDNSRecordResource{}
	_ resource.ResourceWithImportState = &staticDNSRecordResource{}
)

// NewStaticDNSRecordResource is a helper function to simplify the provider implementation.
func NewStaticDNSRecordResource() resource.Resource {
	return &staticDNSRecordResource{}
}

// staticDNSRecordResource is the resource implementation.
type staticDNSRecordResource struct {
	client *api.Client
}

type staticDNSRecordResourceModel struct {
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

func (r *staticDNSRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *staticDNSRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_dns_record"
}

// Schema defines the schema for the resource.
func (r *staticDNSRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
func (r *staticDNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan staticDNSRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	record := api.StaticDNSRecord{
		Key:        plan.Key.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		Value:      plan.Value.ValueString(),
	}
	if !plan.Enabled.IsNull() {
		record.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.Port.IsNull() {
		record.Port = int(plan.Port.ValueInt32())
	}
	if !plan.Priority.IsNull() {
		record.Priority = int(plan.Priority.ValueInt32())
	}
	if !plan.TTL.IsNull() {
		record.TTL = int(plan.TTL.ValueInt32())
	}
	if !plan.Weight.IsNull() {
		record.Weight = int(plan.Weight.ValueInt32())
	}

	// create the record
	createdRecord, err := r.client.CreateStaticDNSRecord(ctx, record)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Create Static DNS Record",
			fmt.Sprintf("Failed to create static DNS record using the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	plan.ID = types.StringValue(createdRecord.ID)
	plan.Enabled = types.BoolValue(createdRecord.Enabled)
	plan.Key = types.StringValue(createdRecord.Key)
	plan.Port = types.Int32Value(int32(createdRecord.Port))
	plan.Priority = types.Int32Value(int32(createdRecord.Priority))
	plan.RecordType = types.StringValue(createdRecord.RecordType)
	plan.TTL = types.Int32Value(int32(createdRecord.TTL))
	plan.Value = types.StringValue(createdRecord.Value)
	plan.Weight = types.Int32Value(int32(createdRecord.Weight))

	// save state with the populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *staticDNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// retrieve the current state
	var state staticDNSRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh value from the API
	record, err := r.client.GetStaticDNSRecord(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Retrieve Static DNS Record",
			fmt.Sprintf("Failed to retrieve the static DNS record with the ID '%s': %s",
				state.ID.ValueString(), err.Error()),
		)
		return
	}

	// update the state
	state.Enabled = types.BoolValue(record.Enabled)
	state.Key = types.StringValue(record.Key)
	state.Port = types.Int32Value(int32(record.Port))
	state.Priority = types.Int32Value(int32(record.Priority))
	state.RecordType = types.StringValue(record.RecordType)
	state.TTL = types.Int32Value(int32(record.TTL))
	state.Value = types.StringValue(record.Value)
	state.Weight = types.Int32Value(int32(record.Weight))

	// set the refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *staticDNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan staticDNSRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	record := api.StaticDNSRecord{
		Key:        plan.Key.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		Value:      plan.Value.ValueString(),
	}
	if !plan.Enabled.IsNull() {
		record.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.Port.IsNull() {
		record.Port = int(plan.Port.ValueInt32())
	}
	if !plan.Priority.IsNull() {
		record.Priority = int(plan.Priority.ValueInt32())
	}
	if !plan.TTL.IsNull() {
		record.TTL = int(plan.TTL.ValueInt32())
	}
	if !plan.Weight.IsNull() {
		record.Weight = int(plan.Weight.ValueInt32())
	}

	// create the record
	updatedRecord, err := r.client.UpdateStaticDNSRecord(ctx, plan.ID.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Update Static DNS Record",
			fmt.Sprintf("Failed to update static DNS record using the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	plan.ID = types.StringValue(updatedRecord.ID)
	plan.Enabled = types.BoolValue(updatedRecord.Enabled)
	plan.Key = types.StringValue(updatedRecord.Key)
	plan.Port = types.Int32Value(int32(updatedRecord.Port))
	plan.Priority = types.Int32Value(int32(updatedRecord.Priority))
	plan.RecordType = types.StringValue(updatedRecord.RecordType)
	plan.TTL = types.Int32Value(int32(updatedRecord.TTL))
	plan.Value = types.StringValue(updatedRecord.Value)
	plan.Weight = types.Int32Value(int32(updatedRecord.Weight))

	// save state with the populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *staticDNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// retrieve the current state
	var state staticDNSRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// delete the record
	if err := r.client.DeleteStaticDNSRecord(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Update Static DNS Record",
			fmt.Sprintf("Failed to update static DNS record using the UDM API:\n\t%s", err.Error()),
		)
		return
	}
}

func (r *staticDNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
