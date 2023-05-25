package resources

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"

	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	OutputFile            types.String `tfsdk:"output_file"`
	OverwriteExistingFile types.Bool   `tfsdk:"overwrite_existing_file"`
	PackageId             types.String `tfsdk:"package_id"`
	SHA1                  types.String `tfsdk:"sha1"`
	SiteId                types.String `tfsdk:"site_id"`
	Version               types.String `tfsdk:"version"`
}

// NewPackageDownload creates a new PackageDownload object.
func NewPackageDownload() resource.Resource {
	return &PackageDownload{}
}

// PackageDownload is a resource used to download update/agent packages using the API.
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
		Description: "This resource is used for downloading an update/agent package from the server and saving it	" +
			"locally.",
		MarkdownDescription: `This resource is used for downloading an update/agent package from the server and saving it
			locally.

		TODO: add more of a description on how to use this data source...
		`,
		Attributes: map[string]schema.Attribute{
			"directory_mode": schema.StringAttribute{
				Description: "The permissions to set on any folders created when saving the file. " +
					"Changing this value has no effect on existing folders. Ignored on Windows. [Default: 0755]",
				MarkdownDescription: "The permissions to set on any folders created when saving the file. " +
					"Changing this value has no effect on existing folders. Ignored on Windows. [Default: `0755`]",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("0755"),
				Validators: []validator.String{
					validators.FileModeIsValid(),
				},
			},
			"file_mode": schema.StringAttribute{
				Description: "The permissions to set on the file once it has been downloaded. Ignored on Windows. " +
					"[Default: 0644]",
				MarkdownDescription: "The permissions to set on the file once it has been downloaded. Ignored on Windows. " +
					"[Default: `0644`]",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("0644"),
				Validators: []validator.String{
					validators.FileModeIsValid(),
				},
			},
			"file_size": schema.Int64Attribute{
				Description:         "The size of the package file that was downloaded.",
				MarkdownDescription: "The size of the package file that was downloaded.",
				Computed:            true,
			},
			"local_filename": schema.StringAttribute{
				Description:         "The name of the file to save the downloaded package as.",
				MarkdownDescription: "The name of the file to save the downloaded package as.",
				Required:            true,
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
			"output_file": schema.StringAttribute{
				Description:         "The absolute path of the downloaded file once it has been saved.",
				MarkdownDescription: "The absolute path of the downloaded file once it has been saved.",
				Computed:            true,
			},
			"overwrite_existing_file": schema.BoolAttribute{
				Description: "Whether or not to overwrite any existing file with the same name in the same " +
					"folder. [Default: true]",
				MarkdownDescription: "Whether or not to overwrite any existing file with the same name in the same " +
					"folder. [Default: `true`]",
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
			},
			"site_id": schema.StringAttribute{
				Description:         "The ID of the site in which the package can be found.",
				MarkdownDescription: "The ID of the site in which the package can be found.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description:         "The version of the downloaded package file.",
				MarkdownDescription: "The version of the downloaded package file.",
				Computed:            true,
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
			"Received: %T", expectedType, req.ProviderData)
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	r.data = providerData
}

// ModifyPlan is called to modify the Terraform plan.
func (r *PackageDownload) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse) {

	// retrieve values from plan and state
	var packageId, sha1 types.String
	var fileSize types.Int64
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, tfpath.Root("package_id"), &packageId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, tfpath.Root("file_size"), &fileSize)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, tfpath.Root("sha1"), &sha1)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// we need to do some extra work here regarding the SHA1 and size checks - otherwise if the file is
	// replaced, no changes will be detected
	if !packageId.IsNull() && !packageId.IsUnknown() &&
		!fileSize.IsNull() && !fileSize.IsUnknown() && !sha1.IsNull() && !sha1.IsUnknown() {
		// refresh package data
		pkg, diags := api.Client().GetPackage(ctx, packageId.ValueString())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// compare API file size and SHA1 with state
		if pkg.FileSize != fileSize.ValueInt64() {
			resp.RequiresReplace.Append(tfpath.Root("file_size"))
			resp.Plan.SetAttribute(ctx, tfpath.Root("file_size"), types.Int64Value(pkg.FileSize))
		}
		if pkg.SHA1 != sha1.ValueString() {
			resp.RequiresReplace.Append(tfpath.Root("sha1"))
			resp.Plan.SetAttribute(ctx, tfpath.Root("sha1"), types.StringValue(pkg.SHA1))
		}
	}
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

	// download the package file
	outputFile, fileSize, sha1, version, diags := api.Client().DownloadPackage(ctx, packageId, siteId,
		path.Join(plan.LocalFolder.ValueString(), plan.LocalFilename.ValueString()),
		plan.DirectoryMode.ValueString(), plan.FileMode.ValueString(),
		plan.OverwriteExistingFile.ValueBool())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.OutputFile = types.StringValue(outputFile)
	plan.Version = types.StringValue(version)

	// compare the downloaded file size and SHA1 to make sure they match what are expected
	if fileSize != pkg.FileSize {
		msg := fmt.Sprintf("The downloaded package size (%d) does not match the expected package size (%d). "+
			"This may be a transient error. Please try again in a few minutes.", fileSize, pkg.FileSize)
		tflog.Error(ctx, msg, map[string]interface{}{
			"downloaded_package_size": fileSize,
			"expected_package_size":   pkg.FileSize,
			"internal_error_code":     plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CREATE,
		})
		resp.Diagnostics.AddError("Download Package Creation Error", msg)
		return
	}
	if sha1 != pkg.SHA1 {
		msg := fmt.Sprintf("The downloaded package SHA1 (%s) does not match the expected package SHA1 (%s). "+
			"This may be a transient error. Please try again in a few minutes.", sha1, pkg.SHA1)
		tflog.Error(ctx, msg, map[string]interface{}{
			"downloaded_package_sha1": sha1,
			"expected_package_sha1":   pkg.SHA1,
			"internal_error_code":     plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CREATE,
		})
		resp.Diagnostics.AddError("Download Package Creation Error", msg)
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

	// get the version from the API
	pkg, diags := api.Client().GetPackage(ctx, state.PackageId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Version = types.StringValue(pkg.Version)

	// gather information about the package file
	absPath := state.OutputFile.ValueString()
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		tflog.Debug(ctx, "Package file no longer exists on the local system.", map[string]interface{}{
			"file": absPath,
		})
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while trying to get information on the downloaded "+
			"package file.\n\nError: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"file":                absPath,
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_READ,
		})
		resp.Diagnostics.AddError("Download Package Refresh Error", msg)
		return
	} else if fileInfo.IsDir() {
		err = fmt.Errorf("the file path given is actually a folder")
		msg := fmt.Sprintf("An unexpected error occurred while trying to get information on the downloaded "+
			"package file.\n\nError: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"file":                absPath,
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_READ,
		})
		resp.Diagnostics.AddError("Download Package Refresh Error", msg)
	}

	// update state values
	if runtime.GOOS != "windows" { // we only care about file mode on non-windows systems
		state.FileMode = types.StringValue(fmt.Sprintf("%04o", fileInfo.Mode()))
	}
	state.FileSize = types.Int64Value(fileInfo.Size())
	sha1, diags := plugin.GetFileSHA1(ctx, absPath)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SHA1 = types.StringValue(sha1)

	// save refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update modifies the Terraform resource in place without destroying it.
