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
ASD=$(EXEC) asd
DB=$(EXEC) db

DBNAME := cyphonic
TESTDBNAME := test_cyphonic



# runs
#===============================================================
.PHONY: run
run: ## start as daemon
	$(GORUN) ./cmd/asd/main.go

.PHONY: apply/prod-env
apply/prod-env: ## apply production env
	direnv allow



# builds
#===============================================================
.PHONY: build
build: ## build as daemon
	GOOS=linux GOARCH=amd64 $(GOBUILD) cmd/asd/main.go



# initializes
#===============================================================
add/account: ## Add CYPHONIC account to Cockroach
	$(GORUN) ./cmd/account/create_account.go test01@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test02@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test03@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test04@cyphonic.org cyphonic


add/device: ## Add CYPHONIC account device to Cockroach
	$(GORUN) ./cmd/device/create_device.go test01@cyphonic.org cyphonic node
	$(GORUN) ./cmd/device/create_device.go test02@cyphonic.org cyphonic node2
	$(GORUN) ./cmd/device/create_device.go test03@cyphonic.org cyphonic node3
	$(GORUN) ./cmd/device/create_device.go test04@cyphonic.org cyphonic node4
	cp node_cert.pem ../noded
	cp node2_cert.pem ../noded
	cp node3_cert.pem ../noded
	cp node4_cert.pem ../noded
	rm node.pem node2.pem node3.pem node4.pem

create/certificate/server: ## create ca and server certificate files
	$(GORUN) ./cmd/certificate/main.go
	cp ca.pem ../noded
	cp as.local.cyphonic.org.pem ../noded

create/certificate/node: ## create node certificate files
	$(GORUN) ./cmd/certificate/main.go test01@cyphonic.org node_client
	$(GORUN) ./cmd/certificate/main.go test02@cyphonic.org node2_client
	$(GORUN) ./cmd/certificate/main.go test03@cyphonic.org node3_client
	$(GORUN) ./cmd/certificate/main.go test04@cyphonic.org node4_client
	cp node_client_cert.pem ../noded
	cp node2_client_cert.pem ../noded
	cp node3_client_cert.pem ../noded
	cp node4_client_cert.pem ../noded
	rm node_client.pem node2_client.pem node3_client.pem node4_client.pem

create/certificate/node/authentication: ## create node certificate authentication files
	$(GORUN) ./cmd/certificate/main.go test01@cyphonic.org node
	$(GORUN) ./cmd/certificate/main.go test02@cyphonic.org node2
	$(GORUN) ./cmd/certificate/main.go test03@cyphonic.org node3
	$(GORUN) ./cmd/certificate/main.go test04@cyphonic.org node4
	cp node_cert.pem ../noded
	cp node2_cert.pem ../noded
	cp node3_cert.pem ../noded
	cp node4_cert.pem ../noded
	rm node.pem node2.pem node3.pem node4.pem



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
