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
	_ datasource.DataSource              = &Groups{}
	_ datasource.DataSourceWithConfigure = &Groups{}
)

// tfGroupsModel defines the Terraform model for groups.
type tfGroupsModel struct {
	Groups []tfGroupModel       `tfsdk:"groups"`
	Filter *tfGroupsModelFilter `tfsdk:"filter"`
}

// tfGroupsModelFilter defines the Terraform model for group filtering.
type tfGroupsModelFilter struct {
	AccountIds        []types.String `tfsdk:"account_ids"`
	Description       types.String   `tfsdk:"description"`
	GroupIds          []types.String `tfsdk:"group_ids"`
	IsDefault         types.Bool     `tfsdk:"is_default"`
	Name              types.String   `tfsdk:"name"`
	Query             types.String   `tfsdk:"query"`
	Rank              types.Int64    `tfsdk:"rank"`
	RegistrationToken types.String   `tfsdk:"registration_token"`
	SiteIds           []types.String `tfsdk:"site_ids"`
	SortBy            types.String   `tfsdk:"sort_by"`
	SortOrder         types.String   `tfsdk:"sort_order"`
	Types             []types.String `tfsdk:"types"`
	UpdatedAfter      types.String   `tfsdk:"updated_after"`
	UpdatedAtOrAfter  types.String   `tfsdk:"updated_at_or_after"`
	UpdatedAtOrBefore types.String   `tfsdk:"updated_at_or_before"`
	UpdatedBefore     types.String   `tfsdk:"updated_before"`
}

// NewGroups creates a new Groups object.
func NewGroups() datasource.DataSource {
	return &Groups{}
}

