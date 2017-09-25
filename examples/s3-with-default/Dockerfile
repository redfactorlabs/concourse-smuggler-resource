# Build concourse smuggler
FROM golang:1.8-alpine
RUN apk add -U git && rm -rf /var/cache/apk/*
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go get github.com/redfactorlabs/concourse-smuggler-resource

# Build the s3 resource
RUN go get -d github.com/concourse/s3-resource/ && \
    go get -d github.com/concourse/s3-resource/... && \
    cd /go/src/github.com/concourse/s3-resource && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        sh ./scripts/build

# Modify the upstream resource
FROM alpine:3.6

# Add the smuggler binary compiled previously
COPY --from=0 /go/bin/concourse-smuggler-resource /opt/resource/smuggler

# Get the s3 wrapped resource
RUN mkdir -p /opt/resource/wrapped/s3

COPY --from=0 \
    /go/src/github.com/concourse/s3-resource/assets/check \
    /opt/resource/wrapped/s3/check

COPY --from=0 \
    /go/src/github.com/concourse/s3-resource/assets/in \
    /opt/resource/wrapped/s3/in

COPY --from=0 \
    /go/src/github.com/concourse/s3-resource/assets/check \
    /opt/resource/wrapped/s3/out

# Link the /opt/resource{check,in,out} commands to smuggler
RUN ln /opt/resource/smuggler /opt/resource/check \
    && ln /opt/resource/smuggler /opt/resource/in \
    && ln /opt/resource/smuggler /opt/resource/out

# Add the config file for this resource
ADD ./smuggler.yml /opt/resource/

ENV PACKAGES "jq ca-certificates"

RUN apk add --update $PACKAGES && rm -rf /var/cache/apk/*
