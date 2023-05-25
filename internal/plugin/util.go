package plugin

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// CreateDirectory creates the given path along with any parent directories setting the permissions using the
// given permissions mode.
func CreateDirectory(ctx context.Context, path, mode string) diag.Diagnostics {
	var diags diag.Diagnostics

	// create the destination folder if it does not exist
	folderExists, diags := PathExists(ctx, path)
	if diags.HasError() {
		return diags
	}
	if !folderExists {
		fsmode, diags := ParseFilesystemMode(ctx, mode)
		if diags.HasError() {
			return diags
		}
		if err := os.MkdirAll(path, fsmode); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while creating one or more folders in the path.\n\n"+
				"Error: %s\nPath: %s", err.Error(), path)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": ERR_UTIL_CREATE_DIRECTORY,
			})
			diags.AddError("Unexpected Internal Error", msg)
			return diags
		}
	}
	return diags
}

// CreateFile creates a new file (or truncates an existing file) at the given path and opens it for writing.
//
// Any parent folders are automatically created for you with the given folder mode. When the file is created, its
// mode will be set to the given file mode except on Windows, where it is ignored.
//
// If overwrite is false, an existing file will not be overwritten and an error will occur.
func CreateFile(ctx context.Context, path, folderMode, fileMode string, overwrite bool) (*os.File, diag.Diagnostics) {
	// convert the path to an absolute path
	absPath, diags := ToAbsolutePath(ctx, path)
	if diags.HasError() {
		return nil, diags
	}
	folder, file := filepath.Split(absPath)
	ctx = tflog.SetField(ctx, "parent_folder", folder)
	ctx = tflog.SetField(ctx, "file", file)

	// create the destination folder if it does not exist
	diags = CreateDirectory(ctx, folder, folderMode)
	if diags.HasError() {
		return nil, diags
	}

	// check to see if file exists
	if !overwrite {
		exists, diags := PathExists(ctx, absPath)
		if diags.HasError() {
			return nil, diags
		}
		if exists {
			msg := fmt.Sprintf("The destination file already exists and should not be overwritten.\n\nFile: %s", absPath)
			tflog.Error(ctx, msg, map[string]interface{}{
				"internal_error_code": ERR_UTIL_CREATE_FILE,
			})
			diags.AddError("File Exists", msg)
			return nil, diags
		}
	}

	// create the destination file for writing
	outfile, err := os.Create(absPath)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to open the file for writing.\n\n"+
			"Error: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": ERR_UTIL_CREATE_FILE,
		})
		diags.AddError("Unexpected Internal Error", msg)
		return nil, diags
	}

	// set file permissions (ignored on Windows systems)
	if runtime.GOOS != "windows" {
		fsmode, diags := ParseFilesystemMode(ctx, fileMode)
		if diags.HasError() {
			return nil, diags
		}
		if err := os.Chmod(absPath, fsmode); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while setting permissions on the file.\n\n"+
				"Error: %s\nMode: %s\nFile: %s", err.Error(), fileMode, absPath)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"file_mode":           fileMode,
				"internal_error_code": ERR_UTIL_CREATE_FILE,
			})
			diags.AddError("Unexpected Internal Error", msg)
			return nil, diags
		}
	}
	return outfile, diags
}

// GetFileSHA1 calculates the SHA1 hash of a file.
//
// If an error occurs, the function returns an empty string with an error in the diag.Diagnostics object.
func GetFileSHA1(ctx context.Context, file string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	// convert the path to an absolute path
	absPath, diags := ToAbsolutePath(ctx, file)
	if diags.HasError() {
		return "", diags
	}
	ctx = tflog.SetField(ctx, "file", absPath)

	// open the file for reading
	f, err := os.Open(absPath)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to open the given file for computing "+
			"the SHA1 checksum.\n\nError: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": ERR_UTIL_GET_FILE_SHA1,
		})
		diags.AddError("Unexpected Internal Error", msg)
		return "", diags
	}
	defer f.Close()

	// calculate the SHA1
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		msg := fmt.Sprintf("Failed to read file for computing SHA1.\n\n"+
			"Error: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": ERR_UTIL_GET_FILE_SHA1,
		})
		diags.AddError("Unexpected Internal Error", msg)
		return "", diags
	}
	return fmt.Sprintf("%x", h.Sum(nil)), diags
}

// GetWorkDir returns the path to the current working directory.
//
// This function will return "." in the case where os.Getwd() fails.
func GetWorkDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// ParseFilesystemMode converts a filesystem mode string into the corresponding octal mode.
func ParseFilesystemMode(ctx context.Context, mode string) (fs.FileMode, diag.Diagnostics) {
	var diags diag.Diagnostics

	fsmode, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the given filesystem mode string.\n\n"+
			"Error: %s\nMode: %s", err.Error(), mode)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"mode":                mode,
			"internal_error_code": ERR_UTIL_PARSE_FILESYSTEM_MODE,
		})
		diags.AddError("Unexpected Internal Error", msg)
		return 0, diags
	}
	return fs.FileMode(fsmode), diags
}

// PathExists determines whether or not the given path exists. The path may be a folder or a file.
//
// If an error occurs, the function returns false with an error in the diag.Diagnostics object.
func PathExists(ctx context.Context, path string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// convert the path to an absolute path
	absPath, diags := ToAbsolutePath(ctx, path)
	if diags.HasError() {
		return false, diags
	}
	ctx = tflog.SetField(ctx, "path", absPath)

	_, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return false, diags
	} else if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to get information on the given path.\n\n"+
			"Error: %s\nPath: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": ERR_UTIL_PATH_EXISTS,
		})
		diags.AddError("Unexpected Internal Error", msg)
	}
	return true, diags
}

// ToAbsolutePath converts the given path to an absolute path if it's not already an absolute path.
//
// If an error occurs the original path is returned and the diag.Diagnostics object will contain
// an error.
func ToAbsolutePath(ctx context.Context, path string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	// nothing to do - path is already absolute
	if filepath.IsAbs(path) {
		return path, diags
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while getting the absolute path to the file or folder.\n\n"+
			"Error: %s\nPath: %s", err.Error(), path)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"path":                path,
			"internal_error_code": ERR_UTIL_TO_ABSOLUTE_PATH,
		})
		diags.AddError("Unexpected Internal Error", msg)
		return path, diags
	}
	return absPath, diags
}
