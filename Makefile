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
	docker-compose -f .docker/docker-compose.yml down --remove-orphans

install:
	go get -u golang.org/x/lint/golint
	go mod download
	cd tools && npm i

test:
	go test -covermode=atomic -coverprofile=coverage.out -race -failfast ./...

testcov:
	go tool cover -html=coverage.out

tools:
	cd tools && npm run build
	go run ./generator/tools

up:
	docker-compose -f .docker/docker-compose.yml up -d

.PHONY: bootstrap codecheck down install test testcov tools up
