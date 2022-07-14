VERSION := "0.0.0-alpha.1"

OS = ${shell uname | tr [:upper:] [:lower:]}
ARCH = ${shell uname -m | tr [:upper:] [:lower:]}

build:
	go build -v -o terraform-provider-credstash
	sleep 1

test:
	go test ./...

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)
	rm -f ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)/terraform-provider-credstash
	cp terraform-provider-credstash ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

uninstall: 
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/granular-oss/credstash/$(VERSION)/$(OS)_$(ARCH)

release:
	GOOS=darwin go build -v -o terraform-provider-credstash_darwin_amd64
	GOOS=linux go build -v -o terraform-provider-credstash_linux_amd64

integration_test: | install run_integration_test # uninstall

run_integration_test:
	cd test; rm -f .terraform.lock.hcl
	cd test; rm -f *.tfstate*
	credstash delete terraform-provider-credstash-integration-test-1
	credstash delete terraform-provider-credstash-integration-test-2
	credstash delete terraform-provider-credstash-integration-test-3
	cd test; terraform init
	cd test; TF_LOG=debug TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_2=IAmABadPassword2 TF_VAR_secret_5=IamNew terraform plan
	cd test; TF_LOG=debug TF_VAR_secret_1_version=1 TF_VAR_secret_1=IAmABadPassword1 TF_VAR_secret_2=IAmABadPassword2 TF_VAR_secret_5=IamNew terraform apply -auto-approve
	cd test; terraform show -json |jq