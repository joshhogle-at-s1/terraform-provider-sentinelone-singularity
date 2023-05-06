package client

import "github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"

// SingularityProvider describes the provider data model.
type SingularityProvider struct {
	// APIClient is a handle to the Singularity REST API HTTP client.
	APIClient *api.Client
}
