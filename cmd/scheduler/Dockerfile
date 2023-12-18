FROM golang:1.21.0@sha256:12359f96c43077803959dab88ce60c87cd2543f3b907bf54a34a5104270f404f as build
WORKDIR /src

# First cache the dependencies to avoid downloading again on code change
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o scheduler ./cmd/scheduler/main.go


FROM gcr.io/distroless/base:nonroot@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490

COPY --from=build /src/scheduler /usr/local/bin/scheduler

ENTRYPOINT ["/usr/local/bin/scheduler"]
