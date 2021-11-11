FROM golang@sha256:3c4de86eec9cbc619cdd72424abd88326ffcf5d813a8338a7743c55e5898734f as build
WORKDIR /src
COPY . ./
RUN go build -v -o server ./cmd/server/main.go

FROM debian:buster-slim@sha256:f077cd32bfea6c4fa8ddeea05c53b27e90c7fad097e2011c9f5f11a8668f8db4
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /src/server /app/server
CMD ["/app/server"]
