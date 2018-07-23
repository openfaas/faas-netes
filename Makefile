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

.PHONY: charts
charts:
	cd chart && helm package openfaas/
	mv chart/*.tgz docs/
	helm repo index docs --url https://openfaas.github.io/faas-netes/ --merge ./docs/index.yaml

