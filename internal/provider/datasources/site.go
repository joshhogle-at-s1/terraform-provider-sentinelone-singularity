package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/client"
)

// ensure implementation satisfied expected interfaces.
var (
	_ datasource.DataSource              = &Site{}
	_ datasource.DataSourceWithConfigure = &Site{}
)

// apiSiteModel defines the API model for a site.
type apiSiteModel struct {
	AccountId           string              `json:"accountId"`
	AccountName         string              `json:"accountName"`
	ActiveLicenses      int                 `json:"activeLicenses"`
	CreatedAt           string              `json:"createdAt"`
	Creator             string              `json:"creator"`
	CreatorId           string              `json:"creatorId"`
	Description         string              `json:"description"`
	Expiration          string              `json:"expiration"`
	ExternalId          string              `json:"externalId"`
	Id                  string              `json:"id"`
	IsDefault           bool                `json:"isDefault"`
	Licenses            apiSiteLicenseModel `json:"licenses"`
	Name                string              `json:"name"`
	RegistrationToken   string              `json:"registrationToken"`
	SiteType            string              `json:"siteType"`
	State               string              `json:"state"`
	TotalLicenses       int                 `json:"totalLicenses"`
	UnlimitedExpiration bool                `json:"unlimitedExpiration"`
	UnlimitedLicenses   bool                `json:"unlimitedLicenses"`
	UpdatedAt           string              `json:"updatedAt"`
}

// apiSiteLicenseModel defines the API model for a site's license.
type apiSiteLicenseModel struct {
	Bundles  []apiSiteLicenseBundleModel  `json:"bundles"`
	Modules  []apiSiteLicenseModuleModel  `json:"modules"`
	Settings []apiSiteLicenseSettingModel `json:"settings"`
}

// apiSiteLicenseBundleModel defines the API model for a site license's bundle.
type apiSiteLicenseBundleModel struct {
	DisplayName   string                             `json:"displayName"`
	MajorVersion  int                                `json:"majorVersion"`
	MinorVersion  int                                `json:"minorVersion"`
	Name          string                             `json:"name"`
	Surfaces      []apiSiteLicenseBundleSurfaceModel `json:"surfaces"`
	TotalSurfaces int                                `json:"totalSurfaces"`
}

// apiSiteLicenseBundleSurfaceModel defines the API model for a site license bundle's surface.
type apiSiteLicenseBundleSurfaceModel struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

// apiSiteLicenseBundleSurfaceModel defines the API model for a site license's module.
type apiSiteLicenseModuleModel struct {
	DisplayName  string `json:"displayName"`
	MajorVersion int    `json:"majorVersion"`
	Name         string `json:"name"`
}

// apiSiteLicenseBundleSurfaceModel defines the API model for a site license's setting.
type apiSiteLicenseSettingModel struct {
	GroupName               string `json:"groupName"`
	Setting                 string `json:"setting"`
	SettingGroupDisplayName string `json:"settingGroupDisplayName"`
}

// tfSiteModel defines the Terraform model for a site.
type tfSiteModel struct {
	AccountId           types.String        `tfsdk:"account_id"`
	AccountName         types.String        `tfsdk:"account_name"`
	ActiveLicenses      types.Int64         `tfsdk:"active_licenses"`
	CreatedAt           types.String        `tfsdk:"created_at"`
	Creator             types.String        `tfsdk:"creator"`
	CreatorId           types.String        `tfsdk:"creator_id"`
	Description         types.String        `tfsdk:"description"`
	Expiration          types.String        `tfsdk:"expiration"`
	ExternalId          types.String        `tfsdk:"external_id"`
	Id                  types.String        `tfsdk:"id"`
	IsDefault           types.Bool          `tfsdk:"is_default"`
	Licenses            *tfSiteLicenseModel `tfsdk:"licenses"`
	Name                types.String        `tfsdk:"name"`
	RegistrationToken   types.String        `tfsdk:"registration_token"`
	SiteType            types.String        `tfsdk:"site_type"`
	State               types.String        `tfsdk:"state"`
	TotalLicenses       types.Int64         `tfsdk:"total_licenses"`
	UnlimitedExpiration types.Bool          `tfsdk:"unlimited_expiration"`
	UnlimitedLicenses   types.Bool          `tfsdk:"unlimited_licenses"`
	UpdatedAt           types.String        `tfsdk:"updated_at"`
}

