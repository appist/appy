benchmark\:diff:
	cob --base origin/master --threshold 0.2 --bench-args "test -run=NONE -bench . -benchmem -benchtime 1s -failfast ./..."

benchmark\:pack:
	go test -run=NONE -bench . -benchmem -benchtime 10s -failfast ./pack

benchmark\:record:
	go test -run=NONE -bench . -benchmem -benchtime 1s -failfast ./record

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
	GO111MODULE=off go get -u golang.org/x/lint/golint
	(which cob >/dev/null) || (curl -sfL https://raw.githubusercontent.com/knqyf263/cob/master/install.sh | sudo sh -s -- -b /usr/local/bin)
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
