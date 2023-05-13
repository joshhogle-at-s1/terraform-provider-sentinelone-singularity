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
	_ datasource.DataSource              = &Groups{}
	_ datasource.DataSourceWithConfigure = &Groups{}
)

// tfGroups defines the Terraform model for groups.
type tfGroups struct {
	Groups []tfGroup       `tfsdk:"groups"`
	Filter *tfGroupsFilter `tfsdk:"filter"`
}

// tfGroupsFilter defines the Terraform model for group filtering.
type tfGroupsFilter struct {
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
	data *data.SingularityProvider
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
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*data.SingularityProvider)
	if !ok {
		expectedType := reflect.TypeOf(&data.SingularityProvider{})
		msg := fmt.Sprintf("The provider data sent in the request does not match the type expected. This is always an "+
			"error with the provider and should be reported to the provider developers.\n\nExpected Type: %s\nData Type "+
			"Received Type: %T", expectedType, req.ProviderData)
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_DATASOURCE_GROUPS_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	d.data = providerData
}

// Read retrieves data from the API.
func (d *Groups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfGroups

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := api.GroupQueryParams{}
	if data.Filter != nil {
		queryParams = d.queryParamsFromFilter(*data.Filter)
	}

	// find the matching groups
	groups, diags := api.Client().FindGroups(ctx, queryParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert API objects into Terraform objects
	tfgroups := tfGroups{
		Filter: data.Filter,
		Groups: []tfGroup{},
	}
	for _, group := range groups {
		tfgroups.Groups = append(tfgroups.Groups, tfGroupFromAPI(ctx, &group))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfgroups)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Groups) queryParamsFromFilter(filter tfGroupsFilter) api.GroupQueryParams {
	queryParams := api.GroupQueryParams{}

	if len(filter.AccountIds) > 0 {
		queryParams.AccountIds = []string{}
		for _, e := range filter.AccountIds {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.AccountIds = append(queryParams.AccountIds, e.ValueString())
			}
		}
	}

	if !filter.Description.IsNull() && !filter.Description.IsUnknown() {
		value := filter.Description.ValueString()
		queryParams.Description = &value
	}

	if len(filter.GroupIds) > 0 {
		queryParams.GroupIds = []string{}
		for _, e := range filter.GroupIds {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.GroupIds = append(queryParams.GroupIds, e.ValueString())
			}
		}
	}

	if !filter.IsDefault.IsNull() && !filter.IsDefault.IsUnknown() {
		value := filter.IsDefault.ValueBool()
		queryParams.IsDefault = &value
	}

	if !filter.Name.IsNull() && !filter.Name.IsUnknown() {
		value := filter.Name.ValueString()
		queryParams.Name = &value
	}

	if !filter.Query.IsNull() && !filter.Query.IsUnknown() {
		value := filter.Query.ValueString()
		queryParams.Query = &value
	}

	if !filter.Rank.IsNull() && !filter.Rank.IsUnknown() {
		value := filter.Rank.ValueInt64()
		queryParams.Rank = &value
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

	if !filter.SortBy.IsNull() && !filter.SortBy.IsUnknown() {
		value := filter.SortBy.ValueString()
		queryParams.SortBy = &value
	}

	if !filter.SortOrder.IsNull() && !filter.SortOrder.IsUnknown() {
		value := filter.SortOrder.ValueString()
		queryParams.SortOrder = &value
	}

	if len(filter.Types) > 0 {
		queryParams.Types = []string{}
		for _, e := range filter.Types {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.Types = append(queryParams.Types, e.ValueString())
			}
		}
	}

	if !filter.UpdatedAfter.IsNull() && !filter.UpdatedAfter.IsUnknown() {
		value := filter.UpdatedAfter.ValueString()
		queryParams.UpdatedAfter = &value
	}

	if !filter.UpdatedAtOrAfter.IsNull() && !filter.UpdatedAtOrAfter.IsUnknown() {
		value := filter.UpdatedAtOrAfter.ValueString()
		queryParams.UpdatedAtOrAfter = &value
	}

	if !filter.UpdatedAtOrBefore.IsNull() && !filter.UpdatedAtOrBefore.IsUnknown() {
		value := filter.UpdatedAtOrBefore.ValueString()
		queryParams.UpdatedAtOrBefore = &value
	}

	if !filter.UpdatedBefore.IsNull() && !filter.UpdatedBefore.IsUnknown() {
		value := filter.UpdatedBefore.ValueString()
		queryParams.UpdatedBefore = &value
	}
	return queryParams
}
