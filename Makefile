.PHONY: fmt build release mod $(PLATFORMS)
GOFMT_FILES ?= $$(find . -name '?*.go' -maxdepth 2)
NAME := terraform-provider-namecom
PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64
VERSION ?= $(shell git describe --tags --always)
VER ?= $(shell echo $(VERSION)|sed "s/^v\([0-9.]*\).*/\1/")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

BASE := $(NAME)_$(VER)
RELEASE_DIR := ./release

default: build

fmt:
	go fmt ./...

mod:
	go mod download
	go mod tidy
	#go mod vendor

build: fmt
	GOPROXY="off" GOFLAGS="-mod=vendor" go build -ldflags="-X github.com/arthurpro/$(NAME)/version.ProviderVersion=$(VERSION)"

release: $(PLATFORMS)

$(PLATFORMS):
	GOPROXY="off" GOOS=$(os) GOARCH=$(arch) go build -trimpath \
	    -o "$(RELEASE_DIR)/$(BASE)_$(os)_$(arch)/$(NAME)_v$(VER)" \
	    -ldflags="-X github.com/arthurpro/$(NAME)/version.ProviderVersion=$(VERSION)"

	cp README.md $(RELEASE_DIR)/$(BASE)_$(os)_$(arch)
	cd $(RELEASE_DIR)/$(BASE)_$(os)_$(arch)/ && zip -qmr ../$(BASE)_$(os)_$(arch).zip .
	rm -rf $(RELEASE_DIR)/$(BASE)_$(os)_$(arch)

