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
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
)

// client is the HTTP client used for interacting with the S1 REST API.
type client struct {
	apiToken string
	baseURL  string
	conn     *http.Client
}

// Client returns the one and only global REST API client object.
//
// Note that you must call Init() to set the endpoint and API token before using the client for the first time.
func Client() *client {
	_once.Do(func() {
		_client = &client{
			conn: http.DefaultClient,
		}
	})
	return _client
}

// Get executes an HTTP GET query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) Get(ctx context.Context, uri string, queryParams map[string]string) (*apiResponse, diag.Diagnostics) {
	return c.doAndParse(ctx, http.MethodGet, uri, queryParams, map[string]interface{}{})
}

// GetStream executes an HTTP GET query and writes the response body directly to the given writer.
//
// This function should be used when you are expecting a binary response from the API.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) GetStream(ctx context.Context, uri string, queryParams map[string]string,
	writer io.Writer) diag.Diagnostics {

	return c.doAndStream(ctx, http.MethodGet, uri, queryParams, map[string]interface{}{}, writer)
}

// Init sets the base URL and API token to use in any API queries.
func (c *client) Init(endpoint, apiToken string) {
	c.baseURL = fmt.Sprintf("https://%s%s", strings.TrimPrefix(endpoint, "https://"), API_BASE_URI)
	c.apiToken = apiToken
}

// Post executes an HTTP POST query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) Post(ctx context.Context, uri string, body map[string]interface{}) (*apiResponse, diag.Diagnostics) {
	return c.doAndParse(ctx, http.MethodPost, uri, map[string]string{}, body)
}

// Put executes an HTTP PUT query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) Put(ctx context.Context, uri string, body map[string]interface{}) (*apiResponse, diag.Diagnostics) {
	return c.doAndParse(ctx, http.MethodPut, uri, map[string]string{}, body)
}

// Patch executes an HTTP PATCH query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) Patch(ctx context.Context, uri string, body map[string]interface{}) (
	*apiResponse, diag.Diagnostics) {
	return c.doAndParse(ctx, http.MethodPatch, uri, map[string]string{}, body)
}

// Delete executes an HTTP DELETE query.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) Delete(ctx context.Context, uri string, body map[string]interface{}) (
	*apiResponse, diag.Diagnostics) {
	return c.doAndParse(ctx, http.MethodDelete, uri, map[string]string{}, body)
}

// do is responsible for preparing and executing a request and checking the HTTP response code from the
// API server.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
//
// If this function does not return errors in the Diagnostics object, it is the caller's responsibility
// to close the response body.
func (c *client) do(ctx context.Context, method, url string, queryParams map[string]string,
	body map[string]interface{}) (*http.Response, diag.Diagnostics) {

	var diags diag.Diagnostics

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
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_API_CLIENT_DO,
			})
			diags.AddError("API Request Error", msg)
			return nil, diags
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
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_CLIENT_DO,
		})
		diags.AddError("API Request Error", msg)
		return nil, diags
	}

	// add headers to the request
	req.Header.Set("Authorization", fmt.Sprintf("ApiToken %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, application/octet-stream")
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
	tflog.Debug(ctx, "executing REST API query")
	resp, err := c.conn.Do(req)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while executing a request to the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s", err.Error(), url, method)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_CLIENT_DO,
		})
		diags.AddError("API Request Error", msg)
		return nil, diags
	}
	tflog.SetField(ctx, "status_code", resp.StatusCode)

	// status code >= 400 means there was an error
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// read the response body from the call
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while reading the response from the API Server.\n\n"+
				"Error: %s\nURL: %s\nMethod: %s\nHTTP Status Code: %d", err.Error(), url, method, resp.StatusCode)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_API_CLIENT_DO,
			})
			diags.AddError("API Response Error", msg)
			return nil, diags
		}
		tflog.Debug(ctx, "response received from API server", map[string]interface{}{"body": string(respBody)})

		// parse the response - note that we do not use json.NewDecoder().Decode() because it may think invalid json
		// is actually valid without using a loop so this is more efficient
		var result apiResponse
		unmarshalErr := json.Unmarshal(respBody, &result)
		if unmarshalErr != nil || result.Errors == nil || len(result.Errors) == 0 {
			msg := fmt.Sprintf("The request to the API server returned a non-successful error code.\n\n"+
				"URL: %s\nMethod: %s\nHTTP Status Code: %d\nResponse: %s\n",
				url, method, resp.StatusCode, respBody)
			tflog.Error(ctx, msg, map[string]interface{}{
				"response_body":       string(respBody),
				"internal_error_code": plugin.ERR_API_CLIENT_DO,
			})
			diags.AddError("API Response Error", msg)
		} else {
			// add a diagnostic error for every error in the API response
			for _, e := range result.Errors {
				msg := fmt.Sprintf("The request to the API server returned a non-successful error code.\n\n"+
					"URL: %s\nMethod: %s\nHTTP Status Code: %d\nAPI Code: %d\nSummary: %s\nDetails: %s",
					url, method, resp.StatusCode, e.Code, e.Title, e.Detail)
				tflog.Error(ctx, msg, map[string]interface{}{
					"api_code":            e.Code,
					"summary":             e.Title,
					"details":             e.Detail,
					"internal_error_code": plugin.ERR_API_CLIENT_DO,
				})
				diags.AddError("API Response Error", msg)
			}
		}
		return nil, diags
	}
	return resp, diags
}

