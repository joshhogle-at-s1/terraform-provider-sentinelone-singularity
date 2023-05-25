package resources

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/provider/data"
)

// Docker image constants
const (
	DOCKER_IMAGE_BASE_REPOSITORY = "cwpp_agent"
	DOCKER_IMAGE_S1_HELPER       = "s1helper"
	DOCKER_IMAGE_S1_AGENT        = "s1agent"
)

// ensure implementation satisfied expected interfaces
var (
	_ resource.Resource              = &K8sAgentPackageLoader{}
	_ resource.ResourceWithConfigure = &K8sAgentPackageLoader{}
)

// tfK8sAgentPackageLoader defines the Terrform model for loading a package image into Docker.
type tfK8sAgentPackageLoader struct {
	DockerAPIVersion    types.String `tfsdk:"docker_api_version"`
	DockerCertPath      types.String `tfsdk:"docker_cert_path"`
	DockerHost          types.String `tfsdk:"docker_host"`
	DockerTLSVerify     types.Bool   `tfsdk:"docker_tls_verify"`
	PackageFile         types.String `tfsdk:"package_file"`
	Images              types.List   `tfsdk:"images"`
	RemoteRegistryImage types.List   `tfsdk:"remote_registry_image"`
}

// tfK8sAgentPackageLoaderRemoteRegistryImage defines the Terraform model for a pushing the k8s agent image to a
// remote Docker registry.
type tfK8sAgentPackageLoaderRemoteRegistryImage struct {
	/*
		CredentialHelper types.String   `tfsdk:"credential_helper"`
		Hostname         types.String   `tfsdk:"hostname"`
		Images           []types.String `tfsdk:"images"`
		ImageTag         types.String   `tfsdk:"image_tag"`
		Password         types.String   `tfsdk:"password"`
		Platforms        []types.String `tfsdk:"platforms"`
		RepoPath         types.String   `tfsdk:"repo_path"`
		Username         types.String   `tfsdk:"username"`
	*/
}

// tfK8sAgentPackageLoaderImage contains details on a Docker image.
type tfK8sAgentPackageLoaderImage struct {
	Id           types.String `tfsdk:"id"`
	RepoTags     types.List   `tfsdk:"repo_tags"`
	Purpose      types.String `tfsdk:"purpose"`
	Architecture types.String `tfsdk:"architecture"`
	Variant      types.String `tfsdk:"variant"`
	Size         types.Int64  `tfsdk:"size"`
}

// NewK8sAgentPackageLoader creates a new K8sAgentPackageLoader object.
func NewK8sAgentPackageLoader() resource.Resource {
	return &K8sAgentPackageLoader{}
}

// K8sAgentPackageLoader is a resource used to for importing k8s agent images into a local Docker image repository.
type K8sAgentPackageLoader struct {
	data *data.SingularityProvider
}

// Metadata returns metadata about the data source.
func (r *K8sAgentPackageLoader) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_agent_package_loader"
}

