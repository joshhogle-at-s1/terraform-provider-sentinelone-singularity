package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/validators"
)

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Sites{}
	_ datasource.DataSourceWithConfigure = &Sites{}
)

// apiSitesModel defines the API model for a list of sites.
type apiSitesModel struct {
	AllSites apiAllSitesModel `json:"all_sites"`
	Sites    []apiSiteModel   `json:"sites"`
}

// apiAllSitesModel defines the API model for metadata about all sites returned in a request.
type apiAllSitesModel struct {
	ActiveLicenses int `json:"active_licenses"`
	TotalLicenses  int `json:"total_licenses"`
}

// tfSitesModel defines the Terraform model for sites.
type tfSitesModel struct {
	Sites  []tfSiteModel       `tfsdk:"sites"`
	Filter *tfSitesModelFilter `tfsdk:"filter"`
}

// tfSitesModelFilter defines the Terraform model for site filtering.
type tfSitesModelFilter struct {
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
	client *client.SingularityProvider
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
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.SingularityProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Type",
			fmt.Sprintf("Expected *client.SingularityProvider, got: %T. Please report this issue to the provider "+
				"developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

// Read retrieves data from the API.
func (d *Sites) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfSitesModel

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := map[string]string{}
	if data.Filter != nil {
		queryParams = d.queryParamsFromFilter(*data.Filter)
	}

	// find all matching sites querying for additional pages until results are exhausted
	var sites apiSitesModel
	for {
		// get a page of results
		result, diag := d.client.APIClient.Get(ctx, "/sites", queryParams)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		// parse the response
		var page apiSitesModel
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"Site object.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
			resp.Diagnostics.AddError("API Query Failed", msg)
			return
		}
		sites.Sites = append(sites.Sites, page.Sites...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		queryParams["cursor"] = result.Pagination.NextCursor
	}

	// convert API objects into Terraform objects
	tfsites := tfSitesModel{
		Filter: data.Filter,
		Sites:  []tfSiteModel{},
	}
	for _, site := range sites.Sites {
		tfsites.Sites = append(tfsites.Sites, terraformSiteFromAPI(ctx, site))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfsites)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Sites) queryParamsFromFilter(filter tfSitesModelFilter) map[string]string {
	queryParams := map[string]string{}

	if len(filter.AccountIds) > 0 {
		values := []string{}
		for _, e := range filter.AccountIds {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["accountIds"] = strings.Join(values, ",")
	}

	if len(filter.AccountNameContains) > 0 {
		values := []string{}
		for _, e := range filter.AccountNameContains {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["accountName__contains"] = strings.Join(values, ",")
	}

	if !filter.ActiveLicenses.IsNull() && !filter.ActiveLicenses.IsUnknown() {
		queryParams["activeLicenses"] = fmt.Sprintf("%d", filter.ActiveLicenses.ValueInt64())
	}

	if !filter.AdminOnly.IsNull() && !filter.AdminOnly.IsUnknown() {
		queryParams["adminOnly"] = fmt.Sprintf("%t", filter.AdminOnly.ValueBool())
	}

	if !filter.AvailableMoveSites.IsNull() && !filter.AvailableMoveSites.IsUnknown() {
		queryParams["availableMoveSites"] = fmt.Sprintf("%t", filter.AvailableMoveSites.ValueBool())
	}

	if !filter.CreatedAt.IsNull() && !filter.CreatedAt.IsUnknown() {
		queryParams["createdAt"] = filter.CreatedAt.ValueString()
	}

	if !filter.Description.IsNull() && !filter.Description.IsUnknown() {
		queryParams["description"] = filter.Description.ValueString()
	}

	if len(filter.DescriptionContains) > 0 {
		values := []string{}
		for _, e := range filter.DescriptionContains {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["description__contains"] = strings.Join(values, ",")
	}

	if !filter.Expiration.IsNull() && !filter.Expiration.IsUnknown() {
		queryParams["expiration"] = filter.Expiration.ValueString()
	}

	if !filter.ExternalId.IsNull() && !filter.ExternalId.IsUnknown() {
		queryParams["externalId"] = filter.ExternalId.ValueString()
	}

	if len(filter.Features) > 0 {
		values := []string{}
		for _, e := range filter.Features {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["features"] = strings.Join(values, ",")
	}

	if !filter.IsDefault.IsNull() && !filter.IsDefault.IsUnknown() {
		queryParams["isDefault"] = fmt.Sprintf("%t", filter.IsDefault.ValueBool())
	}

	if len(filter.Modules) > 0 {
		values := []string{}
		for _, e := range filter.Modules {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["modules"] = strings.Join(values, ",")
	}

	if !filter.Name.IsNull() && !filter.Name.IsUnknown() {
		queryParams["name"] = filter.Name.ValueString()
	}

	if len(filter.NameContains) > 0 {
		values := []string{}
		for _, e := range filter.NameContains {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["name__contains"] = strings.Join(values, ",")
	}

	if !filter.Query.IsNull() && !filter.Query.IsUnknown() {
		queryParams["query"] = filter.Query.ValueString()
	}

	if !filter.RegistrationToken.IsNull() && !filter.RegistrationToken.IsUnknown() {
		queryParams["registrationToken"] = filter.RegistrationToken.ValueString()
	}

	if len(filter.SiteIds) > 0 {
		values := []string{}
		for _, e := range filter.SiteIds {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["siteIds"] = strings.Join(values, ",")
	}

	if !filter.SiteType.IsNull() && !filter.SiteType.IsUnknown() {
		queryParams["siteType"] = filter.SiteType.ValueString()
	}

	if !filter.SortBy.IsNull() && !filter.SortBy.IsUnknown() {
		queryParams["sortBy"] = filter.SortBy.ValueString()
	}

	if !filter.SortOrder.IsNull() && !filter.SortOrder.IsUnknown() {
		queryParams["sortOrder"] = filter.SortOrder.ValueString()
	}

	if len(filter.States) > 0 {
		values := []string{}
		for _, e := range filter.States {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["states"] = strings.Join(values, ",")
	}

	if !filter.TotalLicenses.IsNull() && !filter.TotalLicenses.IsUnknown() {
		queryParams["totalLicenses"] = fmt.Sprintf("%d", filter.TotalLicenses.ValueInt64())
	}

	if !filter.UpdatedAt.IsNull() && !filter.UpdatedAt.IsUnknown() {
		queryParams["updatedAt"] = filter.UpdatedAt.ValueString()
	}
	return queryParams
}
