terraform {
  required_providers {
    credstash = {
      source = "granular-oss/credstash"
    }
  }

  required_version = "~> 1.0"

  backend "local" {
    path = "./terraform.tfstate"
  }
}

provider "credstash" {
  table  = "terraform-provider-credstash-test-table"
  region = "us-east-1"
}

resource "credstash_secret" "invalid_test_1" {
  name  = "invalid_test_1"
  value = "foobar"
  generate {
    length = 10
  }
}

resource "credstash_secret" "invalid_test_2" {
  name = "invalid_test_2"
}
