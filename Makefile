benchmark\:pack:
	go test -run=NONE -bench=Benchmark -benchmem -failfast ./pack

benchmark\:record:
	go test -run=NONE -bench=Benchmark -benchmem -failfast ./record

bootstrap:
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-update --all
	asdf install
	asdf reshim golang
	asdf reshim nodejs

codecheck:
	golint -set_exit_status ./... || exit 1
	go vet ./...

down:
	docker-compose -p appy -f .docker/docker-compose.yml down --remove-orphans

install:
	GO111MODULE=off go get -u golang.org/x/lint/golint github.com/gojp/goreportcard/cmd/goreportcard-cli github.com/golangci/golangci-lint/cmd/golangci-lint@v1.25.1
	go mod download

restart:
	docker-compose -p appy -f .docker/docker-compose.yml restart

test:
	mkdir -p tmp
	go test -covermode=atomic -coverprofile=tmp/coverage.out -race -failfast -v ./...

testcov:
	go tool cover -html=tmp/coverage.out

up:
	docker-compose -p appy -f .docker/docker-compose.yml up -d

.PHONY: benchmark\:pack benchmark\:record bootstrap codecheck down install restart test testcov up
