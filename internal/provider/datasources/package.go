package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
)

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Package{}
	_ datasource.DataSourceWithConfigure = &Package{}
)

// apiPackageModel defines the API model for a package.
type apiPackageModel struct {
	Accounts      []apiPackageAccountModel `json:"accounts"`
	CreatedAt     string                   `json:"createdAt"`
	FileExtension string                   `json:"fileExtension"`
	FileName      string                   `json:"fileName"`
	FileSize      int64                    `json:"fileSize"`
	Id            string                   `json:"id"`
	Link          string                   `json:"link"`
	MajorVersion  string                   `json:"majorVersion"`
	MinorVersion  string                   `json:"minorVersion"`
	OSArch        string                   `json:"osArch"`
	OSType        string                   `json:"osType"`
	PackageType   string                   `json:"packageType"`
	PlatformType  string                   `json:"platformType"`
	RangerVersion string                   `json:"rangerVersion"`
	ScopeLevel    string                   `json:"scopeLevel"`
	SHA1          string                   `json:"sha1"`
	Sites         []apiPackageSiteModel    `json:"sites"`
	Status        string                   `json:"status"`
	UpdatedAt     string                   `json:"updatedAt"`
	Version       string                   `json:"version"`
}

// apiPackageAccountModel defines the API model for package accounts.
type apiPackageAccountModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// apiPackageSiteModel defines the API model for package accounts.
type apiPackageSiteModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// tfPackageModel defines the Terraform model for a package.
type tfPackageModel struct {
	Accounts      []tfPackageAccountModel `tfsdk:"accounts"`
	CreatedAt     types.String            `tfsdk:"created_at"`
	FileExtension types.String            `tfsdk:"file_extension"`
	FileName      types.String            `tfsdk:"file_name"`
	FileSize      types.Int64             `tfsdk:"file_size"`
	Id            types.String            `tfsdk:"id"`
	Link          types.String            `tfsdk:"link"`
	MajorVersion  types.String            `tfsdk:"major_version"`
	MinorVersion  types.String            `tfsdk:"minor_version"`
	OSArch        types.String            `tfsdk:"os_arch"`
	OSType        types.String            `tfsdk:"os_type"`
	PackageType   types.String            `tfsdk:"package_type"`
	PlatformType  types.String            `tfsdk:"platform_type"`
	RangerVersion types.String            `tfsdk:"ranger_version"`
	ScopeLevel    types.String            `tfsdk:"scope_level"`
	SHA1          types.String            `tfsdk:"sha1"`
	Sites         []tfPackageSiteModel    `tfsdk:"sites"`
	Status        types.String            `tfsdk:"status"`
	UpdatedAt     types.String            `tfsdk:"updated_at"`
	Version       types.String            `tfsdk:"version"`
}

// tfPackageAccountModel defines the Terraform model for package accounts.
type tfPackageAccountModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// tfPackageSiteModel defines the Terraform model for package sites.
type tfPackageSiteModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewPackage creates a new Package object.
func NewPackage() datasource.DataSource {
	return &Package{}
}

// Package is a data source used to store details about an single package available on the server.
type Package struct {
	client *client.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Package) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Package) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "This data source is used for getting details on a specific package.",
		MarkdownDescription: "This data source is used for getting details on a specific package.",
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
				Required:            true,
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
	}
}

// Configure initializes the configuration for the data source.
func (d *Package) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *Package) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfPackageModel

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// query for the given package
	pkg, diag := getPackage(ctx, d.client.APIClient, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, pkg)...)
}

// apiPackage2TFPackage converts an API package object to a Terrform package object.
func apiPackage2TFPackage(ctx context.Context, pkg apiPackageModel) *tfPackageModel {
	tfpkg := tfPackageModel{
		CreatedAt:     types.StringValue(pkg.CreatedAt),
		FileExtension: types.StringValue(pkg.FileExtension),
		FileName:      types.StringValue(pkg.FileName),
		FileSize:      types.Int64Value(pkg.FileSize),
		Id:            types.StringValue(pkg.Id),
		Link:          types.StringValue(pkg.Link),
		MajorVersion:  types.StringValue(pkg.MajorVersion),
		MinorVersion:  types.StringValue(pkg.MinorVersion),
		OSArch:        types.StringValue(pkg.OSArch),
		OSType:        types.StringValue(pkg.OSType),
		PackageType:   types.StringValue(pkg.PackageType),
		PlatformType:  types.StringValue(pkg.PlatformType),
		RangerVersion: types.StringValue(pkg.RangerVersion),
		ScopeLevel:    types.StringValue(pkg.ScopeLevel),
		SHA1:          types.StringValue(pkg.SHA1),
		Status:        types.StringValue(pkg.Status),
		UpdatedAt:     types.StringValue(pkg.UpdatedAt),
		Version:       types.StringValue(pkg.Version),
	}
	for _, acct := range pkg.Accounts {
		tfpkg.Accounts = append(tfpkg.Accounts, tfPackageAccountModel{
			Id:   types.StringValue(acct.Id),
			Name: types.StringValue(acct.Name),
		})
	}
	for _, site := range pkg.Sites {
		tfpkg.Sites = append(tfpkg.Sites, tfPackageSiteModel{
			Id:   types.StringValue(site.Id),
			Name: types.StringValue(site.Name),
		})
	}
	tflog.Trace(ctx, fmt.Sprintf("converted API package to TF package: %+v", tfpkg), map[string]interface{}{
		"api_package": pkg,
	})
	return &tfpkg
}

// getPackage retrieves an update package from the server.
//
// This function expects exactly 1 matching package to be found.
func getPackage(ctx context.Context, client *api.Client, data tfPackageModel) (*tfPackageModel, diag.Diagnostics) {
	// generate query parameters from data
	queryParams := queryParamFromTFData(data)

	// find the matching package
	result, diag := client.Get(ctx, "/update/agent/packages", queryParams)
	if diag.HasError() {
		return nil, diag
	}

	// parse the response - we are expecting exactly 1 package to be returned
	numPkgs := result.Pagination.TotalItems
	if numPkgs == 0 {
		msg := "No matching package was found. Try expanding your search."
		tflog.Error(ctx, msg, map[string]interface{}{"packages_found": numPkgs})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	} else if numPkgs > 1 {
		msg := fmt.Sprintf("This data source expects 1 matching package but %d were found. Please narrow your search.",
			numPkgs)
		tflog.Error(ctx, msg, map[string]interface{}{"packages_found": numPkgs})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	}
	var pkgs []apiPackageModel
	if err := json.Unmarshal(result.Data, &pkgs); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Package object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	}

	// convert the package into a Terraform object
	return apiPackage2TFPackage(ctx, pkgs[0]), diag
}

// queryParamFromTFData returns API query parameters based on Terraform inputs.
func queryParamFromTFData(data tfPackageModel) map[string]string {
	queryParams := map[string]string{}
	if len(data.Accounts) > 0 {
		ids := []string{}
		for _, acct := range data.Accounts {
			if !acct.Id.IsNull() {
				ids = append(ids, acct.Id.ValueString())
			}
		}
		queryParams["accountIds"] = strings.Join(ids, ",")
	}
	if !data.FileExtension.IsNull() {
		queryParams["fileExtension"] = data.FileExtension.ValueString()
	}
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
	}
	return queryParams
}
