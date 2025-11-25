# syntax directive enables BuildKit features like RUN --mount
# usage example (from repo root):
# DOCKER_BUILDKIT=1 docker build \
#   --build-arg SERVICE=forum \
#   --build-arg EXPOSE_PORT=8082 \
#   -f backend/docker/Dockerfile -t forum-service:latest .

# syntax=docker/dockerfile:1.4
FROM docker.io/library/golang:1.24.9 AS development
ARG SERVICE=user
ARG EXPOSE_PORT=8081

RUN go install github.com/air-verse/air@v1.62.0

WORKDIR /src
COPY . .
EXPOSE ${EXPOSE_PORT}

WORKDIR /src/cmd/${SERVICE}-service
CMD ["air", "-c", "./.air.toml"]
# CMD ["tail", "-f", "/dev/null"]

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
    cd /src/cmd/${SERVICE}-service && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /go-bin/${SERVICE}-service && cp /go-bin/${SERVICE}-service /bin/myservice

FROM alpine:3.18 AS production
ARG SERVICE=user
ARG EXPOSE_PORT=8081
RUN apk add --no-cache ca-certificates

# Copy the built binary (we always write it as /go-bin/service in builder)
COPY --from=builder /bin/myservice /bin/myservice
WORKDIR /app

# Note: Dockerfile cannot conditionally COPY based on ARG at build time.
# If you don't want to copy a config, omit --build-arg CONFIG_PATH or make sure
# the ARG points to a valid path inside the build context. If CONFIG_PATH is
# empty the following COPY will fail.
COPY ./cmd/${SERVICE}-service/${SERVICE}.yaml /app/
EXPOSE ${EXPOSE_PORT}
ENV GIN_MODE=release

ENTRYPOINT ["/bin/myservice"]