FROM golang:1.23.5-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /src/cmd/user-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/user-service

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/local/bin /usr/local/bin

WORKDIR /app
COPY cmd/user-service/user.yaml /app/user.yaml
EXPOSE 8081
ENV GIN_MODE=release
ENTRYPOINT ["/usr/local/bin/user-service"]
CMD ["/usr/local/bin/user-service"]