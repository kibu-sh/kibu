go.generate:
	@echo "Generating mocks..."
	@go generate ./...
.PHONY: go.generate


go.build:
	@echo "Building..."
	@go build ./...
.PHONY: go.build

generate.cmd:
	go generate ./cmd/...
.PHONY: generate.cmd

install: generate.cmd
	@go install ./cmd/kibu
.PHONY: install

install.kibugenv2:
	 go install ./internal/toolchain/kibugenv2/cmd/kibugenv2
.PHONY: install.kibugenv2

install.kibugen_ts:
	go install ./internal/toolchain/kibugen_ts/cmd/kibugen_ts
.PHONY: install.kibugen_ts


install.kibuwire:
	 go install ./internal/toolchain/kibuwire/cmd/kibuwire
.PHONY: install.kibuwire