func (r *PackageDownload) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from state
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve values from plan
	var plan tfPackageDownload
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// directory mode and overwrite flag updates require no changes locally
	if !plan.DirectoryMode.IsNull() && !plan.DirectoryMode.IsUnknown() {
		state.DirectoryMode = plan.DirectoryMode
	}
	if !plan.OverwriteExistingFile.IsNull() && !plan.OverwriteExistingFile.IsUnknown() {
		state.OverwriteExistingFile = plan.OverwriteExistingFile
	}

	// update source/dest file paths based on state and plan
	srcPath := state.OutputFile.ValueString()
	folder := state.LocalFolder.ValueString()
	if !plan.LocalFolder.IsNull() && !plan.LocalFolder.IsUnknown() {
		folder = plan.LocalFolder.ValueString()
		state.LocalFolder = plan.LocalFolder
	}
	filename := state.LocalFilename.ValueString()
	if !plan.LocalFilename.IsNull() && !plan.LocalFilename.IsUnknown() {
		filename = plan.LocalFilename.ValueString()
		state.LocalFilename = plan.LocalFilename
	}
	destPath, diags := plugin.ToAbsolutePath(ctx, path.Join(folder, filename))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if filename or folder has changed, move the file
	if srcPath != destPath {
		state.OutputFile = types.StringValue(destPath)

		// make sure the destination folder exists
		folder, _ := filepath.Split(destPath)
		resp.Diagnostics.Append(plugin.CreateDirectory(ctx, folder, state.DirectoryMode.ValueString())...)
		if resp.Diagnostics.HasError() {
			return
		}

		// move the file
		if err := os.Rename(srcPath, destPath); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while moving the package file.\n\n"+
				"Error: %s\nSource: %s\nDestination: %s", err.Error(), srcPath, destPath)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_UPDATE,
				"src_path":            srcPath,
				"dest_path":           destPath,
			})
			resp.Diagnostics.AddError("Download Package Update Error", msg)
			return
		}
		tflog.Debug(ctx, "Moved package file", map[string]interface{}{
			"src_path":  srcPath,
			"dest_path": destPath,
		})
	}

	// if file mode has changed, update the file mode
	if !plan.FileMode.IsNull() && !plan.FileMode.IsUnknown() {
		state.FileMode = plan.FileMode

		// get the new file mode
		newMode, diags := plugin.ParseFilesystemMode(ctx, plan.FileMode.ValueString())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// update the file mode
		if err := os.Chmod(destPath, newMode); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while changing permissions on the package file.\n\n"+
				"Error: %s\nFile: %s\nNew Mode: %s", err.Error(), destPath, fmt.Sprintf("%04o", newMode))
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_UPDATE,
				"new_mode":            fmt.Sprintf("%04o", newMode),
			})
			resp.Diagnostics.AddError("Download Package Update Error", msg)
			return
		}
		tflog.Debug(ctx, "Updated file mode for package file", map[string]interface{}{
			"file":     destPath,
			"new_mode": fmt.Sprintf("%04o", newMode),
		})
	}

	// save the the plan to the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete removes the Terraform resource.
func (r *PackageDownload) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// get the current state
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if output file is empty, nothing to remove
	if state.OutputFile.IsNull() {
		return
	}
	absPath := state.OutputFile.ValueString()

	// make sure the path is actually a file that exists
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		// nothing to do - file no longer exists
		return
	} else if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to open the package file.\n\n"+
			"Error: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"file":                absPath,
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_DELETE,
		})
		resp.Diagnostics.AddError("Download Package Removal Error", msg)
		return
	}
	if fileInfo.IsDir() {
		err := fmt.Errorf("the destination path is a directory, not a file")
		msg := fmt.Sprintf("An unexpected error occurred while attempting to remove the package file.\n\n"+
			"Error: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"file":                absPath,
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_DELETE,
		})
		resp.Diagnostics.AddError("Download Package Removal Error", msg)
		return
	}

	// remove the file
	if err := os.Remove(absPath); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while removing the package file.\n\nError: %s\nFile: %s",
			err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_DELETE,
		})
		resp.Diagnostics.AddError("Download Package Removal Error", msg)
		return
	}
	tflog.Debug(ctx, "Removed package file", map[string]interface{}{
		"file": absPath,
	})
}
