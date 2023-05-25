package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
)

// Package defines the API model for a package.
type Package struct {
	Accounts      []packageAccount `json:"accounts"`
	CreatedAt     string           `json:"createdAt"`
	FileExtension string           `json:"fileExtension"`
	FileName      string           `json:"fileName"`
	FileSize      int64            `json:"fileSize"`
	Id            string           `json:"id"`
	Link          string           `json:"link"`
	MajorVersion  string           `json:"majorVersion"`
	MinorVersion  string           `json:"minorVersion"`
	OSArch        string           `json:"osArch"`
	OSType        string           `json:"osType"`
	PackageType   string           `json:"packageType"`
	PlatformType  string           `json:"platformType"`
	RangerVersion string           `json:"rangerVersion"`
	ScopeLevel    string           `json:"scopeLevel"`
	SHA1          string           `json:"sha1"`
	Sites         []packageSite    `json:"sites"`
	Status        string           `json:"status"`
	UpdatedAt     string           `json:"updatedAt"`
	Version       string           `json:"version"`
}

// packageAccount defines the API model for package accounts.
type packageAccount struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// packageSite defines the API model for package accounts.
type packageSite struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DownloadPackage is responsible for downloading the package with the given ID to a local path.
func (c *client) DownloadPackage(ctx context.Context, id, siteId, path, folderMode, fileMode string,
	overwrite bool) (string, int64, string, string, diag.Diagnostics) {

	// convert the path to an absolute path
	absPath, diags := plugin.ToAbsolutePath(ctx, path)
	if diags.HasError() {
		return "", 0, "", "", diags
	}
	ctx = tflog.SetField(ctx, "file", absPath)

	// create the file for writing
	outfile, diags := plugin.CreateFile(ctx, absPath, folderMode, fileMode, overwrite)
	if diags.HasError() {
		return "", 0, "", "", diags
	}

	// stream the download package into the output file
	diags = c.GetStream(ctx, fmt.Sprintf("/update/agent/download/%s/%s", siteId, id), map[string]string{},
		outfile)
	if diags.HasError() {
		return "", 0, "", "", diags
	}
	outfile.Close()

	// get the SHA1 and size of the destination file
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while retrieving information about the package file.\n\n"+
			"Error: %s\nFile: %s", err.Error(), absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_PACKAGE_DOWNLOAD_PACKAGE,
		})
		diags.AddError("Unexpected Internal Error", msg)
		os.Remove(absPath)
		return "", 0, "", "", diags
	}
	sha1, diags := plugin.GetFileSHA1(ctx, absPath)
	if diags.HasError() {
		os.Remove(absPath)
		return "", 0, "", "", diags
	}

	// finally get the version of the downloaded package
	pkg, diags := c.GetPackage(ctx, id)
	if diags.HasError() {
		os.Remove(absPath)
		return "", 0, "", "", diags
	}
	return absPath, fileInfo.Size(), sha1, pkg.Version, diags
}

// FindPackages returns a list of packages found based on the given query parameters.
func (c *client) FindPackages(ctx context.Context, queryParams PackageQueryParams) ([]Package, diag.Diagnostics) {
	var pkgs []Package
	var diags diag.Diagnostics
	getQueryParams := queryParams.toStringMap()
	for {
		// get a page of results
		result, diags := c.Get(ctx, "/update/agent/packages", getQueryParams)
		if diags.HasError() {
			return nil, diags
		}

		// parse the response
		var page []Package
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"list of Package objects.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_API_PACKAGE_FIND_PACKAGES,
			})
			diags.AddError("API Response Error", msg)
			return nil, diags
		}
		pkgs = append(pkgs, page...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		getQueryParams["cursor"] = result.Pagination.NextCursor
	}
	return pkgs, diags
}

