.PHONY: local build build-amd64 build-armhf build-arm64 build-all push-amd64 push-armhf push-arm64 push-manifest-list push-all namespaces install install-armhf charts
TAG?=latest

all: build

# Maintain build as alias target for build-amd64 target for backwards compatibility.
build: build-amd64

local:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes

build-amd64:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-netes:$(TAG)-amd64 .

build-armhf:
	docker build -t openfaas/faas-netes:$(TAG)-armhf . -f Dockerfile.armhf

build-arm64:
	docker build -t openfaas/faas-netes:$(TAG)-arm64 . -f Dockerfile.arm64

push-amd64:
	docker push openfaas/faas-netes:$(TAG)-amd64

push-armhf:
	docker push openfaas/faas-netes:$(TAG)-armhf

push-arm64:
	docker push openfaas/faas-netes:$(TAG)-arm64

push-manifest-list:
	docker manifest create openfaas/faas-netes:$(TAG) openfaas/faas-netes:$(TAG)-amd64 openfaas/faas-netes:$(TAG)-armhf openfaas/faas-netes:$(TAG)-arm64
	docker manifest push openfaas/faas-netes:$(TAG)

build-all: build-arm64 build-armhf build-amd64

push-all: push-amd64 push-armhf push-arm64 push-manifest-list

namespaces:
	kubectl apply -f namespaces.yml

install: namespaces
	kubectl apply -f yaml/

install-armhf: namespaces
	kubectl apply -f yaml_armhf/

charts:
	cd chart && helm package openfaas/ && helm package kafka-connector/ && helm package cron-connector/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml
	./contrib/create-static-manifest.sh
	./contrib/create-static-manifest.sh ./chart/openfaas ./yaml_arm64 ./chart/openfaas/values-arm64.yaml
	./contrib/create-static-manifest.sh ./chart/openfaas ./yaml_armhf ./chart/openfaas/values-armhf.yaml

start-kind: ## attempt to start a new dev environment
	@./contrib/create_dev.sh \
		&& echo "" \
		&& printf '%-15s:\t %s\n' 'Web UI' 'http://localhost:31112/ui' \
		&& printf '%-15s:\t %s\n' 'User' 'admin' \
		&& printf '%-15s:\t %s\n' 'Password' $(shell cat password.txt) \
		&& printf '%-15s:\t %s\n' 'CLI Login' "faas-cli login --username admin --password $(shell cat password.txt)"

stop-kind: ## attempt to stop the dev environment
	@./contrib/stop_dev.sh
