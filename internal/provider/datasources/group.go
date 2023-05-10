package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
)

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Group{}
	_ datasource.DataSourceWithConfigure = &Group{}
)

// apiGroupModel defines the API model for a group.
type apiGroupModel struct {
	CreatedAt         string `json:"createdAt"`
	Creator           string `json:"creator"`
	CreatorId         string `json:"creatorId"`
	Description       string `json:"description"`
	FilterId          string `json:"filterId"`
	FilterName        string `json:"filterName"`
	Id                string `json:"id"`
	Inherits          bool   `json:"inherits"`
	IsDefault         bool   `json:"isDefault"`
	Name              string `json:"name"`
	Rank              int    `json:"rank"`
	RegistrationToken string `json:"registrationToken"`
	SiteId            string `json:"siteId"`
	TotalAgents       int    `json:"totalAgents"`
	Type              string `json:"type"`
	UpdatedAt         string `json:"updatedAt"`
}

// tfGroupModel defines the Terraform model for a group.
type tfGroupModel struct {
	CreatedAt         types.String `tfsdk:"created_at"`
	Creator           types.String `tfsdk:"creator"`
	CreatorId         types.String `tfsdk:"creator_id"`
	Description       types.String `tfsdk:"description"`
	FilterId          types.String `tfsdk:"filter_id"`
	FilterName        types.String `tfsdk:"filter_name"`
	Id                types.String `tfsdk:"id"`
	Inherits          types.Bool   `tfsdk:"inherits"`
	IsDefault         types.Bool   `tfsdk:"is_default"`
	Name              types.String `tfsdk:"name"`
	Rank              types.Int64  `tfsdk:"rank"`
	RegistrationToken types.String `tfsdk:"registration_token"`
	SiteId            types.String `tfsdk:"site_id"`
	TotalAgents       types.Int64  `tfsdk:"total_agents"`
	Type              types.String `tfsdk:"type"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

// NewGroup creates a new Group object.
func NewGroup() datasource.DataSource {
	return &Group{}
}

// Group is a data source used to store details about a single group.
type Group struct {
	client *client.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Group) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Group) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	groupSchema := getGroupSchema(ctx)

	// override the default schema
	groupSchema.Attributes["id"] = schema.StringAttribute{
		Description:         "ID of the group.",
		MarkdownDescription: "ID of the group.",
		Required:            true,
	}
	resp.Schema = groupSchema
}

// Configure initializes the configuration for the data source.
func (d *Group) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *Group) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfGroupModel

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := map[string]string{
		"groupIds": data.Id.ValueString(), // 'id' is required so no need to check
	}

	// find the matching group
	result, diag := d.client.APIClient.Get(ctx, "/groups", queryParams)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the response - we are expecting exactly 1 group to be returned
	numGroups := result.Pagination.TotalItems
	if numGroups == 0 {
		msg := "No matching group was found. Try expanding your search or check that your group ID is valid."
		tflog.Error(ctx, msg, map[string]interface{}{"groups_found": numGroups})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	} else if numGroups > 1 {
		msg := fmt.Sprintf("This data source expects 1 matching group but %d were found. Please narrow your search.",
			numGroups)
		tflog.Error(ctx, msg, map[string]interface{}{"groups_found": numGroups})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	}
	var groups []apiGroupModel
	if err := json.Unmarshal(result.Data, &groups); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Group object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	}

	// convert the API object to the Terraform object
	resp.Diagnostics.Append(resp.State.Set(ctx, terraformGroupFromAPI(ctx, groups[0]))...)
}

// getGroupSchema returns a default Terraform schema where all values are computed.
func getGroupSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "This data source is used for getting details on a specific group.",
		MarkdownDescription: "This data source is used for getting details on a specific group.",
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Description:         "Timestamp of when the group was created.",
				MarkdownDescription: "Timestamp of when the group was created.",
				Computed:            true,
			},
			"creator": schema.StringAttribute{
				Description:         "Full name of the user who created the group.",
				MarkdownDescription: "Full name of the user who created the group.",
				Computed:            true,
			},
			"creator_id": schema.StringAttribute{
				Description:         "ID of the user who created the group.",
				MarkdownDescription: "ID of the user who created the group.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				Description:         "User-defined description of the group.",
				MarkdownDescription: "User-defined description of the group.",
				Computed:            true,
			},
			"filter_id": schema.StringAttribute{
				Description:         "If the group is dynamic, the ID of the filter which is used to associate the agents.",
				MarkdownDescription: "If the group is dynamic, the ID of the filter which is used to associate the agents.",
				Computed:            true,
			},
			"filter_name": schema.StringAttribute{
				Description:         "If the group is dynamic, the name of the filter which is used to associate the agents.",
				MarkdownDescription: "If the group is dynamic, the name of the filter which is used to associate the agents.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Description:         "ID of the group.",
				MarkdownDescription: "ID of the group.",
				Computed:            true,
			},
			"inherits": schema.BoolAttribute{
				Description:         "Whether or not the group inherits policies from its parent site.",
				MarkdownDescription: "Whether or not the group inherits policies from its parent site.",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				Description:         "Whether or not the group is the default group for the parent site.",
				MarkdownDescription: "Whether or not the group is the default group for the parent site.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the group.",
				MarkdownDescription: "Name of the group.",
				Computed:            true,
			},
			"rank": schema.Int64Attribute{
				Description:         "Rank sets the priority of a dynamic group over others.",
				MarkdownDescription: "Rank sets the priority of a dynamic group over others.",
				Computed:            true,
			},
			"registration_token": schema.StringAttribute{
				Description:         "Registration token for the group.",
				MarkdownDescription: "Registration token for the group.",
				Computed:            true,
			},
			"site_id": schema.StringAttribute{
				Description:         "ID of site to which the group belongs.",
				MarkdownDescription: "ID of site to which the group belongs.",
				Computed:            true,
			},
			"total_agents": schema.Int64Attribute{
				Description:         "Total number of agents in the group.",
				MarkdownDescription: "Total number of agents in the group.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				Description:         "Type of group (eg: dynamic, pinned, static).",
				MarkdownDescription: "Type of group (eg: `dynamic`, `pinned`, `static`)",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				Description:         "Timestamp of when the group was last updated.",
				MarkdownDescription: "Timestamp of when the group was last updated.",
				Computed:            true,
			},
		},
	}
}

// terraformGroupFromAPI converts an API group into a Terraform group.
func terraformGroupFromAPI(ctx context.Context, group apiGroupModel) tfGroupModel {
	return tfGroupModel{
		CreatedAt:         types.StringValue(group.CreatedAt),
		Creator:           types.StringValue(group.Creator),
		CreatorId:         types.StringValue(group.CreatorId),
		Description:       types.StringValue(group.Description),
		FilterId:          types.StringValue(group.FilterId),
		FilterName:        types.StringValue(group.FilterName),
		Id:                types.StringValue(group.Id),
		Inherits:          types.BoolValue(group.Inherits),
		IsDefault:         types.BoolValue(group.IsDefault),
		Name:              types.StringValue(group.Name),
		Rank:              types.Int64Value(int64(group.Rank)),
		RegistrationToken: types.StringValue(group.RegistrationToken),
		SiteId:            types.StringValue(group.SiteId),
		TotalAgents:       types.Int64Value(int64(group.TotalAgents)),
		Type:              types.StringValue(group.Type),
		UpdatedAt:         types.StringValue(group.UpdatedAt),
	}
}
