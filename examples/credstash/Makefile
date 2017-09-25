#!/bin/sh

DOCKER_HUB_ACCOUNT=redfactorlabs

all: build push

build:
	docker build .  -t $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-resource

push:
	docker push $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-resource
