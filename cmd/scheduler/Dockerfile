FROM golang:1.23.1@sha256:4a3c2bcd243d3dbb7b15237eecb0792db3614900037998c2cd6a579c46888c1e AS build
WORKDIR /src

# First cache the dependencies to avoid downloading again on code change
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o scheduler ./cmd/scheduler/main.go


FROM gcr.io/distroless/base:nonroot@sha256:746b9dbe3065a124395d4a7698241dbd6f3febbf01b73e48f942aabd7b8e5eac

COPY --from=build /src/scheduler /usr/local/bin/scheduler

ENTRYPOINT ["/usr/local/bin/scheduler"]
