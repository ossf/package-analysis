FROM golang:1.26.3@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS build
RUN apt-get update && apt-get install -y libpcap-dev
WORKDIR /src

# First cache the dependencies to avoid downloading again on code change
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY . ./
RUN go build -o analyze ./cmd/analyze && go build -o worker ./cmd/worker

FROM ubuntu:22.04@sha256:962f6cadeae0ea6284001009daa4cc9a8c37e75d1f5191cf0eb83fe565b63dd7

ENV DEBIAN_FRONTEND="noninteractive"
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        iptables \
        iproute2 \
        podman \
        software-properties-common && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# Install gVisor.
RUN mkdir -m 0700 -p /etc/apt/keyrings && \
    curl -fsSL https://gvisor.dev/archive.key -o /etc/apt/keyrings/gvisor.key && \
    echo "deb [signed-by=/etc/apt/keyrings/gvisor.key] https://storage.googleapis.com/gvisor/releases 20260216 main" > /etc/apt/sources.list.d/gvisor.list && \
    apt-get update && apt-get install -y runsc

COPY --from=build /src/analyze /usr/local/bin/analyze
COPY --from=build /src/worker /usr/local/bin/worker
COPY --from=build /src/tools/gvisor/runsc_compat.sh /usr/local/bin/runsc_compat.sh
COPY --from=build /src/tools/network/iptables.rules /usr/local/etc/iptables.rules
COPY --from=build /src/tools/network/podman-analysis.conflist /etc/cni/net.d/podman-analysis.conflist
RUN chmod 755 /usr/local/bin/runsc_compat.sh && \
    chmod 644 /usr/local/etc/iptables.rules /etc/cni/net.d/podman-analysis.conflist

ARG SANDBOX_IMAGE_TAG
ENV OSSF_SANDBOX_IMAGE_TAG=${SANDBOX_IMAGE_TAG}
