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

// paleta Bootstrap 5 (alineada con la web principal)
var (
	colorHeaderBg   = color.NRGBA{R: 33, G: 37, B: 41, A: 255}   // --bs-dark  #212529
	colorHeaderText = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // blanco
	colorSubtitle   = color.NRGBA{R: 173, G: 181, B: 189, A: 255} // --bs-gray-500 #adb5bd
	colorGreen      = color.NRGBA{R: 25, G: 135, B: 84, A: 255}   // --bs-success #198754
	colorYellow     = color.NRGBA{R: 255, G: 193, B: 7, A: 255}   // --bs-warning #ffc107
	colorRed        = color.NRGBA{R: 220, G: 53, B: 69, A: 255}   // --bs-danger  #dc3545
	colorGray       = color.NRGBA{R: 108, G: 117, B: 125, A: 255} // --bs-secondary #6c757d
	colorCardBg     = color.NRGBA{R: 52, G: 58, B: 64, A: 230}    // --bs-gray-800 #343a40 semitransparente
	colorCardText   = color.NRGBA{R: 33, G: 37, B: 41, A: 255}    // --bs-dark (no usado directamente)
	colorWeightBg   = color.NRGBA{R: 13, G: 110, B: 253, A: 255}  // --bs-primary #0d6efd
	colorWeightText = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // blanco sobre azul
	colorWeightSub  = color.NRGBA{R: 207, G: 226, B: 255, A: 255} // --bs-primary-bg-subtle #cfe2ff
	colorOverlay    = color.NRGBA{R: 0, G: 0, B: 0, A: 90}        // overlay neutro oscuro
)

// colorRed se usa en tray; evitar duplicado.
var _ = colorRed

