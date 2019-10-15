bootstrap:
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-update --all
	asdf install
	asdf reshim golang
	asdf reshim nodejs

clean:
	rm -rf assets coverage.out wewatch **/*.log

codecheck:
	go vet ./...
	golint -set_exit_status ./...
	npm run lint

install:
	go get -u golang.org/x/lint/golint
	go mod download
	npm i

test:
	go test -covermode=atomic -coverprofile=coverage.out -race -failfast ./...

testcov:
	go tool cover -html=coverage.out

.PHONY: bootstrap clean codecheck install test testcov
