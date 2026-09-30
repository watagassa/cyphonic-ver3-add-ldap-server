CURRENT_SHORTHASH := $(shell git rev-parse --short HEAD)
STAGE := dev
HELM := $(shell basename $$(which helm))
SHELL_NAME := $(shell basename $${SHELL:-bash})
TEMPLATE = $(HELM) template
INSTALL = $(HELM) install
LIST = $(HELM) list --all-namespaces

AS := ./as
NMS := ./nms
TRS := ./trs



# helm
#===============================================================
.PHONY: helm-version
helm-version: ## helm version
	$(HELM) version

.PHONY: helm-list
helm-list: ## installed chart list
	$(LIST)

.PHONY: as-template
as-template: ## as helm template
	@$(TEMPLATE) $(AS) -f as/values/dev.yaml -f as/values/dev-image-tag.yaml

.PHONY: nms-template
nms-template: ## nms helm template
	@$(TEMPLATE) $(NMS) -f nms/values/dev.yaml -f nms/values/dev-image-tag.yaml

.PHONY: trs-template
trs-template: ## trs helm template
	@$(TEMPLATE) $(TRS) -f trs/values/dev.yaml -f trs/values/dev-image-tag.yaml



# Makefile config
#===============================================================
help:
	echo "Usage: make [task]\n\nTasks:"
	perl -nle 'printf("    \033[33m%-30s\033[0m %s\n",$$1,$$2) if /^([a-zA-Z0-9_\/-]*?):(?:.+?## )?(.*?)$$/' $(MAKEFILE_LIST)

.SILENT: help

.PHONY: $(shell egrep -o '^(\._)?[a-z_-]+:' $(MAKEFILE_LIST) | sed 's/://')