// GetPackage returns the package with the matching ID.
func (c *client) GetPackage(ctx context.Context, id string) (*Package, diag.Diagnostics) {
	// query the API
	result, diags := c.Get(ctx, "/update/agent/packages", map[string]string{
		"ids": id,
	})
	if diags.HasError() {
		return nil, diags
	}

	// we are expecting exactly 1 package to be returned
	totalItems := result.Pagination.TotalItems
	if totalItems == 0 {
		msg := "No matching package was found. Try expanding your search or check that your package ID is valid."
		tflog.Error(ctx, msg, map[string]interface{}{
			"packages_found":      totalItems,
			"internal_error_code": plugin.ERR_API_PACKAGE_GET_PACKAGE,
		})
		diags.AddError("Package Not Found", msg)
		return nil, diags
	} else if totalItems > 1 {
		// this shouldn't happen but we want to be sure
		msg := fmt.Sprintf("This data source expects 1 matching package but %d were found. Please narrow your search.",
			totalItems)
		tflog.Error(ctx, msg, map[string]interface{}{
			"packages_found":      totalItems,
			"internal_error_code": plugin.ERR_API_PACKAGE_GET_PACKAGE,
		})
		diags.AddError("Multiple Packages Found", msg)
		return nil, diags
	}

	// parse the data returned
	var pkgs []Package
	if err := json.Unmarshal(result.Data, &pkgs); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Package object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_PACKAGE_GET_PACKAGE,
		})
		diags.AddError("API Response Error", msg)
		return nil, diags
	}
	return &pkgs[0], diags
}

// PackageQueryParams is used to hold query parameters for finding packages.
type PackageQueryParams struct {
	AccountIds    []string `json:"accountIds"`
	FileExtension *string  `json:"fileExtension"`
	Ids           []string `json:"ids"`
	MinorVersion  *string  `json:"minorVersion"`
	OSArches      []string `json:"osArches"`
	OSTypes       []string `json:"osTypes"`
	PackageTypes  []string `json:"packageTypes"`
	PlatformTypes []string `json:"platformTypes"`
	RangerVersion *string  `json:"rangerVersion"`
	Sha1          *string  `json:"sha1"`
	SiteIds       []string `json:"siteIds"`
	SortBy        *string  `json:"sortBy"`
	SortOrder     *string  `json:"sortOrder"`
	Status        []string `json:"status"`
	Version       *string  `json:"version"`
}

// toStringMap converts the object into a string map for actual query parameters.
func (p *PackageQueryParams) toStringMap() map[string]string {
	queryString := map[string]string{}
	if len(p.AccountIds) > 0 {
		queryString["accountIds"] = strings.Join(p.AccountIds, ",")
	}
	if p.FileExtension != nil {
		queryString["fileExtension"] = *p.FileExtension
	}
	if len(p.Ids) > 0 {
		queryString["ids"] = strings.Join(p.Ids, ",")
	}
	if p.MinorVersion != nil {
		queryString["minorVersion"] = *p.MinorVersion
	}
	if len(p.OSArches) > 0 {
		queryString["osArches"] = strings.Join(p.OSArches, ",")
	}
	if len(p.OSTypes) > 0 {
		queryString["osTypes"] = strings.Join(p.OSTypes, ",")
	}
	if len(p.PackageTypes) > 0 {
		queryString["packageTypes"] = strings.Join(p.PackageTypes, ",")
	}
	if len(p.PlatformTypes) > 0 {
		queryString["platformTypes"] = strings.Join(p.PlatformTypes, ",")
	}
	if p.RangerVersion != nil {
		queryString["rangerVersion"] = *p.RangerVersion
	}
	if p.Sha1 != nil {
		queryString["sha1"] = *p.Sha1
	}
	if len(p.SiteIds) > 0 {
		queryString["siteIds"] = strings.Join(p.SiteIds, ",")
	}
	if p.SortBy != nil {
		queryString["sortBy"] = *p.SortBy
	}
	if p.SortOrder != nil {
		queryString["sortOrder"] = *p.SortOrder
	}
	if len(p.Status) > 0 {
		queryString["status"] = strings.Join(p.Status, ",")
	}
	if p.Version != nil {
		queryString["version"] = *p.Version
	}
	return queryString
}
