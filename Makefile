# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES.  All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include $(CURDIR)/versions.mk

PWD := $(shell pwd)
GEN_DIR := $(PWD)/gen
PKG_DIR := $(PWD)/pkg
GEN_BINDINGS_DIR := $(GEN_DIR)/nvfm
PKG_BINDINGS_DIR := $(PKG_DIR)/nvfm
GO ?= env GOCACHE=$(PWD)/.cache/go GOPATH=$(PWD)/.cache/gopath GOTOOLCHAIN=local go

CHECK_TARGETS := validate-modules golangci-lint

MAKE_TARGETS := build fmt generate test

GENERATE_TARGETS := bindings clean clean-bindings

TARGETS := $(MAKE_TARGETS) $(GENERATE_TARGETS) $(CHECK_TARGETS)

.PHONY: $(TARGETS)

build:
	$(GO) build $(MODULE)/pkg/...

check: $(CHECK_TARGETS)

fmt:
	$(GO) list -f '{{.Dir}}' $(MODULE)/pkg/... $(MODULE)/gen/... | xargs gofmt -s -l -w

golangci-lint:
	golangci-lint run ./pkg/...

generate:
	go generate $(MODULE)/...

COVERAGE_FILE := coverage.out
test: build
	$(GO) test -v -coverprofile=$(COVERAGE_FILE) $(MODULE)/pkg/... 

coverage: test
	go tool cover -func=$(COVERAGE_FILE)

clean: clean-bindings

validate-modules:
	@echo "- Verifying that the dependencies have expected content..."
	go mod verify
	@echo "- Checking for any unused/missing packages in go.mod..."
	go mod tidy
	git diff --exit-code HEAD -- go.mod go.sum

$(PKG_BINDINGS_DIR):
	mkdir -p $(@)

bindings: .create-bindings .strip-autogen-comment .strip-nvfm-h-linenumber
.create-bindings: $(GEN_BINDINGS_DIR)/nv_fm_agent.h $(GEN_BINDINGS_DIR)/nv_fm_types.h $(GEN_BINDINGS_DIR)/nvfm.yml | $(PKG_BINDINGS_DIR)
	cp $(GEN_BINDINGS_DIR)/nv_fm_agent.h $(PKG_BINDINGS_DIR)
	cp $(GEN_BINDINGS_DIR)/nv_fm_types.h $(PKG_BINDINGS_DIR)
	cp $(GEN_BINDINGS_DIR)/nvfm.yml $(PKG_BINDINGS_DIR)
	c-for-go -out $(PKG_DIR) $(PKG_BINDINGS_DIR)/nvfm.yml
	cd $(PKG_BINDINGS_DIR); $(GO) tool cgo -godefs types.go > types_gen.go
	cd $(PKG_BINDINGS_DIR); $(GO) fmt types_gen.go
	rm -rf $(PKG_BINDINGS_DIR)/nvfm.yml $(PKG_BINDINGS_DIR)/cgo_helpers.go $(PKG_BINDINGS_DIR)/types.go $(PKG_BINDINGS_DIR)/_obj
	rm -f $(PKG_BINDINGS_DIR)/_cgo_*.o
	$(GO) run $(GEN_BINDINGS_DIR)/generateapi.go --sourceDir $(PKG_BINDINGS_DIR) --output $(PKG_BINDINGS_DIR)/zz_generated.api.go
	make fmt

.strip-autogen-comment: SED_SEARCH_STRING := // WARNING: This file has automatically been generated on
.strip-autogen-comment: SED_REPLACE_STRING := // WARNING: THIS FILE WAS AUTOMATICALLY GENERATED.
.strip-autogen-comment: | .create-bindings
	grep -l -R "$(SED_SEARCH_STRING)" pkg | xargs -r sed -i -E 's#$(SED_SEARCH_STRING).*$$#$(SED_REPLACE_STRING)#g'

.strip-nvfm-h-linenumber: SED_SEARCH_STRING := // (.*) nvfm/nv_fm_.*\.h:[0-9]+
.strip-nvfm-h-linenumber: SED_REPLACE_STRING := // \1 nvfm/nv_fm_*.h
.strip-nvfm-h-linenumber: | .create-bindings
	grep -l -RE "$(SED_SEARCH_STRING)" pkg | xargs -r sed -i -E 's#:[0-9]+$$##g'

clean-bindings:
	rm -f $(PKG_BINDINGS_DIR)/cgo_helpers.go
	rm -f $(PKG_BINDINGS_DIR)/cgo_helpers.h
	rm -f $(PKG_BINDINGS_DIR)/const.go
	rm -f $(PKG_BINDINGS_DIR)/doc.go
	rm -f $(PKG_BINDINGS_DIR)/nvfm.go
	rm -f $(PKG_BINDINGS_DIR)/nv_fm_agent.h
	rm -f $(PKG_BINDINGS_DIR)/nv_fm_types.h
	rm -f $(PKG_BINDINGS_DIR)/types.go
	rm -f $(PKG_BINDINGS_DIR)/types_gen.go
	rm -f $(PKG_BINDINGS_DIR)/zz_generated.api.go
	rm -f $(PKG_BINDINGS_DIR)/_cgo_*.o
