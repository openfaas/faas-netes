TAG?=latest

local-fmt:
	gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")
local-go:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes
local: 	local-fmt 	local-go

build:
	docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t alexellis2/faas-netes:$(TAG) .

push:
	docker push alexellis2/faas-netes:$(TAG)
