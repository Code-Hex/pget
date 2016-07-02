test: deps
	go test

deps:
	go get -d -v -t ./...
	go get github.com/golang/lint/golint
	go get github.com/mattn/goveralls
	go get github.com/jessevdk/go-flags
	go get github.com/pkg/errors
	go get github.com/ricochet2200/go-disk-usage/du
	go get gopkg.in/cheggaaa/pb.v1
	go get github.com/stretchr/testify

lint: deps
	golint ./...

cover: deps
	go get github.com/axw/gocov/gocov
	goveralls

.PHONY: test deps lint cover
