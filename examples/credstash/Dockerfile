FROM alpine:3.5

ADD assets/smuggler-linux-amd64 /opt/resource/smuggler

RUN ln /opt/resource/smuggler /opt/resource/check && \
    ln /opt/resource/smuggler /opt/resource/in && \
    ln /opt/resource/smuggler /opt/resource/out

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

# Add the config file for this resource
ADD ./smuggler.yml /opt/resource/


