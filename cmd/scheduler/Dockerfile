FROM golang:1.26.3@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS build
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
