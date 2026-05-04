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
SERVER_PORT=8080
SERIAL_PORT=COM1
BAUD_RATE=9600
ALLOWED_ORIGIN=*
MOCK_SCALE=true
SCALE_ID=bascula-1
ENABLED_DEBUG_MODE=true
STATUS_ENDPOINT_URL=https://cfc.fresa.com.ar/pushBasculaStatus
EOF
    echo "   - .env generado por defecto"
  fi
}

# ── GUI Windows 64 bits (sin consola) ─────────────────────────────────────────
# Requiere mingw-w64 para cross-compile desde macOS/Linux:
#   brew install mingw-w64
echo ">> [GUI windows-amd64] Compilando..."
mkdir -p "${DIST_DIR}/windows-amd64"
if command -v x86_64-w64-mingw32-gcc &>/dev/null; then
  CGO_ENABLED=1 \
  CC=x86_64-w64-mingw32-gcc \
  GOOS=windows GOARCH=amd64 \
  go build \
    -ldflags="-H windowsgui -s -w" \
    -o "${DIST_DIR}/windows-amd64/bascula-connector.exe" \
    ./cmd/gui
  echo "   ✅ GUI compilado (sin consola)"
else
  echo "   ⚠️  mingw-w64 no encontrado — compilando CLI como fallback"
  GOOS=windows GOARCH=amd64 \
  go build \
    -ldflags="-s -w" \
    -o "${DIST_DIR}/windows-amd64/bascula-connector.exe" \
    ./cmd/cli
  echo "   ✅ CLI compilado (instalar mingw-w64 para GUI)"
fi
create_env "${DIST_DIR}/windows-amd64"

# ── CLI Windows 64 bits (headless, para servidores) ───────────────────────────
echo ">> [CLI windows-amd64] Compilando..."
GOOS=windows GOARCH=amd64 \
go build \
  -ldflags="-s -w" \
  -o "${DIST_DIR}/windows-amd64/bascula-connector-cli.exe" \
  ./cmd/cli
echo "   ✅ CLI compilado"

# ── GUI Windows 32 bits ───────────────────────────────────────────────────────
echo ">> [GUI windows-386] Compilando..."
mkdir -p "${DIST_DIR}/windows-386"
if command -v i686-w64-mingw32-gcc &>/dev/null; then
  CGO_ENABLED=1 \
  CC=i686-w64-mingw32-gcc \
  GOOS=windows GOARCH=386 \
  go build \
    -ldflags="-H windowsgui -s -w" \
    -o "${DIST_DIR}/windows-386/bascula-connector.exe" \
    ./cmd/gui
  echo "   ✅ GUI 32-bit compilado"
else
  GOOS=windows GOARCH=386 \
  go build \
    -ldflags="-s -w" \
    -o "${DIST_DIR}/windows-386/bascula-connector.exe" \
    ./cmd/cli
  echo "   ✅ CLI 32-bit compilado (instalar mingw-w64 para GUI)"
fi
create_env "${DIST_DIR}/windows-386"

# ── CLI Windows 32 bits ───────────────────────────────────────────────────────
GOOS=windows GOARCH=386 \
go build \
  -ldflags="-s -w" \
  -o "${DIST_DIR}/windows-386/bascula-connector-cli.exe" \
  ./cmd/cli
echo "   ✅ CLI 32-bit compilado"

echo ""
echo "✅ Build completado."
echo "Archivos en: ${DIST_DIR}"
ls -lh "${DIST_DIR}"/windows-*/*.exe 2>/dev/null
