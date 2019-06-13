export GOPATH := $(shell pwd)

.PHONY: all clean test

all:
	@go install -v org/gtunnel/gtunnel org/gtunnel/hack-echo org/gtunnel/hack-test

clean:
	@rm -rfv ./bin ./pkg

test:
	@go test org/gtunnel/gtunnel
