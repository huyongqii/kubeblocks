#
# Copyright ApeCloud, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# To use buildx: https://github.com/docker/buildx#docker-ce
export DOCKER_CLI_EXPERIMENTAL=enabled

DEBIAN_MIRROR=mirrors.aliyun.com


# Docker image build and push setting
DOCKER:=docker
DOCKERFILE_DIR?=./docker

# Image URL to use all building/pushing image targets
IMG ?= docker.io/apecloud/$(APP_NAME)
CLI_IMG ?= docker.io/apecloud/kbcli
CLI_TAG ?= v$(CLI_VERSION)

# Update whenever you upgrade dev container image
DEV_CONTAINER_VERSION_TAG ?= latest
DEV_CONTAINER_IMAGE_NAME = docker.io/apecloud/$(APP_NAME)-dev

DEV_CONTAINER_DOCKERFILE = Dockerfile-dev
DOCKERFILE_DIR = ./docker
BUILDX_ARGS ?=

##@ Docker containers

.PHONY: build-dev-container-image
build-dev-image: DOCKER_BUILD_ARGS += --build-arg DEBIAN_MIRROR=$(DEBIAN_MIRROR) --build-arg GITHUB_PROXY=$(GITHUB_PROXY) --build-arg GOPROXY=$(GOPROXY)
build-dev-image: ## Build dev container image.
ifneq ($(BUILDX_ENABLED), true)
	docker build $(DOCKERFILE_DIR)/. $(DOCKER_BUILD_ARGS) -f $(DOCKERFILE_DIR)/${DEV_CONTAINER_DOCKERFILE} -t $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)
else
	docker buildx build $(DOCKERFILE_DIR)/.  $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -f $(DOCKERFILE_DIR)/$(DEV_CONTAINER_DOCKERFILE) -t $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG) $(BUILDX_ARGS)
endif


.PHONY: push-dev-container-image
push-dev-image: DOCKER_BUILD_ARGS += --build-arg DEBIAN_MIRROR=$(DEBIAN_MIRROR) --build-arg GITHUB_PROXY=$(GITHUB_PROXY) --build-arg GOPROXY=$(GOPROXY)
push-dev-image: ## Push dev container image.
ifneq ($(BUILDX_ENABLED), true)
	docker push $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)
else
	docker buildx build . $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -f $(DOCKERFILE_DIR)/$(DEV_CONTAINER_DOCKERFILE) -t $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG) --push $(BUILDX_ARGS)
endif


.PHONY: build-manager-image
build-manager-image: generate ## Build Operator manager container image.
ifneq ($(BUILDX_ENABLED), true)
	docker build . -t ${IMG}:${VERSION} -f $(DOCKERFILE_DIR)/Dockerfile -t ${IMG}:latest
else
ifeq ($(TAG_LATEST), true)
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:latest $(BUILDX_ARGS)
else
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:${VERSION} $(BUILDX_ARGS)
endif
endif


.PHONY: push-manager-image
push-manager-image: generate ## Push Operator manager container image.
ifneq ($(BUILDX_ENABLED), true)
ifeq ($(TAG_LATEST), true)
	docker push ${IMG}:latest
else
	docker push ${IMG}:${VERSION}
endif
else
ifeq ($(TAG_LATEST), true)
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:latest --push $(BUILDX_ARGS)
else
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:${VERSION} --push $(BUILDX_ARGS)
endif
endif

.PHONY: build-tools-image
build-tools-image: generate ## Build tools container image.
ifneq ($(BUILDX_ENABLED), true)
	docker build . -t ${IMG}:${VERSION} -f $(DOCKERFILE_DIR)/Dockerfile-tools -t ${IMG}:latest
else
ifeq ($(TAG_LATEST), true)
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile-tools $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:latest $(BUILDX_ARGS)
else
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile-tools $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:${VERSION} $(BUILDX_ARGS)
endif
endif

.PHONY: push-tools-image
push-tools-image: generate ## Push tools container image.
ifneq ($(BUILDX_ENABLED), true)
ifeq ($(TAG_LATEST), true)
	docker push ${IMG}:latest
else
	docker push ${IMG}:${VERSION}
endif
else
ifeq ($(TAG_LATEST), true)
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile-tools $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:latest --push $(BUILDX_ARGS)
else
	docker buildx build . -f $(DOCKERFILE_DIR)/Dockerfile-tools $(DOCKER_BUILD_ARGS) --platform $(BUILDX_PLATFORMS) -t ${IMG}:${VERSION} --push $(BUILDX_ARGS)
endif
endif