// tfSiteLicenseModel defines the Terraform model for a site's license.
type tfSiteLicenseModel struct {
	Bundles  []tfSiteLicenseBundleModel  `tfsdk:"bundles"`
	Modules  []tfSiteLicenseModuleModel  `tfsdk:"modules"`
	Settings []tfSiteLicenseSettingModel `tfsdk:"settings"`
}

// tfSiteLicenseBundleModel defines the Terraform model for a site license's bundle.
type tfSiteLicenseBundleModel struct {
	DisplayName   types.String                      `tfsdk:"display_name"`
	MajorVersion  types.Int64                       `tfsdk:"major_version"`
	MinorVersion  types.Int64                       `tfsdk:"minor_version"`
	Name          types.String                      `tfsdk:"name"`
	Surfaces      []tfSiteLicenseBundleSurfaceModel `tfsdk:"surfaces"`
	TotalSurfaces types.Int64                       `tfsdk:"total_surfaces"`
}

// tfSiteLicenseBundleSurfaceModel defines the Terraform model for a site license bundle's surface.
type tfSiteLicenseBundleSurfaceModel struct {
	Count types.Int64  `tfsdk:"count"`
	Name  types.String `tfsdk:"name"`
}

// tfSiteLicenseBundleSurfaceModel defines the Terraform model for a site license's module.
type tfSiteLicenseModuleModel struct {
	DisplayName  types.String `tfsdk:"display_name"`
	MajorVersion types.Int64  `tfsdk:"major_version"`
	Name         types.String `tfsdk:"name"`
}

// tfSiteLicenseBundleSurfaceModel defines the Terraform model for a site license's setting.
type tfSiteLicenseSettingModel struct {
	GroupName               types.String `tfsdk:"group_name"`
	Setting                 types.String `tfsdk:"setting"`
	SettingGroupDisplayName types.String `tfsdk:"setting_group_display_name"`
}

// NewSite creates a new Site object.
func NewSite() datasource.DataSource {
	return &Site{}
}

// Site is a data source used to store details about a single site.
type Site struct {
	client *client.SingularityProvider
}

// Metadata returns metadata about the data source.
func (d *Site) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

// Schema defines the parameters for the data sources's configuration.
func (d *Site) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	siteSchema := getSiteSchema(ctx)

	// override the default schema
	siteSchema.Attributes["id"] = schema.StringAttribute{
		Description:         "ID of the site.",
		MarkdownDescription: "ID of the site.",
		Required:            true,
	}
	resp.Schema = siteSchema
}

// Configure initializes the configuration for the data source.
func (d *Site) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.SingularityProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Type",
			fmt.Sprintf("Expected *client.SingularityProvider, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)
		return
	}
	d.client = client
}

// Read retrieves data from the API.
func (d *Site) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tfSiteModel

	// read configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// construct query parameters
	queryParams := map[string]string{
		"siteIds": data.Id.ValueString(), // 'id' is required so no need to check
	}

	// find the matching site
	result, diag := d.client.APIClient.Get(ctx, "/sites", queryParams)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the response - we are expecting exactly 1 site to be returned
	numSites := result.Pagination.TotalItems
	if numSites == 0 {
		msg := "No matching site was found. Try expanding your search or check that your site ID is valid."
		tflog.Error(ctx, msg, map[string]interface{}{"sites_found": numSites})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	} else if numSites > 1 {
		msg := fmt.Sprintf("This data source expects 1 matching site but %d were found. Please narrow your search.",
			numSites)
		tflog.Error(ctx, msg, map[string]interface{}{"sites_found": numSites})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	}
	var sites apiSitesModel
	if err := json.Unmarshal(result.Data, &sites); err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while parsing the response from the API Server into a "+
			"Site object.\n\nError: %s", err.Error())
		tflog.Error(ctx, msg, map[string]interface{}{"error": err.Error()})
		resp.Diagnostics.AddError("API Query Failed", msg)
		return
	}

	// convert the API object to the Terraform object
	resp.Diagnostics.Append(resp.State.Set(ctx, terraformSiteFromAPI(ctx, sites.Sites[0]))...)
}

