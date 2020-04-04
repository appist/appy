benchmarkhttp:
	go test -run=NONE -bench=Benchmark -tags=benchmarkhttp -failfast .

bootstrap:
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-update --all
	asdf install
	asdf reshim golang
	asdf reshim nodejs

codecheck:
	golint -set_exit_status . || exit 1
	go vet .

down:
	docker-compose -p appy -f .docker/docker-compose.yml down --remove-orphans

install:
	go get -u golang.org/x/lint/golint
	go mod download

restart:
	docker-compose -p appy -f .docker/docker-compose.yml restart

test:
	mkdir -p tmp
	go test -covermode=atomic -coverprofile=tmp/coverage.out -tags=test -race -failfast -v .

testcov:
	go tool cover -html=tmp/coverage.out

up:
	docker-compose -p appy -f .docker/docker-compose.yml up -d

.PHONY: benchmarkhttp bootstrap codecheck down install restart test testcov up
