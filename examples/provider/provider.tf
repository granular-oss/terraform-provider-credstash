terraform {
  required_providers {
    credstash = {
      version = "0.7"
      source  = "granular-oss/credstash"
    }
  }
}


provider "credstash" {
  table  = "credential-store"
  region = "us-east-1"
}
