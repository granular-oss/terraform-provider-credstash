VERSION := "0.0.0-alpha.1"
OS = ${shell uname | tr [:upper:] [:lower:]}
ARCH = ${shell uname -m | tr [:upper:] [:lower:]}
PLUGIN_BINARY_NAME = terraform-provider-credstash
# On MacOS, Terraform expects amd64
ifeq ($(ARCH), x86_64)
	ARCH = amd64
endif

# On Windows, Terraform expects windows_amd64
ifneq (,$(findstring mingw64,$(OS)))
    OS = windows
	PLUGIN_BINARY_NAME := $(PLUGIN_BINARY_NAME).exe
endif

build:
	@echo "Building go package"
	@go build -v -o $(PLUGIN_BINARY_NAME)

test:
	go test -race -v github.com/granular-oss/terraform-provider-credstash/credstash

install: uninstall build
	@echo "Installing TF Plugin locally"
	@mkdir -p integration_test/tf/providers/
	@cp terraform-provider-credstash integration_test/tf/providers/$(PLUGIN_BINARY_NAME)
	@mkdir -p integration_test/invalid_tf/providers/
	@cp terraform-provider-credstash integration_test/invalid_tf/providers/$(PLUGIN_BINARY_NAME)

uninstall:
	@echo "Uninstalling TF Plugin locally"
	@rm -rf integration_test/tf/providers/

release:
	GOOS=darwin go build -v -o terraform-provider-credstash_darwin_amd64
	GOOS=linux go build -v -o terraform-provider-credstash_linux_amd64

run_integration_test: install
	@echo "Running Intergation Test"
	@go clean -testcache ./...
	@rm -rf integration_test/tf/.terraform
	@rm -rf integration_test/tf/.terraform.lock.hcl
	@rm -rf integration_test/invalid_tf/.terraform
	@rm -rf integration_test/invalid_tf/.terraform.lock.hcl
	pip3 install credstash
	go test github.com/granular-oss/terraform-provider-credstash/integration_test
	@rm -rf integration_test/tf/providers/
	@rm -rf integration_test/invalid_tf/providers/
