INTERNAL_BIN_DIR=sub_bin
GOVERSION=$(shell go version)
GOOS=$(word 1,$(subst /, ,$(lastword $(GOVERSION))))
GOARCH=$(word 2,$(subst /, ,$(lastword $(GOVERSION))))
GO15VENDOREXPERIMENT=1
HAS_GLIDE:=$(shell which glide)

test: deps
	@PATH=$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH):$(PATH) go test -v $(shell glide nv)

deps: glide
	@PATH=$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH):$(PATH) glide install
	go get github.com/golang/lint/golint
	go get github.com/mattn/goveralls
	go get github.com/axw/gocov/gocov

$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide:
ifndef HAS_GLIDE
	@mkdir -p $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)
	@curl -L https://github.com/Masterminds/glide/releases/download/v0.11.0/glide-v0.11.0-$(GOOS)-$(GOARCH).zip -o glide.zip
	@unzip glide.zip
	@mv ./$(GOOS)-$(GOARCH)/glide $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide
	@rm -rf ./$(GOOS)-$(GOARCH)
	@rm ./glide.zip
endif

glide: $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide

lint: deps
	@for dir in $$(glide novendor); do \
	golint $$dir; \
	done;

cover: deps
	goveralls

.PHONY: test deps lint cover
