#!/bin/sh

DOCKER_HUB_ACCOUNT=redfactorlabs

all: build

build:
	docker build . -f Dockerfile.git-resource -t $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-git-resource
	docker build . -f Dockerfile.s3-resource -t $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-s3-resource

push:
	docker push $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-git-resource
	docker push $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-s3-resource
