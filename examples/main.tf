terraform {
  required_providers {
    singularity = {
      version = "0.1"
      source  = "joshhogle-at-s1/sentinelone-singularity"
    }
  }
}

provider "singularity" {
  #api_token = ""       # set SINGULARITY_API_TOKEN environment variable instead
  #api_endpoint = ""    # set SINGULARITY_API_ENDPOINT environment variable instead
}


data "singularity_sites" "dest" {
  filter {
    is_default = true
    states     = ["active"]
  }
}

data "singularity_packages" "k8s" {
  filter {
    file_extension = ".gz"
    platform_types = ["linux_k8s"]
    site_ids       = [data.singularity_sites.dest.sites[0].id]
    sort_by        = "version"
    sort_order     = "desc"
  }
}

data "singularity_groups" "dest" {
  filter {
    site_ids   = [data.singularity_sites.dest.sites[0].id]
    is_default = true
  }
}

resource "singularity_package_download" "k8s_agent" {
  package_id     = data.singularity_packages.k8s.packages[0].id
  site_id        = data.singularity_sites.dest.sites[0].id
  local_folder   = pathexpand("~/.singularity/packages/k8s")
  local_filename = data.singularity_packages.k8s.packages[0].file_name
}

/*
resource "github_repository" "singularity_agent" {

}
*/

resource "singularity_k8s_agent_package_loader" "k8s_agent" {
  package_file       = singularity_package_download.k8s_agent.output_file
  docker_host        = "unix:///Users/joshhogle/.rd/docker.sock"
  docker_api_version = null
  docker_cert_path   = null
  docker_tls_verify  = false

  /*
  remote_registry_image {
    
    platforms         = ["arm64", "amd64"]
    images            = ["agent", "helper"]
    hostname          = "ghcr.io"
    credential_helper = "none"
    repo_path         = joshhogle-at-s1 / cwpp-k8s-agent / helper
    image_tag         = singularity_package_download.k8s_agent.version
   
  }
  */
}

/*
resource "kubernetes_secret" "registry_creds" {

}

resource "helm_chart" "k8s_agent" {

}
*/
