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
	_ datasource.DataSource              = &Packages{}
	_ datasource.DataSourceWithConfigure = &Packages{}
)

// tfPackagesModel defines the Terraform model for packages.
type tfPackagesModel struct {
	Packages []tfPackageModel       `tfsdk:"packages"`
	Filter   *tfPackagesModelFilter `tfsdk:"filter"`
}

// tfPackagesModelFilter defines the Terraform model for package filtering.
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
	SortBy        types.String   `tfsdk:"sort_by"`
	SortOrder     types.String   `tfsdk:"sort_order"`
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
			"packages": schema.ListNestedAttribute{
				Description:         "List of matching package IDs that were found",
				MarkdownDescription: "List of matching package IDs that were found",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"accounts": schema.ListNestedAttribute{
							Description:         "List of accounts to which the package belongs.",
							MarkdownDescription: "List of accounts to which the package belongs.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description:         "ID of the account.",
										MarkdownDescription: "ID of the account.",
										Computed:            true,
									},
									"name": schema.StringAttribute{
										Description:         "Name of the account.",
										MarkdownDescription: "Name of the account.",
										Computed:            true,
									},
								},
							},
						},
						"created_at": schema.StringAttribute{
							Description:         "Date and time the package was created.",
							MarkdownDescription: "Date and time the package was created.",
							Computed:            true,
						},
						"file_extension": schema.StringAttribute{
							Description:         "Extension of the package file.",
							MarkdownDescription: "Extension of the package file.",
							Computed:            true,
						},
						"file_name": schema.StringAttribute{
							Description:         "Name of the package file",
							MarkdownDescription: "Name of the package file",
							Computed:            true,
						},
						"file_size": schema.Int64Attribute{
							Description:         "Size of the package file.",
							MarkdownDescription: "Size of the package file.",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							Description:         "ID for the package.",
							MarkdownDescription: "ID for the package.",
							Computed:            true,
						},
						"link": schema.StringAttribute{
							Description:         "Link to the package file download.",
							MarkdownDescription: "Link to the package file download.",
							Computed:            true,
						},
						"major_version": schema.StringAttribute{
							Description:         "Major version of the package.",
							MarkdownDescription: "Major version of the package.",
							Computed:            true,
						},
						"minor_version": schema.StringAttribute{
							Description:         "Minor version of the package.",
							MarkdownDescription: "Minor version of the package.",
							Computed:            true,
						},
						"os_arch": schema.StringAttribute{
							Description:         "Architecture of OS on which the package runs.",
							MarkdownDescription: "Architecture of OS on which the package runs.",
							Computed:            true,
						},
						"os_type": schema.StringAttribute{
							Description:         "Type of OS on which the package runs.",
							MarkdownDescription: "Type of OS on which the package runs.",
							Computed:            true,
						},
						"package_type": schema.StringAttribute{
							Description:         "The type of packagee.",
							MarkdownDescription: "The type of packagee.",
							Computed:            true,
						},
						"platform_type": schema.StringAttribute{
							Description:         "Platform on which the package runs.",
							MarkdownDescription: "Platform on which the package runs.",
							Computed:            true,
						},
						"ranger_version": schema.StringAttribute{
							Description:         "Ranger version, if applicable.",
							MarkdownDescription: "Ranger version, if applicable.",
							Computed:            true,
						},
						"scope_level": schema.StringAttribute{
							Description:         "Package scope.",
							MarkdownDescription: "Package scope.",
							Computed:            true,
						},
						"sha1": schema.StringAttribute{
							Description:         "SHA1 hash of the package.",
							MarkdownDescription: "SHA1 hash of the package.",
							Computed:            true,
						},
						"sites": schema.ListNestedAttribute{
							Description:         "List of sites to which the package belongs.",
							MarkdownDescription: "List of sites to which the package belongs.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description:         "ID of the site.",
										MarkdownDescription: "ID of the site.",
										Computed:            true,
									},
									"name": schema.StringAttribute{
										Description:         "Name of the site.",
										MarkdownDescription: "Name of the site.",
										Computed:            true,
									},
								},
							},
						},
						"status": schema.StringAttribute{
							Description:         "Status of the package.",
							MarkdownDescription: "Status of the package.",
							Computed:            true,
						},
						"updated_at": schema.StringAttribute{
							Description:         "Date and time the package was last updated.",
							MarkdownDescription: "Date and time the package was last updated.",
							Computed:            true,
						},
						"version": schema.StringAttribute{
							Description:         "Version of the package.",
							MarkdownDescription: "Version of the package.",
							Computed:            true,
						},
					},
				},
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
							validators.EnumStringValueOneOf(false,
								".bsx", ".deb", ".exe", ".gz", ".img", ".msi",
								".pkg", ".rpm", ".tar", ".xz", ".zip", "unknown",
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
							validators.EnumStringListValuesAre(false,
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
							validators.EnumStringListValuesAre(false,
								"linux", "linux_k8s", "macos", "sdk", "windows", "windows_legacy",
							),
						},
					},
					"package_types": schema.ListAttribute{
						Description:         "Package type (eg: agent).",
						MarkdownDescription: "Package type (eg: `agent`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
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
							validators.EnumStringListValuesAre(false,
								"linux", "linux_k8s", "macos", "sdk", "windows", "windows_legacy",
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
					"sort_by": schema.StringAttribute{
						Description:         "Field on which to sort results (eg: version).",
						MarkdownDescription: "Field on which to sort results(eg: `version`).",
						Optional:            true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"createdAt", "fileExtension", "fileName", "fileSize", "id", "link", "majorVersion",
								"minorVersion", "osType", "packageType", "platformType", "rangerVersion", "scopeLevel",
								"sha1", "status", "updatedAt", "version",
							),
						},
					},
					"sort_order": schema.StringAttribute{
						Description:         "File extension (eg: .msi).",
						MarkdownDescription: "File extension (eg: `.msi`).",
						Optional:            true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"asc", "desc",
							),
						},
					},
					"status": schema.ListAttribute{
						Description:         "Package status (eg: GA).",
						MarkdownDescription: "Package status (eg: `GA`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"beta", "ea", "ga", "other",
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

	// find all matching packages querying for additional pages until results are exhausted
	var pkgs []apiPackageModel
	for {
		// get a page of results
		result, diag := d.client.APIClient.Get(ctx, "/update/agent/packages", queryParams)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		// parse the response
		var page []apiPackageModel
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"Package object.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
			resp.Diagnostics.AddError("API Query Failed", msg)
			return
		}
		pkgs = append(pkgs, page...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		queryParams["cursor"] = result.Pagination.NextCursor
	}

	// convert API objects into Terraform objects
	var tfpkgs tfPackagesModel
	for _, pkg := range pkgs {
		tfpkg := terraformPackageFromAPI(ctx, pkg)
		tfpkgs.Packages = append(tfpkgs.Packages, tfpkg)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfpkgs)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Packages) queryParamsFromFilter(filter tfPackagesModelFilter) map[string]string {
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

	if !filter.FileExtension.IsNull() && !filter.FileExtension.IsUnknown() {
		queryParams["fileExtension"] = filter.FileExtension.ValueString()
	}

	if len(filter.Ids) > 0 {
		values := []string{}
		for _, e := range filter.Ids {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["ids"] = strings.Join(values, ",")
	}

	if !filter.MinorVersion.IsNull() && !filter.MinorVersion.IsUnknown() {
		queryParams["minorVersion"] = filter.MinorVersion.ValueString()
	}

	if len(filter.OSArches) > 0 {
		values := []string{}
		for _, e := range filter.OSArches {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["osArches"] = strings.Join(values, ",")
	}

	if len(filter.OSTypes) > 0 {
		values := []string{}
		for _, e := range filter.OSTypes {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["osTypes"] = strings.Join(values, ",")
	}

	if len(filter.PackageTypes) > 0 {
		values := []string{}
		for _, e := range filter.PackageTypes {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["packageTypes"] = strings.Join(values, ",")
	}

	if len(filter.PlatformTypes) > 0 {
		values := []string{}
		for _, e := range filter.PlatformTypes {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["platformTypes"] = strings.Join(values, ",")
	}

	if !filter.RangerVersion.IsNull() && !filter.RangerVersion.IsUnknown() {
		queryParams["rangerVersion"] = filter.RangerVersion.ValueString()
	}

	if !filter.Sha1.IsNull() && !filter.Sha1.IsUnknown() {
		queryParams["sha1"] = filter.Sha1.ValueString()
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

	if len(filter.Status) > 0 {
		values := []string{}
		for _, e := range filter.Status {
			if !e.IsNull() && !e.IsUnknown() {
				values = append(values, e.ValueString())
			}
		}
		queryParams["status"] = strings.Join(values, ",")
	}

	if !filter.Version.IsNull() && !filter.Version.IsUnknown() {
		queryParams["version"] = filter.Version.ValueString()
	}
	return queryParams
}
