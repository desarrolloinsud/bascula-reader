# Conector de Báscula Local

Este proyecto es un **conector local** que permite leer el peso de una báscula
conectada por puerto serie (COM/USB) a una PC y exponerlo a una aplicación web
a través de un servidor HTTP en `http://127.0.0.1`.

## Arquitectura (resumen)

- `cmd/main.go`  
  Entry point del binario. Carga configuración, inicializa la báscula (real o mock)
  y arranca el servidor HTTP.

- `internal/config`  
  Carga la configuración desde un archivo `.env` ubicado junto al ejecutable.  
  Si el `.env` no existe, se crea uno por defecto.

- `internal/domain`  
  Define el modelo de dominio (`WeightReading`) y el puerto (`Scale`) que representa
  el contrato de una báscula.

- `internal/scale`
  - `serial_scale.go`: implementación de `Scale` que lee de un puerto serie real
    usando `github.com/tarm/serial`.
  - `mock_scale.go`: implementación de `Scale` que genera pesos simulados para pruebas.

- `internal/server/http_server.go`  
  Servidor HTTP que expone:
  - `GET /status`  → estado básico del conector
  - `GET /weight`  → última lectura de peso
  - `GET /stream`  → streaming en tiempo real (SSE)

- `web/demo-bascula.html`  
  Página de prueba para validar la instalación y ver el streaming en tiempo real.

---

## Requisitos

- Go >= 1.25 instalado.
- (Para Windows cliente) PC con Windows 10+ y una báscula serie/USB.

---

## Instalación para desarrollo

1. Clonar el repo y entrar al directorio:

   ```bash
   git clone <tu-repo> bascula-connector
   cd bascula-connector