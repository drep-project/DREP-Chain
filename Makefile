# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: drep-linux drep-linux-386 drep-linux-amd64 drep-linux-mips64 drep-linux-mips64le
.PHONY: drep-darwin drep-darwin-386 drep-darwin-amd64
.PHONY: drep-windows drep-windows-386 drep-windows-amd64

BIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go build

drep:
	$(GORUN) ./cmds/drep/drep.go
	@echo "Done building."
	@echo "Run \"$(GOBIN)/drep\" to launch drep."


GORUN-WIN = env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GO111MODULE=on go build
GORUN-LINXU64 = env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build
GORUN-DARWIN64 = env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GO111MODULE=on go build
all:
	#win64
	$(GORUN-WIN) -o $(BIN)/drep-win-amd64-v1.0.0/drep.exe ./cmds/drep/drep.go
	$(GORUN-WIN) -o $(BIN)/drep-win-amd64-v1.0.0/drep-cli.exe ./cmds/drepClient/main.go
	$(GORUN-WIN) -o $(BIN)/drep-win-amd64-v1.0.0/genconfig/genaccount.exe ./cmds/genaccount/main.go
	#$(GORUN-WIN) -o $(BIN)/drep-win-amd64-v1.0.0/genapicode.exe ./cmds/genapicode/main.go
	#$(GORUN-WIN) -o $(BIN)/drep-win-amd64-v1.0.0/gendoc.exe ./cmds/gendoc/*.go
	mkdir $(BIN)/drep-win-amd64-v1.0.0/mainnet-config/
	mkdir $(BIN)/drep-win-amd64-v1.0.0/testnet-config/

	#linux 64
	$(GORUN-LINXU64) -o $(BIN)/drep-linux-amd64-v1.0.0/drep ./cmds/drep/drep.go
	$(GORUN-LINXU64) -o $(BIN)/drep-linux-amd64-v1.0.0/drep-cli ./cmds/drepClient/main.go
	$(GORUN-LINXU64) -o $(BIN)/drep-linux-amd64-v1.0.0/genconfig/genaccount ./cmds/genaccount/main.go
	#$(GORUN-LINXU64) -o $(BIN)/linux64/genapicode ./cmds/genapicode/main.go
	#$(GORUN-LINXU64) -o $(BIN)/linux64/gendoc ./cmds/gendoc/*.go
	mkdir $(BIN)/drep-linux-amd64-v1.0.0/mainnet-config/
	mkdir $(BIN)/drep-linux-amd64-v1.0.0/testnet-config/

	# mac 64
	$(GORUN-DARWIN64) -o $(BIN)/drep-darwin-amd64-v1.0.0/drep ./cmds/drep/drep.go
	$(GORUN-DARWIN64) -o $(BIN)/drep-darwin-amd64-v1.0.0/drepClient ./cmds/drepClient/main.go
	$(GORUN-DARWIN64) -o $(BIN)/drep-darwin-amd64-v1.0.0/genconfig/genaccount ./cmds/genaccount/main.go
	#$(GORUN-DARWIN64) -o $(BIN)/darwin64/genapicode ./cmds/genapicode/main.go
	#$(GORUN-DARWIN64) -o $(BIN)/darwin64/gendoc ./cmds/gendoc/*.go
	mkdir $(BIN)/drep-darwin-amd64-v1.0.0/mainnet-config/
	mkdir $(BIN)/drep-darwin-amd64-v1.0.0/testnet-config/


lint: ## Run linters.
	env GO111MODULE=on go run ./cmds/drep/drep.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr $(BIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

