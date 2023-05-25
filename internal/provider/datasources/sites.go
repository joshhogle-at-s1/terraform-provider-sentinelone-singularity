package datasources

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/data"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/validators"
)

// ensure implementation satisfied expected interfaces
var (
	_ datasource.DataSource              = &Sites{}
	_ datasource.DataSourceWithConfigure = &Sites{}
)

// tfSites defines the Terraform model for sites.
type tfSites struct {
	Sites  []tfSite       `tfsdk:"sites"`
	Filter *tfSitesFilter `tfsdk:"filter"`
}

// tfSitesFilter defines the Terraform model for site filtering.
type tfSitesFilter struct {
	AccountIds          []types.String `tfsdk:"account_ids"`
	AccountNameContains []types.String `tfsdk:"account_name_contains"`
	ActiveLicenses      types.Int64    `tfsdk:"active_licenses"`
	AdminOnly           types.Bool     `tfsdk:"admin_only"`
	AvailableMoveSites  types.Bool     `tfsdk:"available_move_sites"`
	CreatedAt           types.String   `tfsdk:"created_at"`
	Description         types.String   `tfsdk:"description"`
	DescriptionContains []types.String `tfsdk:"description_contains"`
	Expiration          types.String   `tfsdk:"expiration"`
	ExternalId          types.String   `tfsdk:"external_id"`
	Features            []types.String `tfsdk:"features"`
	IsDefault           types.Bool     `tfsdk:"is_default"`
	Modules             []types.String `tfsdk:"modules"`
	Name                types.String   `tfsdk:"name"`
	NameContains        []types.String `tfsdk:"name_contains"`
	Query               types.String   `tfsdk:"query"`
	RegistrationToken   types.String   `tfsdk:"registration_token"`
	SiteIds             []types.String `tfsdk:"site_ids"`
	SiteType            types.String   `tfsdk:"site_type"`
	SortBy              types.String   `tfsdk:"sort_by"`
	SortOrder           types.String   `tfsdk:"sort_order"`
	States              []types.String `tfsdk:"states"`
	TotalLicenses       types.Int64    `tfsdk:"total_licenses"`
	UpdatedAt           types.String   `tfsdk:"updated_at"`
}

// NewSites creates a new Sites object.
func NewSites() datasource.DataSource {
	return &Sites{}
}

