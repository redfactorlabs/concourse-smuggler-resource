
SMUGGLER_DOCKER_TAG:=ubuntu-14.04
SMUGGLER_DOCKER_REPOSITORY:=redfactorlabs/concourse-smuggler-resource
SMUGGLER_DOCKER_IMAGE:=$(SMUGGLER_DOCKER_REPOSITORY):$(SMUGGLER_DOCKER_TAG)

test:
	test -z "$$(go fmt ./...)"
	godep go test ./... -v

build: test assets/smuggler-darwin-amd64 assets/smuggler-linux-amd64

assets/smuggler-darwin-amd64:
	mkdir -p assets
	CGO_ENABLED=1 \
	GOOS=darwin \
	GOARCH=amd64 \
		godep go build -o $@ .

assets/smuggler-linux-amd64:
	mkdir -p assets
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
		godep go build -o $@ .

build-docker: test
	docker build -t "${SMUGGLER_DOCKER_IMAGE}" .

push-docker: build-docker
	docker push "${SMUGGLER_DOCKER_IMAGE}"

install-deps:
	go get -u github.com/tools/godep
	go get -u github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
	go get -u github.com/onsi/gomega         # fetches the matcher library
