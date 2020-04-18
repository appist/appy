clean:
	rm -rf dist coverage.out mercury **/*.log

docker\:build:
	DOCKER_BUILDKIT=1 docker build -f .docker/Dockerfile --rm .

format:
	go vet ./...
	golint -set_exit_status ./...

i install:
	go mod download

test:
	mkdir -p tmp/coverage/backend
	APPY_ENV=test go test -covermode=atomic -coverprofile=tmp/coverage/backend/cover.out -race -failfast -v ./...

.PHONY: clean docker\:build format i install test
