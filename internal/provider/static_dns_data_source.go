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
	_ datasource.DataSource              = &staticDNSEntriesDataSource{}
	_ datasource.DataSourceWithConfigure = &staticDNSEntriesDataSource{}
)

func NewStaticDNSEntriesDataSource() datasource.DataSource {
	return &staticDNSEntriesDataSource{}
}

type staticDNSEntriesDataSource struct {
	client *api.Client
}

type staticDNSEntriesDataSourceModel struct {
	Entries []staticDNSEntryDataSourceModel `tfsdk:"entries"`
	Filter  *staticDNSFilterDataSourceModel `tfsdk:"filter"`
}

type staticDNSFilterDataSourceModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	ID         types.String `tfsdk:"id"`
	Key        types.String `tfsdk:"key"`
	RecordType types.String `tfsdk:"record_type"`
	Value      types.String `tfsdk:"value"`
}

type staticDNSEntryDataSourceModel struct {
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

func (d *staticDNSEntriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *staticDNSEntriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_dns_entries"
}

func (d *staticDNSEntriesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"entries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"key": schema.StringAttribute{
							Computed: true,
						},
						"port": schema.Int32Attribute{
							Computed: true,
						},
						"priority": schema.Int32Attribute{
							Computed: true,
						},
						"record_type": schema.StringAttribute{
							Computed: true,
						},
						"ttl": schema.Int32Attribute{
							Computed: true,
						},
						"value": schema.StringAttribute{
							Computed: true,
						},
						"weight": schema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
			"filter": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional: true,
					},
					"id": schema.StringAttribute{
						Optional: true,
					},
					"key": schema.StringAttribute{
						Optional: true,
					},
					"record_type": schema.StringAttribute{
						Optional: true,
					},
					"value": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (d *staticDNSEntriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// read configuration
	var config staticDNSEntriesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// query for all entries
	entries, err := d.client.GetStaticDNSEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Failed to Retrieve Static DNS Entries",
			fmt.Sprintf("Failed to retrieve static DNS entries from the UDM API:\n\t%s", err.Error()),
		)
		return
	}

	// map the response to the model
	state := staticDNSEntriesDataSourceModel{
		Entries: []staticDNSEntryDataSourceModel{},
		Filter:  config.Filter,
	}
	for _, entry := range entries {
		if config.Filter != nil {
			// filter (en/dis)abled entries
			if !config.Filter.Enabled.IsNull() && entry.Enabled != config.Filter.Enabled.ValueBool() {
				continue

			}
			// filter non-matching IDs
			if !config.Filter.ID.IsNull() && entry.ID != config.Filter.ID.ValueString() {
				continue
			}

			// filter non-matching keys
			if !config.Filter.Key.IsNull() && entry.Key != config.Filter.Key.ValueString() {
				continue
			}

			// filter non-matching record types
			if !config.Filter.RecordType.IsNull() && entry.RecordType != config.Filter.RecordType.ValueString() {
				continue
			}

			// filter non-matching values
			if !config.Filter.Value.IsNull() && entry.Value != config.Filter.Value.ValueString() {
				continue
			}
		}

		entryState := staticDNSEntryDataSourceModel{
			ID:         types.StringValue(entry.ID),
			Enabled:    types.BoolValue(entry.Enabled),
			Key:        types.StringValue(entry.Key),
			Port:       types.Int32Value(int32(entry.Port)),
			Priority:   types.Int32Value(int32(entry.Priority)),
			RecordType: types.StringValue(entry.RecordType),
			TTL:        types.Int32Value(int32(entry.TTL)),
			Value:      types.StringValue(entry.Value),
			Weight:     types.Int32Value(int32(entry.Weight)),
		}
		state.Entries = append(state.Entries, entryState)
	}

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
