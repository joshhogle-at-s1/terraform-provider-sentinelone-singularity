terraform {
  required_providers {
    singularity = {
      version = "0.1"
      source  = "joshhogle-at-s1/sentinelone-singularity"
    }
  }
}

provider "singularity" {}

data "singularity_package" "test" {
  id = 1076647574506860121
}

data "singularity_packages" "k8s" {
  filter {
    file_extension = ".deb"
  }
}
