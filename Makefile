benchmark\:diff:
	cob --base origin/master --threshold 0.1 --bench-args "test -run=NONE -bench . -benchmem -benchtime 10s -failfast ./..."

benchmark\:pack:
	go test -run=NONE -bench . -benchmem -benchtime 10s -failfast ./pack

benchmark\:record:
	go test -run=NONE -bench . -benchmem -benchtime 10s -failfast ./record

bootstrap:
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-update --all
	asdf install
	asdf reshim golang
	asdf reshim nodejs
	brew install vektra/tap/mockery && brew upgrade mockery

codecheck:
	golint -set_exit_status ./... || exit 1
	go vet ./...

down:
	docker-compose -p appy -f .docker/docker-compose.yml down --remove-orphans

install:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	(which cob >/dev/null) || (curl -sfL https://raw.githubusercontent.com/knqyf263/cob/master/install.sh | sudo sh -s -- -b /usr/local/bin)
	go mod download

genmock:
	mockery -name DBer -structname DB -filename db.go -dir ./record
	mockery -name Modeler -structname Model -filename model.go -dir ./record
	mockery -name Stmter -structname Stmt -filename stmt.go -dir ./record
	mockery -name Txer -structname Tx -filename tx.go -dir ./record

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
