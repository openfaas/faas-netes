.PHONY: build local push namespaces install charts start-kind stop-kind
TAG?=latest

all: build

local:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-netes:$(TAG) .

push:
	docker push openfaas/faas-netes:$(TAG)

namespaces:
	kubectl apply -f namespaces.yml

install: namespaces
	kubectl apply -f yaml/

charts:
	cd chart && helm package openfaas/ && helm package kafka-connector/ && helm package cron-connector/ && helm package nats-connector/ && helm package mqtt-connector/
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
