export GOPATH := $(shell pwd)

.PHONY: all clean test rpm install uninstall

all:
	@go install -v org/gtunnel/gtunnel org/gtunnel/hack-echo \
			org/gtunnel/hack-test-throughput org/gtunnel/hack-test-latency

clean:
	@rm -rfv ./bin ./build ./pkg

test:
	@go test org/gtunnel/gtunnel

rpm: all
	make -f dist/Makefile rpm

install: all
	install -D bin/gtunnel $(DESTDIR)/opt/gtunnel/gtunnel
	install -D bin/gtunnel $(DESTDIR)/opt/gtunnel/hack-echo
	install -D bin/gtunnel $(DESTDIR)/opt/gtunnel/hack-test-throughput
	install -D bin/gtunnel $(DESTDIR)/opt/gtunnel/hack-test-latency
	ln -s $(DESTDIR)/opt/gtunnel/gtunnel $(DESTDIR)/usr/bin/gtunnel

uninstall: all
	rm -f $(DESTDIR)/opt/gtunnel/gtunnel
	rm $(DESTDIR)/opt/gtunnel/hack-echo
	rm $(DESTDIR)/opt/gtunnel/hack-test-throughput
	rm $(DESTDIR)/opt/gtunnel/hack-test-latency

distclean: clean
