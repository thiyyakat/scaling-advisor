# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: tidy
tidy:
	@env GO111MODULE=on go mod tidy

.PHONY: build
build:
	go build ./...

.PHONY: format
format: $(GOIMPORTS_REVISER)
	@$(REPO_ROOT)/hack/format.sh ./core

.PHONY: check
check: $(GOLANGCI_LINT) $(GOIMPORTS) format
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=../.golangci.yaml ./...