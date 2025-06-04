# API Gateway Proxy Boilerplate

## Overview
Basic Go proxy engine for API Gateway assignment. Forwards `/api/v1/*` requests to a backend, ready for middleware/routing.

## Project Structure

- `cmd/apigw/` — main app entry point
- `internal/` — for private app packages (to be created)
- `pkg/` — for reusable packages (to be created)

## Quickstart

```sh
# Set backend target
export BACKEND_URL=http://localhost:9000

# Build
cd cmd/apigw
go build -o apigw main.go

# Run
./apigw
```

## Env Vars
- `BACKEND_URL`: URL to forward requests to (required)

## Next Steps
- Add middleware for auth, rate limiting, etc.
- Implement Redis integration
- Add tests and Dockerfile
