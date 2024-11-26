package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.joshhogle.dev/terraform-provider-ubiquiti-udm/internal/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &clientDeviceResource{}
	_ resource.ResourceWithConfigure   = &clientDeviceResource{}
	_ resource.ResourceWithImportState = &clientDeviceResource{}
)

// NewClientDeviceResource is a helper function to simplify the provider implementation.
func NewClientDeviceResource() resource.Resource {
	return &clientDeviceResource{}
}

// clientDeviceResource is the resource implementation.
type clientDeviceResource struct {
	client *api.Client
}

type clientDeviceResourceModel struct {
	FixedIP               types.String `tfsdk:"fixed_ip"`
	HardwareAddress       types.String `tfsdk:"mac_address"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	LocalDNSRecord        types.String `tfsdk:"local_dns_record"`
	LocalDNSRecordEnabled types.Bool   `tfsdk:"local_dns_record_enabled"`
	UseFixedIP            types.Bool   `tfsdk:"use_fixed_ip"`
}

func (r *clientDeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *clientDeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_device"
}

// Schema defines the schema for the resource.
func (r *clientDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"fixed_ip": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"mac_address": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"local_dns_record": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"local_dns_record_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"use_fixed_ip": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *clientDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan clientDeviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// generate API request body from plan
	device := api.ClientDevice{
		HardwareAddress: plan.HardwareAddress.ValueString(),
	}
	if !plan.FixedIP.IsNull() {
		device.FixedIP = plan.FixedIP.ValueString()
	}
	if !plan.Name.IsNull() {
		device.Name = plan.Name.ValueString()
	}
	if !plan.LocalDNSRecord.IsNull() {
		device.LocalDNSRecord = plan.LocalDNSRecord.ValueString()
	}
	if !plan.LocalDNSRecordEnabled.IsNull() {
		device.LocalDNSRecordEnabled = plan.LocalDNSRecordEnabled.ValueBool()
	}
	if !plan.UseFixedIP.IsNull() {
		device.UseFixedIP = plan.UseFixedIP.ValueBool()
	}

	// create the record
	createdDevice, err := r.client.CreateClientDevice(ctx, device)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Create Client Device",
			fmt.Sprintf("Failed to create client device using the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	plan.ID = types.StringValue(createdDevice.ID)
	plan.HardwareAddress = types.StringValue(createdDevice.HardwareAddress)
	plan.FixedIP = types.StringValue(createdDevice.FixedIP)
	plan.LocalDNSRecord = types.StringValue(createdDevice.LocalDNSRecord)
	plan.LocalDNSRecordEnabled = types.BoolValue(createdDevice.LocalDNSRecordEnabled)
	plan.Name = types.StringValue(createdDevice.Name)
	plan.UseFixedIP = types.BoolValue(createdDevice.UseFixedIP)

	// save state with the populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *clientDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// retrieve the current state
	var state clientDeviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh value from the API
	device, err := r.client.GetClientDevice(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Retrieve Client Device",
			fmt.Sprintf("Failed to retrieve the client device with the ID '%s': %s",
				state.ID.ValueString(), err.Error()),
		)
		return
	}

	// update the state
	state.HardwareAddress = types.StringValue(device.HardwareAddress)
	state.FixedIP = types.StringValue(device.FixedIP)
	state.LocalDNSRecord = types.StringValue(device.LocalDNSRecord)
	state.LocalDNSRecordEnabled = types.BoolValue(device.LocalDNSRecordEnabled)
	state.Name = types.StringValue(device.Name)
	state.UseFixedIP = types.BoolValue(device.UseFixedIP)

	// set the refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *clientDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *clientDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *clientDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
