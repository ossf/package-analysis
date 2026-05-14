FROM golang:1.26.3@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS build

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

FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
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
