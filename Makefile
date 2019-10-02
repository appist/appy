bootstrap:
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-update --all
	asdf install
	asdf reshim golang
	asdf reshim nodejs

codecheck:
	go vet ./...
	golint -set_exit_status ./...

down:
	docker-compose down --remove-orphans

install:
	go get -u golang.org/x/lint/golint
	go mod download

test:
	go test -covermode=atomic -coverprofile=coverage.out -race ./...

testcov:
	go tool cover -html=coverage.out

up:
	docker-compose up -d

.PHONY: bootstrap codecheck install test
.SILENT: bootstrap codecheck install test
