
installmockery:
	go install github.com/vektra/mockery/v3@v3.3.2

generatemocks: installmockery
	mockery

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total