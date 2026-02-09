#!/usr/bin/env bash
set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
ENV_TEMPLATE="${ROOT_DIR}/.env.example"

echo "=================================================="
echo ">> Construyendo binarios..."
echo "Salida: ${DIST_DIR}"
echo "=================================================="

# Limpiar y crear dist
echo ">> Limpiando carpeta dist/"
rm -rf "${DIST_DIR}" 2>/dev/null || true
mkdir -p "${DIST_DIR}"

# Función para crear .env
create_env() {
  local target_dir="$1"
  if [[ -f "${ENV_TEMPLATE}" ]]; then
    cp "${ENV_TEMPLATE}" "${target_dir}/.env"
    echo "   - .env copiado desde .env.example"
  else
    cat > "${target_dir}/.env" <<'EOF'
# Configuración del conector de báscula
SERVER_PORT=7070
SERIAL_PORT=COM1
BAUD_RATE=9600
ALLOWED_ORIGIN=*
MOCK_SCALE=true
SCALE_ID=bascula-1
ENABLED_DEBUG_MODE=true
# LOG_DIR vacío = logs en la misma carpeta del ejecutable
# LOG_DIR=logs
STATUS_ENDPOINT_URL=https://cfc.fresa.com.ar/pushBasculaStatus
EOF
    echo "   - .env generado por defecto"
  fi
}

# Windows 64 bits
echo ">> [windows-amd64] Compilando..."
mkdir -p "${DIST_DIR}/windows-amd64"
GOOS=windows GOARCH=amd64 go build -o "${DIST_DIR}/windows-amd64/bascula-windows-amd64.exe" ./cmd
create_env "${DIST_DIR}/windows-amd64"
echo "   ✅ Compilado"

# Windows 32 bits
echo ">> [windows-386] Compilando..."
mkdir -p "${DIST_DIR}/windows-386"
GOOS=windows GOARCH=386 go build -o "${DIST_DIR}/windows-386/bascula-windows-386.exe" ./cmd
create_env "${DIST_DIR}/windows-386"
echo "   ✅ Compilado"

echo ""
echo "✅ Build completado."
echo "Archivos en: ${DIST_DIR}"
ls -lh "${DIST_DIR}"/windows-*/*.exe 2>/dev/null
