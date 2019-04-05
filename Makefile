# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

include Makefile.common

PROMTOOL_VERSION ?= 2.5.0
PROMTOOL_URL     ?= https://github.com/prometheus/prometheus/releases/download/v$(PROMTOOL_VERSION)/prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM).tar.gz
PROMTOOL         ?= $(FIRST_GOPATH)/bin/promtool

DOCKER_IMAGE_NAME ?= shell-exit-status-exporter
MACH              ?= $(shell uname -m)
DOCKERFILE        ?= Dockerfile

STATICCHECK_IGNORE =

# Use CGO for non-Linux builds.
ifeq ($(GOOS), linux)
	PROMU_CONF ?= .promu.yml
else
	ifndef GOOS
		ifeq ($(GOHOSTOS), linux)
			PROMU_CONF ?= .promu.yml
		else
			PROMU_CONF ?= .promu-cgo.yml
		endif
	else
		PROMU_CONF ?= .promu-cgo.yml
	endif
endif

PROMU := $(FIRST_GOPATH)/bin/promu --config $(PROMU_CONF)

# 64bit -> 32bit mapping for cross-checking. At least for amd64/386, the 64bit CPU can execute 32bit code but not the other way around, so we don't support cross-testing upwards.
cross-test = skip-test-32bit

# By default, "cross" test with ourselves to cover unknown pairings.
$(eval $(call goarch_pair,amd64,386))
$(eval $(call goarch_pair,mips64,mips))
$(eval $(call goarch_pair,mips64el,mipsel))

all:: vet common-all $(cross-test)

.PHONY: skip-test-32bit
skip-test-32bit:
	@echo ">> SKIP running tests in 32-bit mode: not supported on $(GOHOSTOS)/$(GOHOSTARCH)"

.PHONY: docker
docker:
ifeq ($(MACH), ppc64le)
	$(eval DOCKERFILE=Dockerfile.ppc64le)
endif
	@echo ">> building docker image from $(DOCKERFILE)"
	@docker build --file $(DOCKERFILE) -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

.PHONY: test-docker
test-docker:
	@echo ">> testing docker image"
	./test_image.sh "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" 9100

.PHONY: promtool
promtool: $(PROMTOOL)

$(PROMTOOL):
	$(eval PROMTOOL_TMP := $(shell mktemp -d))
	curl -s -L $(PROMTOOL_URL) | tar -xvzf - -C $(PROMTOOL_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMTOOL_TMP)/prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM)/promtool $(FIRST_GOPATH)/bin/promtool
	rm -r $(PROMTOOL_TMP)

