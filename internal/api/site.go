package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
)

// Site defines the API model for a site.
type Site struct {
	AccountId           string      `json:"accountId"`
	AccountName         string      `json:"accountName"`
	ActiveLicenses      int         `json:"activeLicenses"`
	CreatedAt           string      `json:"createdAt"`
	Creator             string      `json:"creator"`
	CreatorId           string      `json:"creatorId"`
	Description         string      `json:"description"`
	Expiration          string      `json:"expiration"`
	ExternalId          string      `json:"externalId"`
	Id                  string      `json:"id"`
	IsDefault           bool        `json:"isDefault"`
	Licenses            siteLicense `json:"licenses"`
	Name                string      `json:"name"`
	RegistrationToken   string      `json:"registrationToken"`
	SiteType            string      `json:"siteType"`
	State               string      `json:"state"`
	TotalLicenses       int         `json:"totalLicenses"`
	UnlimitedExpiration bool        `json:"unlimitedExpiration"`
	UnlimitedLicenses   bool        `json:"unlimitedLicenses"`
	UpdatedAt           string      `json:"updatedAt"`
}

// siteLicense defines the API model for a site's license.
type siteLicense struct {
	Bundles  []siteLicenseBundle  `json:"bundles"`
	Modules  []siteLicenseModule  `json:"modules"`
	Settings []siteLicenseSetting `json:"settings"`
}

// siteLicenseBundle defines the API model for a site license's bundle.
type siteLicenseBundle struct {
	DisplayName   string                     `json:"displayName"`
	MajorVersion  int                        `json:"majorVersion"`
	MinorVersion  int                        `json:"minorVersion"`
	Name          string                     `json:"name"`
	Surfaces      []siteLicenseBundleSurface `json:"surfaces"`
	TotalSurfaces int                        `json:"totalSurfaces"`
}

// siteLicenseBundleSurface defines the API model for a site license bundle's surface.
type siteLicenseBundleSurface struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

// siteLicenseBundleSurface defines the API model for a site license's module.
type siteLicenseModule struct {
	DisplayName  string `json:"displayName"`
	MajorVersion int    `json:"majorVersion"`
	Name         string `json:"name"`
}

// siteLicenseBundleSurface defines the API model for a site license's setting.
type siteLicenseSetting struct {
	GroupName               string `json:"groupName"`
	Setting                 string `json:"setting"`
	SettingGroupDisplayName string `json:"settingGroupDisplayName"`
}

// Sites defines the API model for a list of sites.
type Sites struct {
	AllSites allSites `json:"all_sites"`
	Sites    []Site   `json:"sites"`
}

// allSites defines the API model for metadata about all sites returned in a request.
type allSites struct {
	ActiveLicenses int `json:"active_licenses"`
	TotalLicenses  int `json:"total_licenses"`
}

// FindSites returns a list of sites found based on the given query parameters.
func (c *client) FindSites(ctx context.Context, queryParams SiteQueryParams) ([]Site, diag.Diagnostics) {
	var sites []Site
	var diags diag.Diagnostics
	getQueryParams := queryParams.toStringMap()
	for {
		// get a page of results
		result, diags := c.Get(ctx, "/sites", getQueryParams)
		if diags.HasError() {
			return nil, diags
		}

		// parse the response
		var page Sites
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"list of Site objects.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_API_SITE_FIND_SITES,
			})
			diags.AddError("API Response Error", msg)
			return nil, diags
		}
		sites = append(sites, page.Sites...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		getQueryParams["cursor"] = result.Pagination.NextCursor
	}
	return sites, diags
}

// GetSite returns the site with the matching ID.
func (c *client) GetSite(ctx context.Context, id string) (*Site, diag.Diagnostics) {
	// query the API
	result, diags := c.Get(ctx, "/sites", map[string]string{
		"ids": id,
	})
	if diags.HasError() {
		return nil, diags
	}

	// we are expecting exactly 1 package to be returned
	totalItems := result.Pagination.TotalItems
	if totalItems == 0 {
		msg := "No matching site was found. Try expanding your search or check that your site ID is valid."
		tflog.Error(ctx, msg, map[string]interface{}{
			"sites_found":         totalItems,
			"internal_error_code": plugin.ERR_API_SITE_FIND_SITES,
		})
		diags.AddError("Site Not Found", msg)
		return nil, diags
	} else if totalItems > 1 {
		// this shouldn't happen but we want to be sure
		msg := fmt.Sprintf("This data source expects 1 matching site but %d were found. Please narrow your search.",
			totalItems)
		tflog.Error(ctx, msg, map[string]interface{}{
			"sites_found":         totalItems,
			"internal_error_code": plugin.ERR_API_SITE_FIND_SITES,
		})
		diags.AddError("Multiple Sites Found", msg)
		return nil, diags
	}

	// parse the data returned
	var sites []Site
	if err := json.Unmarshal(result.Data, &sites); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Site object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_SITE_FIND_SITES,
		})
		diags.AddError("API Response Error", msg)
		return nil, diags
	}
	return &sites[0], diags
}

