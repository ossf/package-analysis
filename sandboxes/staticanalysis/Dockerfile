FROM golang:1.23.1@sha256:4a3c2bcd243d3dbb7b15237eecb0792db3614900037998c2cd6a579c46888c1e as build

# Note: Dockerfile uses paths relative to the top-level project directory,
# so it should be built from that directory, i.e:
# $ cd package-analysis
# $ docker build -f sandboxes/staticanalysis/Dockerfile .

# Cache dependencies to avoid downloading again on code change
WORKDIR /src
# Dependencies for package analysis
COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . ./

WORKDIR /src/sandboxes/staticanalysis
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
COPY --from=build /src/internal/staticanalysis/parsing/package.json /src/internal/staticanalysis/parsing/package-lock.json /npm_deps/

# cache NPM installs in /npm_cache so that static analysis binary can use them
RUN mkdir -m 755 /npm_cache && \
	npm ci --prefix /npm_deps --cache /npm_cache

WORKDIR /app

ENTRYPOINT [ "sleep" ]
CMD [ "30m" ]
