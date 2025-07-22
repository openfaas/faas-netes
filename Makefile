.PHONY: local publish-buildx publish-buildx-all namespaces install charts start-kind stop-kind build-buildx build-buildx-all render-charts verify-charts verify-chart charts-only bump-charts

IMG_NAME?=faas-netes

VERBOSE?=false

TAG?=latest
OWNER?=openfaas
SERVER?=ttl.sh
export DOCKER_CLI_EXPERIMENTAL=enabled
export DOCKER_BUILDKIT=1

VERSION := $(shell git describe --tags --dirty --always)
GIT_COMMIT := $(shell git rev-parse HEAD)

all: build-docker

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
		--output "type=image,push=false" \
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

.PHONY: publish-buildx
publish-buildx:
	@echo  $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) && \
	docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform linux/amd64 \
		--push=true \
        --build-arg GIT_COMMIT=$(GIT_COMMIT) \
        --build-arg VERSION=$(VERSION) \
		--tag $(SERVER)/$(OWNER)/$(IMG_NAME):$(TAG) \
		.

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
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/sns-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/rabbitmq-connector/values.yaml && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/headroom-controller/values.yaml

verify-chart:
	@echo Verifying helm chart images in remote registries && \
	arkade chart verify --verbose=$(VERBOSE) -f ./chart/openfaas/values.yaml

# Only upgrade the openfaas chart, for speed
upgrade-chart:
	@echo Upgrading openfaas helm chart images && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/openfaas/values.yaml

upgrade-charts:
	@echo Upgrading images for all helm charts && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/openfaas/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/kafka-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/cron-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/nats-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/mqtt-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/pro-builder/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/sqs-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/postgres-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/queue-worker/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/sns-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/rabbitmq-connector/values.yaml && \
	arkade chart upgrade --verbose=$(VERBOSE) -w -f ./chart/headroom-controller/values.yaml

bump-charts:
	arkade chart bump --file ./chart/openfaas/Chart.yaml -w && \
	arkade chart bump --file ./chart/kafka-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/cron-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/nats-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/mqtt-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/pro-builder/Chart.yaml -w && \
	arkade chart bump --file ./chart/sqs-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/postgres-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/queue-worker/Chart.yaml -w && \
	arkade chart bump --file ./chart/sns-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/rabbitmq-connector/Chart.yaml -w && \
	arkade chart bump --file ./chart/headroom-controller/Chart.yaml -w

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
		helm package sns-connector/ && \
		helm package rabbitmq-connector/ && \
		helm package headroom-controller/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml


start-kind: ## attempt to start a new dev environment
	@./contrib/create_dev.sh

stop-kind: ## attempt to stop the dev environment
	@./contrib/stop_dev.sh

.PHONY: verify-codegen
verify-codegen:
	./hack/verify-codegen.sh

.PHONY: update-codegen
update-codegen:
	./hack/update-codegen.sh