// getSiteSchema returns a default Terraform schema where all values are computed.
func getSiteSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "This data source is used for getting details on a specific site.",
		MarkdownDescription: "This data source is used for getting details on a specific site.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description:         "ID of account to which the site belongs.",
				MarkdownDescription: "ID of account to which the site belongs.",
				Computed:            true,
			},
			"account_name": schema.StringAttribute{
				Description:         "Name of account to which the site belongs.",
				MarkdownDescription: "Name of account to which the site belongs.",
				Computed:            true,
			},
			"active_licenses": schema.Int64Attribute{
				Description:         "Number of active licenses for the site.",
				MarkdownDescription: "Number of active licenses for the site.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				Description:         "Timestamp of when the site was created.",
				MarkdownDescription: "Timestamp of when the site was created.",
				Computed:            true,
			},
			"creator": schema.StringAttribute{
				Description:         "Full name of the user who created the site.",
				MarkdownDescription: "Full name of the user who created the site.",
				Computed:            true,
			},
			"creator_id": schema.StringAttribute{
				Description:         "ID of the user who created the site.",
				MarkdownDescription: "ID of the user who created the site.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				Description:         "User-defined description of the site.",
				MarkdownDescription: "User-defined description of the site.",
				Computed:            true,
			},
			"expiration": schema.StringAttribute{
				Description:         "Date and time that the site expires.",
				MarkdownDescription: "Date and time that the site expires.",
				Computed:            true,
			},
			"external_id": schema.StringAttribute{
				Description:         "ID of CRM external system.",
				MarkdownDescription: "ID of CRM external system.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Description:         "ID of the site.",
				MarkdownDescription: "ID of the site.",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				Description:         "Whether or not the site is the default site.",
				MarkdownDescription: "Whether or not the site is the default site.",
				Computed:            true,
			},
			"licenses": schema.SingleNestedAttribute{
				Description:         "List of licenses associated with the site.",
				MarkdownDescription: "List of licenses associated with the site.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"bundles": schema.ListNestedAttribute{
						Description:         "License bundles.",
						MarkdownDescription: "License bundles.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"display_name": schema.StringAttribute{
									Description:         "Bundle display name.",
									MarkdownDescription: "Bundle display name.",
									Computed:            true,
								},
								"major_version": schema.Int64Attribute{
									Description:         "Bundle major version.",
									MarkdownDescription: "Bundle major version.",
									Computed:            true,
								},
								"minor_version": schema.Int64Attribute{
									Description:         "Bundle minor version.",
									MarkdownDescription: "Bundle minor version.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									Description:         "Bundle API name.",
									MarkdownDescription: "Bundle API name.",
									Computed:            true,
								},
								"surfaces": schema.ListNestedAttribute{
									Description:         "Surfaces in the bundle.",
									MarkdownDescription: "Surfaces in the bundle.",
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"count": schema.Int64Attribute{
												Description:         "Surface count or -1 for unlimited.",
												MarkdownDescription: "Surface count or -1 for unlimited.",
												Computed:            true,
											},
											"name": schema.StringAttribute{
												Description:         "Surface name.",
												MarkdownDescription: "Surface name.",
												Computed:            true,
											},
										},
									},
								},
								"total_surfaces": schema.Int64Attribute{
									Description:         "Total number of surfaces in the bundle or -1 for unlimited.",
									MarkdownDescription: "Total number of surfaces in the bundle or -1 for unlimited.",
									Computed:            true,
								},
							},
						},
					},
					"modules": schema.ListNestedAttribute{
						Description:         "License add-ons.",
						MarkdownDescription: "License add-ons.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"display_name": schema.StringAttribute{
									Description:         "Add-on display name.",
									MarkdownDescription: "Add-on display name.",
									Computed:            true,
								},
								"major_version": schema.Int64Attribute{
									Description:         "Add-on major version.",
									MarkdownDescription: "Add-on major version.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									Description:         "Add-on API name.",
									MarkdownDescription: "Add-on API name.",
									Computed:            true,
								},
							},
						},
					},
					"settings": schema.ListNestedAttribute{
						Description:         "License settings.",
						MarkdownDescription: "License Settings.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"group_name": schema.StringAttribute{
									Description:         "Setting group name.",
									MarkdownDescription: "Setting group name.",
									Computed:            true,
								},
								"setting": schema.StringAttribute{
									Description:         "Setting display name.",
									MarkdownDescription: "Setting display name.",
									Computed:            true,
								},
								"setting_group_display_name": schema.StringAttribute{
									Description:         "Setting group display name.",
									MarkdownDescription: "Setting group display name.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the site.",
				MarkdownDescription: "Name of the site.",
				Computed:            true,
			},
			"registration_token": schema.StringAttribute{
				Description:         "Registration token for the site.",
				MarkdownDescription: "Registration token for the site.",
				Computed:            true,
			},
			"site_type": schema.StringAttribute{
				Description:         "Type of site.",
				MarkdownDescription: "Type of site.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				Description:         "State of site.",
				MarkdownDescription: "State of site.",
				Computed:            true,
			},
			"total_licenses": schema.Int64Attribute{
				Description:         "Number of licenses.",
				MarkdownDescription: "Number of licenses.",
				Computed:            true,
			},
			"unlimited_expiration": schema.BoolAttribute{
				Description:         "Whether or not the site expires.",
				MarkdownDescription: "Whether or not the site expires.",
				Computed:            true,
			},
			"unlimited_licenses": schema.BoolAttribute{
				Description:         "Whether or not the site has unlimited licenses.",
				MarkdownDescription: "Whether or not the site has unlimited licenses.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				Description:         "Timestamp of when the site was last updated.",
				MarkdownDescription: "Timestamp of when the site was last updated.",
				Computed:            true,
			},
		},
	}
}

