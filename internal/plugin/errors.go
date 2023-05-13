package plugin

// Error codes
const (
	ERR_PROVIDER_CONFIGURE = 400

	ERR_VALIDATOR_ENUM_STRING     = 450
	ERR_VALIDATOR_ENUM_STRINGLIST = 451

	ERR_UTIL_CREATE_FILE           = 500
	ERR_UTIL_GET_FILE_SHA1         = 501
	ERR_UTIL_PATH_EXISTS           = 502
	ERR_UTIL_PARSE_FILESYSTEM_MODE = 503
	ERR_UTIL_TO_ABSOLUTE_PATH      = 504
	ERR_UTIL_CREATE_DIRECTORY      = 505

	ERR_API_CLIENT_DO                = 1000
	ERR_API_CLIENT_DO_AND_PARSE      = 1001
	ERR_API_CLIENT_DO_AND_STREAM     = 1002
	ERR_API_PACKAGE_FIND_PACKAGES    = 1003
	ERR_API_PACKAGE_DOWNLOAD_PACKAGE = 1004
	ERR_API_PACKAGE_GET_PACKAGE      = 1005
	ERR_API_GROUP_FIND_GROUPS        = 1006
	ERR_API_GROUP_GET_GROUP          = 1007
	ERR_API_SITE_FIND_SITES          = 1008
	ERR_API_SITE_GET_SITES           = 1009

	ERR_DATASOURCE_GROUP_CONFIGURE    = 2000
	ERR_DATASOURCE_PACKAGE_CONFIGURE  = 2001
	ERR_DATASOURCE_SITE_CONFIGURE     = 2002
	ERR_DATASOURCE_GROUPS_CONFIGURE   = 2003
	ERR_DATASOURCE_PACKAGES_CONFIGURE = 2004
	ERR_DATASOURCE_SITES_CONFIGURE    = 2005

	ERR_RESOURCE_PACKAGE_DOWNLOAD_CONFIGURE = 3000
	ERR_RESOURCE_PACKAGE_DOWNLOAD_CREATE    = 3001
	ERR_RESOURCE_PACKAGE_DOWNLOAD_READ      = 3002
	ERR_RESOURCE_PACKAGE_DOWNLOAD_UPDATE    = 3003
	ERR_RESOURCE_PACKAGE_DOWNLOAD_DELETE    = 3004
	ERR_RESOURCE_PACKAGE_DOWNLOAD_MODIFIERS = 3005
)
