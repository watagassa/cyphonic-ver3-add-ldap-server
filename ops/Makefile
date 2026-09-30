GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build



# node certificate
#===============================================================
create/certificate: create/certificate/server create/certificate/node/authentication

create/certificate/server: ## create ca and server certificate files
	go run ./cmd/certificate/main.go
	cp ca.pem ../asd
	cp ca.pem ../psd
	cp ca.pem ../cmsd
	cp ca.pem ../fsd
	cp ca.pem ../noded
	cp ca_private_key.pem ../asd
	cp ca_private_key.pem ../psd
	cp as.local.cyphonic.org.pem ../asd
	cp as.local.cyphonic.org.key ../asd
	cp as.local.cyphonic.org.pem ../noded
	cp ps.local.cyphonic.org.pem ../psd
	cp ps.local.cyphonic.org.key ../psd
	cp ps.local.cyphonic.org.pem ../noded
	cp cms.local.cyphonic.org.pem ../cmsd
	cp cms.local.cyphonic.org.key ../cmsd
	cp cms.local.cyphonic.org.key ../noded
	cp fs.local.cyphonic.org.pem ../fsd
	cp fs.local.cyphonic.org.key ../fsd
	cp fs.local.cyphonic.org.key ../noded

create/certificate/node/authentication: ## create node certificate authentication files
	go run ./cmd/certificate/main.go test01@cyphonic.org node
	go run ./cmd/certificate/main.go test02@cyphonic.org node2
	go run ./cmd/certificate/main.go test03@cyphonic.org node3
	go run ./cmd/certificate/main.go test04@cyphonic.org node4
	cp node_cert.pem ../noded
	cp node2_cert.pem ../noded
	cp node3_cert.pem ../noded
	cp node4_cert.pem ../noded



# accounts
#===============================================================
add/account: ## Add CYPHONIC account to Cockroach
	$(GORUN) ./cmd/account/create_account.go test01@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test02@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test03@cyphonic.org cyphonic
	$(GORUN) ./cmd/account/create_account.go test04@cyphonic.org cyphonic



# Makefile config
#===============================================================
help: ## Display this help screen
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-28s\033[0m %s\n", $$1, $$2}'
