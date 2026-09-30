GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODOC=$(GOCMD) doc
GOTOOL=$(GOCMD) tool
GOLANGCI=golangci-lint
COMPOSE=docker compose
EXEC=$(COMPOSE) exec
BUILD=$(COMPOSE) build
UP=$(COMPOSE) up -d
LOGS=$(COMPOSE) logs
STOP=$(COMPOSE) stop
RM=$(COMPOSE) rm
DOWN=$(COMPOSE) down
cmsd=$(EXEC) cmsd
DB=$(EXEC) db

DBNAME := cyphonic
TESTDBNAME := test_cyphonic



# runs
#===============================================================
.PHONY: run
run: ## start cms daemon
	$(GORUN) ./cmd/cmsd/main.go

.PHONY: apply/prod-env
apply/prod-env: ## apply production env
	direnv allow

# builds
#===============================================================
.PHONY: build
build: ## build cms daemon
	GOOS=linux GOARCH=amd64 $(GOBUILD) cmd/cmsd/main.go


# pkg daemon
#===============================================================
node_exporter: node_exporter/init node_exporter/run/background ## Install and run node_exporter

.PHONY: node_exporter/init
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

# tests
#===============================================================
gtest: ## go test
	$(GOTEST) -v ./...

lint: ## golang-ci lint
	$(GOLANGCI) run --config=.golangci.yaml ./...

benchmark: ## go benchmark
	$(GOTEST) -bench -v ./...

doc: ## godoc http:6060
	$(GODOC) -http=:6060

coverage: ## coverage
	$(GOTEST) -coverprofile=cover.out ./...
	$(GOTOOL) cover -html=cover.out -o cover.html



# Makefile config
#===============================================================
help:
	echo "Usage: make [task]\n\nTasks:"
	perl -nle 'printf("    \033[33m%-30s\033[0m %s\n",$$1,$$2) if /^([a-zA-Z0-9_\/-]*?):(?:.+?## )?(.*?)$$/' $(MAKEFILE_LIST)

.SILENT: help

.PHONY: $(shell egrep -o '^(\._)?[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/://')
