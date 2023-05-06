package api

import "encoding/json"

// apiError holds a single error returned by an API call.
type apiError struct {
	// Code is the S1 error code returned by the API.
	Code int `json:"code"`

	// Detail contains details around the error that occurred.
	Detail string `json:"detail"`

	// Title is the title or summary of the error that occurred.
	Title string `json:"title"`
}

// pagination defines information on the current page of results.
type pagination struct {
	// TotalItems holds the total number of items returned by the query.
	TotalItems int `json:"totalItems"`

	// NextCursor holds the cursor to the next page of results.
	NextCursor string `json:"nextCursor"`
}

// apiResponse defines the generic response to any API query.
type apiResponse struct {
	// Pagination holds data regarding which page of results the API has returned.
	Pagination pagination `json:"pagination"`

	// Data holds the actual data returned from the query.
	Data json.RawMessage `json:"data"`

	// Errors holds any errors that occurred during the query.
	Errors []apiError `json:"errors"`
}
