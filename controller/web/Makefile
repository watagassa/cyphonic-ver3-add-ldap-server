NPM=npm
INSTALL=$(NPM) install
RUN=$(NPM) run
BUILD=$(RUN) build
START=$(NPM) start
UPDATE=$(NPM) update
PACKAGES=$(NPM) list
CACHE=$(NPM) cache
CONFIG=$(NPM) config



init: install run

install: ## module install
	$(INSTALL)

run: ## npm run
	$(RUN) dev

env: ## display the environment variable
	$(RUN) env

build: ## npm build
	$(BUILD)

start: ## npm start
	$(START)

list: ## package list
	$(PACKAGES) --depth=0

update: ## all packages update
	$(UPDATE)

cache-list: ## cache list
	$(CACHE) ls

cache-verify: ## cache verify
	$(CACHE) verify

config: ## config list
	$(CONFIG) list



# Makefile config
#===============================================================
help: ## display this help screen
	echo "Usage: make [task]\n\nTasks:"
	perl -nle 'printf("    \033[33m%-30s\033[0m %s\n",$$1,$$2) if /^([a-zA-Z_-]*?):(?:.+?## )?(.*?)$$/' $(MAKEFILE_LIST)

.SILENT: help

.PHONY: $(shell egrep -o '^(\._)?[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/://')
