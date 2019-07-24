GOBIN ?= $(GOPATH)/bin

export GO111MODULE = on

########################################
### Tools

GOLANGCI_LINT_VERSION := v1.15.0
GOLANGCI_LINT_HASHSUM := ac897cadc180bf0c1a4bf27776c410debad27205b22856b861d41d39d06509cf

get_tools: tools-stamp
tools-stamp: $(GOBIN)/golangci-lint
	touch $@

$(GOBIN)/golangci-lint: contrib/install-golangci-lint.sh
	bash contrib/install-golangci-lint.sh $(GOBIN) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)


########################################
### Dependencies

go-mod-cache:
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

distclean: clean
	rm -rf vendor/


########################################
### Build/Install

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly -o build/drep.exe ./cmds/drep
	go build -mod=readonly -o build/drepClient.exe ./cmds/drepClient
	go build -mod=readonly -o build/genaccount.exe ./cmds/genaccount
    go build -mod=readonly -o build/genapicode.exe ./cmds/genapicode
	go build -mod=readonly -o build/gendoc.exe ./cmds/gendoc
else
	go build -mod=readonly -o build/drep ./cmds/drep
	go build -mod=readonly -o build/drepClient ./cmds/drepClient
	go build -mod=readonly -o build/genaccount ./cmds/genaccount
	go build -mod=readonly -o build/genapicode ./cmds/genapicode
	go build -mod=readonly -o build/gendoc ./cmds/gendoc
endif


install: go.sum
	go install -mod=readonly ./cmds/drep
	go install -mod=readonly ./cmds/drepClient
	go install -mod=readonly ./cmds/genaccount
	go install -mod=readonly ./cmds/genapicode
	go install -mod=readonly ./cmds/gendoc


########################################
### Testing

ci-lint:
	golangci-lint run
	go vet -composites=false -tests=false ./...
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify