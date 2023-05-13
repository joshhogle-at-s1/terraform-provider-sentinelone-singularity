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
  local_folder   = pathexpand("~/.singularity/packages")
  local_filename = data.singularity_packages.k8s.packages[0].file_name
}

/*
resource "singularity_docker_local_image" "k8s_agent" {
  source = singularity_package_download.k8s_agent.local_path
  platform = "x86_64|arm64"
}

resource "github_repository" "singularity_agent" {

}

resource "singularity_docker_registry_image" "k8s_helper" {
  local_docker_image =

  registry_url =
  registry_username =
  registry_password =
  registry_image_name = 
  registry_tag =

}

resource "kubernetes_secret" "registry_creds" {

}

resource "helm_chart" "k8s_agent" {

}
*/
