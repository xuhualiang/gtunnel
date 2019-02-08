export GOPATH := $(shell pwd)

.PHONY: all clean

all:
	@go install -v org/gtunnel

clean:
	@rm -rfv ./bin