// Sites is a data source used to store details about sites.
type Sites struct {
	data *data.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Sites) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Sites) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used for getting a list of sites based on filters.",
		MarkdownDescription: `This data source can be used for getting a list of sites based on filters.

		TODO: add more of a description on how to use this data source...
		`,
		Attributes: map[string]schema.Attribute{
			"sites": schema.ListNestedAttribute{
				Description:         "List of matching sites that were found.",
				MarkdownDescription: "List of matching sites that were found.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: getSiteSchema(ctx).Attributes,
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description:         "Defines the query filters to use when searching for sites.",
				MarkdownDescription: "Defines the query filters to use when searching for sites.",
				Attributes: map[string]schema.Attribute{
					"account_ids": schema.ListAttribute{
						Description:         "List of account IDs to filter by.",
						MarkdownDescription: "List of account IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"account_name_contains": schema.ListAttribute{
						Description:         "Free-text filter by account name.",
						MarkdownDescription: "Free-text filter by account name.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"active_licenses": schema.Int64Attribute{
						Description:         "Number of active licenses tied to the site.",
						MarkdownDescription: "Number of active licenses tied to the site.",
						Optional:            true,
					},
					"admin_only": schema.BoolAttribute{
						Description:         "Only return sites the user has admin privileges to.",
						MarkdownDescription: "Only return sites the user has admin privileges to.",
						Optional:            true,
					},
					"available_move_sites": schema.BoolAttribute{
						Description:         "Only return sites the user can move agents to.",
						MarkdownDescription: "Only return sites the user can move agents to.",
						Optional:            true,
					},
					"created_at": schema.StringAttribute{
						Description:         "Site was created at the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Site was created at the given timestamp (eg: 2023-01-01T00:00:00Z).",
						Optional:            true,
					},
					"description": schema.StringAttribute{
						Description:         "Description of the site.",
						MarkdownDescription: "Description of the site.",
						Optional:            true,
					},
					"description_contains": schema.ListAttribute{
						Description:         "Free-text filter by description.",
						MarkdownDescription: "Free-text filter by description.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"expiration": schema.StringAttribute{
						Description:         "Site expires at the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Site expires at the given timestamp (eg: 2023-01-01T00:00:00Z).",
						Optional:            true,
					},
					"external_id": schema.StringAttribute{
						Description:         "ID of site in external CRM system.",
						MarkdownDescription: "ID of site in external CRM system.",
						Optional:            true,
					},
					"features": schema.ListAttribute{
						Description: "Only return sites with the given features (valid values: device-control, " +
							"firewall-control, ioc).",
						MarkdownDescription: "Only return sites with the given features (valid values: `device-control`, " +
							"`firewall-control`, `ioc`).",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"device-control", "firewall-control", "ioc",
							),
						},
					},
					"is_default": schema.BoolAttribute{
						Description:         "Whether or not the site is the default site.",
						MarkdownDescription: "Whether or not the site is the default site.",
						Optional:            true,
					},
					"modules": schema.ListAttribute{
						Description:         "Only return sites licensed for the given modules (eg: star, rso)",
						MarkdownDescription: "Only return sites licensed for the given modules (eg: `star`, `rso`).",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"name": schema.StringAttribute{
						Description:         "Name of the site.",
						MarkdownDescription: "Name of the site.",
						Optional:            true,
					},
					"name_contains": schema.ListAttribute{
						Description:         "Free-text filter by name.",
						MarkdownDescription: "Free-text filter by name.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"query": schema.StringAttribute{
						Description:         "A free-text search term, will match applicable attributes.",
						MarkdownDescription: "A free-text search term, will match applicable attributes.",
						Optional:            true,
					},
					"registration_token": schema.StringAttribute{
						Description:         "The registration token for the site.",
						MarkdownDescription: "The registration token for the site.",
						Optional:            true,
					},
					"site_ids": schema.ListAttribute{
						Description:         "List of site IDs to filter by.",
						MarkdownDescription: "List of site IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"site_type": schema.StringAttribute{
						Description:         "Type of site (valid values: trial, paid).",
						MarkdownDescription: "Type of site (valid values: `trial`, `paid`).",
						Optional:            true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false, "trial", "paid"),
						},
					},
					"sort_by": schema.StringAttribute{
						Description: "Field on which to sort results (valid values: accountName, activeLicenses, " +
							"createdAt, description, expiration, id, name, siteType, state, totalLicenses, updatedAt).",
						MarkdownDescription: "Field on which to sort results (valid values: `accountName`, `activeLicenses`, " +
							"`createdAt`, `description`, `expiration`, `id`, `name`, `siteType`, `state`, `totalLicenses`, " +
							"`updatedAt`).",
						Optional: true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"accountName", "activeLicenses", "createdAt", "description", "expiration", "id", "name",
								"siteType", "state", "totalLicenses", "updatedAt",
							),
						},
					},
					"sort_order": schema.StringAttribute{
						Description:         "Order in which to sort results (valid values: asc, desc).",
						MarkdownDescription: "Order in which to sort results (valid values: `asc`, `desc`).",
						Optional:            true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"asc", "desc",
							),
						},
					},
					"states": schema.ListAttribute{
						Description:         "State of the site (valid values: active, deleted, expired).",
						MarkdownDescription: "State of the site (valid values: `active`, `deleted`, `expired`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"active", "deleted", "expired",
							),
						},
					},
					"total_licenses": schema.Int64Attribute{
						Description:         "Total number of licenses associated with the site.",
						MarkdownDescription: "Total number of licenses associated with the site.",
						Optional:            true,
					},
					"updated_at": schema.StringAttribute{
						Description:         "Site was updated at the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Site was updated at the given timestamp (eg: `2023-01-01T00:00:00Z`).",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Configure initializes the configuration for the data source.
func (d *Sites) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*data.SingularityProvider)
	if !ok {
		expectedType := reflect.TypeOf(&data.SingularityProvider{})
		msg := fmt.Sprintf("The provider data sent in the request does not match the type expected. This is always an "+
			"error with the provider and should be reported to the provider developers.\n\nExpected Type: %s\nData Type "+
			"Received: %T", expectedType, req.ProviderData)
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_DATASOURCE_SITES_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	d.data = providerData
}