// Schema defines the parameters for the data sources's configuration.
func (r *K8sAgentPackageLoader) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource is used for loading a downloaded Singularity Agent for Kubernetes package into " +
			"a local Docker image repository so it can then be pushed to a destination registry for deployment.",
		MarkdownDescription: `This resource is used for loading a downloaded Singularity agent for Kubernetes package
			into a local Docker image repository so it can then be pushed to a destination registry for deployment.

			TODO: add more of a description on how to use this data source...
			`,
		Attributes: map[string]schema.Attribute{
			"docker_api_version": schema.StringAttribute{
				Description: "The version of the Docker API to use when communicating with the Docker host. If empty, " +
					"use the latest version available. [Default: none].",
				MarkdownDescription: "The version of the Docker API to use when communicating with the Docker host. If " +
					"empty, use the latest version available. [Default: none].",
				Optional: true,
				Default:  nil,
				Computed: true,
			},
			"docker_cert_path": schema.StringAttribute{
				Description: "If a TLS connection to the Docker host is enabled, the full path in which to find the " +
					"CA certificate and client certificate and key used to connect to the host. [Default: none].",
				MarkdownDescription: "If a TLS connection to the Docker host is enabled, the full path in which to find " +
					"the CA certificate and client certificate and key used to connect to the host. [Default: none].",
				Optional: true,
				Default:  nil,
				Computed: true,
			},
			"docker_host": schema.StringAttribute{
				Description: "The URL to use for the Docker host where the agent and helper images will be loaded. " +
					"[Defualt: unix:///var/run/docker.sock]",
				MarkdownDescription: "The URL to use for the Docker host where the agent and helper images will be " +
					"loaded. [Defualt: `unix:///var/run/docker.sock`]",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("unix:///var/run/docker.sock"),
			},
			"docker_tls_verify": schema.BoolAttribute{
				Description: "If a TLS connection to the Docker host is enabled, whether or not to perform TLS " +
					"verification during the handshake. [Default: false].",
				MarkdownDescription: "If a TLS connection to the Docker host is enabled, whether or not to perform TLS " +
					"verification during the handshake. [Default: `false`].",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"images": schema.ListNestedAttribute{
				Description:         "",
				MarkdownDescription: "",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "",
							MarkdownDescription: "",
							Computed:            true,
						},
						"repo_tags": schema.ListAttribute{
							Description:         "",
							MarkdownDescription: "",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"purpose": schema.StringAttribute{
							Description:         "",
							MarkdownDescription: "",
							Computed:            true,
						},
						"architecture": schema.StringAttribute{
							Description:         "",
							MarkdownDescription: "",
							Computed:            true,
						},
						"variant": schema.StringAttribute{
							Description:         "",
							MarkdownDescription: "",
							Computed:            true,
						},
						"size": schema.Int64Attribute{
							Description:         "",
							MarkdownDescription: "",
							Computed:            true,
						},
					},
				},
			},
			"package_file": schema.StringAttribute{
				Description:         "The path to the downloaded Singularity Agent for Kubernetes package file.",
				MarkdownDescription: "The path to the downloaded Singularity Agent for Kubernetes package file.",
				Required:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"remote_registry_image": schema.ListNestedBlock{
				Description:         "Defines a remote repository to push the image to once it has been loaded.",
				MarkdownDescription: "Defines a remote repository to push the image to once it has been loaded.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						/*
							"credential_helper": schema.StringAttribute{
								Description: "If the remote registry requires a Docker credential helper for authentication, set " +
									"this to the appropriate value (valid values: none, aws-ecr, google-gcr, osxkeychain, pass, " +
									"secretservice, wincred) [Default: none].",
								MarkdownDescription: "If the remote registry requires a Docker credential helper for authentication, " +
									"set this to the appropriate value (valid values: `none`, `aws-ecr`, `google-gcr`, `osxkeychain`, " +
									"`pass`, `secretservice`, `wincred`) [Default: `none`].",
								Optional: true,
								Computed: true,
								Default:  stringdefault.StaticString("none"),
								Validators: []validator.String{
									validators.EnumStringValueOneOf(false, "none", "aws-ecr", "google-gcr", "osxkeychain", "pass",
										"secretservice", "wincred"),
								},
							},
							"hostname": schema.ListAttribute{
								Description:         "The hostname of the remote registry (eg: ghcr.io).",
								MarkdownDescription: "The hostname of the remote registry (eg: `ghcr.io`).",
								Required:            true,
							},
							"image_tag": schema.StringAttribute{
								Description:         "The actual tag to use for the image (eg: latest).",
								MarkdownDescription: "The actual tag to use for the image (eg: `latest`).",
								Required:            true,
							},
							"images": schema.ListAttribute{
								Description: "The image(s) to push to the remote repository (valid values: agent, helper) " +
									"[Default: [agent, helper] ].",
								MarkdownDescription: "The image(s) to push to the remote repository (valid values: agent, helper) " +
									"[Default: `[agent, helper]`].",
								Optional: true,
								Computed: true,
								Default: listdefault.StaticValue(types.ListValueMust(
									types.StringType, []attr.Value{
										types.StringValue("agent"),
										types.StringValue("helper"),
									},
								)),
								ElementType: types.StringType,
								Validators: []validator.List{
									validators.EnumStringListValuesAre(false, "agent", "helper"),
								},
							},
							"password": schema.StringAttribute{
								Description: "If not using a credential helper, the password to use for authentication with the " +
									"remote registry.",
								MarkdownDescription: "If not using a credential helper, the password to use for authentication with " +
									"the remote registry.",
								Optional:  true,
								Sensitive: true,
							},
							"platforms": schema.ListAttribute{
								Description: "CPU platform(s) of image to push to remote repository (valid values: " +
									"amd64, arm64) [Default: [amd64, arm64] ].",
								MarkdownDescription: "CPU platform(s) of image to push to remote repository(valid values: " +
									"amd64, arm64) [Default: `[amd64, arm64]` ].",
								Optional: true,
								Computed: true,
								Default: listdefault.StaticValue(types.ListValueMust(
									types.StringType, []attr.Value{
										types.StringValue("amd64"),
										types.StringValue("arm64"),
									},
								)),
								ElementType: types.StringType,
								Validators: []validator.List{
									validators.EnumStringListValuesAre(false, "amd64", "arm64"),
								},
							},
							"repo_path": schema.StringAttribute{
								Description: "The repository path within the remote registry in which to store the container " +
									"(eg: joshhogle-at-s1/cwpp-k8s-agent).",
								MarkdownDescription: "The repository path within the remote registry in which to store the container " +
									"(eg: `joshhogle-at-s1/cwpp-k8s-agent`).",
								Required: true,
							},
							"username": schema.StringAttribute{
								Description: "If not using a credential helper, the username to use for authentication with the " +
									"remote registry.",
								MarkdownDescription: "If not using a credential helper, the username to use for authentication with " +
									"the remote registry.",
								Optional: true,
							},
						*/
					},
				},
			},
		},
	}
}

