package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/validators"
)

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Packages{}
	_ datasource.DataSourceWithConfigure = &Packages{}
)

// tfPackagesModel defines the Terraform model for packages.
type tfPackagesModel struct {
	PackageIds []types.String         `tfsdk:"ids"`
	Filter     *tfPackagesModelFilter `tfsdk:"filter"`
}

type tfPackagesModelFilter struct {
	AccountIds    []types.String `tfsdk:"account_ids"`
	FileExtension types.String   `tfsdk:"file_extension"`
	Ids           []types.String `tfsdk:"ids"`
	MinorVersion  types.String   `tfsdk:"minor_version"`
	OSArches      []types.String `tfsdk:"os_arches"`
	OSTypes       []types.String `tfsdk:"os_types"`
	PackageTypes  []types.String `tfsdk:"package_types"`
	PlatformTypes []types.String `tfsdk:"platform_types"`
	Query         types.String   `tfsdk:"query"`
	RangerVersion types.String   `tfsdk:"ranger_version"`
	Sha1          types.String   `tfsdk:"sha1"`
	SiteIds       []types.String `tfsdk:"site_ids"`
	Status        []types.String `tfsdk:"status"`
	Version       types.String   `tfsdk:"version"`
}

// NewPackage creates a new Packages object.
func NewPackages() datasource.DataSource {
	return &Packages{}
}

// Packages is a data source used to store details about agent packages available on the server.
type Packages struct {
	client *client.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Packages) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packages"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Packages) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "This resource can be used for getting a list of package IDs based on filters.",
		MarkdownDescription: "This resource can be used for getting a list of package IDs based on filters.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				Description:         "List of matching package IDs that were found",
				MarkdownDescription: "List of matching package IDs that were found",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description:         "Defines the query filters to use when searching for packages.",
				MarkdownDescription: "Defines the query filters to use when searching for packages.",
				Attributes: map[string]schema.Attribute{
					"account_ids": schema.ListAttribute{
						Description:         "List of account IDs to filter by.",
						MarkdownDescription: "List of account IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"file_extension": schema.StringAttribute{
						Description:         "File extension (eg: .msi).",
						MarkdownDescription: "File extension (eg: `.msi`).",
						Optional:            true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(true,
								".bsx", ".deb", ".exe", ".gz", ".img", ".msi",
								".pkg", ".rpm", ".tar", ".xz", ".zip", "Unknown",
							),
						},
					},
					"ids": schema.ListAttribute{
						Description:         "List of package IDs to filter by.",
						MarkdownDescription: "List of package IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"minor_version": schema.StringAttribute{
						Description:         "Minor version of the package.",
						MarkdownDescription: "Minor version of the package.",
						Optional:            true,
					},
					"os_arches": schema.ListAttribute{
						Description:         "Package OS architecture, applicable to Windows packages only (eg: 32-bit).",
						MarkdownDescription: "Package OS architecture, applicable to Windows packages only (eg: `32-bit`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(true,
								"32 bit", "32/64 bit", "64 bit", "N/A",
							),
						},
					},
					"os_types": schema.ListAttribute{
						Description:         "Package OS type (eg: macos).",
						MarkdownDescription: "Package OS type (eg: `macos`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(true,
								"Linux", "Linux_k8s", "Macos", "Sdk", "Windows", "Windows_legacy",
							),
						},
					},
					"package_types": schema.ListAttribute{
						Description:         "Package type (eg: agent).",
						MarkdownDescription: "Package type (eg: `agent`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(true,
								"Agent", "AgentAndRanger", "Ranger",
							),
						},
					},
					"platform_types": schema.ListAttribute{
						Description:         "Package platform (eg: macos).",
						MarkdownDescription: "Package platform (eg: `macos`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(true,
								"Linux", "Linux_k8s", "Macos", "Sdk", "Windows", "Windows_legacy",
							),
						},
					},
					"query": schema.StringAttribute{
						Description:         "A free-text search term, will match applicable attributes.",
						MarkdownDescription: "A free-text search term, will match applicable attributes.",
						Optional:            true,
					},
					"ranger_version": schema.StringAttribute{
						Description:         "Ranger version (eg: 2.5.1.1320).",
						MarkdownDescription: "Ranger version (eg: `2.5.1.1320`).",
						Optional:            true,
					},
					"sha1": schema.StringAttribute{
						Description:         "Package hash (eg: 2fd4e1c67a2d28fced849ee1bb76e7391b93eb12).",
						MarkdownDescription: "Package hash (eg: `2fd4e1c67a2d28fced849ee1bb76e7391b93eb12`).",
						Optional:            true,
					},
					"site_ids": schema.ListAttribute{
						Description:         "List of site IDs to filter by.",
						MarkdownDescription: "List of site IDs to filter by.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"status": schema.ListAttribute{
						Description:         "Package status (eg: GA).",
						MarkdownDescription: "Package status (eg: `GA`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(true,
								"Beta", "EA", "GA", "Other",
							),
						},
					},
					"version": schema.StringAttribute{
						Description:         "Agent version (eg: 2.5.1.1320).",
						MarkdownDescription: "Agent version (eg: `2.5.1.1320`).",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Configure initializes the configuration for the data source.
func (d *Packages) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.SingularityProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Type",
			fmt.Sprintf("Expected *client.SingularityProvider, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}
	d.client = client
}

// Read retrieves data from the API.
func (d *Packages) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfPackagesModel

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
	queryParams["sortBy"] = "updatedAt"
	queryParams["sortOrder"] = "desc"

	//resp.Diagnostics.Append(resp.State.Set(ctx, pkgs)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Packages) queryParamsFromFilter(filter tfPackagesModelFilter) map[string]string {
	queryParams := map[string]string{}

	if len(filter.AccountIds) > 0 {
		ids := []string{}
		for _, acct := range filter.AccountIds {
			if !acct.IsNull() && !acct.IsUnknown() {
				ids = append(ids, acct.ValueString())
			}
		}
		queryParams["accountIds"] = strings.Join(ids, ",")
	}

	if !filter.FileExtension.IsNull() && !filter.FileExtension.IsUnknown() {
		queryParams["fileExtension"] = filter.FileExtension.ValueString()
	}

	/*
		if !data.Id.IsNull() {
			queryParams["ids"] = data.Id.ValueString()
		}
		if !data.MinorVersion.IsNull() {
			queryParams["minorVersion"] = data.MinorVersion.ValueString()
		}
		if !data.OSArch.IsNull() {
			queryParams["osArches"] = data.OSArch.ValueString()
		}
		if !data.OSType.IsNull() {
			queryParams["osTypes"] = data.OSType.ValueString()
		}
		if !data.PackageType.IsNull() {
			queryParams["packageTypes"] = data.PackageType.ValueString()
		}
		if !data.PlatformType.IsNull() {
			queryParams["platformTypes"] = data.PlatformType.ValueString()
		}
		if !data.RangerVersion.IsNull() {
			queryParams["rangerVersion"] = data.RangerVersion.ValueString()
		}
		if !data.SHA1.IsNull() {
			queryParams["sha1"] = data.SHA1.ValueString()
		}
		if len(data.Sites) > 0 {
			ids := []string{}
			for _, site := range data.Sites {
				if !site.Id.IsNull() {
					ids = append(ids, site.Id.ValueString())
				}
			}
			queryParams["siteIds"] = strings.Join(ids, ",")
		}
		if !data.Status.IsNull() {
			queryParams["status"] = data.Status.ValueString()
		}
		if !data.Version.IsNull() {
			queryParams["version"] = data.Version.ValueString()
	*/
	return queryParams
}

