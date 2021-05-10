FROM golang:1.16.3 as build
WORKDIR /src
COPY analysis/* ./
RUN go mod download
RUN go build -o analyze

FROM ubuntu:20.04@sha256:bc2f7250f69267c9c6b66d7b6a81a54d3878bb85f1ebb5f951c896d13e6ba537

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        software-properties-common

# Install Docker.
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    add-apt-repository \
        "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable" && \
    apt-get install -y docker-ce

# Install gVisor.
RUN curl -fsSL https://gvisor.dev/archive.key | apt-key add - && \
    add-apt-repository "deb https://storage.googleapis.com/gvisor/releases master main" && \
    apt-get update && apt-get install -y runsc

# Make gVisor the default Docker runtime.
COPY daemon.json /etc/docker/daemon.json

COPY --from=build /src/analyze /usr/local/bin/analyze
