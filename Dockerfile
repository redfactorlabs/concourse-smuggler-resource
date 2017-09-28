#
# Example Dockerfile for smuggler concourse resource.
#
# Build concourse smuggler
FROM golang:1.8-alpine

RUN apk add -U git && rm -rf /var/cache/apk/*
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go get github.com/redfactorlabs/concourse-smuggler-resource

# Use your favorite base image
FROM alpine:3.6

# Add some stuff to your container
# Our base container will have some handy tooling
ARG INSTALLED_PACKAGES="\
    bash                \
    zip                 \
    curl                \
    wget                \
    openssl             \
    ca-certificates     \
    jq                  \
    git                 \
    openssh-client      \
"
RUN apk add --update ${INSTALLED_PACKAGES} \
    && rm -rf /var/cache/apk/*

# Add the smuggler binary compiled previously
COPY --from=0 /go/bin/concourse-smuggler-resource /opt/resource/smuggler

# Link it to the /opt/resource{check,in,out} commands
RUN ln /opt/resource/smuggler /opt/resource/check \
    && ln /opt/resource/smuggler /opt/resource/in \
    && ln /opt/resource/smuggler /opt/resource/out

# Add a example default configuration of the commands
ADD example-smuggler.yml /opt/resource/smuggler.yml
