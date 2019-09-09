Name:           gtunnel
Version:        1.0
Release:        1%{?dist}
Summary:        gtunnel

License:        GPL v2.0
Source0:        gtunnel-%{version}.tar.gz

%description
gtunnel

%prep
%setup -q

%install
rm -rf $RPM_BUILD_ROOT
install -d $RPM_BUILD_ROOT/opt/gtunnel
install gtunnel $RPM_BUILD_ROOT/opt/gtunnel/gtunnel
install hack-echo $RPM_BUILD_ROOT/opt/gtunnel/hack-echo
install hack-test-throughput $RPM_BUILD_ROOT/opt/gtunnel/hack-test-throughput
install hack-test-latency $RPM_BUILD_ROOT/opt/gtunnel/hack-test-latency

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/opt/gtunnel/gtunnel
/opt/gtunnel/hack-echo
/opt/gtunnel/hack-test-throughput
/opt/gtunnel/hack-test-latency

%post
ln -s /opt/gtunnel/gtunnel /usr/bin/gtunnel

%preun
rm /usr/bin/gtunnel

%postun

%changelog
