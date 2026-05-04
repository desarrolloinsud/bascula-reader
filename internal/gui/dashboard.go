package gui

import (
	"image/color"

	"bascula-connector/internal/app"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ShowDashboard abre la ventana principal del conector.
// connector debe haber arrancado ya (StartAll llamado por el caller).
func ShowDashboard(fa fyne.App, connector *app.App) fyne.Window {
	w := fa.NewWindow("Báscula Connector")
	w.Resize(fyne.NewSize(500, 420))
	w.SetMaster()

	// ── Bindings ──────────────────────────────────────────────────────────────
	weightBind := binding.NewString()
	lastReadBind := binding.NewString()
	httpStatusBind := binding.NewString()
	comStatusBind := binding.NewString()
	portBind := binding.NewString()
	serialBind := binding.NewString()
	scaleIDBind := binding.NewString()

	weightBind.Set("--")
	lastReadBind.Set("Sin lecturas")
	httpStatusBind.Set("Detenido")
	comStatusBind.Set("Detenido")

	// ── Título ────────────────────────────────────────────────────────────────
	scaleIDLabel := widget.NewLabelWithData(scaleIDBind)
	scaleIDLabel.TextStyle = fyne.TextStyle{Bold: true}

	// ── Indicadores de estado ─────────────────────────────────────────────────
	httpDot := canvas.NewCircle(colorGray)
	httpDot.Resize(fyne.NewSize(14, 14))
	comDot := canvas.NewCircle(colorGray)
	comDot.Resize(fyne.NewSize(14, 14))

	httpStatusLabel := widget.NewLabelWithData(httpStatusBind)
	comStatusLabel := widget.NewLabelWithData(comStatusBind)

	httpBtn := widget.NewButton("Iniciar HTTP", nil)
	comBtn := widget.NewButton("Iniciar COM", nil)

	httpBtn.OnTapped = func() {
		st := connector.GetStatus()
		if st.HTTPRunning {
			connector.StopHTTP()
		} else {
			connector.StartHTTP()
		}
	}
	comBtn.OnTapped = func() {
		st := connector.GetStatus()
		if st.COMConnected || st.COMReconnecting {
			connector.StopCOM()
		} else {
			connector.StartCOM()
		}
	}

	portLabel := widget.NewLabelWithData(portBind)
	serialLabel := widget.NewLabelWithData(serialBind)

	statusGrid := container.NewVBox(
		container.NewHBox(httpDot, widget.NewLabel("HTTP"), httpStatusLabel, layout.NewSpacer(), httpBtn),
		container.NewHBox(comDot, widget.NewLabel("COM "), comStatusLabel, layout.NewSpacer(), comBtn),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Puerto HTTP:"), portLabel),
		container.NewHBox(widget.NewLabel("Puerto COM: "), serialLabel),
	)

	// ── Peso en vivo ──────────────────────────────────────────────────────────
	weightLabel := widget.NewLabelWithData(weightBind)
	weightLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	weightLabel.Alignment = fyne.TextAlignCenter

	weightFrame := container.NewPadded(
		container.NewCenter(weightLabel),
	)
	bg := canvas.NewRectangle(color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	bg.CornerRadius = 6
	weightBox := container.NewStack(bg, weightFrame)

	lastReadLabel := widget.NewLabelWithData(lastReadBind)
	lastReadLabel.Alignment = fyne.TextAlignCenter

	// ── Botón reconfigurar ────────────────────────────────────────────────────
	reconfigBtn := widget.NewButton("Reconfigurar", func() {
		connector.StopAll()
		ShowWizard(fa, connector, func() {
			refreshBindings(connector, weightBind, lastReadBind, httpStatusBind, comStatusBind,
				portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn)
		})
	})

	// ── Layout ────────────────────────────────────────────────────────────────
	root := container.NewVBox(
		container.NewPadded(scaleIDLabel),
		widget.NewSeparator(),
		container.NewPadded(statusGrid),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(
			widget.NewLabel("Peso actual"),
			weightBox,
			lastReadLabel,
		)),
		widget.NewSeparator(),
		container.NewHBox(reconfigBtn, layout.NewSpacer()),
	)
	w.SetContent(container.NewPadded(root))

	// Cerrar ventana → minimizar al tray (no salir).
	w.SetCloseIntercept(func() { w.Hide() })

	// ── Goroutine de actualización ────────────────────────────────────────────
	go func() {
		ch := connector.Subscribe()
		defer connector.Unsubscribe(ch)
		// Render inicial
		refreshBindings(connector, weightBind, lastReadBind, httpStatusBind, comStatusBind,
			portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn)
		for st := range ch {
			applyStatus(st, weightBind, lastReadBind, httpStatusBind, comStatusBind,
				portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn)
		}
	}()

	w.Show()
	return w
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func refreshBindings(connector *app.App,
	weightBind, lastReadBind, httpStatusBind, comStatusBind,
	portBind, serialBind, scaleIDBind binding.String,
	httpDot, comDot *canvas.Circle,
	httpBtn, comBtn *widget.Button,
) {
	applyStatus(connector.GetStatus(), weightBind, lastReadBind, httpStatusBind, comStatusBind,
		portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn)
}

func applyStatus(st app.Status,
	weightBind, lastReadBind, httpStatusBind, comStatusBind,
	portBind, serialBind, scaleIDBind binding.String,
	httpDot, comDot *canvas.Circle,
	httpBtn, comBtn *widget.Button,
) {
	// Peso
	w := st.LastWeight
	if w == "" {
		w = "--"
	}
	weightBind.Set(w)

	t := st.LastReadAt
	if t == "" {
		t = "Sin lecturas"
	}
	lastReadBind.Set(t)

	// HTTP
	if st.HTTPRunning {
		httpStatusBind.Set("Activo")
		httpDot.FillColor = colorGreen
		httpBtn.SetText("Detener HTTP")
	} else {
		httpStatusBind.Set("Detenido")
		httpDot.FillColor = colorGray
		httpBtn.SetText("Iniciar HTTP")
	}
	httpDot.Refresh()

	// COM
	switch {
	case st.COMConnected:
		comStatusBind.Set("Conectado")
		comDot.FillColor = colorGreen
		comBtn.SetText("Detener COM")
	case st.COMReconnecting:
		comStatusBind.Set("Reconectando...")
		comDot.FillColor = colorYellow
		comBtn.SetText("Detener COM")
	default:
		if st.COMError != "" {
			comStatusBind.Set("Error")
		} else {
			comStatusBind.Set("Detenido")
		}
		comDot.FillColor = colorGray
		comBtn.SetText("Iniciar COM")
	}
	comDot.Refresh()

	portBind.Set(st.ActivePort)
	serialBind.Set(st.SerialPort)
	scaleIDBind.Set(st.ScaleID)
}

// ── Colores ───────────────────────────────────────────────────────────────────

var (
	colorGreen  = color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	colorYellow = color.NRGBA{R: 255, G: 193, B: 7, A: 255}
	colorRed    = color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	colorGray   = color.NRGBA{R: 158, G: 158, B: 158, A: 255}
)

// colorRed se usa en el tray; declarada aquí para evitar duplicados.
var _ = colorRed
