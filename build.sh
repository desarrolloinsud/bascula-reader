#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
ENV_TEMPLATE="${ROOT_DIR}/.env.example"

echo ">> Limpiando carpeta dist/"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

# Crea un .env dentro del target usando .env.example si existe
create_env() {
  local target_dir="$1"

  if [[ -f "${ENV_TEMPLATE}" ]]; then
    cp "${ENV_TEMPLATE}" "${target_dir}/.env"
    echo "   - Copiado .env desde .env.example -> ${target_dir}/.env"
  else
    cat > "${target_dir}/.env" <<'EOF'
# Configuración del conector de báscula
SERVER_PORT=7070
SERIAL_PORT=COM1
BAUD_RATE=9600
ALLOWED_ORIGIN=*
MOCK_SCALE=true
SCALE_ID=bascula-1
EOF
    echo "   - .env.example NO encontrado, generado .env por defecto en ${target_dir}/.env"
  fi
}

echo "=================================================="
echo ">> Construyendo binarios y armando carpetas de instalación..."
echo "Salida: ${DIST_DIR}"
echo "=================================================="

#
# 1) Binario local (SO actual) - útil para desarrollo
#
LOCAL_DIR="${DIST_DIR}/local"
mkdir -p "${LOCAL_DIR}"

echo ">> [local] Construyendo bascula-local (SO actual)..."
go build -o "${LOCAL_DIR}/bascula-local" ./cmd
create_env "${LOCAL_DIR}"

#
# 2) Windows 64 bits (amd64)
#
WIN64_DIR="${DIST_DIR}/windows-amd64"
mkdir -p "${WIN64_DIR}"

echo ">> [windows-amd64] Construyendo bascula-windows-amd64.exe..."
GOOS=windows GOARCH=amd64 go build -o "${WIN64_DIR}/bascula-windows-amd64.exe" ./cmd
create_env "${WIN64_DIR}"

#
# 3) Windows 32 bits (386)
#
WIN32_DIR="${DIST_DIR}/windows-386"
mkdir -p "${WIN32_DIR}"

echo ">> [windows-386] Construyendo bascula-windows-386.exe..."
GOOS=windows GOARCH=386 go build -o "${WIN32_DIR}/bascula-windows-386.exe" ./cmd
create_env "${WIN32_DIR}"

#
# 4) macOS ARM (M1/M2/M3)
#
MAC_ARM_DIR="${DIST_DIR}/macos-arm64"
mkdir -p "${MAC_ARM_DIR}"

echo ">> [macos-arm64] Construyendo bascula-macos-arm64..."
GOOS=darwin GOARCH=arm64 go build -o "${MAC_ARM_DIR}/bascula-macos-arm64" ./cmd
create_env "${MAC_ARM_DIR}"

#
# 5) macOS Intel (amd64)
#
MAC_AMD64_DIR="${DIST_DIR}/macos-amd64"
mkdir -p "${MAC_AMD64_DIR}"

echo ">> [macos-amd64] Construyendo bascula-macos-amd64..."
GOOS=darwin GOARCH=amd64 go build -o "${MAC_AMD64_DIR}/bascula-macos-amd64" ./cmd
create_env "${MAC_AMD64_DIR}"

#
# 6) Linux 64 bits (amd64)
#
LINUX_AMD64_DIR="${DIST_DIR}/linux-amd64"
mkdir -p "${LINUX_AMD64_DIR}"

echo ">> [linux-amd64] Construyendo bascula-linux-amd64..."
GOOS=linux GOARCH=amd64 go build -o "${LINUX_AMD64_DIR}/bascula-linux-amd64" ./cmd
create_env "${LINUX_AMD64_DIR}"

echo
echo "✅ Build completado."
echo "Carpetas de instalación generadas en: ${DIST_DIR}"
find "${DIST_DIR}" -maxdepth 2 -type f -print
