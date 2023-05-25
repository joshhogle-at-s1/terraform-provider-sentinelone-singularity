	"path/filepath"
	"runtime"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/validators"
// ensure implementation satisfied expected interfaces
	OutputFile            types.String `tfsdk:"output_file"`
	Version               types.String `tfsdk:"version"`
// NewPackageDownload creates a new PackageDownload object.
// PackageDownload is a resource used to download update/agent packages using the API.
		Description: "This resource is used for downloading an update/agent package from the server and saving it	" +
		MarkdownDescription: `This resource is used for downloading an update/agent package from the server and saving it
					"Changing this value has no effect on existing folders. Ignored on Windows. [Default: 0755]",
					"Changing this value has no effect on existing folders. Ignored on Windows. [Default: `0755`]",
				Validators: []validator.String{
					validators.FileModeIsValid(),
				},
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
				Description:         "The name of the file to save the downloaded package as.",
				MarkdownDescription: "The name of the file to save the downloaded package as.",
				Required:            true,
			"output_file": schema.StringAttribute{
				Description:         "The absolute path of the downloaded file once it has been saved.",
				MarkdownDescription: "The absolute path of the downloaded file once it has been saved.",
				Computed:            true,
			},
					"folder. [Default: `true`]",
			"version": schema.StringAttribute{
				Description:         "The version of the downloaded package file.",
				MarkdownDescription: "The version of the downloaded package file.",
				Computed:            true,
			},
			"Received: %T", expectedType, req.ProviderData)
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CONFIGURE,
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

	outputFile, fileSize, sha1, version, diags := api.Client().DownloadPackage(ctx, packageId, siteId,
	plan.OutputFile = types.StringValue(outputFile)
	plan.Version = types.StringValue(version)
			"internal_error_code":     plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CREATE,
		resp.Diagnostics.AddError("Download Package Creation Error", msg)
			"internal_error_code":     plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_CREATE,
		resp.Diagnostics.AddError("Download Package Creation Error", msg)
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
			"file": absPath,
			"package file.\n\nError: %s\nFile: %s", err.Error(), absPath)
			"error":               err.Error(),
			"file":                absPath,
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_READ,
		resp.Diagnostics.AddError("Download Package Refresh Error", msg)
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
	sha1, diags := plugin.GetFileSHA1(ctx, absPath)
	// save refreshed state
	// retrieve values from state
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
	// retrieve values from plan
	var plan tfPackageDownload
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
	// directory mode and overwrite flag updates require no changes locally
	if !plan.DirectoryMode.IsNull() && !plan.DirectoryMode.IsUnknown() {
		state.DirectoryMode = plan.DirectoryMode
	if !plan.OverwriteExistingFile.IsNull() && !plan.OverwriteExistingFile.IsUnknown() {
		state.OverwriteExistingFile = plan.OverwriteExistingFile
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
// Delete removes the Terraform resource.
func (r *PackageDownload) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// get the current state
	var state tfPackageDownload
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
	// if output file is empty, nothing to remove
	if state.OutputFile.IsNull() {
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
	// remove the file
	if err := os.Remove(absPath); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while removing the package file.\n\nError: %s\nFile: %s",
			err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_RESOURCE_PACKAGE_DOWNLOAD_DELETE,
		})
		resp.Diagnostics.AddError("Download Package Removal Error", msg)
	tflog.Debug(ctx, "Removed package file", map[string]interface{}{
		"file": absPath,