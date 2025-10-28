FROM golang:1.23.5-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /src/cmd/forum-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/forum-service

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/local/bin/forum-service /usr/local/bin/forum-service
COPY cmd/forum-service/forum.yaml /app/forum.yaml
WORKDIR /app
EXPOSE 8082
ENV GIN_MODE=release
ENTRYPOINT ["/usr/local/bin/forum-service"]
CMD ["tail", "-f", "/dev/null"]
