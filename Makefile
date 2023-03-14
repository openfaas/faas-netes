.PHONY: build local push namespaces install charts start-kind stop-kind build-buildx render-charts verify-charts charts-only
IMG_NAME?=faas-netes

VERBOSE?=false

TAG?=latest
OWNER?=openfaas
SERVER?=ghcr.io
export DOCKER_CLI_EXPERIMENTAL=enabled
export DOCKER_BUILDKIT=1

TOOLS_DIR := .tools

GOPATH := $(shell go env GOPATH)
CODEGEN_VERSION := $(shell hack/print-codegen-version.sh)
CODEGEN_PKG := $(GOPATH)/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}

VERSION := $(shell git describe --tags --dirty --always)
GIT_COMMIT := $(shell git rev-parse HEAD)

all: build-docker

$(TOOLS_DIR)/code-generator.mod: go.mod
	@echo "syncing code-generator tooling version"
	@cd $(TOOLS_DIR) && go mod edit -require "k8s.io/code-generator@${CODEGEN_VERSION}"

${CODEGEN_PKG}: $(TOOLS_DIR)/code-generator.mod
	@echo "(re)installing k8s.io/code-generator-${CODEGEN_VERSION}"
	@cd $(TOOLS_DIR) && go mod download -modfile=code-generator.mod

local:
	CGO_ENABLED=0 GOOS=linux go build -o faas-netes

build-docker:
	docker build \
	--build-arg GIT_COMMIT=$(GIT_COMMIT) \
	--build-arg VERSION=$(VERSION) \
	-t $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) .

.PHONY: build-buildx
build-buildx:
	@echo $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) && \
	docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--push \
		--platform linux/amd64 \
        --build-arg GIT_COMMIT=$(GIT_COMMIT) \
        --build-arg VERSION=$(VERSION) \
		--tag $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) \
		.

.PHONY: build-buildx-all
build-buildx-all:
	@docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform linux/amd64,linux/arm/v7,linux/arm64 \
		--output "type=image,push=false" \
        --build-arg GIT_COMMIT=$(GIT_COMMIT) \
        --build-arg VERSION=$(VERSION) \
		--tag $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) \
		.

.PHONY: publish-buildx-all
publish-buildx-all:
	@echo  $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) && \
	docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform linux/amd64,linux/arm/v7,linux/arm64 \
		--push=true \
        --build-arg GIT_COMMIT=$(GIT_COMMIT) \
        --build-arg VERSION=$(VERSION) \
		--tag $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) \
		.

push:
	docker push $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG)

charts: verify-charts charts-only

verify-charts:
	@echo Verifying helm charts images in remote registries && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/openfaas/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/kafka-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/cron-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/nats-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/mqtt-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/pro-builder/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/sqs-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/postgres-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/queue-worker/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/sns-connector/values.yaml

upgrade-charts:
	@echo Upgrading helm charts images && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/openfaas/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/kafka-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/cron-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/nats-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/mqtt-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/pro-builder/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/sqs-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/postgres-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/queue-worker/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/sns-connector/values.yaml

charts-only:
	@cd chart && \
		helm package openfaas/ && \
		helm package kafka-connector/ && \
		helm package cron-connector/ && \
		helm package nats-connector/ && \
		helm package mqtt-connector/ && \
		helm package pro-builder/ && \
		helm package sqs-connector/ && \
		helm package postgres-connector/ && \
		helm package queue-worker/ && \
		helm package sns-connector/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml
	./contrib/create-static-manifest.sh

render-charts:
	./contrib/create-static-manifest.sh

start-kind: ## attempt to start a new dev environment
	@./contrib/create_dev.sh

stop-kind: ## attempt to stop the dev environment
	@./contrib/stop_dev.sh

.PHONY: verify-codegen
verify-codegen: ${CODEGEN_PKG}
	./hack/verify-codegen.sh

.PHONY: update-codegen
update-codegen: ${CODEGEN_PKG}
	./hack/update-codegen.sh
