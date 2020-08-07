# Maintainer: Andrei Costescu <andrei@costescu.no>

# shellcheck shell=bash

pkgname=fah-exporter-git
_pkgname="${pkgname%-git}"
pkgver=9879e6b
pkgrel=1
pkgdesc="Prometheus exporter for FAH client"
arch=("any")
url="https://github.com/cosandr/fah-exporter"
license=("MIT")
provides=("${_pkgname}")
conflicts=("${_pkgname}")
makedepends=("git" "go")
source=("git+$url")
md5sums=("SKIP")

_fah_service="foldingathome.service"

pkgver() {
    cd "${_pkgname}"
  ( set -o pipefail
    git describe --long 2>/dev/null | sed 's/\([^-]*-g\)/r\1/;s/-/./g' ||
    printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
  )
}

build() {
    cd "${_pkgname}"
    ./setup.sh systemd --pkg-name "${_pkgname}" --systemd-path ./ --systemd-after "$_fah_service" --systemd-requires "$_fah_service"
    go mod vendor
    go build -o "${_pkgname}"
}

package() {
    cd "${_pkgname}"
    install -d "${pkgdir}/usr/lib/systemd/system"
    install -Dm 755 "${_pkgname}" "${pkgdir}/usr/bin/${_pkgname}"
    install -m 644 "${_pkgname}.service" "${pkgdir}/usr/lib/systemd/system/"
    install -m 644 "${_pkgname}.socket" "${pkgdir}/usr/lib/systemd/system/"
    install -Dm 644 LICENSE "${pkgdir}/usr/share/licenses/${_pkgname}/LICENSE"
}
