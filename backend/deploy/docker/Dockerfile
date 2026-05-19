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

FROM alpine:3.18 AS production
ARG SERVICE=user
ARG EXPOSE_PORT=8081
RUN apk add --no-cache ca-certificates

# Copy the prebuilt host binary.
COPY ./services/${SERVICE}/bin/${SERVICE}-service /bin/myservice
WORKDIR /app

# Copy the service config file
COPY ./services/${SERVICE}/config/${SERVICE}.yaml /app/config/${SERVICE}.yaml
EXPOSE ${EXPOSE_PORT}
ENV GIN_MODE=release

ENTRYPOINT ["/bin/myservice"]
