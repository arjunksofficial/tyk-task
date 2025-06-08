# Docker image of sampleserver in alpine after build in multistage build
FROM golang:1.24.3 AS builder
WORKDIR /app
COPY . .
RUN ls -al /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -o sampleserver cmd/sampleserver/main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sampleserver /app/sampleserver
CMD ["/app/sampleserver"]
EXPOSE 8080