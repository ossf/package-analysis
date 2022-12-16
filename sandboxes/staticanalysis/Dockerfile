#FROM golang:1.19-alpine as build
FROM golang@sha256:122f3484f844467ebe0674cf57272e61981770eb0bc7d316d1f0be281a88229f as build

WORKDIR /src
COPY . ./

WORKDIR /src/sandboxes/staticanalysis
RUN go mod download
# If CGO is disabled then we don't need glibc
RUN CGO_ENABLED=0 go build -o staticanalyze staticanalyze.go

FROM alpine:latest
RUN apk add --no-cache nodejs && \
	apk add --no-cache npm && \
	apk add --no-cache python3

COPY --from=build /src/sandboxes/staticanalysis/staticanalyze /usr/local/bin/staticanalyze
RUN chmod 755 /usr/local/bin/staticanalyze

RUN mkdir /npm_deps
COPY --from=build /src/internal/staticanalysis/parsing/js/package.json /npm_deps
COPY --from=build /src/internal/staticanalysis/parsing/js/package-lock.json /npm_deps

# cache NPM installs in /npm_cache so that static analysis binary can use them
RUN mkdir /npm_cache && \
	chmod 755 /npm_cache && \
	npm ci --prefix /npm_deps --cache /npm_cache


WORKDIR /app

ENTRYPOINT [ "sleep" ]
CMD [ "30m" ]