// Configure initializes the configuration for the data source.
func (r *K8sAgentPackageLoader) Configure(ctx context.Context, req resource.ConfigureRequest,
	resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*data.SingularityProvider)
	if !ok {
		expectedType := reflect.TypeOf(&data.SingularityProvider{})
		msg := fmt.Sprintf("The provider data sent in the request does not match the type expected. This is always an "+
			"error with the provider and should be reported to the provider developers.\n\nExpected Type: %s\nData Type "+
			"Received: %T", expectedType, req.ProviderData)
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_CONFIGURE,
			"expected_type":       fmt.Sprintf("%T", expectedType),
			"received_type":       fmt.Sprintf("%T", req.ProviderData),
		})
		resp.Diagnostics.AddError("Unexpected Configuration Error", msg)
		return
	}
	r.data = providerData
}

// ModifyPlan is called to modify the Terraform plan.
func (r *K8sAgentPackageLoader) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse) {
}

// Create is used to create the Terraform resource.
func (r *K8sAgentPackageLoader) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan tfK8sAgentPackageLoader
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// make sure the package file exists
	absPath, diags := plugin.ToAbsolutePath(ctx, plan.PackageFile.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	exists, diags := plugin.PathExists(ctx, absPath)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !exists {
		msg := fmt.Sprintf("The package file specified does not exist.\n\nFile: %s", absPath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"package_file":        absPath,
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_CREATE,
		})
		resp.Diagnostics.AddError("K8s Agent Package Loader Creation Error", msg)
		return
	}

	// load the image to the Docker host
	dockerClient, diags := r.newDockerClient(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, diags = r.dockerLoad(ctx, dockerClient, absPath)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save the the plan to the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the current state of the Terraform resource.
func (r *K8sAgentPackageLoader) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update modifies the Terraform resource in place without destroying it.
func (r *K8sAgentPackageLoader) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete removes the Terraform resource.
func (r *K8sAgentPackageLoader) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// newDockerClient constructs the Docker API client from the given configuration.
func (r *K8sAgentPackageLoader) newDockerClient(ctx context.Context, cfg tfK8sAgentPackageLoader) (
	*client.Client, diag.Diagnostics) {

	var diags diag.Diagnostics

	// construct the client
	os.Setenv("DOCKER_HOST", cfg.DockerHost.ValueString())
	if cfg.DockerTLSVerify.IsNull() || cfg.DockerTLSVerify.IsUnknown() || !cfg.DockerTLSVerify.ValueBool() {
		os.Unsetenv("DOCKER_TLS_VERIFY")
	} else {
		os.Setenv("DOCKER_TLS_VERIFY", "true")
	}
	if cfg.DockerAPIVersion.IsNull() || cfg.DockerAPIVersion.IsUnknown() || cfg.DockerAPIVersion.ValueString() == "" {
		os.Unsetenv("DOCKER_API_VERSION")
	} else {
		os.Setenv("DOCKER_API_VERSION", cfg.DockerAPIVersion.ValueString())
	}
	if cfg.DockerCertPath.IsNull() || cfg.DockerCertPath.IsUnknown() || cfg.DockerCertPath.ValueString() == "" {
		os.Unsetenv("DOCKER_CERT_PATH")
	} else {
		os.Setenv("DOCKER_CERT_PATH", cfg.DockerCertPath.ValueString())
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to connect to the Docker host.\n\n"+
			"Error: %s\nDocker Host: %s", err.Error(), cfg.DockerHost.ValueString())
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_INIT,
			"error":               err.Error(),
			"docker_host":         cfg.DockerHost.ValueString(),
		})
		diags.AddError("Docker Connection Error", msg)
		return nil, diags
	}

	// ping to make sure connection is established
	ping, err := cli.Ping(ctx)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to connect to the Docker host.\n\n"+
			"Error: %s\nDocker Host: %s", err.Error(), cfg.DockerHost.ValueString())
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_INIT,
			"error":               err.Error(),
			"docker_host":         cfg.DockerHost.ValueString(),
		})
		diags.AddError("Docker Connection Error", msg)
		return nil, diags
	}
	tflog.Debug(ctx, fmt.Sprintf("Docker ping response: %+v", ping))
	return cli, diags
}

