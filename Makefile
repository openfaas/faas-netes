.PHONY: build local build-arm64 build-armhf push namespaces install install-armhf charts ci-armhf-build ci-armhf-push ci-arm64-build ci-arm64-push
TAG?=latest

all: build

local:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes

build-arm64:
	docker build -t openfaas/faas-netes:$(TAG)-arm64 . -f Dockerfile.arm64

build-armhf:
	docker build -t openfaas/faas-netes:$(TAG)-armhf . -f Dockerfile.armhf

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-netes:$(TAG) .

push:
	docker push alexellis2/faas-netes:$(TAG)

namespaces:
	kubectl apply -f namespaces.yml

install: namespaces
	kubectl apply -f yaml/

install-armhf: namespaces
	kubectl apply -f yaml_armhf/

charts:
	cd chart && helm package openfaas/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml

ci-armhf-build:
	docker build -t openfaas/faas-netes:$(TAG)-armhf . -f Dockerfile.armhf

ci-armhf-push:
	docker push openfaas/faas-netes:$(TAG)-armhf

ci-arm64-build:
	docker build -t openfaas/faas-netes:$(TAG)-arm64 . -f Dockerfile.arm64

ci-arm64-push:
	docker push openfaas/faas-netes:$(TAG)-arm64
