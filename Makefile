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

install:
	go get -u golang.org/x/lint/golint
	go mod download

test:
	go test -covermode=atomic -coverprofile=coverage.out -race ./...

test-cov:
	go tool cover -html=coverage.out

.PHONY: bootstrap codecheck install test
.SILENT: bootstrap codecheck install test
