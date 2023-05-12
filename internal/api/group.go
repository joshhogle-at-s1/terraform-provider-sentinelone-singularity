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

// Group defines the API model for a group.
type Group struct {
	CreatedAt         string `json:"createdAt"`
	Creator           string `json:"creator"`
	CreatorId         string `json:"creatorId"`
	Description       string `json:"description"`
	FilterId          string `json:"filterId"`
	FilterName        string `json:"filterName"`
	Id                string `json:"id"`
	Inherits          bool   `json:"inherits"`
	IsDefault         bool   `json:"isDefault"`
	Name              string `json:"name"`
	Rank              int    `json:"rank"`
	RegistrationToken string `json:"registrationToken"`
	SiteId            string `json:"siteId"`
	TotalAgents       int    `json:"totalAgents"`
	Type              string `json:"type"`
	UpdatedAt         string `json:"updatedAt"`
}

// FindGroups returns a list of groups found based on the given query parameters.
func (c *client) FindGroups(ctx context.Context, queryParams GroupQueryParams) ([]Group, diag.Diagnostics) {
	var groups []Group
	var diags diag.Diagnostics
	getQueryParams := queryParams.toStringMap()
	for {
		// get a page of results
		result, diags := c.Get(ctx, "/groups", getQueryParams)
		if diags.HasError() {
			return nil, diags
		}

		// parse the response
		var page []Group
		if err := json.Unmarshal(result.Data, &page); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
				"list of Group objects.\n\nError: %s", err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_API_GROUP_FIND_GROUPS,
			})
			diags.AddError("API Response Error", msg)
			return nil, diags
		}
		groups = append(groups, page...)

		// get the next page of results until there is no next cursor
		if result.Pagination.NextCursor == "" {
			break
		}
		getQueryParams["cursor"] = result.Pagination.NextCursor
	}
	return groups, diags
}

// GetGroup returns the group with the matching ID.
func (c *client) GetGroup(ctx context.Context, id string) (*Group, diag.Diagnostics) {
	// query the API
	result, diags := c.Get(ctx, "/groups", map[string]string{
		"ids": id,
	})
	if diags.HasError() {
		return nil, diags
	}

	// we are expecting exactly 1 package to be returned
	totalItems := result.Pagination.TotalItems
	if totalItems == 0 {
		msg := "No matching group was found. Try expanding your search or check that your group ID is valid."
		tflog.Error(ctx, msg, map[string]interface{}{
			"groups_found":        totalItems,
			"internal_error_code": plugin.ERR_API_GROUP_GET_GROUP,
		})
		diags.AddError("Group Not Found", msg)
		return nil, diags
	} else if totalItems > 1 {
		// this shouldn't happen but we want to be sure
		msg := fmt.Sprintf("This data source expects 1 matching group but %d were found. Please narrow your search.",
			totalItems)
		tflog.Error(ctx, msg, map[string]interface{}{
			"groups_found":        totalItems,
			"internal_error_code": plugin.ERR_API_GROUP_GET_GROUP,
		})
		diags.AddError("Multiple Groups Found", msg)
		return nil, diags
	}

	// parse the data returned
	var groups []Group
	if err := json.Unmarshal(result.Data, &groups); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Group object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_GROUP_GET_GROUP,
		})
		diags.AddError("API Response Error", msg)
		return nil, diags
	}
	return &groups[0], diags
}

// GroupQueryParams is used to hold query parameters for finding groups.
type GroupQueryParams struct {
	AccountIds        []string `json:"accountIds"`
	Description       *string  `json:"description"`
	GroupIds          []string `json:"groupIds"`
	IsDefault         *bool    `json:"isDefault"`
	Name              *string  `json:"name"`
	Query             *string  `json:"query"`
	Rank              *int64   `json:"rank"`
	RegistrationToken *string  `json:"registrationToken"`
	SiteIds           []string `json:"siteIds"`
	SortBy            *string  `json:"sortBy"`
	SortOrder         *string  `json:"sortOrder"`
	Types             []string `json:"types"`
	UpdatedAfter      *string  `json:"updatedAt__gt"`
	UpdatedAtOrAfter  *string  `json:"updatedAt__gte"`
	UpdatedAtOrBefore *string  `json:"updatedAt__lte"`
	UpdatedBefore     *string  `json:"updatedAt__lt"`
}

// toStringMap converts the object into a string map for actual query parameters.
func (p *GroupQueryParams) toStringMap() map[string]string {
	queryString := map[string]string{}
	if len(p.AccountIds) > 0 {
		queryString["accountIds"] = strings.Join(p.AccountIds, ",")
	}
	if p.Description != nil {
		queryString["description"] = *p.Description
	}
	if len(p.GroupIds) > 0 {
		queryString["ids"] = strings.Join(p.GroupIds, ",")
	}
	if p.IsDefault != nil {
		queryString["isDefault"] = fmt.Sprintf("%t", *p.IsDefault)
	}
	if p.Name != nil {
		queryString["name"] = *p.Name
	}
	if p.Query != nil {
		queryString["query"] = *p.Query
	}
	if p.Rank != nil {
		queryString["rank"] = fmt.Sprintf("%d", *p.Rank)
	}
	if p.RegistrationToken != nil {
		queryString["registrationToken"] = *p.RegistrationToken
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
	if len(p.Types) > 0 {
		queryString["types"] = strings.Join(p.Types, ",")
	}
	if p.UpdatedAfter != nil {
		queryString["updatedAt__gt"] = *p.UpdatedAfter
	}
	if p.UpdatedAtOrAfter != nil {
		queryString["updatedAt__gte"] = *p.UpdatedAtOrAfter
	}
	if p.UpdatedAtOrBefore != nil {
		queryString["updatedAt__lte"] = *p.UpdatedAtOrAfter
	}
	if p.UpdatedBefore != nil {
		queryString["updatedAt__lt"] = *p.UpdatedBefore
	}
	return queryString
}
