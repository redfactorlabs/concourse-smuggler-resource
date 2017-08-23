#!/bin/sh

DOCKER_HUB_ACCOUNT=redfactorlabs

all: build push

clean:
	rm -rf assets

assets/smuggler:
	cd ../../ \
	  && ./scripts/build
	cp ../../assets/smuggler-linux-amd64 ./assets

get_assets: assets/smuggler

build: get_assets
	docker build .  -t $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-resource

push:
	docker push $(DOCKER_HUB_ACCOUNT)/smuggler-credstash-resource

