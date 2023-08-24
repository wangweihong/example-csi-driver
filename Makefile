# .DEFAULT_GOAL为makefile自带变量, 用于设置默认目标
# https://www.gnu.org/software/make/manual/html_node/Special-Variables.html
.DEFAULT_GOAL := all

# Build options
# 代码根目录
ROOT_PACKAGE=github.com/wangweihong/example-csi-driver
# 程序版本代码所在目录
VERSION_PACKAGE=github.com/wangweihong/eazycloud/pkg/version

.PHONY: all
all: tidy format lint cover build

include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/golang.mk
include scripts/make-rules/image.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/gen.mk
include scripts/make-rules/dependencies.mk
include scripts/make-rules/swagger.mk
include scripts/make-rules/proto.mk
include scripts/make-rules/certs.mk
include scripts/make-rules/template.mk


# Usage

define USAGE_OPTIONS

Options:
  DEBUG            Whether to generate debug symbols. Default is 0.
  BINS             The binaries to build. Default is all of cmd.
                   This option is available when using: make build/build.multiarch
                   Example: make build BINS="eazycloud-apiserver hubctl"
  IMAGES           Backend images to make. Default is all of cmd starting with iam-.
                   This option is available when using: make image/image.multiarch/push/push.multiarch
                   Example: make image.multiarch IMAGES="eazycloud-apiserver hubctl"
  REGISTRY_PREFIX  Docker registry prefix. Default is "".
                   Example: make push REGISTRY_PREFIX=harbor.registry.wang/exampled VERSION=v1.6.2
  PLATFORMS        The multiple platforms to build. Default is linux_amd64 and linux_arm64.
                   This option is available when using: make build.multiarch/image.multiarch/push.multiarch
                   Example: make image.multiarch IMAGES="eazycloud-apiserver hubctl" PLATFORMS="linux/amd64 linux/arm64".
                   Support PLATFORMS check `go tool dist list` shows.
  VERSION          The version information compiled into binaries.
                   The default is obtained from gsemver or git.
  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

## build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) go.build

## build.multiarch: Build source code for multiple platforms. See option PLATFORMS.
.PHONY: build.multiarch
build.multiarch:
	@$(MAKE) go.build.multiarch

## image: Build docker images for host arch.
.PHONY: image
image:
	@$(MAKE) image.build

## image.multiarch: Build docker images for multiple platforms. See option PLATFORMS.
.PHONY: image.multiarch
image.multiarch:
	@$(MAKE) image.build.multiarch

## clean: Remove all files that are created by building.
.PHONY: clean
clean:
	@echo "===========> Cleaning all build output"
	@-rm -vrf $(OUTPUT_DIR)

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint:
	@$(MAKE) go.lint

## test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test

## cover: Run unit test and get test coverage.
.PHONY: cover
cover:
	@$(MAKE) go.test.cover

## format: Gofmt (reformat) package sources (exclude vendor dir if existed).
.PHONY: format
format: tools.verify.golines tools.verify.goimports
	@echo "===========> Formatting codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt


## dependencies: Install necessary dependencies.
.PHONY: dependencies
dependencies:
	@$(MAKE) dependencies.run

## tools: Install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## check-updates: Check outdated dependencies of the go projects.
.PHONY: check-updates
check-updates:
	@$(MAKE) go.updates

## tidy: Go mod tidy
.PHONY: tidy
tidy:
	@echo "===========> Run go mod tidy"
	@$(GO) mod tidy -compat=1.17

## gen: Generate all necessary files, such as error code files.
.PHONY: gen
gen:
	@echo "===========> Run gen"
	@$(MAKE) gen.run

## ca: Generate CA files for all components.
# 可以通过make ca CERTIFICATES_SUBJECT=192.168.134.139,127.0.0.1来覆写证书主体
# 可以通过make ca CERTIFICATES=apiserver来覆写证书对象
.PHONY: ca
ca:
	@$(MAKE) gen.ca

## proto: Generate Proto file for gRPC service.
.PHONY: proto
proto:
	@$(MAKE) proto.gen

## configs: Generate application default configs files.
.PHONY: configs
configs:
	@$(MAKE) gen.defaultconfigs

## help: Show this help info.
# 这里会提取target上一行的\#\#注释并生成到Makefile help文档中
.PHONY: help
help: Makefile
	@echo -e "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"
