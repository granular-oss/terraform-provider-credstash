VERSION := "0.0.0-alpha.1"
OS = ${shell uname | tr [:upper:] [:lower:]}
ARCH = ${shell uname -m | tr [:upper:] [:lower:]}
ifeq ($(ARCH), x86_64)
	ARCH = amd64
endif

build:
	@echo "Building go package"
	@go build -v -o terraform-provider-credstash

test:
	go test ./...

install: build
	@echo "Installing TF Plugin locally"
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)
	@rm -f ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)/terraform-provider-credstash
	@cp terraform-provider-credstash ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

uninstall: 
	@echo "Uninstalling TF Plugin locally"
	@rm -rf ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

release:
	GOOS=darwin go build -v -o terraform-provider-credstash_darwin_amd64
	GOOS=linux go build -v -o terraform-provider-credstash_linux_amd64

integration_test: | install run_integration_test # uninstall

run_integration_test: install
	@echo "Running Intergation Test"
	@go clean -testcache ./...
	@rm -rf integration_test/tf/.terraform
	pip3 install credstash
	go test github.com/granular-oss/terraform-provider-credstash/integration_test -v

run_integration_test_bash: install
	@echo "Cleaning Up TF State and Lock Files"
	@cd integration_test/tf; rm -f .terraform.lock.hcl
	@cd integration_test/tf; rm -f *.tfstate*
	@echo "Cleaning Up CredStash"
	@credstash delete terraform-provider-credstash-integration-test-1
	@credstash delete terraform-provider-credstash-integration-test-2
	@credstash delete terraform-provider-credstash-integration-test-3
	@credstash delete terraform-provider-credstash-integration-test-4
	@credstash delete terraform-provider-credstash-integration-test-5
	@credstash delete terraform-provider-credstash-integration-test-6
	@credstash delete terraform-provider-credstash-integration-test-7
	@echo "Creating Credstash Secrets for data and import test"
	@credstash put -a terraform-provider-credstash-integration-test-4 "test-4-1"
	@credstash put -a terraform-provider-credstash-integration-test-4 "test-4-2"
	@credstash put -a terraform-provider-credstash-integration-test-5 "test-5-1"
	@credstash put -a terraform-provider-credstash-integration-test-5 "test-5-2"
	@credstash put -a terraform-provider-credstash-integration-test-7 "test-7-1"
	@credstash put -a terraform-provider-credstash-integration-test-7 "test-7-2"
	@credstash put -a terraform-provider-credstash-integration-test-7 "test-7-3"
	@cd integration_test/tf; terraform init
	@cd integration_test/tf; TF_LOG=$(TF_LOG) TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_3=IAmABadPassword2 TF_VAR_secret_5=test-5-2 terraform import credstash_secret.terraform-provider-credstash-integration-test-5 terraform-provider-credstash-integration-test-5
	@cd integration_test/tf; TF_LOG=$(TF_LOG) TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_3=IAmABadPassword2 TF_VAR_secret_5=test-5-2 terraform plan
	@cd integration_test/tf; TF_LOG=$(TF_LOG) TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_3=IAmABadPassword2 TF_VAR_secret_5=test-5-2 terraform apply -auto-approve