// Read retrieves data from the API.
func (d *Sites) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfSites

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := api.SiteQueryParams{}
	if data.Filter != nil {
		queryParams = d.queryParamsFromFilter(*data.Filter)
	}

	// find the matching sites
	sites, diags := api.Client().FindSites(ctx, queryParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert API objects into Terraform objects
	tfsites := tfSites{
		Filter: data.Filter,
		Sites:  []tfSite{},
	}
	for _, site := range sites {
		tfsites.Sites = append(tfsites.Sites, tfSiteFromAPI(ctx, &site))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfsites)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Sites) queryParamsFromFilter(filter tfSitesFilter) api.SiteQueryParams {
	queryParams := api.SiteQueryParams{}

	if len(filter.AccountIds) > 0 {
		queryParams.AccountIds = []string{}
		for _, e := range filter.AccountIds {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.AccountIds = append(queryParams.AccountIds, e.ValueString())
			}
		}

	}

	if len(filter.AccountNameContains) > 0 {
		queryParams.AccountNameContains = []string{}
		for _, e := range filter.AccountNameContains {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.AccountNameContains = append(queryParams.AccountNameContains, e.ValueString())
			}
		}
	}

	if !filter.ActiveLicenses.IsNull() && !filter.ActiveLicenses.IsUnknown() {
		value := filter.ActiveLicenses.ValueInt64()
		queryParams.ActiveLicenses = &value
	}

	if !filter.AdminOnly.IsNull() && !filter.AdminOnly.IsUnknown() {
		value := filter.AdminOnly.ValueBool()
		queryParams.AdminOnly = &value
	}

	if !filter.AvailableMoveSites.IsNull() && !filter.AvailableMoveSites.IsUnknown() {
		value := filter.AvailableMoveSites.ValueBool()
		queryParams.AvailableMoveSites = &value
	}

	if !filter.CreatedAt.IsNull() && !filter.CreatedAt.IsUnknown() {
		value := filter.CreatedAt.ValueString()
		queryParams.CreatedAt = &value
	}

	if !filter.Description.IsNull() && !filter.Description.IsUnknown() {
		value := filter.Description.ValueString()
		queryParams.Description = &value
	}

	if len(filter.DescriptionContains) > 0 {
		queryParams.DescriptionContains = []string{}
		for _, e := range filter.DescriptionContains {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.DescriptionContains = append(queryParams.DescriptionContains, e.ValueString())
			}
		}
	}

	if !filter.Expiration.IsNull() && !filter.Expiration.IsUnknown() {
		value := filter.Expiration.ValueString()
		queryParams.Expiration = &value
	}

	if !filter.ExternalId.IsNull() && !filter.ExternalId.IsUnknown() {
		value := filter.ExternalId.ValueString()
		queryParams.ExternalId = &value
	}

	if len(filter.Features) > 0 {
		queryParams.Features = []string{}
		for _, e := range filter.Features {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.Features = append(queryParams.Features, e.ValueString())
			}
		}
	}

	if !filter.IsDefault.IsNull() && !filter.IsDefault.IsUnknown() {
		value := filter.IsDefault.ValueBool()
		queryParams.IsDefault = &value
	}

	if len(filter.Modules) > 0 {
		queryParams.Modules = []string{}
		for _, e := range filter.Modules {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.Modules = append(queryParams.Modules, e.ValueString())
			}
		}
	}

	if !filter.Name.IsNull() && !filter.Name.IsUnknown() {
		value := filter.Name.ValueString()
		queryParams.Name = &value
	}

	if len(filter.NameContains) > 0 {
		queryParams.NameContains = []string{}
		for _, e := range filter.NameContains {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.NameContains = append(queryParams.NameContains, e.ValueString())
			}
		}
	}

	if !filter.Query.IsNull() && !filter.Query.IsUnknown() {
		value := filter.Query.ValueString()
		queryParams.Query = &value
	}

	if !filter.RegistrationToken.IsNull() && !filter.RegistrationToken.IsUnknown() {
		value := filter.RegistrationToken.ValueString()
		queryParams.RegistrationToken = &value
	}

	if len(filter.SiteIds) > 0 {
		queryParams.SiteIds = []string{}
		for _, e := range filter.SiteIds {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.SiteIds = append(queryParams.SiteIds, e.ValueString())
			}
		}
	}

	if !filter.SiteType.IsNull() && !filter.SiteType.IsUnknown() {
		value := filter.SiteType.ValueString()
		queryParams.SiteType = &value
	}

	if !filter.SortBy.IsNull() && !filter.SortBy.IsUnknown() {
		value := filter.SortBy.ValueString()
		queryParams.SortBy = &value
	}

	if !filter.SortOrder.IsNull() && !filter.SortOrder.IsUnknown() {
		value := filter.SortOrder.ValueString()
		queryParams.SortOrder = &value
	}

	if len(filter.States) > 0 {
		queryParams.States = []string{}
		for _, e := range filter.States {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.States = append(queryParams.States, e.ValueString())
			}
		}
	}

	if !filter.TotalLicenses.IsNull() && !filter.TotalLicenses.IsUnknown() {
		value := filter.TotalLicenses.ValueInt64()
		queryParams.TotalLicenses = &value
	}

	if !filter.UpdatedAt.IsNull() && !filter.UpdatedAt.IsUnknown() {
		value := filter.UpdatedAt.ValueString()
		queryParams.UpdatedAt = &value
	}
	return queryParams
}
