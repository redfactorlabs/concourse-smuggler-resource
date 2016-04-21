FROM alpine:3.3

ENV PACKAGES "jq ca-certificates"

RUN apk add --update $PACKAGES && rm -rf /var/cache/apk/*

# Add the wrapped resource S3
RUN mkdir -p /opt/resource/wrapped/s3
ADD ./assets/s3/check /opt/resource/wrapped/s3/
ADD ./assets/s3/in /opt/resource/wrapped/s3/
ADD ./assets/s3/out /opt/resource/wrapped/s3/

# Add smuggler binary
ADD ./assets/smuggler-linux-amd64 /opt/resource/smuggler
RUN ln /opt/resource/smuggler /opt/resource/check && \
    ln /opt/resource/smuggler /opt/resource/in && \
    ln /opt/resource/smuggler /opt/resource/out

# Add the config file for this resource
ADD ./smuggler.yml /opt/resource/