/*
// getPackages retrieves multiple update packages from the server which match the given search criteria.
func getPackages(ctx context.Context, client *api.Client, data tfPackagesModel) (*tfPackagesModel, diag.Diagnostics) {
	/*
	   // generate query parameters from data


	// generate query parameters from data
	//queryParams := queryParamFromTFData(data.Packages)
	//queryParams := map[string]string{}

	// keep querying until we've exhausted all pages
	var diag diag.Diagnostics
	//var pkgs []apiPackageModel
	/*
		for {
			// find the matching packages
			result, diag := client.Get(ctx, "/update/agent/packages", queryParams)
			if diag.HasError() {
				return nil, diag
			}

			// parse the response
			var page []apiPackageModel
			if err := json.Unmarshal(result.Data, &page); err != nil {
				msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
					"Package object.\n\nError: %s", err.Error())
				tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
				diag.AddError("API Query Failed", msg)
				return nil, diag
			}
			pkgs = append(pkgs, page...)

			// get the next page of results until there is no next cursor
			if result.Pagination.NextCursor == "" {
				break
			}
			queryParams["cursor"] = result.Pagination.NextCursor
		}

		// convert the packages into a Terraform object
		var tfpkgs tfPackagesModel
		for _, pkg := range pkgs {
			tfpkgs.Packages = append(tfpkgs.Packages, *apiPackage2TFPackage(ctx, pkg))
		}

	return nil, diag
}
*/
