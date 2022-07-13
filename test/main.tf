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

resource "credstash_secret" "ryan-secret" {
  name = "ryan.test.secret.tf"
  # generate {
  #   length = 10
  # }
  value   = "IamABadSecret!"
  version = 1
}
