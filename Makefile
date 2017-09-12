TAG?=latest

build:
	docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t alexellis2/faas-netes:$(TAG) .

push:
	docker push alexellis2/faas-netes:$(TAG)

