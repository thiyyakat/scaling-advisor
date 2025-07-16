# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

main_package_path = .
binary_name = scaling-advisor

.PHONY: tidy
tidy:
	@env GO111MODULE=on go mod tidy

.PHONY: build
build:
	mkdir -p bin
    # Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
	go build -buildvcs -o=bin/${binary_name} ${main_package_path}