// SiteQueryParams is used to hold query parameters for finding sites.
type SiteQueryParams struct {
	AccountIds          []string `json:"accountIds"`
	AccountNameContains []string `json:"accountName__contains"`
	ActiveLicenses      *int64   `json:"activeLicenses"`
	AdminOnly           *bool    `json:"adminOnly"`
	AvailableMoveSites  *bool    `json:"availableMoveSites"`
	CreatedAt           *string  `json:"createdAt"`
	Description         *string  `json:"description"`
	DescriptionContains []string `json:"description__contains"`
	Expiration          *string  `json:"expiration"`
	ExternalId          *string  `json:"externalId"`
	Features            []string `json:"features"`
	IsDefault           *bool    `json:"isDefault"`
	Modules             []string `json:"modules"`
	Name                *string  `json:"name"`
	NameContains        []string `json:"name__contains"`
	Query               *string  `json:"query"`
	RegistrationToken   *string  `json:"registrationToken"`
	SiteIds             []string `json:"siteIds"`
	SiteType            *string  `json:"siteType"`
	SortBy              *string  `json:"sortBy"`
	SortOrder           *string  `json:"sortOrder"`
	States              []string `json:"states"`
	TotalLicenses       *int64   `json:"totalLicenses"`
	UpdatedAt           *string  `json:"updatedAt"`
}

// toStringMap converts the object into a string map for actual query parameters.
func (p *SiteQueryParams) toStringMap() map[string]string {
	queryString := map[string]string{}
	if len(p.AccountIds) > 0 {
		queryString["accountIds"] = strings.Join(p.AccountIds, ",")
	}
	if len(p.AccountNameContains) > 0 {
		queryString["accountName__contains"] = strings.Join(p.AccountNameContains, ",")
	}
	if p.ActiveLicenses != nil {
		queryString["activeLicenses"] = fmt.Sprintf("%d", *p.ActiveLicenses)
	}
	if p.AdminOnly != nil {
		queryString["adminOnly"] = fmt.Sprintf("%t", *p.AdminOnly)
	}
	if p.AvailableMoveSites != nil {
		queryString["availableMoveSites"] = fmt.Sprintf("%t", *p.AvailableMoveSites)
	}
	if p.CreatedAt != nil {
		queryString["createdAt"] = *p.CreatedAt
	}
	if p.Description != nil {
		queryString["description"] = *p.Description
	}
	if len(p.DescriptionContains) > 0 {
		queryString["description__contains"] = strings.Join(p.DescriptionContains, ",")
	}
	if p.Expiration != nil {
		queryString["expiration"] = *p.Expiration
	}
	if p.ExternalId != nil {
		queryString["externalId"] = *p.ExternalId
	}
	if len(p.Features) > 0 {
		queryString["features"] = strings.Join(p.Features, ",")
	}
	if p.IsDefault != nil {
		queryString["isDefault"] = fmt.Sprintf("%t", *p.IsDefault)
	}
	if len(p.Modules) > 0 {
		queryString["modules"] = strings.Join(p.Modules, ",")
	}
	if p.Name != nil {
		queryString["name"] = *p.Name
	}
	if len(p.NameContains) > 0 {
		queryString["name__contains"] = strings.Join(p.NameContains, ",")
	}
	if p.Query != nil {
		queryString["query"] = *p.Query
	}
	if p.RegistrationToken != nil {
		queryString["registrationToken"] = *p.RegistrationToken
	}
	if len(p.SiteIds) > 0 {
		queryString["siteIds"] = strings.Join(p.SiteIds, ",")
	}
	if p.SiteType != nil {
		queryString["siteType"] = *p.SiteType
	}
	if p.SortBy != nil {
		queryString["sortBy"] = *p.SortBy
	}
	if p.SortOrder != nil {
		queryString["sortOrder"] = *p.SortOrder
	}
	if len(p.States) > 0 {
		queryString["states"] = strings.Join(p.States, ",")
	}
	if p.TotalLicenses != nil {
		queryString["totalLicenses"] = fmt.Sprintf("%d", *p.TotalLicenses)
	}
	if p.UpdatedAt != nil {
		queryString["updatedAt"] = *p.UpdatedAt
	}
	return queryString
}
