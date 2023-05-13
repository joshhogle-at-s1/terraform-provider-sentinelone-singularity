package datasources

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/data"
)

// ensure implementation satisfied expected interfaces
var (
	_ datasource.DataSource              = &Package{}
	_ datasource.DataSourceWithConfigure = &Package{}
)

// tfPackage defines the Terraform model for a package.
type tfPackage struct {
	Accounts      []tfPackageAccount `tfsdk:"accounts"`
	CreatedAt     types.String       `tfsdk:"created_at"`
	FileExtension types.String       `tfsdk:"file_extension"`
	FileName      types.String       `tfsdk:"file_name"`
	FileSize      types.Int64        `tfsdk:"file_size"`
	Id            types.String       `tfsdk:"id"`
	Link          types.String       `tfsdk:"link"`
	MajorVersion  types.String       `tfsdk:"major_version"`
	MinorVersion  types.String       `tfsdk:"minor_version"`
	OSArch        types.String       `tfsdk:"os_arch"`
	OSType        types.String       `tfsdk:"os_type"`
	PackageType   types.String       `tfsdk:"package_type"`
	PlatformType  types.String       `tfsdk:"platform_type"`
	RangerVersion types.String       `tfsdk:"ranger_version"`
	ScopeLevel    types.String       `tfsdk:"scope_level"`
	SHA1          types.String       `tfsdk:"sha1"`
	Sites         []tfPackageSite    `tfsdk:"sites"`
	Status        types.String       `tfsdk:"status"`
	UpdatedAt     types.String       `tfsdk:"updated_at"`
	Version       types.String       `tfsdk:"version"`
}

// tfPackageAccount defines the Terraform model for package accounts.
type tfPackageAccount struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// tfPackageSite defines the Terraform model for package sites.
type tfPackageSite struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewPackage creates a new Package object.
func NewPackage() datasource.DataSource {
	return &Package{}
}

// Package is a data source used to store details about a single agent/update package.
type Package struct {
	data *data.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Package) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Package) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	pkgSchema := getPackageSchema(ctx)

	// override the default schema
	pkgSchema.Attributes["id"] = schema.StringAttribute{
		Description:         "ID for the package.",
		MarkdownDescription: "ID for the package.",
		Required:            true,
	}
	resp.Schema = pkgSchema
}

// Configure initializes the configuration for the data source.
func (d *Package) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
			"internal_error_code": plugin.ERR_DATASOURCE_PACKAGE_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	d.data = providerData
}

// Read retrieves data from the API.
func (d *Package) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfPackage

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// find the matching package
	pkg, diags := api.Client().GetPackage(ctx, data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// convert the API object to the Terraform object
	resp.Diagnostics.Append(resp.State.Set(ctx, tfPackageFromAPI(ctx, pkg))...)
}

// getPackageSchema returns a default Terraform schema where all values are computed.
func getPackageSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
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
				Description:         "Timestamp of when the package was created.",
				MarkdownDescription: "Timestamp of when the package was created.",
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
				Description:         "Timestamp of when the package was last updated.",
				MarkdownDescription: "Timestamp of when the package was last updated.",
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

// tfPackageFromAPI converts an API package into a Terraform package.
func tfPackageFromAPI(ctx context.Context, pkg *api.Package) tfPackage {
	tfpkg := tfPackage{
		Accounts:      []tfPackageAccount{},
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
		Sites:         []tfPackageSite{},
		Status:        types.StringValue(pkg.Status),
		UpdatedAt:     types.StringValue(pkg.UpdatedAt),
		Version:       types.StringValue(pkg.Version),
	}
	for _, acct := range pkg.Accounts {
		tfpkg.Accounts = append(tfpkg.Accounts, tfPackageAccount{
			Id:   types.StringValue(acct.Id),
			Name: types.StringValue(acct.Name),
		})
	}
	for _, site := range pkg.Sites {
		tfpkg.Sites = append(tfpkg.Sites, tfPackageSite{
			Id:   types.StringValue(site.Id),
			Name: types.StringValue(site.Name),
		})
	}
	tflog.Debug(ctx, fmt.Sprintf("converted API package to TF package: %+v", tfpkg), map[string]interface{}{
		"api_package": pkg,
	})
	return tfpkg
}
