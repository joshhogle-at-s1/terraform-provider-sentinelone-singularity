package api

import "sync"

const (
	// API_BASE_URI is the base URI for the REST API which indicates the version of the API to use.
	API_BASE_URI = "/web/api/v2.1"

	// USER_AGENT is the User-Agent string sent in HTTP requests to the API server.
	USER_AGENT = "SentinelOne-Singularity-Terraform-Provider"
)

var (
	// _client is the one and only global client
	_client *client

	// _once is used to make the singleton client creation thread-safe.
	_once sync.Once
)
