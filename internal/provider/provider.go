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
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/data"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/datasources"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/resources"
)

// ensure SingularityProvider satisfies various provider interfaces.
var _ provider.Provider = &SingularityProvider{}

// SingularityProviderModel describes the provider data model.
type SingularityProviderModel struct {
	// ApiToken contains the API token used to interact with the REST API.
	ApiToken types.String `tfsdk:"api_token"`

	// ApiEndpoint contains the hostname used in the base URL for querying the REST API.
	ApiEndpoint types.String `tfsdk:"api_endpoint"`
}

// SingularityProvider defines the provider implementation.
type SingularityProvider struct {
	// NOTE: we do not have the REST API client here because in certain cases it is needed before it is available
	//       to data sources / resources so a globally accessible singleton is used instead.
}

// New creates a new instance of the provider.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &SingularityProvider{}
	}
}

// Metadata returns metadata about the provider.
func (p *SingularityProvider) Metadata(ctx context.Context, req provider.MetadataRequest,
	resp *provider.MetadataResponse) {

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
func (p *SingularityProvider) Configure(ctx context.Context, req provider.ConfigureRequest,
	resp *provider.ConfigureResponse) {

	// environment variables take precedence over configuration variables
	apiToken := os.Getenv("SINGULARITY_API_TOKEN")
	apiEndpoint := os.Getenv("SINGULARITY_API_ENDPOINT")

	// read configuration
	var config SingularityProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if apiToken == "" {
		apiToken = config.ApiToken.ValueString()
	}
	if apiEndpoint == "" {
		apiEndpoint = config.ApiEndpoint.ValueString()
	}

	// check required configuration variables
	if apiToken == "" {
		msg := "While configuring the provider, the API token was not found in the " +
			"SINGULARITY_API_TOKEN environment variable nor was it defined in the " +
			"provider configuration block's 'api_token' attribute."
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_PROVIDER_CONFIGURE,
		})
		resp.Diagnostics.AddError("Missing API Token Configuration", msg)
	}
	if apiEndpoint == "" {
		msg := "While configuring the provider, the API endpoint was not found in the " +
			"SINGULARITY_API_ENDPOINT environment variable nor was it defined in the " +
			"provider configuration block's 'api_endpoint' attribute."
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_PROVIDER_CONFIGURE,
		})
		resp.Diagnostics.AddError("Missing API Endpoint Configuration", msg)
	}

	// share the configuration with resources and data sources
	d := &data.SingularityProvider{}
	resp.DataSourceData = d
	resp.ResourceData = d

	// initialize the global REST API client singleton
	// NOTE: we are not storing the API client in the provider because in some instances the client may be needed before
	//       the provider data is available to the specific data source or resource
	api.Client().Init(apiEndpoint, apiToken)
	tflog.Debug(ctx, "REST API client has been initialized.")
}

// DataSources defines the various data sources from which the provider can read data.
func (p *SingularityProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewGroup,
		datasources.NewGroups,
		datasources.NewPackage,
		datasources.NewPackages,
		datasources.NewSite,
		datasources.NewSites,
	}
}

// Resources defines the various resources that the provider can create.
func (p *SingularityProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewK8sAgentPackageLoader,
		resources.NewPackageDownload,
	}
}
