#
# Example Dockerfile for smuggler concourse resource.
#
# You only need to:
#  1. Pick your favourite base image
FROM ubuntu:14.04

#  2. with your favourite tools
RUN apt-get update && \
    apt-get install -y ssh-client && \
    rm -rf /var/lib/apt/lists/*

#  3. add smuggler command in /opt/resource/
ADD assets/smuggler-linux-amd64 /opt/resource/smuggler

#  4. and link that command to  /opt/resource/{check,in,out}
RUN ln /opt/resource/smuggler /opt/resource/check && \
    ln /opt/resource/smuggler /opt/resource/in && \
    ln /opt/resource/smuggler /opt/resource/out