// ShowDashboard abre la ventana principal del conector.
func ShowDashboard(fa fyne.App, connector *app.App) fyne.Window {
	w := fa.NewWindow("Lector de básculas")
	w.Resize(fyne.NewSize(520, 500))
	w.SetMaster()

	// ── Bindings ──────────────────────────────────────────────────────────────
	weightBind    := binding.NewString()
	lastReadBind  := binding.NewString()
	httpStatusBind := binding.NewString()
	comStatusBind := binding.NewString()
	portBind      := binding.NewString()
	serialBind    := binding.NewString()
	scaleIDBind   := binding.NewString()

	weightBind.Set("--")
	lastReadBind.Set("Sin lecturas")
	httpStatusBind.Set("Detenido")
	comStatusBind.Set("Detenido")

	// ── Header ────────────────────────────────────────────────────────────────
	headerTitle := canvas.NewText("Lector de básculas", colorHeaderText)
	headerTitle.TextSize = 15
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}

	scaleIDText := canvas.NewText("", colorSubtitle)
	scaleIDText.TextSize = 11

	headerContent := container.NewHBox(
		container.NewPadded(container.NewVBox(
			container.NewWithoutLayout(headerTitle),
			container.NewWithoutLayout(scaleIDText),
		)),
	)

	headerBg := canvas.NewRectangle(colorHeaderBg)
	header := container.NewStack(headerBg, container.NewPadded(headerContent))

	// ── Indicadores de estado ─────────────────────────────────────────────────
	httpDot := canvas.NewCircle(colorGray)
	httpDot.Resize(fyne.NewSize(12, 12))
	comDot := canvas.NewCircle(colorGray)
	comDot.Resize(fyne.NewSize(12, 12))

	httpStatusLabel := widget.NewLabelWithData(httpStatusBind)
	comStatusLabel  := widget.NewLabelWithData(comStatusBind)

	httpBtn := widget.NewButton("Iniciar HTTP", nil)
	comBtn  := widget.NewButton("Iniciar COM", nil)
	httpBtn.Importance = widget.LowImportance
	comBtn.Importance  = widget.LowImportance

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

	portLabel   := widget.NewLabelWithData(portBind)
	serialLabel := widget.NewLabelWithData(serialBind)

	statusRows := container.NewVBox(
		container.NewHBox(httpDot, widget.NewLabel("HTTP"), httpStatusLabel, layout.NewSpacer(), httpBtn),
		container.NewHBox(comDot,  widget.NewLabel("COM "), comStatusLabel,  layout.NewSpacer(), comBtn),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Puerto HTTP:"), portLabel),
		container.NewHBox(widget.NewLabel("Puerto COM: "), serialLabel),
	)

	statusCardBg := canvas.NewRectangle(colorCardBg)
	statusCardBg.CornerRadius = 8
	statusCard := container.NewStack(statusCardBg, container.NewPadded(statusRows))

	// ── Peso en vivo ──────────────────────────────────────────────────────────
	weightLabel := widget.NewLabelWithData(weightBind)
	weightLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	weightLabel.Alignment = fyne.TextAlignCenter

	weightTitle := canvas.NewText("PESO ACTUAL", colorWeightSub)
	weightTitle.TextSize = 10
	weightTitle.TextStyle = fyne.TextStyle{Bold: true}

	lastReadLabel := widget.NewLabelWithData(lastReadBind)
	lastReadLabel.Alignment = fyne.TextAlignCenter

	weightInner := container.NewVBox(
		container.NewCenter(container.NewWithoutLayout(weightTitle)),
		container.NewCenter(weightLabel),
		container.NewCenter(lastReadLabel),
	)
	weightBg := canvas.NewRectangle(colorWeightBg)
	weightBg.CornerRadius = 8
	weightCard := container.NewStack(weightBg, container.NewPadded(weightInner))

	// ── Botones inferiores ────────────────────────────────────────────────────
	reconfigBtn := widget.NewButton("⚙  Reconfigurar", func() {
		connector.StopAll()
		ShowWizard(fa, connector, func() {
			refreshBindings(connector, weightBind, lastReadBind, httpStatusBind, comStatusBind,
				portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn, scaleIDText)
		})
	})
	reconfigBtn.Importance = widget.LowImportance

	var logWin fyne.Window
	logsBtn := widget.NewButton("📋  Logs", func() {
		if logWin != nil {
			logWin.Show()
			logWin.RequestFocus()
			return
		}
		logWin = ShowLogs(fa, connector.Cfg.LogDir)
		logWin.SetOnClosed(func() { logWin = nil })
	})
	logsBtn.Importance = widget.LowImportance

	// ── Contenido sobre el fondo ───────────────────────────────────────────────
	mainContent := container.NewVBox(
		container.NewPadded(statusCard),
		container.NewPadded(weightCard),
		container.NewPadded(container.NewHBox(reconfigBtn, logsBtn, layout.NewSpacer())),
	)

	// ── Fondo: imagen patio + overlay oscuro ──────────────────────────────────
	bgImg := canvas.NewImageFromResource(resPatio)
	bgImg.FillMode = canvas.ImageFillStretch
	bgImg.Translucency = 0.45

	bgOverlay := canvas.NewRectangle(colorOverlay)

	body := container.NewStack(bgImg, bgOverlay, mainContent)

	// ── Layout final ──────────────────────────────────────────────────────────
	root := container.NewBorder(header, nil, nil, nil, body)
	w.SetContent(root)
	w.SetCloseIntercept(func() { w.Hide() })

	// ── Goroutine de actualización ─────────────────────────────────────────────
	go func() {
		ch := connector.Subscribe()
		defer connector.Unsubscribe(ch)
		refreshBindings(connector, weightBind, lastReadBind, httpStatusBind, comStatusBind,
			portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn, scaleIDText)
		for st := range ch {
			applyStatus(st, weightBind, lastReadBind, httpStatusBind, comStatusBind,
				portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn, scaleIDText)
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
	scaleIDText *canvas.Text,
) {
	applyStatus(connector.GetStatus(), weightBind, lastReadBind, httpStatusBind, comStatusBind,
		portBind, serialBind, scaleIDBind, httpDot, comDot, httpBtn, comBtn, scaleIDText)
}

func applyStatus(st app.Status,
	weightBind, lastReadBind, httpStatusBind, comStatusBind,
	portBind, serialBind, scaleIDBind binding.String,
	httpDot, comDot *canvas.Circle,
	httpBtn, comBtn *widget.Button,
	scaleIDText *canvas.Text,
) {
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

	// bindings son thread-safe
	portBind.Set(st.ActivePort)
	serialBind.Set(st.SerialPort)
	scaleIDBind.Set(st.ScaleID)

	httpRunning := st.HTTPRunning
	comConnected := st.COMConnected
	comReconnecting := st.COMReconnecting
	comError := st.COMError
	scaleID := st.ScaleID

	if httpRunning {
		httpStatusBind.Set("Activo")
	} else {
		httpStatusBind.Set("Detenido")
	}
	switch {
	case comConnected:
		comStatusBind.Set("Conectado")
	case comReconnecting:
		comStatusBind.Set("Reconectando...")
	default:
		if comError != "" {
			comStatusBind.Set("Error")
		} else {
			comStatusBind.Set("Detenido")
		}
	}

	// canvas y widget updates deben ir en el hilo principal
	fyne.Do(func() {
		if httpRunning {
			httpDot.FillColor = colorGreen
			httpBtn.SetText("Detener HTTP")
		} else {
			httpDot.FillColor = colorGray
			httpBtn.SetText("Iniciar HTTP")
		}
		httpDot.Refresh()

		switch {
		case comConnected:
			comDot.FillColor = colorGreen
			comBtn.SetText("Detener COM")
		case comReconnecting:
			comDot.FillColor = colorYellow
			comBtn.SetText("Detener COM")
		default:
			comDot.FillColor = colorGray
			comBtn.SetText("Iniciar COM")
		}
		comDot.Refresh()

		if scaleIDText != nil {
			scaleIDText.Text = scaleID
			scaleIDText.Refresh()
		}
	})
}
