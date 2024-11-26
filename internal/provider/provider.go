// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.joshhogle.dev/terraform-provider-ubiquiti-udm/internal/api"
)

// Ensure udmProvider satisfies various provider interfaces.
var (
	_ provider.Provider              = &udmProvider{}
	_ provider.ProviderWithFunctions = &udmProvider{}
)

// udmProvider defines the provider implementation.
type udmProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// udmProviderModel describes the provider data model.
type udmProviderModel struct {
	Hostname                      types.String `tfsdk:"hostname"`
	IgnoreUntrustedSSLCertificate types.Bool   `tfsdk:"ignore_untrusted_ssl_certificate"`
	Password                      types.String `tfsdk:"password"`
	Site                          types.String `tfsdk:"site"`
	Username                      types.String `tfsdk:"username"`
}

func (p *udmProvider) Metadata(ctx context.Context, req provider.MetadataRequest,
	resp *provider.MetadataResponse) {

	resp.TypeName = "udm"
	resp.Version = p.version
}

func (p *udmProvider) Schema(ctx context.Context, req provider.SchemaRequest,
	resp *provider.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description:         "UDM hostname or IP address",
				MarkdownDescription: "UDM hostname or IP address",
				Required:            true,
				//Validators:          []validator.String{},
			},
			"ignore_untrusted_ssl_certificate": schema.BoolAttribute{
				Description:         "Ignore any untrusted / self-signed certificate from the UDM host",
				MarkdownDescription: "Ignore any untrusted / self-signed certificate from the UDM host",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description:         "UDM password to use for authentication",
				MarkdownDescription: "UDM password to use for authentication",
				Required:            true,
				Sensitive:           true,
				//Validators:          []validator.String{},
			},
			"site": schema.StringAttribute{
				Description:         "Name of the UDM site (uses 'default' if not supplied)",
				MarkdownDescription: "Name of the UDM site (uses 'default' if not supplied)",
				Optional:            true,
				//Validators:          []validator.String{},
			},
			"username": schema.StringAttribute{
				Description:         "UDM username to use for authentication",
				MarkdownDescription: "UDM username to use for authentication",
				Required:            true,
				Sensitive:           true,
				//Validators:          []validator.String{},
			},
		},
	}
}

func (p *udmProvider) Configure(ctx context.Context, req provider.ConfigureRequest,
	resp *provider.ConfigureResponse) {

	// retrieve provider data from configuration
	var config udmProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if the caller provided a configuration value for any of the attributes, it must be a known value
	if config.Hostname.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Unknown UDM API Host",
			"The provider cannot create the UDM API client as there is an unknown configuration value for "+
				"the UDM API hostname. Either target apply the source of the value first, set the value "+
				"statically in the configuration, or use a variable in the configuration.",
		)
	}
	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown UDM API Password",
			"The provider cannot create the UDM API client as there is an unknown configuration value for "+
				"the UDM API password. Either target apply the source of the value first, set the value "+
				"statically in the configuration, or use a variable in the configuration.",
		)
	}
	if config.Site.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("site"),
			"Unknown UDM API Site",
			"The provider cannot create the UDM API client as there is an unknown configuration value for "+
				"the UDM API site. Either target apply the source of the value first, set the value "+
				"statically in the configuration, or use a variable in the configuration.",
		)
	}
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown UDM API Username",
			"The provider cannot create the UDM API client as there is an unknown configuration value for "+
				"the HashiCups API username. Either target apply the source of the value first, set the value "+
				"statically in the configuration, or use a variable in the configuration.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// if any of the configurations are missing, return errors with guidance
	hostname := config.Hostname.ValueString()
	if hostname == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Missing UDM API Host",
			"The provider cannot create the UDM API client as there is a missing or empty value for "+
				"the UDM API hostname. Set the value statically in the configuration or use a variable in the "+
				"configuration. If either is set already, ensure the value is not empty.",
		)
	}
	password := config.Password.ValueString()
	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing UDM API Password",
			"The provider cannot create the UDM API client as there is a missing or empty value for "+
				"the UDM API password. Set the value statically in the configuration or use a variable in the "+
				"configuration. If either is set already, ensure the value is not empty.",
		)
	}
	site := config.Site.ValueString()
	if site == "" {
		site = api.DefaultSite
	}
	username := config.Username.ValueString()
	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing UDM API Username",
			"The provider cannot create the UDM API client as there is a missing or empty value for "+
				"the UDM API username. Set the value statically in the configuration or use a variable in the "+
				"configuration. If either is set already, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// create the API client
	client := api.NewClient(hostname, site, config.IgnoreUntrustedSSLCertificate.ValueBool())
	if err := client.Login(ctx, username, password); err != nil {
		resp.Diagnostics.AddError(
			"UDM API: Authentication Failed",
			fmt.Sprintf("Failed to authenticate to the UDM API:\n\t%s", err.Error()),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *udmProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClientDeviceResource,
		NewStaticDNSRecordResource,
	}
}

func (p *udmProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClientDevicesDataSource,
		NewStaticDNSRecordsDataSource,
	}
}

func (p *udmProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &udmProvider{
			version: version,
		}
	}
}
