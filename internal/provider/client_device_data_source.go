package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.joshhogle.dev/terraform-provider-ubiquiti-udm/internal/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clientDevicesDataSource{}
	_ datasource.DataSourceWithConfigure = &clientDevicesDataSource{}
)

func NewClientDevicesDataSource() datasource.DataSource {
	return &clientDevicesDataSource{}
}

type clientDevicesDataSource struct {
	client *api.Client
}

type clientDevicesDataSourceModel struct {
	Devices []clientDeviceDataSourceModel      `tfsdk:"devices"`
	Filter  *clientDeviceFilterDataSourceModel `tfsdk:"filter"`
}

type clientDeviceFilterDataSourceModel struct {
	FixedIP               types.String `tfsdk:"fixed_ip"`
	HardwareAddress       types.String `tfsdk:"mac_address"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	LocalDNSRecord        types.String `tfsdk:"local_dns_record"`
	LocalDNSRecordEnabled types.Bool   `tfsdk:"local_dns_record_enabled"`
	UseFixedIP            types.Bool   `tfsdk:"use_fixed_ip"`
}

type clientDeviceDataSourceModel struct {
	FixedIP               types.String `tfsdk:"fixed_ip"`
	HardwareAddress       types.String `tfsdk:"mac_address"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	LocalDNSRecord        types.String `tfsdk:"local_dns_record"`
	LocalDNSRecordEnabled types.Bool   `tfsdk:"local_dns_record_enabled"`
	UseFixedIP            types.Bool   `tfsdk:"use_fixed_ip"`
}

func (d *clientDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse) {

	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *clientDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest,
	resp *datasource.MetadataResponse) {

	resp.TypeName = req.ProviderTypeName + "_client_devices"
}

func (d *clientDevicesDataSource) Schema(_ context.Context, req datasource.SchemaRequest,
	resp *datasource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"devices": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"fixed_ip": schema.StringAttribute{
							Computed: true,
						},
						"mac_address": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"local_dns_record": schema.StringAttribute{
							Computed: true,
						},
						"local_dns_record_enabled": schema.BoolAttribute{
							Computed: true,
						},
						"use_fixed_ip": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"filter": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Optional: true,
					},
					"fixed_ip": schema.StringAttribute{
						Optional: true,
					},
					"mac_address": schema.StringAttribute{
						Optional: true,
					},
					"name": schema.StringAttribute{
						Optional: true,
					},
					"local_dns_record": schema.StringAttribute{
						Optional: true,
					},
					"local_dns_record_enabled": schema.BoolAttribute{
						Optional: true,
					},
					"use_fixed_ip": schema.BoolAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (d *clientDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest,
	resp *datasource.ReadResponse) {

	// read configuration
	var config clientDevicesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// query for all devices
	devices, err := d.client.GetClientDevices(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Retrieve Client Device Devices",
			fmt.Sprintf("Failed to retrieve client devices from the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	state := clientDevicesDataSourceModel{
		Devices: []clientDeviceDataSourceModel{},
		Filter:  config.Filter,
	}
	for _, device := range devices {
		if config.Filter != nil {
			// filter non-matching fixed IP devices
			if !config.Filter.FixedIP.IsNull() && device.FixedIP != config.Filter.FixedIP.ValueString() {
				continue
			}

			// filter non-matching hardware address devices
			if !config.Filter.HardwareAddress.IsNull() &&
				device.HardwareAddress != config.Filter.HardwareAddress.ValueString() {
				continue
			}

			// filter non-matching IDs
			if !config.Filter.ID.IsNull() && device.ID != config.Filter.ID.ValueString() {
				continue
			}

			// filter non-matching names
			if !config.Filter.Name.IsNull() && device.Name != config.Filter.Name.ValueString() {
				continue
			}

			// filter non-matching local DNS records
			if !config.Filter.LocalDNSRecord.IsNull() &&
				device.LocalDNSRecord != config.Filter.LocalDNSRecord.ValueString() {
				continue
			}

			// filter non-matching local DNS record status devices
			if !config.Filter.LocalDNSRecordEnabled.IsNull() &&
				device.LocalDNSRecordEnabled != config.Filter.LocalDNSRecordEnabled.ValueBool() {
				continue
			}

			// filter non-matching use fixed IP status devices
			if !config.Filter.UseFixedIP.IsNull() &&
				device.UseFixedIP != config.Filter.UseFixedIP.ValueBool() {
				continue
			}
		}

		deviceState := clientDeviceDataSourceModel{
			FixedIP:               types.StringValue(device.FixedIP),
			HardwareAddress:       types.StringValue(device.HardwareAddress),
			ID:                    types.StringValue(device.ID),
			Name:                  types.StringValue(device.Name),
			LocalDNSRecord:        types.StringValue(device.LocalDNSRecord),
			LocalDNSRecordEnabled: types.BoolValue(device.LocalDNSRecordEnabled),
			UseFixedIP:            types.BoolValue(device.UseFixedIP),
		}
		state.Devices = append(state.Devices, deviceState)
	}

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
