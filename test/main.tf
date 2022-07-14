terraform {
  required_providers {
    credstash = {
      version = "0.0.0-alpha.1"
      source  = "granular-oss/credstash"
    }
  }

  backend "local" {
    path = "./terraform.tfstate"
  }
}

provider "credstash" {
  region = "us-east-1"
}

variable "secret_1" {
  type = string
}

variable "secret_1_version" {
  type = number
}

variable "secret_2" {
  type = string
}
resource "credstash_secret" "terraform-provider-credstash-integration-test-1" {
  name = "terraform-provider-credstash-integration-test-1"
  # generate {
  #   length = 10
  # }
  value   = var.secret_1
  version = var.secret_1_version
}

resource "credstash_secret" "terraform-provider-credstash-integration-test-2" {
  name = "terraform-provider-credstash-integration-test-2"
  generate {
    length = 10
  }
}

resource "credstash_secret" "terraform-provider-credstash-integration-test-3" {
  name  = "terraform-provider-credstash-integration-test-3"
  value = var.secret_2
}

data "credstash_secret" "terraform-provider-credstash-integration-test-4" {
  name = "terraform-provider-credstash-integration-test-4"
}