// dockerLoad uses the Docker client to load the given image archive file into the local Docker image cache.
func (r *K8sAgentPackageLoader) dockerLoad(ctx context.Context, dockerClient *client.Client, imagePath string) (
	[]tfK8sAgentPackageLoaderImage, diag.Diagnostics) {

	var diags diag.Diagnostics

	// open the archive and load the image(s)
	file, err := os.Open(imagePath)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to open the package file for reading.\n\n"+
			"Error: %s\nFile: %s", err.Error(), imagePath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"package_file":        imagePath,
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
		})
		diags.AddError("Docker Image Load Error", msg)
		return nil, diags
	}
	defer file.Close()
	result, err := dockerClient.ImageLoad(ctx, file, true)
	if err != nil {
		msg := fmt.Sprintf("An unexpected error occurred while attempting to load the package into the Docker image "+
			"cache.\n\nError: %s\nFile: %s", err.Error(), imagePath)
		tflog.Error(ctx, msg, map[string]interface{}{
			"package_file":        imagePath,
			"error":               err.Error(),
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
		})
		diags.AddError("Docker Image Load Error", msg)
		return nil, diags
	}
	defer result.Body.Close()

	// response body should always be JSON
	if !result.JSON {
		body, _ := io.ReadAll(result.Body)
		msg := fmt.Sprintf("The Docker API unexpectedly returned a non-JSON response body.\n\nBody: %s", body)
		tflog.Error(ctx, msg, map[string]interface{}{
			"body":                body,
			"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
		})
		diags.AddError("Docker Image Load Error", msg)
		return nil, diags
	}

	// parse the output to get the image(s) loaded
	var images []tfK8sAgentPackageLoaderImage
	var responseLine struct {
		Stream  string `json:"stream"`
		Message string `json:"message"`
	}
	scanner := bufio.NewScanner(result.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// unmarshal the line
		if err := json.Unmarshal(scanner.Bytes(), &responseLine); err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while attempting to parse Docker API output.\n\nError: %s",
				err.Error())
			tflog.Error(ctx, msg, map[string]interface{}{
				"output_line":         scanner.Text(),
				"error":               err.Error(),
				"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
			})
			diags.AddError("Docker Image Load Error", msg)
			return nil, diags
		}

		// if there's a "message", that's typcially an error
		if responseLine.Message != "" {
			msg := fmt.Sprintf("An unexpected error message was returned in the Docker API output.\n\nError: %s",
				responseLine.Message)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               responseLine.Message,
				"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
			})
			diags.AddError("Docker Image Load Error", msg)
			return nil, diags
		}

		// verify format of the "stream" matches an expected image name
		imageName := strings.TrimPrefix(strings.TrimSpace(responseLine.Stream), "Loaded image: ")
		imageFormat := regexp.MustCompile(fmt.Sprintf(`^%s\/(%s|%s):([a-zA-Z0-9\-_].*)$`, DOCKER_IMAGE_BASE_REPOSITORY,
			DOCKER_IMAGE_S1_AGENT, DOCKER_IMAGE_S1_HELPER))
		if !imageFormat.MatchString(imageName) {
			tflog.Warn(ctx, fmt.Sprintf("response line from Docker API was not the expected 'Loaded image' "+
				"message or a maching container image name: ignoring\n\nLine: %s", responseLine.Stream))
			continue
		}

		// inspect the image and save its details
		details, _, err := dockerClient.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			msg := fmt.Sprintf("An unexpected error occurred while attempting to retrieve information on the "+
				"container image.\n\nError: %s\nImage: %s", err.Error(), imageName)
			tflog.Error(ctx, msg, map[string]interface{}{
				"error":               err.Error(),
				"image":               imageName,
				"internal_error_code": plugin.ERR_RESOURCE_K8S_AGENT_PACKAGE_LOADER_DOCKER_LOAD,
			})
			diags.AddError("Docker Image Load Error", msg)
			return nil, diags
		}
		image := tfK8sAgentPackageLoaderImage{
			Id:           types.StringValue(details.ID),
			RepoTags:     types.ListNull(types.StringType),
			Architecture: types.StringValue(details.Architecture),
			Variant:      types.StringValue(details.Variant),
			Size:         types.Int64Value(details.Size),
		}
		image.RepoTags, diags = types.ListValueFrom(ctx, types.StringType, details.RepoTags)
		if diags.HasError() {
			return nil, diags
		}
		matches := imageFormat.FindStringSubmatch(imageName)
		if matches[1] == DOCKER_IMAGE_S1_HELPER {
			image.Purpose = types.StringValue("helper")
		} else if matches[1] == DOCKER_IMAGE_S1_AGENT {
			image.Purpose = types.StringValue("agent")
		} else {
			image.Purpose = types.StringNull()
		}
		images = append(images, image)
		tflog.Debug(ctx, fmt.Sprintf("loaded Docker image: %s", imageName), map[string]interface{}{
			"image":        imageName,
			"id":           image.Id.ValueString(),
			"architecture": image.Architecture.ValueString(),
			"variant":      image.Variant.ValueString(),
			"size":         image.Size.ValueInt64(),
			"purpose":      image.Purpose.ValueString(),
		})
	}
	return images, diags
}
