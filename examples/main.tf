terraform {
  required_providers {
    singularity = {
      version = "0.1"
      source  = "joshhogle-at-s1/sentinelone-singularity"
    }
  }
}

provider "singularity" {}

data "singularity_packages" "k8s" {
  filter {
    file_extension = ".gz"
    platform_types = ["linux_k8s"]
    sort_by        = "version"
    sort_order     = "desc"
  }
}

/*
data "singularity_sites" "dest" {

}

data "singularity_groups" "dest" {
  
}
*/

/*
resource "singularity_package_download" "k8s_agent" {
  package_id = data.singularity_packages.k8s.packages[0].id
}

resource "singularity_k8s_image" "k8s_agent" {

}

resource "singularity_k8s_image" "k8s_helper" {

}

resource "kubernetes_secret" "registry_creds" {

}

resource "helm_chart" "k8s_agent" {

}
*/
