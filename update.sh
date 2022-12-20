#!/usr/bin/env bash
set -Eeuo pipefail
go build  -o terraform-provider-credstash
mv terraform-provider-credstash ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/0.9/darwin_arm64

pushd test
rm -f .terraform.lock.hcl
terraform init
TF_LOG=debug terraform plan
TF_LOG=debug terraform apply -auto-approve
popd