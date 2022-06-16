SMUGGLER_DOCKER_TAG:=alpine3.16
SMUGGLER_DOCKER_REPOSITORY:=redfactorlabs/concourse-smuggler-resource
SMUGGLER_DOCKER_IMAGE:=$(SMUGGLER_DOCKER_REPOSITORY):$(SMUGGLER_DOCKER_TAG)

SMUGGLER_GIT_URL:=https://github.com/redfactorlabs/concourse-smuggler-resource
SMUGGLER_GIT_BRANCH:=master

GO_PACKAGES = $(shell go list ./... | grep -v vendor)
GO_FILES = $(shell find . -name "*.go" | grep -v vendor | uniq)

test:
	gofmt -s -l -w $(GO_FILES)
	go vet $(GO_PACKAGES)
	go test $(GO_PACKAGES) -v

build: test assets/smuggler-darwin-amd64 assets/smuggler-linux-amd64

assets/smuggler-darwin-amd64: $(GO_FILES)
	mkdir -p assets
	CGO_ENABLED=1 \
	GOOS=darwin \
	GOARCH=amd64 \
		go build -o $@ .

assets/smuggler-linux-amd64: $(GO_FILES)
	mkdir -p assets
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
		go build -o $@ .

build-docker:
	docker build --no-cache \
		--build-arg SMUGGLER_GIT_URL \
		--build-arg SMUGGLER_GIT_BRANCH \
		-t "${SMUGGLER_DOCKER_IMAGE}" .

push-docker: build-docker
	docker push "${SMUGGLER_DOCKER_IMAGE}"

install-deps:
	go get -u github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
	go get -u github.com/onsi/gomega         # fetches the matcher library
