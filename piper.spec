###############################################################################

# rpmbuilder:relative-pack true

###############################################################################

%define  debug_package %{nil}

###############################################################################

Summary:         Utility for log rotation for 12-factor apps
Name:            piper
Version:         1.1.1
Release:         0%{?dist}
Group:           Development/Tools
License:         MIT
URL:             https://github.com/gongled/piper

Source0:         %{name}-%{version}.tar.bz2

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.8

Provides:        %{name} = %{version}-%{release}

###############################################################################

%description
%{name} is a tiny and ease-to-use utility for log rotation for 12-factor apps
that write their logs to stdout.

###############################################################################

%prep
%setup -q

%build
mkdir src && mv {github.com,pkg.re} src

export GOPATH=$(pwd)
pushd src/github.com/gongled/%{name}/
%{__make} %{?_smp_mflags} all
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}/

install -pm 755 src/github.com/gongled/%{name}/%{name} \
                %{buildroot}%{_bindir}/

%clean
rm -rf %{buildroot}

###############################################################################

%files
%defattr(-,root,root,-)
%{_bindir}/%{name}

###############################################################################

%changelog
* Mon Nov 20 2017 Gleb Goncharov <gongled@gongled.ru> - 1.1.1-0
- Updated to the latest release.

* Fri Oct 27 2017 Gleb Goncharov <gongled@gongled.ru> - 1.0.0-0
- Initial build.
