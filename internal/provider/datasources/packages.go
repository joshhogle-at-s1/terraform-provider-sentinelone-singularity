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

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Packages{}
	_ datasource.DataSourceWithConfigure = &Packages{}
)

// tfPackages defines the Terraform model for packages.
type tfPackages struct {
	Packages []tfPackage       `tfsdk:"packages"`
	Filter   *tfPackagesFilter `tfsdk:"filter"`
}

// tfPackagesFilter defines the Terraform model for package filtering.
type tfPackagesFilter struct {
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

// NewPackages creates a new Packages object.
func NewPackages() datasource.DataSource {
	return &Packages{}
}

// Packages is a data source used to store details about agent/update packages.
type Packages struct {
	data *data.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Packages) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packages"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Packages) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source can be used for getting a list of packages based on filters.",
		MarkdownDescription: `This data source can be used for getting a list of packages based on filters.

		TODO: add more of a description on how to use this data source...
		`,
		Attributes: map[string]schema.Attribute{
			"packages": schema.ListNestedAttribute{
				Description:         "List of matching packages that were found.",
				MarkdownDescription: "List of matching packages that were found.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: getPackageSchema(ctx).Attributes,
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
						Description: "File extension (valid values: .bsx, .deb, .exe, .gz, .img, .msi, .pkg, .rpm, .tar " +
							".xz, .zip, unknown).",
						MarkdownDescription: "File extension (valid values: `.bsx`, `.deb`, `.exe`, `.gz`, `.img`, `.msi`, " +
							"`.pkg`, `.rpm`, `.tar` `.xz`, `.zip`, `unknown`).",
						Optional: true,
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
						Description: "Package OS architecture, applicable to Windows packages only " +
							"(valid values: 32 bit, 32/64 bit, 64 bit, N/A).",
						MarkdownDescription: "Package OS architecture, applicable to Windows packages only " +
							"(valid values: `32 bit`, `32/64 bit`, `64 bit`, `N/A`).",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"32 bit", "32/64 bit", "64 bit", "N/A",
							),
						},
					},
					"os_types": schema.ListAttribute{
						Description: "Package OS type (valid values: linux, linux_k8s, macos, sdk, windows, " +
							"windows_legacy).",
						MarkdownDescription: "Package OS type (valid values: `linux`, `linux_k8s`, `macos`, `sdk` " +
							"`windows`, `windows_legacy`).",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"linux", "linux_k8s", "macos", "sdk", "windows", "windows_legacy",
							),
						},
					},
					"package_types": schema.ListAttribute{
						Description:         "Package type (valid values: Agent, AgentAndRanger, Ranger).",
						MarkdownDescription: "Package type (valid values: `Agent`, `AgentAndRanger`, `Ranger`).",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							validators.EnumStringListValuesAre(false,
								"Agent", "AgentAndRanger", "Ranger",
							),
						},
					},
					"platform_types": schema.ListAttribute{
						Description: "Package platform (valid values: linux, linux_k8s, macos, sdk, windows, " +
							"windows_legacy).",
						MarkdownDescription: "Package platform (valid values: `linux`, `linux_k8s`, `macos`, `sdk` " +
							"`windows`, `windows_legacy`).",
						Optional:    true,
						ElementType: types.StringType,
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
						Description: "Field on which to sort results (valid values: createdAt, fileExtension, fileName, " +
							"fileSize, id, majorVersion, minorVersion, osType, packageType, platformType, rangerVersion, " +
							"scopeLevel, sha1, status, updatedAt, version).",
						MarkdownDescription: "Field on which to sort results (valid values: `createdAt`, `fileExtension`, " +
							"`fileName`, `fileSize`, `id`, `majorVersion`, `minorVersion`, `osType`, `packageType`, " +
							"`platformType`, `rangerVersion`, `scopeLevel`, `sha1`, `status`, `updatedAt`, `version`).",
						Optional: true,
						Validators: []validator.String{
							validators.EnumStringValueOneOf(false,
								"createdAt", "fileExtension", "fileName", "fileSize", "id", "majorVersion",
								"minorVersion", "osType", "packageType", "platformType", "rangerVersion", "scopeLevel",
								"sha1", "status", "updatedAt", "version",
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
					"status": schema.ListAttribute{
						Description:         "Package status (valid values: beta, ea, ga, other).",
						MarkdownDescription: "Package status (valid values: `beta`, `ea`, `ga`, `other`).",
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
			"internal_error_code": plugin.ERR_DATASOURCE_PACKAGES_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	d.data = providerData
}

// Read retrieves data from the API.
func (d *Packages) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfPackages

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := api.PackageQueryParams{}
	if data.Filter != nil {
		queryParams = d.queryParamsFromFilter(*data.Filter)
	}

	// find the matching packages
	pkgs, diags := api.Client().FindPackages(ctx, queryParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert API objects into Terraform objects
	tfpkgs := tfPackages{
		Filter:   data.Filter,
		Packages: []tfPackage{},
	}
	for _, pkg := range pkgs {
		tfpkgs.Packages = append(tfpkgs.Packages, tfPackageFromAPI(ctx, &pkg))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tfpkgs)...)
}

// queryParamsFromFilter converts the TF filter block into API query parameters.
func (d *Packages) queryParamsFromFilter(filter tfPackagesFilter) api.PackageQueryParams {
	queryParams := api.PackageQueryParams{}

	if len(filter.AccountIds) > 0 {
		queryParams.AccountIds = []string{}
		for _, e := range filter.AccountIds {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.AccountIds = append(queryParams.AccountIds, e.ValueString())
			}
		}
	}

	if !filter.FileExtension.IsNull() && !filter.FileExtension.IsUnknown() {
		value := filter.FileExtension.ValueString()
		queryParams.FileExtension = &value
	}

	if len(filter.Ids) > 0 {
		queryParams.Ids = []string{}
		for _, e := range filter.Ids {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.Ids = append(queryParams.Ids, e.ValueString())
			}
		}
	}

	if !filter.MinorVersion.IsNull() && !filter.MinorVersion.IsUnknown() {
		value := filter.MinorVersion.ValueString()
		queryParams.MinorVersion = &value
	}

	if len(filter.OSArches) > 0 {
		queryParams.OSArches = []string{}
		for _, e := range filter.OSArches {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.OSArches = append(queryParams.OSArches, e.ValueString())
			}
		}
	}

	if len(filter.OSTypes) > 0 {
		queryParams.OSTypes = []string{}
		for _, e := range filter.OSTypes {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.OSTypes = append(queryParams.OSTypes, e.ValueString())
			}
		}
	}

	if len(filter.PackageTypes) > 0 {
		queryParams.PackageTypes = []string{}
		for _, e := range filter.PackageTypes {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.PackageTypes = append(queryParams.PackageTypes, e.ValueString())
			}
		}
	}

	if len(filter.PlatformTypes) > 0 {
		queryParams.PlatformTypes = []string{}
		for _, e := range filter.PlatformTypes {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.PlatformTypes = append(queryParams.PlatformTypes, e.ValueString())
			}
		}
	}

	if !filter.RangerVersion.IsNull() && !filter.RangerVersion.IsUnknown() {
		value := filter.RangerVersion.ValueString()
		queryParams.RangerVersion = &value
	}

	if !filter.Sha1.IsNull() && !filter.Sha1.IsUnknown() {
		value := filter.Sha1.ValueString()
		queryParams.Sha1 = &value
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

	if len(filter.Status) > 0 {
		queryParams.Status = []string{}
		for _, e := range filter.Status {
			if !e.IsNull() && !e.IsUnknown() {
				queryParams.Status = append(queryParams.Status, e.ValueString())
			}
		}
	}

	if !filter.Version.IsNull() && !filter.Version.IsUnknown() {
		value := filter.Version.ValueString()
		queryParams.Version = &value
	}
	return queryParams
}
