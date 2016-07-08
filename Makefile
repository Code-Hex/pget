test: deps
	go test

deps:
	go get -d -v -t ./...
	go get github.com/golang/lint/golint
	go get github.com/mattn/goveralls
	go get github.com/axw/gocov/gocov
	go get github.com/tools/godep

lint: deps
	golint ./...

cover: deps
	goveralls

.PHONY: test deps lint cover
