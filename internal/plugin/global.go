package plugin

const (
	// PROVIDER_NAME is the name of the Terraform provider.
	PROVIDER_NAME = "singularity"

	// PROVIDER_ADDRESS refers to the URL used for the provider in the Terraform registry.
	PROVIDER_ADDRESS = "registry.terraform.io/joshhogle-at-s1/sentinelone-singularity"
)

var (
	// Build holds the specific git-commit SHA for this build of the provider.
	Build string

	// Version holds the provider version.
	Version string
)