// Groups is a data source used to store details about groups.
type Groups struct {
	client *client.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Groups) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Groups) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used for getting a list of groups based on filters.",
		MarkdownDescription: `This data source can be used for getting a list of groups based on filters.

		TODO: add more of a description on how to use this data source...
		`,
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Description:         "List of matching groups that were found.",
				MarkdownDescription: "List of matching groups that were found.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: getGroupSchema(ctx).Attributes,
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description:         "Defines the query filters to use when searching for groups.",
				MarkdownDescription: "Defines the query filters to use when searching for groups.",
				Attributes: map[string]schema.Attribute{
					"account_ids": schema.ListAttribute{
						Description:         "List of account IDs to filter by.",
						MarkdownDescription: "List of account IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"description": schema.StringAttribute{
						Description:         "Description of the group.",
						MarkdownDescription: "Description of the group.",
						Optional:            true,
					},
					"group_ids": schema.ListAttribute{
						Description:         "List of group IDs to filter by.",
						MarkdownDescription: "List of group IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"is_default": schema.BoolAttribute{
						Description:         "Whether or not the group is the default group.",
						MarkdownDescription: "Whether or not the group is the default group.",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						Description:         "Name of the group.",
						MarkdownDescription: "Name of the group.",
						Optional:            true,
					},
					"query": schema.StringAttribute{
						Description:         "A free-text search term, will match applicable attributes.",
						MarkdownDescription: "A free-text search term, will match applicable attributes.",
						Optional:            true,
					},
					"rank": schema.Int64Attribute{
						Description:         "Priority of one dynamic group over another.",
						MarkdownDescription: "Priority of one dynamic group over another.",
						Optional:            true,
					},
					"registration_token": schema.StringAttribute{
						Description:         "The registration token for the group.",
						MarkdownDescription: "The registration token for the group.",
						Optional:            true,
					},
					"site_ids": schema.ListAttribute{
						Description:         "List of site IDs to filter by.",
						MarkdownDescription: "List of site IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"sort_by": schema.StringAttribute{
						Description: "Field on which to sort results (valid values: createdAt, description, id, inherits, " +
							"name, rank, type, updatedAt).",
						MarkdownDescription: "Field on which to sort results (valid values: `createdAt`, `description`, " +
							"`id`, `inherits`, `name`, `rank`, `type`, `updatedAt`).",
						Optional: true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"createdAt", "description", "id", "inherits", "name", "rank", "type", "updatedAt",
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
					"types": schema.ListAttribute{
						Description:         "Group type (valid values: dynamic, pinned, static).",
						MarkdownDescription: "Group type (valid values: `dynamic`, `pinned`, `static`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"dynamic", "pinned", "static",
							),
						},
					},
					"updated_after": schema.StringAttribute{
						Description:         "Group was updated after the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Group was updated after the given timestamp (eg: `2023-01-01T00:00:00Z`).",
						Optional:            true,
					},
					"updated_at_or_after": schema.StringAttribute{
						Description:         "Group was updated at or after the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Group was updated at or after the given timestamp (eg: `2023-01-01T00:00:00Z`).",
						Optional:            true,
					},
					"updated_at_or_before": schema.StringAttribute{
						Description:         "Group was updated at or before the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Group was updated at or before the given timestamp (eg: `2023-01-01T00:00:00Z`).",
						Optional:            true,
					},
					"updated_before": schema.StringAttribute{
						Description:         "Group was updated before the given timestamp (eg: 2023-01-01T00:00:00Z).",
						MarkdownDescription: "Group was updated before the given timestamp (eg: `2023-01-01T00:00:00Z`).",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Configure initializes the configuration for the data source.
func (d *Groups) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *Groups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfGroupsModel

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

	// find all matching groups querying for additional pages until results are exhausted
	var groups []apiGroupModel
	for {
		// get a page of results
		result, diag := d.client.APIClient.Get(ctx, "/groups", queryParams)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		// parse the response
		var page []apiGroupModel
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"Group object.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
			resp.Diagnostics.AddError("API Query Failed", msg)
			return
		}
		groups = append(groups, page...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		queryParams["cursor"] = result.Pagination.NextCursor
	}

	// convert API objects into Terraform objects
	tfgroups := tfGroupsModel{
		Filter: data.Filter,
		Groups: []tfGroupModel{},
	}
	for _, group := range groups {
		tfgroups.Groups = append(tfgroups.Groups, terraformGroupFromAPI(ctx, group))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfgroups)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Groups) queryParamsFromFilter(filter tfGroupsModelFilter) map[string]string {
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

	if !filter.Description.IsNull() && !filter.Description.IsUnknown() {
		queryParams["description"] = filter.Description.ValueString()
	}

	if len(filter.GroupIds) > 0 {
		values := []string{}
		for _, e := range filter.GroupIds {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["groupIds"] = strings.Join(values, ",")
	}

	if !filter.IsDefault.IsNull() && !filter.IsDefault.IsUnknown() {
		queryParams["isDefault"] = fmt.Sprintf("%t", filter.IsDefault.ValueBool())
	}

	if !filter.Name.IsNull() && !filter.Name.IsUnknown() {
		queryParams["name"] = filter.Name.ValueString()
	}

	if !filter.Query.IsNull() && !filter.Query.IsUnknown() {
		queryParams["query"] = filter.Query.ValueString()
	}

	if !filter.Rank.IsNull() && !filter.Rank.IsUnknown() {
		queryParams["rank"] = fmt.Sprintf("%d", filter.Rank.ValueInt64())
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

	if !filter.SortBy.IsNull() && !filter.SortBy.IsUnknown() {
		queryParams["sortBy"] = filter.SortBy.ValueString()
	}

	if !filter.SortOrder.IsNull() && !filter.SortOrder.IsUnknown() {
		queryParams["sortOrder"] = filter.SortOrder.ValueString()
	}

	if len(filter.Types) > 0 {
		values := []string{}
		for _, e := range filter.Types {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["types"] = strings.Join(values, ",")
	}

	if !filter.UpdatedAfter.IsNull() && !filter.UpdatedAfter.IsUnknown() {
		queryParams["updatedAt__gt"] = filter.UpdatedAfter.ValueString()
	}

	if !filter.UpdatedAtOrAfter.IsNull() && !filter.UpdatedAtOrAfter.IsUnknown() {
		queryParams["updatedAt__gte"] = filter.UpdatedAtOrAfter.ValueString()
	}

	if !filter.UpdatedAtOrBefore.IsNull() && !filter.UpdatedAtOrBefore.IsUnknown() {
		queryParams["updatedAt__lte"] = filter.UpdatedAtOrBefore.ValueString()
	}

	if !filter.UpdatedBefore.IsNull() && !filter.UpdatedBefore.IsUnknown() {
		queryParams["updatedAt__lt"] = filter.UpdatedBefore.ValueString()
	}
	return queryParams
}
