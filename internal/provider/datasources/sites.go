package datasources

// apiSitesModel defines the API model for a list of sites.
type apiSitesModel struct {
	AllSites apiAllSitesModel `json:"all_sites"`
	Sites    []apiSiteModel   `json:"sites"`
}

// apiAllSitesModel defines the API model for metadata about all sites returned in a request.
type apiAllSitesModel struct {
	ActiveLicenses int `json:"active_licenses"`
	TotalLicenses  int `json:"total_licenses"`
}
