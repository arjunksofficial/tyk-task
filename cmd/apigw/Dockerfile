# Docker image of apigw in alpine after build in multistage build
FROM golang:1.24.3 AS builder
WORKDIR /app
COPY . .
RUN ls -al /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -o apigw cmd/apigw/main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/apigw /app/apigw
COPY --from=builder /app/cmd/apigw/config /app/config
CMD ["/app/apigw"]
EXPOSE 8080