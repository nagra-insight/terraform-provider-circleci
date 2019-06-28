VERSION=v0.0.1

TARGET_BINARY=terraform-provider-circleci_$(VERSION)

TERRAFORM_PLUGIN_DIR=$(HOME)/.terraform.d/plugins/$(OS)_$(ARCH)

.PHONY: $(TARGET_BINARY)

build: $(TARGET_BINARY)

$(TARGET_BINARY):
	go build -mod=vendor -ldflags="-s -w" -a -o $(TARGET_BINARY)

test:
	go test -mod=vendor -v -cover ./...

install_plugin_locally: $(TARGET_BINARY)
	mkdir -p $(TERRAFORM_PLUGIN_DIR)
	cp ./$(TARGET_BINARY) $(TERRAFORM_PLUGIN_DIR)/
