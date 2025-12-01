#!/usr/bin/env bash
set -euo pipefail

# Directorio raíz del proyecto
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="${ROOT_DIR}/dist"

mkdir -p "${DIST_DIR}"

echo ">> Construyendo binario local (SO actual)..."
go build -o "${DIST_DIR}/bascula-local" ./cmd

echo ">> Construyendo para Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "${DIST_DIR}/bascula-windows-amd64.exe" ./cmd

echo ">> Construyendo para Windows 32 bits (386)..."
GOOS=windows GOARCH=386 go build -o "${DIST_DIR}/bascula-windows-386.exe" ./cmd

echo ">> Construyendo para macOS (arm64 - M1/M2/M3)..."
GOOS=darwin GOARCH=arm64 go build -o "${DIST_DIR}/bascula-macos-arm64" ./cmd

echo ">> Construyendo para macOS (amd64 - Intel)..."
GOOS=darwin GOARCH=amd64 go build -o "${DIST_DIR}/bascula-macos-amd64" ./cmd

echo ">> Construyendo para Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/bascula-linux-amd64" ./cmd

echo
echo "✅ Build completado. Binarios generados en: ${DIST_DIR}"
ls -1 "${DIST_DIR}"