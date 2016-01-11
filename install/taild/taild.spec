Summary: taild package
Name: taild
Version: 0.1.0
Release: 150417000.el6
License: GPL 2 license
Group: System Environment/Daemons
Source: taild.tar.gz
URL: http://www.dk.wandoulabs.com
Distribution: Linux
Packager: zhouqiang <zhouqiang@wandoujia.com>
BuildRoot: %{_tmppath}/%{name}-%{version}-root

%define userpath /home/op
%define debug_package %{nil}

%description
This is the ng taild

%prep
%setup -c

%build
    #make %{?_smp_mflags}

%install
    [ ${RPM_BUILD_ROOT} != "/" ] && rm -rf ${RPM_BUILD_ROOT}
    install -d ${RPM_BUILD_ROOT}%{userpath}
    %{__cp} -r %{_builddir}/%{name}-%{version}/taild  ${RPM_BUILD_ROOT}%{userpath}/taild

%files
%defattr(-,root,root,755)
%{userpath}/taild

%pre
    ps aux | grep tail.linux | grep -v grep | awk '{print $2}'|xargs kill -9 2>/dev/null || true

%post
    mv %{userpath}/taild/taild /etc/init.d/taild
    chmod +x /etc/init.d/taild
    chkconfig --level 345 taild on
    service taild start
    exit 0

%preun
    if [ "$1" = 0 ];then
      if [ -f "/etc/init.d/taild" ];then
            /etc/init.d/taild stop
            chkconfig --del taild
            rm -f /etc/init.d/taild
      fi
    fi
    echo

%postun
echo

%clean
  [ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
