GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODOC=$(GOCMD)doc
GOLANGCI=golangci-lint

all: gtest lint

run: ## go run main.go
	$(GORUN) ./cmd/u-trsd/main.go

gtest: ## go test
	$(GOTEST) -v ./...

lint: ## golang-ci lint
	$(GOLANGCI) run --config=.golangci.yaml ./...

benchmark: ## go benchmark
	$(GOTEST) -bench -v ./...

doc: ## godoc http:6060
	$(GODOC) -http=:6060

add/trs: ## Add TRS to Cockroach
	$(GORUN) cmd/add/main.go asia trs01.local.cyphonic.org

node_exporter: node_exporter/init node_exporter/run/background ## Install and run node_exporter

node_exporter/init: ## Download node_exporter
	./scripts/exporter/node_exporter.sh

.PHONY: node_exporter/run
node_exporter/run: ## Run node_exporter
	node_exporter

.PHONY: node_exporter/run/background
node_exporter/run/background: ## Run node_exporter in background
	node_exporter &

fluentd: fluentd/init fluentd/run/background ## Install and run fluentd

.PHONY: fluentd/init
fluentd/init: ## Download fluentd
	./scripts/fluentd/fluentd.sh

.PHONY: fluentd/run
fluentd/run: ## Run fluentd
	fluentd -c ./scripts/fluentd/fluent.conf

.PHONY: fluentd/run/background
fluentd/run/background: ## Run fluentd in background
	fluentd -c ./scripts/fluentd/fluent.conf &

build/linux: ## Build to run TRS on linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) cmd/u-trsd/main.go

build/pi: ## Build to run TRS on Raspberry Pi
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) cmd/u-trsd/main.go

help: ## Display this help screen
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-28s\033[0m %s\n", $$1, $$2}'
