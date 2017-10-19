# Build concourse smuggler
FROM golang:1.9-alpine

RUN apk add -U git && rm -rf /var/cache/apk/*
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go get github.com/redfactorlabs/concourse-smuggler-resource

# Use your favorite base image
FROM alpine:3.6

ENV PACKAGES "openssh-client jq ca-certificates"
RUN apk add --update $PACKAGES && rm -rf /var/cache/apk/*

COPY --from=0 /go/bin/concourse-smuggler-resource /opt/resource/smuggler

RUN ln /opt/resource/smuggler /opt/resource/check \
    && ln /opt/resource/smuggler /opt/resource/in \
    && ln /opt/resource/smuggler /opt/resource/out
