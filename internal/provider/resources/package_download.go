package resources

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/data"
)

// ensure implementation satisfied expected interfaces.
var (
	_ resource.Resource              = &PackageDownload{}
	_ resource.ResourceWithConfigure = &PackageDownload{}
)

// tfPackageDownload defines the Terrform model for a package download.
type tfPackageDownload struct {
	DirectoryMode         types.String `tfsdk:"directory_mode"`
	FileMode              types.String `tfsdk:"file_mode"`
	FileSize              types.Int64  `tfsdk:"file_size"`
	LocalFilename         types.String `tfsdk:"local_filename"`
	LocalFolder           types.String `tfsdk:"local_folder"`
	OverwriteExistingFile types.Bool   `tfsdk:"overwrite_existing_file"`
	PackageId             types.String `tfsdk:"package_id"`
	SHA1                  types.String `tfsdk:"sha1"`
	SiteId                types.String `tfsdk:"site_id"`
}

// NewPackageDownload creates a new PacakgeDownload object.
func NewPackageDownload() resource.Resource {
	return &PackageDownload{}
}

// PackageDownload is a resource used to store details about an update/agent package that has been downloaded locally.
type PackageDownload struct {
	data *data.SingularityProvider
}

// Metadata returns metadata about the data source.
func (r *PackageDownload) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package_download"
}

// Schema defines the parameters for the data sources's configuration.
func (r *PackageDownload) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource is used for downloading an update/agent package from the server and saving it " +
			"locally.",
		MarkdownDescription: `resource is used for downloading an update/agent package from the server and saving it 
			locally.

		TODO: add more of a description on how to use this data source...
		`,
		Attributes: map[string]schema.Attribute{
			"directory_mode": schema.StringAttribute{
				Description: "The permissions to set on any folders created when saving the file. " +
					"Changing this value has no effect on existing folders. [Default: 0755]",
				MarkdownDescription: "The permissions to set on any folders created when saving the file. " +
					"Changing this value has no effect on existing folders. [Default: 0755]",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("0755"),
			},
			"file_mode": schema.StringAttribute{
				Description:         "The permissions to set on the file once it has been downloaded. [Default: 0644]",
				MarkdownDescription: "The permissions to set on the file once it has been downloaded. [Default: 0644]",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("0644"),
			},
			"file_size": schema.Int64Attribute{
				Description:         "The size of the package file that was downloaded.",
				MarkdownDescription: "The size of the package file that was downloaded.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					r.requiresPackageSizeUpdate(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"local_filename": schema.StringAttribute{
				Description: "Rename the downloaded package file using this name instead of keeping the original package " +
					"file name. [Default: name of the original package file from the server]",
				MarkdownDescription: "Rename the downloaded package file using this name instead of keeping the original " +
					"package file name. [Default: name of the original package file from the server]",
				Optional: true,
				Computed: true,
			},
			"local_folder": schema.StringAttribute{
				Description: "The full path to the folder in which to store the downloaded package. Use absolute " +
					"paths when possible. Relative paths will be based on the working directory when the Terrform plan is " +
					"applied. [Default: the current working directory]",
				MarkdownDescription: "The full path to the folder in which to store the downloaded package. Use absolute " +
					"paths when possible. Relative paths will be based on the working directory when the Terrform plan is " +
					"applied. [Default: the current working directory]",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(plugin.GetWorkDir()),
			},
			"overwrite_existing_file": schema.BoolAttribute{
				Description: "Whether or not to overwrite any existing file with the same name in the same " +
					"folder. [Default: true]",
				MarkdownDescription: "Whether or not to overwrite any existing file with the same name in the same " +
					"folder. [Default: true]",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"package_id": schema.StringAttribute{
				Description:         "The ID of the package to download.",
				MarkdownDescription: "The ID of the package to download.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sha1": schema.StringAttribute{
				Description:         "The SHA1 checksum of the package file that was downloaded.",
				MarkdownDescription: "The SHA1 checksum of the package file that was downloaded.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					r.requiresPackageSHA1Update(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site_id": schema.StringAttribute{
				Description:         "The ID of the site in which the package can be found.",
				MarkdownDescription: "The ID of the site in which the package can be found.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure initializes the configuration for the data source.
func (r *PackageDownload) Configure(ctx context.Context, req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {
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
			"internal_error_code": plugin.ERR_DATASOURCE_GROUP_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	r.data = providerData
}

// Create is used to create the Terraform resource.
func (r *PackageDownload) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan tfPackageDownload
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// first make sure the package we are going to download exists
	siteId := plan.SiteId.ValueString()       // always required so no need to check
	packageId := plan.PackageId.ValueString() // always required so no need to check
	pkg, diags := api.Client().GetPackage(ctx, packageId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.FileSize = types.Int64Value(pkg.FileSize)
	plan.SHA1 = types.StringValue(pkg.SHA1)

	// set default values in plan if not provided
	if plan.LocalFilename.IsNull() || plan.LocalFilename.IsUnknown() {
		plan.LocalFilename = types.StringValue(pkg.FileName)
	}

	// download the package file
	fileSize, sha1, diags := api.Client().DownloadPackage(ctx, packageId, siteId,
		path.Join(plan.LocalFolder.ValueString(), plan.LocalFilename.ValueString()),
		plan.DirectoryMode.ValueString(), plan.FileMode.ValueString(),
		plan.OverwriteExistingFile.ValueBool())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// compare the downloaded file size and SHA1 to make sure they match what are expected
	if fileSize != pkg.FileSize {
		msg := fmt.Sprintf("The downloaded package size (%d) does not match the expected package size (%d). "+
			"This may be a transient error. Please try again in a few minutes.", fileSize, pkg.FileSize)
		tflog.Error(ctx, msg, map[string]interface{}{
			"downloaded_package_size": fileSize,
			"expected_package_size":   pkg.FileSize,
		})
		resp.Diagnostics.AddError("Download Package Failure", msg)
		return
	}
	if sha1 != pkg.SHA1 {
		msg := fmt.Sprintf("The downloaded package SHA1 (%s) does not match the expected package SHA1 (%s). "+
			"This may be a transient error. Please try again in a few minutes.", sha1, pkg.SHA1)
		tflog.Error(ctx, msg, map[string]interface{}{
			"downloaded_package_sha1": sha1,
			"expected_package_sha1":   pkg.SHA1,
		})
		resp.Diagnostics.AddError("Download Package Failure", msg)
		return
	}

	// save the the plan to the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the current state of the Terraform resource.
func (r *PackageDownload) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get the current state
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh data about the downloaded file
	filePath := path.Join(state.LocalFolder.ValueString(), state.LocalFilename.ValueString())
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		tflog.Debug(ctx, "Package file no longer exists on the local system.", map[string]interface{}{
			"file": filePath,
		})
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while trying to get information on the downloaded "+
			"package file.\n\nError: %s\nFile: %s", err.Error(), filePath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error": err.Error(),
			"file":  filePath,
		})
		resp.Diagnostics.AddError("Download Package Refresh Failed", msg)
		return
	}
	state.FileMode = types.StringValue(fmt.Sprintf("%04o", fileInfo.Mode()))
	state.FileSize = types.Int64Value(fileInfo.Size())
	sha1, diags := plugin.GetFileSHA1(ctx, filePath)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SHA1 = types.StringValue(sha1)

	// set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update modifies the Terraform resource in place without destroying it.
func (r *PackageDownload) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "PackageDownload: updating resource")

}

// Delete removes the Terraform resource.
func (r *PackageDownload) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "PackageDownload: deleting resource")

}

