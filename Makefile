.PHONY: clean
build:
	go build -o bin/pget -trimpath -ldflags "-w -s" -mod=readonly