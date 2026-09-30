GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GODOC=$(GOCMD)doc
GOLANGCI=golangci-lint



# runs
#===============================================================
all: version run

version: ## show version
	@$(GORUN) ./cmd/server version

run: ## run this program
	@$(GORUN) ./cmd/server

.PHONY: apply/prod-env
apply/prod-env: ## apply production env
	direnv allow



# builds
#===============================================================
build:
	GOOS=linux $(GOBUILD) ./cmd/server

build/pi:
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) ./cmd/server

build/pi64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) ./cmd/server



# tests
#===============================================================
.PHONY: gtest
gtest: ## go test
	$(GOTEST) -v ./...

.PHONY: lint
lint: ## golang-ci lint
	$(GOLANGCI) run --config=.golangci.yaml ./...

.PHONY: benchmark
benchmark: ## go benchmark
	$(GOTEST) -bench -v ./...

.PHONY: doc
doc: ## godoc http:6060
	$(GODOC) -http=:6060

.PHONY: coverage
coverage: ## coverage
	$(GOTEST) -coverprofile=cover.out ./...
	$(GOTOOL) cover -html=cover.out -o cover.html



# Makefile config
#===============================================================
help: ## Display this help screen
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
