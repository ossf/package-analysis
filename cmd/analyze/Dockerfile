FROM golang@sha256:3c4de86eec9cbc619cdd72424abd88326ffcf5d813a8338a7743c55e5898734f as build
RUN apt-get update && apt-get install -y libpcap-dev
WORKDIR /src
COPY . ./
RUN go build -o analyze cmd/analyze/main.go && go build -o worker cmd/worker/main.go

FROM ubuntu:21.04@sha256:be154cc2b1211a9f98f4d708f4266650c9129784d0485d4507d9b0fa05d928b6

ENV DEBIAN_FRONTEND noninteractive
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
RUN curl -fsSL https://gvisor.dev/archive.key | apt-key add - && \
    add-apt-repository "deb https://storage.googleapis.com/gvisor/releases 20220425 main" && \
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