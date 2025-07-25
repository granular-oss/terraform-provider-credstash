## v0.7.2 (07 23, 2025)

- Add import documentation.
- Update release.yml to remove `--rm-dist` which has been replaced with the functionally equivalent `--clean` argument.

## v0.7.1 (01 06, 2022)

- Fix credstash resource read to set the length of generate config to the length of the value retrieved from table. Also set import to default update generated values to default config for `use_symbols`

## v0.7.0 (12 20, 2022)

- Add support for credstash resource

## v0.6.0 (03 16, 2022)

IMPROVEMENTS:

- No changes since v0.6.0-alpha.2 release.

## v0.6.0-alpha.2 (03 16, 2022)

IMPROVEMENTS:

- Updated ChangeLog format to follow [HashiCorp Best Practices](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#changelog-specification)

## v0.6.0-alpha.1 (03 08, 2022)

IMPROVEMENTS:

- Add Docs for Terraform Registry
- Update to terraform plugin SDK2
- Update GitHub Actions to include go test
- Add CLA and update Readme

## v0.5.2 (02 25, 2022)

IMPROVEMENTS:

- Forked Project from [terraform-mars](https://github.com/terraform-mars/terraform-provider-credstash)
- This release has been updated to use go 1.16.
- Darwin ARM64 Builds are now available.
