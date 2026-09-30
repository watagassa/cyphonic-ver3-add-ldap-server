GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODOC=$(GOCMD)doc
GENERALUSER=pluslab
DOCKER=docker
MINI-COMPOSE-FILE=mini-compose.yaml
COMPOSE=$(DOCKER) compose
MINI-COMPOSE=$(COMPOSE) -f $(MINI-COMPOSE-FILE)
EXEC=$(COMPOSE) exec
EXECD=$(COMPOSE) exec -d
EXECU=$(COMPOSE) exec -u $(GENERALUSER)
BUILD=$(COMPOSE) build
UP=$(COMPOSE) up -d
MINI-UP=$(MINI-COMPOSE) up -d
LOGS=$(COMPOSE) logs
STOP=$(COMPOSE) stop
RM=$(COMPOSE) rm
DOWN=$(COMPOSE) down
MINI-DOWN=$(MINI-COMPOSE) down
AS=$(EXEC) as
PS=$(EXEC) ps
U-TRS=$(EXEC) u-trs
CMS=$(EXEC) cms
U-NS=$(EXEC) u-ns
DS=$(EXEC) ds
FS=$(EXEC) fs
CONTROLLER=$(EXEC) controller
WEB=$(EXEC) web
NODE=$(EXEC) node
NODE2=$(EXEC) node2
NODE3=$(EXEC) node3
NODE4=$(EXEC) node4
DB=$(EXEC) cockroach
OPS=$(EXEC) ops
SHELL=sh
PRECOMMIT=pre-commit
KUBECTL=kubectl
SERVICES=as ds u-trs node node2 node3 node4
MIDDLEWARE_NAMESPACE=middleware
DB_NAME=cyphonic
DB_USER=cyphonic
DB_PORT=26257


AS_TAG_VERSION=v1.0
PS_TAG_VERSION=v1.0
U-TRS_TAG_VERSION=v1.0



# develop for docker-compose
#===============================================================
define ExporterInitRun
	$(EXEC) $(1) make node_exporter/init
	$(EXECD) $(1) make node_exporter/run

endef

define FluentdInitRun
	$(EXECU) $(1) make fluentd/init
	$(EXECD) $(1) make fluentd/run

endef

all: docker/up init_cert init_exporter init_fluentd

mini: docker/up-mini init_cert

init_cert: create/certificate

init_exporter: ## exporter init
	$(foreach service,$(SERVICES),$(call ExporterInitRun,$(service)))

init_fluentd: ## fluentd init
	$(foreach service,$(SERVICES),$(call FluentdInitRun,$(service)))

docker/build: ## docker build
	$(BUILD)

docker/up: ## docker up
	$(UP)

docker/up-mini: ## docker up mini
	$(MINI-UP)

docker/logs: ## docker logs
	$(LOGS)

docker/stop: ## docker stop
	$(STOP)

docker/rm: ## docker clean
	$(RM)

docker/down: ## docker down
	$(DOWN) -v

docker/down-mini: ## docker down mini
	$(MINI-DOWN) -v

docker/volume/prune: ### docker volume prune
	$(DOCKER) volume prune

as/bash: ## as container bash
	$(AS) bash

ps/bash: ## ps container bash
	$(PS) bash

u-trs/bash: ## trs container bash
	$(U-TRS) bash

cms/bash: ## cms container bash
	$(CMS) bash

u-ns/bash: ## u-ns container bash
	$(U-NS) bash

ds/bash: ## ds container bash
	$(DS) bash

fs/bash: ## fs container bash
	$(FS) bash

controller/bash: ## controller container bash
	$(CONTROLLER) bash

web/sh: ## web container sh
	$(WEB) sh

node/bash: ## node container bash
	$(NODE) bash

node2/bash: ## node2 container bash
	$(NODE2) bash

node3/bash: ## node3 container bash
	$(NODE3) bash

node4/bash: ## node4 container bash
	$(NODE4) bash