## Import Tests
# Test terraform-provider-credstash-integration-test-5 Imported Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-5")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-5 --version 2 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-5 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-5 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-5 Imported Version is 0, so we can autoincrement in the future
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-5")).values.version') E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-5 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-5 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-5 Imported Name is correct
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-5")).values.name') E_VALUE=terraform-provider-credstash-integration-test-5 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mName Correct For terraform-provider-credstash-integration-test-5 \033[0m"; else echo -e "\033[0;31mName Incorrect For terraform-provider-credstash-integration-test-5 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Create with Value And Version
# Test terraform-provider-credstash-integration-test-1 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-1")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-1 --version 1 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-1 Resource Version is 1
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-1")).values.version') E_VALUE=1 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-1 CredStash Version is 1
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-1 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=1 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Create with Generated Value
# Test terraform-provider-credstash-integration-test-2 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-2")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-2 --version 1 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-2 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-2 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-2 Resource Version is 0
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-2")).values.version') E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-2 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-2 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-2 CredStash Version is 1
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-2 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=1 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-2 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-2 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Create with Generated Value and Version
# Test terraform-provider-credstash-integration-test-6 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-6")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-6 --version 10 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-6 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-6 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-6 Resource Version is 10
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-6")).values.version') E_VALUE=10 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-6 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-6 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-6 CredStash Version is 10
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-6 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=10 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-6 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-6 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Create with Value And No Version
# Test terraform-provider-credstash-integration-test-3 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-3")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-3 --version 1 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-3 Resource Version is 0
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-3")).values.version') E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-3 CredStash Version is 1
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-3 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=1 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Data Block Name Only (test-4-2)
# Test terraform-provider-credstash-integration-test-4 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-4")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-4 --version 2 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-4 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-4 Resource Version is 0
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-4")).values.version') E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-4 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-4 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Data Block Name and Version (test-7-2)
# Test terraform-provider-credstash-integration-test-7 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-7")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-7 --version 2 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-7 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-7 Resource Version is 2
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-7")).values.version') E_VALUE=2 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-7 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-7 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

# Test that Plan Shows No Changes
	@T_VALUE=$$(cd integration_test/tf; TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_3=IAmABadPassword2 TF_VAR_secret_5=test-5-2 terraform plan -no-color -detailed-exitcode > /dev/null 2>&1; echo $$?) E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mTerraform Plan Shows No Changes After Apply\033[0m"; else echo -e "\033[0;31mTerraform Plan Shows Changes After Apply\033[0m"; exit 1; fi'

#### Update Passwords and test changes
	@echo "Update TF with new passwords"
	@cd integration_test/tf; TF_LOG=$(TF_LOG) TF_VAR_secret_1_version=3 TF_VAR_secret_1=IAmABadPassword3 TF_VAR_secret_3=IAmABadPassword2Again TF_VAR_secret_5=test-5-2 terraform apply -auto-approve

### Create with Value And Version
# Test terraform-provider-credstash-integration-test-1 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-1")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-1 --version 3 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-1 Resource Version is 1
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-1")).values.version') E_VALUE=3 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-1 CredStash Version is 1
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-1 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=3 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-1 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-1 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'

### Create with Value And No Version
# Test terraform-provider-credstash-integration-test-3 Resource Value Matches credstash latest version 
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-3")).values.value') E_VALUE=$$(credstash get terraform-provider-credstash-integration-test-3 --version 2 -n) bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mValue Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mValue Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-3 Resource Version is 0
	@T_VALUE=$$(cd integration_test/tf; terraform show -json | jq --raw-output '.values.root_module.resources[] | select(.address | contains("terraform-provider-credstash-integration-test-3")).values.version') E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mVersion Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mVersion Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'
# Test terraform-provider-credstash-integration-test-3 CredStash Version is 1
	@T_VALUE=$$(credstash list | grep terraform-provider-credstash-integration-test-3 | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\1/') E_VALUE=2 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mCredStash Version Correct For terraform-provider-credstash-integration-test-3 \033[0m"; else echo -e "\033[0;31mCredStash Version Incorrect For terraform-provider-credstash-integration-test-3 Got: $${T_VALUE} Expected: $${E_VALUE}\033[0m"; exit 1; fi'




# Test that Plan Shows No Changes
	@T_VALUE=$$(cd integration_test/tf; TF_VAR_secret_1_version=3 TF_VAR_secret_1=IAmABadPassword3 TF_VAR_secret_3=IAmABadPassword2Again TF_VAR_secret_5=test-5-2 terraform plan -no-color -detailed-exitcode > /dev/null 2>&1; echo $$?) E_VALUE=0 bash -c 'if [ "$$T_VALUE" = "$$E_VALUE" ]; then echo -e "\033[0;32mTerraform Plan Shows No Changes After Apply\033[0m"; else echo -e "\033[0;31mTerraform Plan Shows Changes After Apply\033[0m"; exit 1; fi'
	@cd integration_test/tf; rm -f .terraform.lock.hcl
	@cd integration_test/tf; rm -f *.tfstate*