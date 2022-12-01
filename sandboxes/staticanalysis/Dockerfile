#FROM golang:1.19-alpine as build
FROM golang@sha256:122f3484f844467ebe0674cf57272e61981770eb0bc7d316d1f0be281a88229f as build
WORKDIR /src

# First cache the dependencies to avoid downloading again on code change
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY staticanalyze.go ./

# If CGO is disabled then we don't need glibc
RUN CGO_ENABLED=0 go build -o staticanalyze ./staticanalyze.go

FROM alpine:latest
RUN apk add --no-cache nodejs && \
	apk add --no-cache npm && \
	apk add --no-cache python3

COPY --from=build /src/staticanalyze /usr/local/bin/staticanalyze
RUN chmod 755 /usr/local/bin/staticanalyze
RUN mkdir -p /app

WORKDIR /app

ENTRYPOINT [ "sleep" ]
CMD [ "10m" ]
