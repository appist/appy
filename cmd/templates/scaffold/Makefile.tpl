clean:
	rm -rf dist coverage.out {{.projectName}} **/*.log

docker\:build:
	DOCKER_BUILDKIT=1 docker build --rm .

codecheck:
	go vet ./...
	golint -set_exit_status ./...

test:
	mkdir -p tmp/coverage/backend
	APPY_ENV=test go test -covermode=atomic -coverprofile=tmp/coverage/backend/cover.out -race -failfast -v ./...

.PHONY: clean codecheck docker\:build test
