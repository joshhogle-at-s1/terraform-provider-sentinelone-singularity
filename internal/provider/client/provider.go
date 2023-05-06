package client

import "github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"

// SingularityProvider describes the provider data model.
type SingularityProvider struct {
	APIClient *api.Client
}