// requiresPackageSizeUpdate returns the plan modifier for updating the package size.
func (r *PackageDownload) requiresPackageSizeUpdate() planmodifier.Int64 {
	return requiresPackageSizeUpdate{}
}

// requiresPackageSHA1Update returns the plan modifier for updating the package file's SHA1 hash.
func (r *PackageDownload) requiresPackageSHA1Update() planmodifier.String {
	return requiresPackageSHA1Update{}
}

// requiresPackageSizeUpdate queries the REST API to retrieve a package's file size so it can be compared to the
// file size on disk.
type requiresPackageSizeUpdate struct {
}

// Description returns a human-readable description of the plan modifier.
func (m requiresPackageSizeUpdate) Description(_ context.Context) string {
	return "refreshes the package file size from the REST API and compares it to the existing file size"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m requiresPackageSizeUpdate) MarkdownDescription(_ context.Context) string {
	return "refreshes the package file size from the REST API and compares it to the existing file size"
}

// PlanModifyInt64 implements the plan modification logic.
func (m requiresPackageSizeUpdate) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request,
	resp *planmodifier.Int64Response) {

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if the plan value is unknown.
	if req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// get the current state and configuration
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh package information from the server
	pkg, diags := api.Client().GetPackage(ctx, state.PackageId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "updating plan file size", map[string]interface{}{
		"file_size_from_state": resp.PlanValue.ValueInt64(),
		"file_size_from_api":   pkg.FileSize,
	})
	resp.PlanValue = types.Int64Value(pkg.FileSize)
}

// requiresPackageSHA1Update queries the REST API to retrieve a package file's SHA1 checksum so it can be compared to
// the SHA1 checksum of the file on disk.
type requiresPackageSHA1Update struct {
}

// Description returns a human-readable description of the plan modifier.
func (m requiresPackageSHA1Update) Description(_ context.Context) string {
	return "refreshes the package file's SHA1 checksum from the REST API and compares it to the SHA1 checksum of " +
		"the file on disk"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m requiresPackageSHA1Update) MarkdownDescription(_ context.Context) string {
	return "refreshes the package file's SHA1 checksum from the REST API and compares it to the SHA1 checksum of " +
		"the file on disk"
}

// PlanModifyInt64 implements the plan modification logic.
func (m requiresPackageSHA1Update) PlanModifyString(ctx context.Context, req planmodifier.StringRequest,
	resp *planmodifier.StringResponse) {

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if the plan value is unknown.
	if req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// get the current state and configuration
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// refresh package information from the server
	pkg, diags := api.Client().GetPackage(ctx, state.PackageId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "updating plan file SHA1", map[string]interface{}{
		"file_sha1_from_state": resp.PlanValue.ValueString(),
		"file_sha1_from_api":   pkg.SHA1,
	})
	resp.PlanValue = types.StringValue(pkg.SHA1)
}
