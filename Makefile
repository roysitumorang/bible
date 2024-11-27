PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
PORT_HTTP=3000

.PHONY: all build

build:
	@go mod tidy
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -race -x -ldflags \
		"-X github.com/roysitumorang/bible/config.AppName=$(PROJECTNAME) \
		-X github.com/roysitumorang/bible/config.Commit=$(shell git rev-list -1 HEAD) \
		-X github.com/roysitumorang/bible/config.Build=$(shell date +%FT%T%:z)"

run: build stop
	@-nohup $(GOBASE)/$(PROJECTNAME) run > /dev/null 2>&1 & echo " > $(PROJECTNAME) is available at port $(PORT_HTTP) and PID $$!"

stop:
	@-lsof -t -i :$(PORT_HTTP) | xargs --no-run-if-empty kill