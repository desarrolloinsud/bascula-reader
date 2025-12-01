Conector de Báscula Local

Este proyecto es un conector local para leer el peso de una báscula conectada por puerto serie (COM/USB) y exponerlo a aplicaciones web a través de un servidor HTTP seguro en http://127.0.0.1.

Permite integrar básculas físicas con una aplicación en la nube sin exponer el hardware a internet. El frontend en la nube consulta el conector local como si fuera un microservicio.

⸻

🚀 Características principales
	•	Lectura en tiempo real desde puerto serie (COM/USB).
	•	Modo simulación (MOCK) para pruebas sin báscula.
	•	API HTTP local:
	•	GET /status
	•	GET /weight
	•	GET /stream (SSE – streaming tiempo real)
	•	.env autogenerado al lado del ejecutable.
	•	Cross-platform: Windows, macOS (Intel/ARM) y Linux.
	•	Puede instalarse como servicio en Linux y Windows.
	•	Incluye página demo (web/demo-bascula.html) para validar instalación.

⸻

📂 Estructura del proyecto

bascula-connector/
├─ cmd/
│  └─ main.go                 # Entry point
├─ internal/
│  ├─ config/                 # Manejo de .env, defaults, rutas
│  ├─ domain/                 # Modelos + puertos
│  ├─ scale/                  # Báscula real + simulador
│  └─ server/                 # HTTP + SSE
├─ web/
│  └─ demo-bascula.html       # Página de prueba
├─ .env.example
├─ build.sh
├─ go.mod
└─ README.md


⸻

⚙️ Variables de entorno (.env)

El ejecutable busca el archivo .env en el mismo directorio donde se encuentra. Si no existe, lo crea automáticamente.

Variables disponibles:

SERVER_PORT=7070        # Puerto HTTP local
SERIAL_PORT=COM3        # Puerto serie de la báscula
BAUD_RATE=9600          # Baudios
ALLOWED_ORIGIN=*        # CORS (dominio del frontend)
MOCK_SCALE=false        # true = simulador, false = báscula real
SCALE_ID=bascula-1      # Identificador lógico (numero de bacula/maquina)


⸻

🔌 API HTTP expuesta

GET /status

{
  "status": "running",
  "last_weight": "12.34 kg",
  "last_read_at": "2025-11-28T16:00:00Z",
  "scale_id": "bascula-1"
  "serial_port": "COM3",
  "baud_rate": "9600",
  "use_mock":  false,
}

GET /weight

{
  "weight": "12.34 kg",
  "time": "2025-11-28T16:00:00Z",
  "scale_id": "bascula-1"
}

GET /stream (Server-Sent Events)

Stream continuo para lecturas en tiempo real:

{
  "weight": "12.34 kg",
  "time": "2025-11-28T16:00:00Z",
  "scale_id": "bascula-1"
}


⸻

Ejecución en desarrollo

Clonar repo:

git clone <repo>
cd bascula-connector
go mod tidy

Crear .env:

cp .env.example .env

Ejecutar en modo desarrollo:

go run ./cmd

Verificar status:

https://cfc.fresa.com.ar/admin/timesheet/scaleStatus

⸻

🛠️ Build multiplataforma

El script build.sh genera binarios para Windows, macOS Intel, macOS ARM y Linux.

Ejecutar:

chmod +x build.sh
./build.sh

Binarios listos en:

dist/
├─ bascula-local
├─ bascula-windows-amd64.exe
├─ bascula-windows-386.exe
├─ bascula-macos-arm64
├─ bascula-macos-amd64
└─ bascula-linux-amd64

⸻

Instalación en Producción

Verifica que la bascula esta conectada y el puerto 7070 esta libre y operativo

    • Descargar archivo segun su distribucion del repositorio en la carpeta dist y 
    • Descarga el archivo .env.example
    • Modifica el nombre del archivo .env.example por .env
    • Edita el archivo .env de ser necesario, con tu editor de texto de confianza segun la configuracion de la bascula y guardalo en la misma carpeta que el archivo de ejecucion segun tu distribucion.
     
Advertencia

La modificacion de la variable SERVER_PORT del archivo .env provocara un fallo en la conexion de la demo en la plataforma de CFC

Windows

Arquitectura x64: bascula-windows-amd64.exe
Arquitectura x32: bascula-windows-386.exe

Archivos necesarios:
	•	bascula-windows-amd64.exe
	•	.env

Copiar a:

C:\bascula-conector\

Ejecutar el .exe.

Autoarranque: Task Scheduler o carpeta shell:startup.

⸻

macOS 

M1, M2...: bascula-macos-arm64
Intel: bascula-macos-amd64

Copiar archivos a:

/Users/<usuario>/bascula-conector/

Dar permisos:

chmod +x bascula-macos-arm64

Ejecutar:

./bascula-macos-arm64

Configurar .env según el puerto serie.

Autoarranque: Login Items o LaunchAgents.

⸻

Linux

Copiar a:

/opt/bascula-conector/

Ejecutar:

chmod +x bascula-linux-amd64
./bascula-linux-amd64

Instalar como servicio systemd (crear bascula.service).

⸻

Integración con Frontend (Vue 3)

const es = new EventSource('http://127.0.0.1:7070/stream')

es.onmessage = (event) => {
  const data = JSON.parse(event.data)
  console.log('Peso:', data.weight, 'ID:', data.scale_id)
}

es.onerror = () => {
  console.error('Error leyendo la báscula')
}


⸻

Verificar instalacion

https://cfc.fresa.com.ar/admin/timesheet/scaleStatus

⸻
Soporte

Para soporte incluir:
	•	.env
	•	salida de consola del ejecutable
	•	captura de web/demo-bascula.html
	•	sistema operativo
	•	modelo de báscula y conexión usada

⸻
