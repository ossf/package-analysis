FROM golang:1.19-alpine@sha256:122f3484f844467ebe0674cf57272e61981770eb0bc7d316d1f0be281a88229f as build

# Note: Dockerfile uses paths relative to the top-level project directory,
# so it should be built from that directory, i.e:
# $ cd package-analysis
# $ docker build -f sandboxes/staticanalysis/Dockerfile .

# Cache dependencies to avoid downloading again on code change
WORKDIR /src/sandboxes/staticanalysis
# Dependencies for package analysis
COPY ./go.mod ./go.sum /src/
# Dependencies for sandbox binary
COPY ./sandboxes/staticanalysis/go.mod ./sandboxes/staticanalysis/go.sum ./

RUN go mod download

COPY . /src/

# If CGO is disabled then we don't need glibc
RUN CGO_ENABLED=0 go build -o staticanalyze staticanalyze.go

FROM alpine:3.17.1@sha256:93d5a28ff72d288d69b5997b8ba47396d2cbb62a72b5d87cd3351094b5d578a0
RUN apk add --no-cache file && \
	apk add --no-cache nodejs && \
	apk add --no-cache npm && \
	apk add --no-cache python3

COPY --from=build /src/sandboxes/staticanalysis/staticanalyze /usr/local/bin/staticanalyze
RUN chmod 755 /usr/local/bin/staticanalyze

RUN mkdir /npm_deps
COPY --from=build /src/internal/staticanalysis/parsing/package.json /npm_deps
COPY --from=build /src/internal/staticanalysis/parsing/package-lock.json /npm_deps

# cache NPM installs in /npm_cache so that static analysis binary can use them
RUN mkdir -m 755 /npm_cache && \
	npm ci --prefix /npm_deps --cache /npm_cache

WORKDIR /app

ENTRYPOINT [ "sleep" ]
CMD [ "30m" ]
