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
NODED=$(EXEC) noded
DB=$(EXEC) db

DBNAME := cyphonic
TESTDBNAME := test_cyphonic



# runs
#===============================================================
.PHONY: run
run: ## start node daemon
	$(GORUN) ./cmd/noded/main.go run

.PHONY: apply/prod-env
apply/prod-env: ## apply production env
	direnv allow



# builds
#===============================================================
.PHONY: build
build: ## build node daemon
	GOOS=linux GOARCH=amd64 $(GOBUILD) cmd/noded/main.go

# coredns
#===============================================================

.PHONY: coredns/init
coredns/init: ## Install CoreDNS
	./scripts/dns/coredns.sh

.PHONY: coredns/run.v4
coredns/run.v4: ## Run CoreDNS in v4 mode
	coredns -conf ./scripts/dns/Corefile.v4

.PHONY: coredns/run
coredns/run: ## Run CoreDNS in v6 mode
	coredns -conf ./scripts/dns/Corefile

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

# /healthz
#===============================================================
.PHONY: /healthz
healthz:
	curl -w "\n\n" -H 'accept: application/json' \
	'http://localhost:8081/healthz'

# status
#===============================================================
.PHONY: status
status:
	curl -w "\n\n" -H 'accept: application/json' \
	'http://localhost:8081/status'

# fqdn
#===============================================================
.PHONY: fqdn
fqdn:
	curl -w "\n\n" -H 'accept: application/json' \
	'http://localhost:8081/fqdn'

# login
#===============================================================
.PHONY: node01/login node02/login
node01/login:
	curl -w "\n\n" -X POST 'http://localhost:8081/login' \
	-H 'accept: applation/json' \
	-H 'Content-Type: application/json' \
	-d '{"username": "f6efb4e0355cae5a37ddd58a286d4bcbe6ff4ca585c67cb02f263aa3", "password": "cyphonic"}'

node02/login:
	curl -w "\n\n" -X POST 'http://localhost:8081/login' \
	-H 'accept: applation/json' \
	-H 'Content-Type: application/json' \
	-d '{"username": "f7b8c180a4d573cb55479f76377f35fbbe58453362736f6640c2612a", "password": "cyphonic"}'

node03/login:
	curl -w "\n\n" -X POST 'http://localhost:8081/login' \
	-H 'accept: applation/json' \
	-H 'Content-Type: application/json' \
	-d '{"username": "204df901511616cf3b9eba8df20003897e54bb3fceaaa3bf4ea85d50", "password": "cyphonic"}'

node04/login:
	curl -w "\n\n" -X POST 'http://localhost:8081/login' \
	-H 'accept: applation/json' \
	-H 'Content-Type: application/json' \
	-d '{"username": "0cd80324265f346233dfcf4318725b141be4d59676fb2836e684c560", "password": "cyphonic"}'

# logout
#===============================================================
.PHONY: logout
logout:
	curl -w "\n\n" -X POST 'http://localhost:8081/logout' \
	-H 'accept: applation/json' \
	-H 'Content-Type: application/json' \

# Makefile config
#===============================================================
help:
	echo "Usage: make [task]\n\nTasks:"
	perl -nle 'printf("    \033[33m%-30s\033[0m %s\n",$$1,$$2) if /^([a-zA-Z0-9_\/-]*?):(?:.+?## )?(.*?)$$/' $(MAKEFILE_LIST)

.SILENT: help

.PHONY: $(shell egrep -o '^(\._)?[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/://')
