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

variable "secret_3" {
  type = string
}
variable "secret_5" {
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
  value = var.secret_3
}

data "credstash_secret" "terraform-provider-credstash-integration-test-4" {
  name = "terraform-provider-credstash-integration-test-4"
}

resource "credstash_secret" "terraform-provider-credstash-integration-test-5" {
  name  = "terraform-provider-credstash-integration-test-5"
  value = var.secret_5
}
resource "credstash_secret" "terraform-provider-credstash-integration-test-6" {
  name    = "terraform-provider-credstash-integration-test-6"
  version = 10
  generate {
    length = 10
  }
}

data "credstash_secret" "terraform-provider-credstash-integration-test-7" {
  name    = "terraform-provider-credstash-integration-test-7"
  version = 2
}
