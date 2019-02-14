export GOPATH := $(shell pwd)

.PHONY: all clean test

all:
	@go install -v org/gtunnel

clean:
	@rm -rfv ./bin

test:
	@go test org/gtunnel
