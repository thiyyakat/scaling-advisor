# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

REPO_ROOT           := $(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))")
HACK_DIR            := $(REPO_ROOT)/hack
SUBMODULES          := minkapi api operator

TOOLS_DIR := $(HACK_DIR)/tools
include $(HACK_DIR)/tools.mk

.PHONY: add-license-headers
add-license-headers: $(GO_ADD_LICENSE)
	@$(HACK_DIR)/addlicenseheaders.sh

.PHONY: tidy
tidy:
	@for dir in $(SUBMODULES); do \
	  $(MAKE) -C $$dir tidy; \
	done

.PHONY: build
build:
	@for dir in $(SUBMODULES); do \
	  $(MAKE) -C $$dir build; \
	done