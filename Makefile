TAG?=latest
SQUASH?=false

local-fmt:
	gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")

local-go:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes

local: 	local-fmt 	local-go

build-arm64:
	docker build -t functions/faas-netesd:$(TAG)-arm64 . -f Dockerfile.arm64

build-armhf:
	docker build -t functions/faas-netesd:$(TAG)-armhf . -f Dockerfile.armhf

build-legacy:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}"  -t functions/faas-netesd:$(TAG) . -f Dockerfile.non-multi --squash=${SQUASH}

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-netesd:$(TAG) . --squash=${SQUASH}

push:
	docker push alexellis2/faas-netes:$(TAG)

install:
	kubectl apply -f faas.yml,monitoring.yml,rbac.yml

install-async:
	kubectl apply -f faas.async.yml,monitoring.yml,rbac.yml,nats.yml

install-armhf:
	kubectl apply -f faas.armhf.yml,monitoring.armhf.yml,rbac.yml
