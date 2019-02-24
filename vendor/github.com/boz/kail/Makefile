
DOCKER_IMAGE ?= kail
DOCKER_REPO  ?= abozanich/$(DOCKER_IMAGE)
DOCKER_TAG   ?= latest

BUILD_ENV = GOOS=linux GOARCH=amd64

ifdef TRAVIS
	LDFLAGS += -X main.version=$(TRAVIS_BRANCH) -X main.commit=$(TRAVIS_COMMIT)
endif

build:
	govendor build -i +program

build-linux:
	$(BUILD_ENV) go build --ldflags '$(LDFLAGS)'  -o kail-linux ./cmd/kail

test:
	govendor test +local

test-full: build image
	govendor test -v -race +local

image: build-linux
	docker build -t $(DOCKER_IMAGE) .

image-minikube: build-linux
	eval $$(minikube docker-env) && docker build -t $(DOCKER_IMAGE) .

image-push: image
	docker tag $(DOCKER_IMAGE) $(DOCKER_REPO):$(DOCKER_TAG)
	docker push $(DOCKER_REPO):$(DOCKER_TAG)

install-libs:
	govendor install +vendor,^program

install-deps:
	go get -u github.com/kardianos/govendor
	govendor sync

release:
	GITHUB_TOKEN=$$GITHUB_REPO_TOKEN goreleaser -f .goreleaser.yml

clean:
	rm kail kail-linux dist 2>/dev/null || true

.PHONY: build build-linux \
	test test-full \
	image image-minikube image-push \
	install-libs install-deps \
	clean