db/bash: ## db(cockroachDB) container bash
	$(DB) bash

cockroach: ## db(cockroachDB) container's cockroachDB access
	$(DB) cockroach sql -u $(DB_USER) -p $(DB_PORT) -d $(DB_NAME) --insecure

create/certificate:
	$(OPS) make create/certificate

# Environment
#===============================================================
.PHONY: plugin-install
plugin-install: ## Install asdf plugins
	@${SHELL} ./bin/scripts/plugin.sh

.PHONY: precommit-install
precommit-install: ## Set pre-commit
	@${PRECOMMIT} install



# production for kubernetes
#===============================================================
.PHONY: login
login: ## docker login
	$(DOCKER) login

.PHONY: prod/build/all
prod/build/all: prod/build/as prod/build/u-trs

.PHONY: prod/build/as
prod/build/as: ## build as' production image
	$(COMPOSE) -f production.yaml build as --progress=plain --no-cache

.PHONY: prod/as/rename-tag
prod/as/rename-tag: ## change the tag of the asd image for production
	$(DOCKER) tag $$($(DOCKER) images | grep pluslab/cyphonic-asd | awk '{print $$3}') pluslab/cyphonic-asd:${AS_TAG_VERSION}

.PHONY: prod/as/push
prod/as/push: ## push to as image registry
	$(DOCKER) push pluslab/cyphonic-asd:${AS_TAG_VERSION}

.PHONY: prod/build/ps
prod/build/ps: ## build ps' production image
	$(COMPOSE) -f production.yaml build ps --progress=plain --no-cache

.PHONY: prod/ps/rename-tag
prod/ps/rename-tag: ## change the tag of the asd image for production
	$(DOCKER) tag $$($(DOCKER) images | grep pluslab/cyphonic-psd | awk '{print $$3}') pluslab/cyphonic-psd:${PS_TAG_VERSION}

.PHONY: prod/ps/push
prod/ps/push: ## push to ps image registry
	$(DOCKER) push pluslab/cyphonic-psd:${PS_TAG_VERSION}

.PHONY: prod/build/u-trs
prod/build/u-trs: ## build u-trs' production image
	$(COMPOSE) -f production.yaml build u-trs --progress=plain --no-cache

.PHONY: prod/u-trs/rename-tag
prod/u-trs/rename-tag: ## change the tag of the u-trsd image for production
	$(DOCKER) tag $$($(DOCKER) images | grep pluslab/cyphonic-u-trsd | awk '{print $$3}') pluslab/cyphonic-u-trsd:${U-TRS_TAG_VERSION}

.PHONY: prod/u-trs/push
prod/u-trs/push: ## push to u-trs image registry
	$(DOCKER) push pluslab/cyphonic-u-trsd:${U-TRS_TAG_VERSION}

.PHONY: kube-cockroach
kube-cockroach: ## kubernetes: db(Cockroach) container's CockroachDB access
	@${KUBECTL} exec -it $$(${KUBECTL} get pod -n ${MIDDLEWARE_NAMESPACE} | grep cyphonic-db- | awk '{print $$1}') -c cockroach sql -u $(DB_USER) -p $(DB_PORT) -d $(DB_NAME) --insecure



# tests
#===============================================================
test: test/as test/u-trs ## go test all

test/as: ## go test(AS)
	$(AS) $(GOTEST) -v ./...

test/u-trs: ## go test(TRS)
	$(U-TRS) $(GOTEST) -v ./...

doc: ## godoc http:6060
	$(GODOC) -http=:6060



# Makefile config
#===============================================================
help:
	echo "Usage: make [task]\n\nTasks:"
	perl -nle 'printf("    \033[33m%-30s\033[0m %s\n",$$1,$$2) if /^([a-zA-Z0-9_\/-]*?):(?:.+?## )?(.*?)$$/' $(MAKEFILE_LIST)

.SILENT: help

.PHONY: $(shell egrep -o '^(\._)?[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/://')
