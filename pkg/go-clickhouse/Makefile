ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)

test:
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "go test in $${dir}"; \
	  (cd "$${dir}" && \
	    go test && \
	    env GOOS=linux GOARCH=386 go test && \
	    go vet); \
	done

.PHONY: go_mod_tidy
go_mod_tidy:
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "go mod tidy in $${dir}"; \
	  (cd "$${dir}" && go mod tidy -go=1.18); \
	done

.PHONY: deps
deps:
	set -e; for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "go get -u ./... && go mod tidy in $${dir}"; \
	  (cd "$${dir}" && \
	    go get -u ./... && \
	    go mod tidy -go=1.18); \
	done

fmt:
	gofmt -w -s ./
	goimports -w  -local github.com/uptrace/go-clickhouse ./

codegen:
	go run ./ch/internal/codegen/ -dir=ch/chschema
