clean:
	rm -rf dist coverage.out {{.Project.Name}} **/*.log

format:
	go vet ./...
	golint -set_exit_status ./...

i install:
	go mod download

test:
	mkdir -p tmp/coverage/backend
	APPY_ENV=test go test -covermode=atomic -coverprofile=tmp/coverage/backend/cover.out -race -failfast -v ./...

.PHONY: clean check i install test
