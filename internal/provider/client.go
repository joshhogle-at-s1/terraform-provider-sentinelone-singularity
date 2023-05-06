package provider

import "github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/api"

// SingularityProviderClient holds data for sharing with the provider's data sources and resources.
type SingularityProviderClient struct {
	APIClient *api.Client
}
