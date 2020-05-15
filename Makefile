# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: drep-linux drep-linux-386 drep-linux-amd64 drep-linux-mips64 drep-linux-mips64le
.PHONY: drep-darwin drep-darwin-386 drep-darwin-amd64
.PHONY: drep-windows drep-windows-386 drep-windows-amd64

GOBIN = $(pwd)/build/bin
GO ?= latest
GORUN = env GO111MODULE=on go install

drep:
	$(GORUN) ./cmds/drep/drep.go
	@echo "Done building."
	@echo "Run \"$(GOBIN)/drep\" to launch drep."

all:
	$(GORUN) ./cmds/drep/drep.go
	$(GORUN) ./cmds/drepClient/main.go
	$(GORUN) ./cmds/genaccount/
	$(GORUN) ./cmds/genapicode/main.go
	$(GORUN) ./cmds/gendoc/*.go


ios:
	$(GORUN) ./cmds/drep/drep.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Drep.framework\" to use the library."

test: all
	$(GORUN) ./cmds/drep/drep.go test

lint: ## Run linters.
	$(GORUN) ./cmds/drep/drep.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr ./cmds/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.


# Cross Compilation Targets (xgo)

drep-cross: drep-linux drep-darwin drep-windows geth-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/geth-*

drep-linux: drep-linux-386 drep-linux-amd64 drep-linux-arm drep-linux-mips64 drep-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/geth-linux-*

drep-linux-386:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/drep
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/geth-linux-* | grep 386

drep-linux-amd64:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/drep
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/geth-linux-* | grep amd64

drep-linux-mips:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/drep
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/drep-linux-* | grep mips

drep-linux-mipsle:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/drep
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/drep-linux-* | grep mipsle

drep-linux-mips64:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/drep
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/drep-linux-* | grep mips64

drep-linux-mips64le:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/drep
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/drep-linux-* | grep mips64le

drep-darwin: drep-darwin-386 drep-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/drep-darwin-*

drep-darwin-386:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/drep
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/drep-darwin-* | grep 386

drep-darwin-amd64:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/drep
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/drep-darwin-* | grep amd64

drep-windows: drep-windows-386 drep-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/drep-windows-*

drep-windows-386:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/drep
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/drep-windows-* | grep 386

drep-windows-amd64:
	$(GORUN) ./cmds/drep/drep.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/drep
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/drep-windows-* | grep amd64