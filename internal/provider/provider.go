package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/datasources"
)

// ensure SingularityProvider satisfies various provider interfaces.
var _ provider.Provider = &SingularityProvider{}

// SingularityProviderModel describes the provider data model.
type SingularityProviderModel struct {
	ApiToken    types.String `tfsdk:"api_token"`
	ApiEndpoint types.String `tfsdk:"api_endpoint"`
}

// SingularityProvider defines the provider implementation.
type SingularityProvider struct {
}

// New creates a new instance of the provider.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &SingularityProvider{}
	}
}

// Metadata returns metadata about the provider.
func (p *SingularityProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = plugin.PROVIDER_NAME
	resp.Version = plugin.Version
}

// Schema defines the parameters for the provider's configuration.
func (p *SingularityProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API key used to query the SentinelOne Singularity API",
				Optional:            true,
			},
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "The FQDN to use for all API queries, excluding 'https://'",
				Optional:            true,
			},
		},
	}
}

// Configure initializes the configuration for the provider.
func (p *SingularityProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// check environment variables
	apiToken := os.Getenv("SINGULARITY_API_TOKEN")
	apiEndpoint := os.Getenv("SINGULARITY_API_ENDPOINT")

	// read configuration into data model
	var data SingularityProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check required configuration variables
	if apiToken == "" {
		apiToken = data.ApiToken.ValueString()
		if apiToken == "" {
			msg := "While configuring the provider, the API token was not found in the " +
				"SINGULARITY_API_TOKEN environment variable nor was it defined in the " +
				"provider configuration block's 'api_token' attribute."
			tflog.Error(ctx, msg)
			resp.Diagnostics.AddError("Missing API Token Configuration", msg)
		}
	}
	if apiEndpoint == "" {
		apiEndpoint = data.ApiEndpoint.ValueString()
		if apiEndpoint == "" {
			msg := "While configuring the provider, the API endpoint was not found in the " +
				"SINGULARITY_API_ENDPOINT environment variable nor was it defined in the " +
				"provider configuration block's 'api_endpoint' attribute."
			tflog.Error(ctx, msg)
			resp.Diagnostics.AddError("Missing API Endpoint Configuration", msg)
		}
	}

	// share the configuration with resources and data sources
	client := &client.SingularityProvider{
		APIClient: api.NewClient(apiToken, apiEndpoint),
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the various data sources from which the provider can read data.
func (p *SingularityProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewPackages,
		datasources.NewPackage,
	}
}

// Resources defines the various resources that the provider can create.
func (p *SingularityProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
