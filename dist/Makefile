RPMDIR=~/rpmbuild
RPMSRC  = $(RPMDIR)/SOURCES
RPMSPEC = $(RPMDIR)/SPECS
O = $(RPMDIR)/RPMS/x86_64

PRGM = bin/gtunnel bin/hack-echo bin/hack-test-throughput bin/hack-test-latency
PKG  = gtunnel-1.0

.PHONY: rpm
rpm: $(PRGM)
	rpmdev-setuptree
	install -d build/$(PKG)
	install $(PRGM) build/$(PKG)
	tar -czvf $(RPMSRC)/$(PKG).tar.gz -C build/ $(PKG)
	install dist/gtunnel-rpm.spec $(RPMSPEC)/gtunnel.spec
	rpmbuild -bb $(RPMSPEC)/gtunnel.spec
	cp $(O)/$(PKG)-1.el7.x86_64.rpm dist/

deb: $(PRGM)
	bzr whoami "Hualiang Xu <hualiang.xu82@aliyun.com>"
	install -d build/$(PKG)
	install $(PRGM) build/$(PKG)
	tar -czvf build/$(PKG).tar.gz -C build/ $(PKG)
	cd build && bzr dh-make gtunnel 1.0 $(PKG).tar.gz
