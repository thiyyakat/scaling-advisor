# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

main_path = cmd/scalingadvisor/main.go
binary_name = scalingadvisor

.PHONY: tidy
tidy:
	@env GO111MODULE=on go mod tidy

.PHONY: build
build:
	mkdir -p bin
	go build -buildvcs -o=bin/${binary_name} ${main_path}
