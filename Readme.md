# API Gateway Proxy

This is a simple API Gateway Proxy that can be used to forward requests to different services based on the request path.

## Features

- Request logging
- Authentication middleware
- Rate limiting
- Redis integration for token storage
- Support for multiple services

## Requirements

- Go 1.18 or later
- Redis server
- Docker (for running Redis)

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/arjunksofficial/tyk-task.git
cd tyk-task
```

2. Edit the file `build/docker/config/master.yaml` to add your services. The file will contain the following by default

```
app:
  port: 9000
  name: "API Gateway"
routes:
  - path: /api/v1/orders/
    host: http://host.docker.internal:8000
  - path: /api/v1/users/
    host: http://host.docker.internal:8001
redis:
  host: redis
  port: 6379
  db: 0
  password: ""
```

You can modify the `routes` section to add or change the services you want to proxy. The `host` should point to the service URL that the gateway will forward requests to.

3. To start gateway, redis and sample server with /api/v1/orders and /api/v1/users endpoints, run the following commands:

```
cd build/docker
docker-compose up
```

4. The API Gateway will be running on `http://localhost:9000`. You can test the endpoints using curl or any API client.
5. To test the endpoints, you can use the following commands:

- Create a token using the command below. This will return a token that can be used to authenticate requests.

```bash
cd cmd/tokengen
go run main.go
```

The above command will generate a token and store it in Redis. The token will be printed in the console.

The token will be generated based on the config in `cmd/tokengen/tokendata.yaml`. You can modify the config file to change the token generation logic.

By default, the tokendata.yaml file contains the following:

```yaml
rate_limit: 5 # 5 requests per minute
allowed_routes:
  - /api/v1/users/*
  - /api/v1/products/*
  - /api/v1/orders/*
duration: 3600 # 1 hour in seconds
```

You can modify the `rate_limit`, `allowed_routes`, and `duration` fields as per your requirements.

- Use the token to make requests to the API Gateway:

```
curl -X GET http://localhost:9000/api/v1/orders/list -H "Authorization: Bearer <your_token_here>"
curl -X GET http://localhost:9000/api/v1/users/list -H "Authorization: Bearer <your_token_here>"
```

This api gateway is using fixed window rate limiting. The rate limit is configured in the `tokendata.yaml` file.

## Unit Tests

To run the unit tests, you can use the following command:

```bash
make test-coverage
```

This will run the tests and generate a coverage report. The coverage report will be saved in the `coverage.out` file.

You can see the coverage report by running the following command:

```bash
go tool cover -html=coverage.out
```

## Health Check

The API Gateway provides a health check endpoint to verify if the service is running correctly. You can access it at:

```
http://localhost:9000/health
```

## Readiness Check

The API Gateway also provides a readiness check endpoint to verify if the service is ready to accept requests. You can access it at:

```
http://localhost:9000/ready
```

## To access metrics

The API Gateway exposes metrics that can be accessed at the following endpoint:

```

http://localhost:9000/metrics

```

```

```
