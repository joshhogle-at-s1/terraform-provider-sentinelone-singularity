package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client is the HTTP client used for interacting with the S1 REST API.
type Client struct {
	apiToken string
	baseURL  string
	conn     *http.Client
}

// NewClient creates a new REST API client.
func NewClient(apiToken, endpoint string) *Client {
	// remove https:// from the endpoint if it exists
	endpoint = strings.TrimPrefix(endpoint, "https://")

	return &Client{
		apiToken: apiToken,
		baseURL:  fmt.Sprintf("https://%s%s", endpoint, API_BASE_URI),
		conn:     http.DefaultClient,
	}
}

// Get executes an HTTP GET query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) Get(ctx context.Context, uri string, queryParams map[string]string) (*apiResponse, diag.Diagnostics) {
	return c.executeQuery(ctx, http.MethodGet, uri, queryParams, map[string]interface{}{})
}

// Post executes an HTTP POST query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) Post(ctx context.Context, uri string, body map[string]interface{}) (*apiResponse, diag.Diagnostics) {
	return c.executeQuery(ctx, http.MethodPost, uri, map[string]string{}, body)
}

// Put executes an HTTP PUT query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) Put(ctx context.Context, uri string, body map[string]interface{}) (*apiResponse, diag.Diagnostics) {
	return c.executeQuery(ctx, http.MethodPut, uri, map[string]string{}, body)
}

// Patch executes an HTTP PATCH query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) Patch(ctx context.Context, uri string, body map[string]interface{}) (
	*apiResponse, diag.Diagnostics) {
	return c.executeQuery(ctx, http.MethodPatch, uri, map[string]string{}, body)
}

// Delete executes an HTTP DELETE query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) Delete(ctx context.Context, uri string, body map[string]interface{}) (
	*apiResponse, diag.Diagnostics) {
	return c.executeQuery(ctx, http.MethodDelete, uri, map[string]string{}, body)
}

// executeQuery handles executing a REST API query and parsing and returning the response body, verifying first if any
// errors have occurred.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *Client) executeQuery(ctx context.Context, method, uri string, queryParams map[string]string,
	body map[string]interface{}) (*apiResponse, diag.Diagnostics) {

	var diag diag.Diagnostics
	var result apiResponse

	// build the request URL
	uri = strings.TrimPrefix(uri, "/")
	url := fmt.Sprintf("%s/%s", c.baseURL, uri)

	// configure log context
	ctx = tflog.SetField(ctx, "method", method)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "api_token", c.apiToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_token")

	// prepare body for the request, if there is any
	var payload *bytes.Buffer
	if len(body) > 0 {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while attempting to create a request to the API Server.\n\n"+
				"Error: %s\nURL: %s\nMethod: %s", err.Error(), url, method)
			tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
			diag.AddError("API Request Error", msg)
			return nil, diag
		}
		payload = bytes.NewBuffer(jsonBody)
		ctx = tflog.SetField(ctx, "body", string(jsonBody))
	}

	// create the request
	var req *http.Request
	var err error
	if payload == nil { // sending a typed nil to NewRequest will cause a panic
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, payload)
	}
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to create a request to the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s", err.Error(), url, method)
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		diag.AddError("API Request Error", msg)
		return nil, diag
	}

	// add headers to the request
	req.Header.Set("Authorization", fmt.Sprintf("ApiToken %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", USER_AGENT)

	// add query parameters, if there are any
	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		ctx = tflog.SetField(ctx, "query_params", req.URL.RawQuery)
	}

	// execute the request
	tflog.Trace(ctx, "executing REST API query")
	resp, err := c.conn.Do(req)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while executing a request to the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s", err.Error(), url, method)
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	}
	defer resp.Body.Close()
	tflog.SetField(ctx, "status_code", resp.StatusCode)

	// read the response body from the call
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while reading the response from the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s\nHTTP Status Code: %d", err.Error(), url, method, resp.StatusCode)
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	}
	tflog.Trace(ctx, "response received from API server", map[string]interface{}{"body": string(respBody)})

	// parse the response - note that we do not use json.NewDecoder().Decode() because it may think invalid json
	// is actually valid without using a loop so this is more efficient
	unmarshalErr := json.Unmarshal(respBody, &result)

	// status code >= 400 means there was an error
	if resp.StatusCode >= 400 {
		// no API errors were parsed from the response
		if unmarshalErr != nil || result.Errors == nil || len(result.Errors) == 0 {
			msg := fmt.Sprintf("The request to the API server returned a non-successful error code.\n\n"+
				"URL: %s\nMethod: %s\nHTTP Status Code: %d\nResponse: %s\n",
				url, method, resp.StatusCode, respBody)
			tflog.Error(ctx, msg, map[string]interface{}{"response_body": string(respBody)})
			diag.AddError("API Query Failed", msg)
		} else {
			// add a diagnostic error for every error in the API response
			for _, e := range result.Errors {
				msg := fmt.Sprintf("The request to the API server returned a non-successful error code.\n\n"+
					"URL: %s\nMethod: %s\nHTTP Status Code: %d\nAPI Code: %d\nSummary: %s\nDetails: %s",
					url, method, resp.StatusCode, e.Code, e.Title, e.Detail)
				tflog.Error(ctx, msg, map[string]interface{}{"api_code": e.Code, "summary": e.Title, "details": e.Detail})
				diag.AddError("API Query Failed", msg)
			}
		}
		return nil, diag
	}

	// if the API returned successfully but the response was not parsed, something is wrong
	// (this shouldn't happen but it's a failsafe)
	if unmarshalErr != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s", unmarshalErr.Error(), url, method)
		tflog.Error(ctx, msg, map[string]interface{}{"error": unmarshalErr.Error()})
		diag.AddError("API Query Failed", msg)
		return nil, diag
	}
	tflog.Trace(ctx, fmt.Sprintf("returning API response to caller: %+v", result))

	// return the response body to be parsed by the caller
	return &result, diag
}
