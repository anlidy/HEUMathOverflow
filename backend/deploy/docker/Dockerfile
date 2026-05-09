# development stage with air for live reloading
FROM docker.io/library/golang:1.24.9 AS development
ARG SERVICE=user
ARG EXPOSE_PORT=8081

RUN go install github.com/air-verse/air@v1.62.0

WORKDIR /src
COPY . .
EXPOSE ${EXPOSE_PORT}

WORKDIR /src/services/${SERVICE}
CMD ["air", "-c", "./.air.toml"]
# CMD ["tail", "-f", "/dev/null"]

# two level for production, alpine only
FROM docker.io/library/golang:1.24.9 AS builder
ARG SERVICE=user

WORKDIR /src

# Use BuildKit mounts to access host caches and source during build.
# Note: when using `--mount=type=bind,source=...`, the `source` path is
# resolved on the builder host (relative to the build context or absolute).
RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod,rw \
    --mount=type=bind,source=./go.mod,target=/src/go.mod \
    --mount=type=bind,source=./go.sum,target=/src/go.sum \
    cd /src && go mod download

RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod,rw \
    --mount=type=bind,source=.,target=/src \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build,rw \
    --mount=type=cache,id=go-bin,target=/go-bin,rw \
    cd /src/services/${SERVICE} && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /go-bin/${SERVICE}-service && cp /go-bin/${SERVICE}-service /bin/myservice

FROM alpine:3.18 AS production
ARG SERVICE=user
ARG EXPOSE_PORT=8081
RUN apk add --no-cache ca-certificates

# Copy the built binary (we always write it as /go-bin/service in builder)
COPY --from=builder /bin/myservice /bin/myservice
WORKDIR /app

# Copy the service config file
COPY ./services/${SERVICE}/config/${SERVICE}.yaml /app/config/${SERVICE}.yaml
EXPOSE ${EXPOSE_PORT}
ENV GIN_MODE=release

ENTRYPOINT ["/bin/myservice"]
