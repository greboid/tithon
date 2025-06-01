# Maintainer: Greboid <greboid@github>
pkgname=tithon
pkgver=0.0.9
pkgrel=1
pkgdesc="Modern IRC client"
arch=('x86_64')
url="https://github.com/greboid/tithon"
license=('MIT')
depends=('electron' 'glibc')
makedepends=('go' 'npm' 'git')
source=("git+https://github.com/greboid/tithon.git#tag=v${pkgver}")
sha256sums=('SKIP')

build() {
  cd "${srcdir}/${pkgname}"
  
  (cd backend && go build -o backend .)
  (cd frontend && npm install --omit=dev)
}

package() {
  cd "${srcdir}/${pkgname}"
  
  install -Dm755 "backend/backend" "${pkgdir}/usr/lib/tithon/backend"
  
  install -Dm644 "frontend/main.js" "${pkgdir}/usr/lib/tithon/main.js"
  install -Dm644 "frontend/icon.png" "${pkgdir}/usr/lib/tithon/icon.png"
  install -Dm644 "frontend/package.json" "${pkgdir}/usr/lib/tithon/package.json"
  
  install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  
  install -Dm644 /dev/stdin "${pkgdir}/usr/share/applications/${pkgname}.desktop" <<EOF
[Desktop Entry]
Name=Tithon
Comment=Modern IRC client
GenericName=IRC Client
Exec=tithon
Icon=/usr/lib/tithon/icon.png
Type=Application
StartupNotify=true
Categories=Network;Chat;IRCClient;
EOF

  # Install launcher script
  install -Dm755 /dev/stdin "${pkgdir}/usr/bin/${pkgname}" <<'EOF'
#!/bin/sh
cd /usr/lib/tithon
export NODE_ENV=production
exec electron /usr/lib/tithon/main.js "$@"
EOF
}
