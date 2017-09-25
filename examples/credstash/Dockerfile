# Build concourse smuggler
FROM golang:1.8-alpine
RUN apk add -U git && rm -rf /var/cache/apk/*
RUN mkdir -p $GOPATH/src/github.com/redfactorlabs && \
    cd $GOPATH/src/github.com/redfactorlabs && \
    git clone https://github.com/redfactorlabs/concourse-smuggler-resource && \
    cd concourse-smuggler-resource && \
    git checkout reorg_project
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go get github.com/redfactorlabs/concourse-smuggler-resource

# Use your favorite base image
FROM alpine:3.6

# Add the smuggler binary compiled previously
COPY --from=0 /go/bin/concourse-smuggler-resource /opt/resource/smuggler

# Link the /opt/resource{check,in,out} commands to smuggler
RUN ln /opt/resource/smuggler /opt/resource/check \
    && ln /opt/resource/smuggler /opt/resource/in \
    && ln /opt/resource/smuggler /opt/resource/out

# Add the config file for this resource
ADD ./smuggler.yml /opt/resource/

# Install all the dependencies
RUN apk -U add --no-cache \
        python \
        py-pip \
        openssl \
        jq \
    && apk -U add --no-cache -t credstash-build-deps \
        python-dev \
        libffi-dev \
        build-base \
        openssl-dev \
    && pip install credstash  \
    && apk del credstash-build-deps  \
    && rm -rf /var/cache/apk/*
