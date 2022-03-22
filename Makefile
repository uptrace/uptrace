VERSION=$(shell git describe --always --tags --match "v[0-9]*" HEAD)
BUILD_INFO_IMPORT_PATH=github.com/uptrace/uptrace/pkg/internal/version
BUILD_INFO=-ldflags "-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)"
GO_BUILD_TAGS=""

.PHONY: uptrace-vue
uptrace-vue:
	cd vue && pnpm build

.PHONY: gomoddownload
gomoddownload:
	go mod download

.PHONY: uptrace
uptrace:
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -o ./bin/uptrace_$(GOOS)_$(GOARCH)$(EXTENSION) \
		$(BUILD_INFO) -tags $(GO_BUILD_TAGS) ./cmd/uptrace

.PHONY: uptrace-all-sys
uptrace-all-sys: uptrace-darwin_amd64 uptrace-darwin_arm64 uptrace-linux_amd64 uptrace-linux_arm64 uptrace-windows_amd64

.PHONY: uptrace-darwin_amd64
uptrace-darwin_amd64:
	GOOS=darwin GOARCH=amd64 $(MAKE) uptrace

.PHONY: uptrace-darwin_arm64
uptrace-darwin_arm64:
	GOOS=darwin GOARCH=arm64 $(MAKE) uptrace

.PHONY: uptrace-linux_amd64
uptrace-linux_amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) uptrace

.PHONY: uptrace-linux_arm64
uptrace-linux_arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) uptrace

.PHONY: uptrace-windows_amd64
uptrace-windows_amd64:
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) uptrace

.PHONY: docker-uptrace
docker-uptrace:
	GOOS=linux GOARCH=amd64 $(MAKE) uptrace
	cp ./bin/uptrace_linux_amd64 ./cmd/uptrace/uptrace
	docker build -t uptrace ./cmd/uptrace/
	rm ./cmd/uptrace/uptrace

.PHONY: deb-rpm-package
%-package: ARCH ?= amd64
%-package:
	$(MAKE) uptrace-linux_$(ARCH)
	docker build -t uptrace-fpm internal/packaging/fpm
	docker run --rm -v $(CURDIR):/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) uptrace-fpm

TOOLS_MOD_DIR := ./pkg/internal/tools
.PHONY: install-tools
install-tools:
	cd $(TOOLS_MOD_DIR) && go install github.com/tcnksm/ghr

.PHONY: clean-repo
clean-repo:
	git filter-repo --path vue/dist --invert-paths
