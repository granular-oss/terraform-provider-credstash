VERSION := "0.0.0-alpha.1"

OS = ${shell uname | tr [:upper:] [:lower:]}
ARCH = ${shell uname -m | tr [:upper:] [:lower:]}

build:
	go build -v -i -o terraform-provider-credstash

test:
	go test ./...

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)
	cp terraform-provider-credstash ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

uninstall: 
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

release:
	GOOS=darwin go build -v -o terraform-provider-credstash_darwin_amd64
	GOOS=linux go build -v -o terraform-provider-credstash_linux_amd64

integration_test: install run_integration_test uninstall

run_integration_test:
	cd test; rm -f .terraform.lock.hcl
	cd test; terraform init
	cd test; TF_LOG=debug terraform plan
	cd test; TF_LOG=debug terraform apply -auto-approve