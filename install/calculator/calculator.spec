Summary: calculator package
Name: calculator
Version: 1.4.0
Release: 150827001.el6
License: GPL 2 license
Group: System Environment/Daemons
Source: calculator.tar.gz
URL: http://www.wandoujia.com
Distribution: Linux
Packager: Qin Guoan <qinguoan@wandoujia.com>
BuildRoot: %{_tmppath}/%{name}-%{version}-root

%define userpath /home/op
%define debug_package %{nil}

%description
This is the taild domain data consumer

%prep
%setup -c

%build
    #make %{?_smp_mflags}

%install
    [ ${RPM_BUILD_ROOT} != "/" ] && rm -rf ${RPM_BUILD_ROOT}
    install -d ${RPM_BUILD_ROOT}%{userpath}
    %{__cp} -r %{_builddir}/%{name}-%{version}/calculator  ${RPM_BUILD_ROOT}%{userpath}/calculator

%files
%defattr(-,root,root,755)
%{userpath}/calculator

%pre
    killall calculator || true

%post
    mv %{userpath}/calculator/run/calculatord /etc/init.d/calculatord
    chmod +x /etc/init.d/calculatord
    chkconfig --level 345 calculatord on
    /etc/init.d/calculatord start
    exit 0

%preun
if [ "$1" = 0 ];then
    :
fi
if [ "$1" = 1 ];then  #包被更新时文件未卸载之前的动作处理
    :
fi
%postun
echo

%clean
  [ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
