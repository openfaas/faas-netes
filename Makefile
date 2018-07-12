TAG?=latest
SQUASH?=false

all: build

local:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes

build-arm64:
	docker build -t openfaas/faas-netes:$(TAG)-arm64 . -f Dockerfile.arm64

build-armhf:
	docker build -t openfaas/faas-netes:$(TAG)-armhf . -f Dockerfile.armhf

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-netes:$(TAG) . --squash=${SQUASH}

push:
	docker push alexellis2/faas-netes:$(TAG)

install:
	kubectl apply -f faas.yml,monitoring.yml,rbac.yml

install-async:
	kubectl apply -f faas.async.yml,monitoring.yml,rbac.yml,nats.yml

install-armhf:
	kubectl apply -f faas.armhf.yml,monitoring.armhf.yml,rbac.yml

.PHONY: charts
charts:
	cd chart && helm package openfaas/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml

