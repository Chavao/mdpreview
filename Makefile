.PHONY: clean build release

BINARY_NAME=mdpreview
BUILD_DIR=bin

build: clean
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME).new ./cmd/$(BINARY_NAME)
	@chmod 755 $(BUILD_DIR)/$(BINARY_NAME).new

clean:
	@rm -rf $(BUILD_DIR)

release: build
	@mv $(BUILD_DIR)/$(BINARY_NAME).new "$(HOME)/bin/$(BINARY_NAME)"
