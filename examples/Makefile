
s3-resource:
	# Build for linux-amd64 docker containers on alpine
	export CGO_ENABLED=1
	export GOOS=linux
	export GOARCH=amd64
	echo "Downloading and building github.com/concourse/s3-resource..."
	go get -d github.com/concourse/s3-resource
	go get -d github.com/concourse/s3-resource/...
	# I need to build in a container to allow cross compiling without
	# the error: ld: unknown option: --build-id=none
	docker run --rm \
		-v "$$GOPATH":/go \
		-w /go/src/github.com/concourse/s3-resource \
		-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=1 \
		golang:1.6-alpine sh -e -c 'sh ./scripts/build'