// doAndParse handles executing a REST API query, verifying if any errors occurred and then parsing the
// API response body.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) doAndParse(ctx context.Context, method, uri string, queryParams map[string]string,
	body map[string]interface{}) (*apiResponse, diag.Diagnostics) {

	// build the request URL
	uri = strings.TrimPrefix(uri, "/")
	url := fmt.Sprintf("%s/%s", c.baseURL, uri)

	// configure log context
	ctx = tflog.SetField(ctx, "method", method)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "api_token", c.apiToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_token")

	// execute the actual request
	resp, diags := c.do(ctx, method, url, queryParams, body)
	if diags.HasError() {
		return nil, diags
	}
	defer resp.Body.Close()
	ctx = tflog.SetField(ctx, "status_code", resp.StatusCode)

	// read the response body from the call
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while reading the response from the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s\nHTTP Status Code: %d", err.Error(), url, method, resp.StatusCode)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_CLIENT_DO_AND_PARSE,
		})
		diags.AddError("API Response Error", msg)
		return nil, diags
	}
	tflog.Debug(ctx, "response received from API server", map[string]interface{}{"body": string(respBody)})

	// parse the response - note that we do not use json.NewDecoder().Decode() because it may think invalid json
	// is actually valid without using a loop so this is more efficient
	var result apiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s\nHTTP Status Code: %d", err.Error(), url, method, resp.StatusCode)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_CLIENT_DO_AND_PARSE,
		})
		diags.AddError("API Response Error", msg)
		return nil, diags
	}
	tflog.Debug(ctx, fmt.Sprintf("returning API response to caller: %+v", result))
	return &result, diags
}

// doAndStream handles executing a REST API query, verifying if any errors occurred and then streaming
// the response body to the given writer.
//
// Callers can check for errors using the HasErrors function on the Diagnostics object returned.
func (c *client) doAndStream(ctx context.Context, method, uri string, queryParams map[string]string,
	body map[string]interface{}, writer io.Writer) diag.Diagnostics {

	// build the request URL
	uri = strings.TrimPrefix(uri, "/")
	url := fmt.Sprintf("%s/%s", c.baseURL, uri)

	// configure log context
	ctx = tflog.SetField(ctx, "method", method)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "api_token", c.apiToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_token")

	// execute the request
	resp, diags := c.do(ctx, http.MethodGet, url, queryParams, map[string]interface{}{})
	if diags.HasError() {
		return diags
	}
	defer resp.Body.Close()
	ctx = tflog.SetField(ctx, "status_code", resp.StatusCode)

	// write the body directly to the stream
	if _, err := io.Copy(writer, resp.Body); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to read a response from the API Server.\n\n"+
			"Error: %s\nURL: %s\nMethod: %s\nHTTP Status Code %d", err.Error(), url, method, resp.StatusCode)
		tflog.Error(ctx, msg, map[string]interface{}{
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_API_CLIENT_DO_AND_STREAM,
		})
		diags.AddError("API Response Error", msg)
		return diags
	}
	return diags
}