// terraformSiteFromAPI converts an API site into a Terraform site.
func terraformSiteFromAPI(ctx context.Context, site apiSiteModel) tfSiteModel {
	tfsite := tfSiteModel{
		AccountId:           types.StringValue(site.AccountId),
		AccountName:         types.StringValue(site.AccountName),
		ActiveLicenses:      types.Int64Value(int64(site.ActiveLicenses)),
		CreatedAt:           types.StringValue(site.CreatedAt),
		Creator:             types.StringValue(site.Creator),
		CreatorId:           types.StringValue(site.CreatorId),
		Description:         types.StringValue(site.Description),
		Expiration:          types.StringValue(site.Expiration),
		ExternalId:          types.StringValue(site.ExternalId),
		Id:                  types.StringValue(site.Id),
		IsDefault:           types.BoolValue(site.IsDefault),
		Name:                types.StringValue(site.Name),
		RegistrationToken:   types.StringValue(site.RegistrationToken),
		SiteType:            types.StringValue(site.SiteType),
		State:               types.StringValue(site.State),
		TotalLicenses:       types.Int64Value(int64(site.TotalLicenses)),
		UnlimitedExpiration: types.BoolValue(site.UnlimitedExpiration),
		UnlimitedLicenses:   types.BoolValue(site.UnlimitedLicenses),
		UpdatedAt:           types.StringValue(site.UpdatedAt),
	}
	tfsite.Licenses = &tfSiteLicenseModel{
		Bundles:  []tfSiteLicenseBundleModel{},
		Modules:  []tfSiteLicenseModuleModel{},
		Settings: []tfSiteLicenseSettingModel{},
	}
	for _, bundle := range site.Licenses.Bundles {
		b := tfSiteLicenseBundleModel{
			DisplayName:   types.StringValue(bundle.DisplayName),
			MajorVersion:  types.Int64Value(int64(bundle.MajorVersion)),
			MinorVersion:  types.Int64Value(int64(bundle.MinorVersion)),
			Name:          types.StringValue(bundle.Name),
			Surfaces:      []tfSiteLicenseBundleSurfaceModel{},
			TotalSurfaces: types.Int64Value(int64(bundle.TotalSurfaces)),
		}
		for _, surface := range bundle.Surfaces {
			b.Surfaces = append(b.Surfaces, tfSiteLicenseBundleSurfaceModel{
				Count: types.Int64Value(int64(surface.Count)),
				Name:  types.StringValue(surface.Name),
			})
		}
		tfsite.Licenses.Bundles = append(tfsite.Licenses.Bundles, b)
	}
	for _, module := range site.Licenses.Modules {
		tfsite.Licenses.Modules = append(tfsite.Licenses.Modules, tfSiteLicenseModuleModel{
			DisplayName:  types.StringValue(module.DisplayName),
			MajorVersion: types.Int64Value(int64(module.MajorVersion)),
			Name:         types.StringValue(module.Name),
		})
	}
	for _, setting := range site.Licenses.Settings {
		tfsite.Licenses.Settings = append(tfsite.Licenses.Settings, tfSiteLicenseSettingModel{
			GroupName:               types.StringValue(setting.GroupName),
			Setting:                 types.StringValue(setting.Setting),
			SettingGroupDisplayName: types.StringValue(setting.SettingGroupDisplayName),
		})
	}
	tflog.Trace(ctx, fmt.Sprintf("converted API site to TF site: %+v", tfsite), map[string]interface{}{
		"api_site": site,
	})
	return tfsite
}
