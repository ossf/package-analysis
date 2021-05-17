FROM golang:1.16.3 as build
WORKDIR /src
COPY analysis/ ./
RUN go build cmd/analyze.go && go build cmd/worker.go

FROM ubuntu:21.04@sha256:be154cc2b1211a9f98f4d708f4266650c9129784d0485d4507d9b0fa05d928b6

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        iptables \
        podman \
        software-properties-common && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# Install gVisor.
RUN curl -fsSL https://gvisor.dev/archive.key | apt-key add - && \
    add-apt-repository "deb https://storage.googleapis.com/gvisor/releases master main" && \
    apt-get update && apt-get install -y runsc

# runsc with our own builtin args.
COPY runsc /usr/local/bin

COPY --from=build /src/analyze /usr/local/bin/analyze
COPY --from=build /src/worker /usr/local/bin/worker
