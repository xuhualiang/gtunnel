export GOPATH := $(shell pwd)

.PHONY: all clean test

all:
	@go install -v org/gtunnel/gtunnel org/gtunnel/dpipe

clean:
	@rm -rfv ./bin ./pkg

test:
	@go test org/gtunnel
