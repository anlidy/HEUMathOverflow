FROM golang:1.23.5-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /src/cmd/audit-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/audit-service

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/local/bin/audit-service /usr/local/bin/audit-service
COPY cmd/audit-service/audit.yaml /app/audit.yaml
WORKDIR /app
EXPOSE 8083
ENV GIN_MODE=release
ENTRYPOINT ["/usr/local/bin/audit-service"]
