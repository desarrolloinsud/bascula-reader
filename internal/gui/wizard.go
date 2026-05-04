package gui

import (
	"fmt"
	"strconv"

	"bascula-connector/internal/app"
	"bascula-connector/internal/config"
	"bascula-connector/internal/scale"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type wizardState struct {
	scaleID  string
	comPort  string
	baudRate string
	httpPort string
}

// ShowWizard abre el wizard de configuración inicial.
// onDone se llama con el conector ya iniciado al finalizar.
func ShowWizard(fa fyne.App, connector *app.App, onDone func()) {
	state := &wizardState{
		scaleID:  connector.Cfg.ScaleID,
		comPort:  connector.Cfg.SerialPort,
		baudRate: strconv.Itoa(connector.Cfg.BaudRate),
		httpPort: connector.Cfg.ServerPort,
	}

	w := fa.NewWindow("Configuración de báscula")
	w.Resize(fyne.NewSize(480, 380))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	stepLabel := widget.NewLabelWithStyle("Paso 1 de 4", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	content := container.NewMax()
	prevBtn := widget.NewButton("← Anterior", nil)
	nextBtn := widget.NewButton("Siguiente →", nil)
	prevBtn.Disable()

	step := 0
	steps := []func() fyne.CanvasObject{
		func() fyne.CanvasObject { return buildStep1(state) },
		func() fyne.CanvasObject { return buildStep2(state) },
		func() fyne.CanvasObject { return buildStep3(state) },
		func() fyne.CanvasObject { return buildStep4(state) },
	}

	render := func() {
		stepLabel.SetText(fmt.Sprintf("Paso %d de %d", step+1, len(steps)))
		content.Objects = []fyne.CanvasObject{steps[step]()}
		content.Refresh()

		if step == 0 {
			prevBtn.Disable()
		} else {
			prevBtn.Enable()
		}
		if step == len(steps)-1 {
			nextBtn.SetText("Guardar y comenzar")
		} else {
			nextBtn.SetText("Siguiente →")
		}
	}

	prevBtn.OnTapped = func() {
		if step > 0 {
			step--
			render()
		}
	}

	nextBtn.OnTapped = func() {
		if err := validateStep(step, state); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if step < len(steps)-1 {
			step++
			render()
			return
		}
		// Último paso: guardar y arrancar.
		if err := saveAndStart(connector, state, w); err != nil {
			dialog.ShowError(err, w)
			return
		}
		w.Hide()
		if onDone != nil {
			onDone()
		}
	}

	render()

	btnRow := container.NewGridWithColumns(2, prevBtn, nextBtn)
	root := container.NewBorder(
		stepLabel, btnRow, nil, nil,
		container.NewPadded(content),
	)
	w.SetContent(root)
	w.Show()
}

// ── Pasos ────────────────────────────────────────────────────────────────────

func buildStep1(s *wizardState) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Identificación de la báscula", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Asignale un nombre único a esta báscula.\nEjemplo: bascula-fresa-01")
	desc.Wrapping = fyne.TextWrapWord

	idEntry := widget.NewEntry()
	idEntry.SetText(s.scaleID)
	idEntry.SetPlaceHolder("bascula-fresa-01")
	idEntry.OnChanged = func(v string) { s.scaleID = v }

	return container.NewVBox(title, desc, widget.NewSeparator(), widget.NewForm(
		widget.NewFormItem("ID / Nombre", idEntry),
	))
}

func buildStep2(s *wizardState) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Puerto COM y velocidad", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Seleccioná el puerto al que está conectada la báscula.")
	desc.Wrapping = fyne.TextWrapWord

	ports := scale.ListAvailablePorts()
	if len(ports) == 0 {
		ports = []string{"COM1", "COM2", "COM3", "COM4"}
	}

	portSelect := widget.NewSelect(ports, func(v string) { s.comPort = v })
	portSelect.SetSelected(s.comPort)
	if portSelect.Selected == "" {
		portSelect.SetSelected(ports[0])
		s.comPort = ports[0]
	}

	bauds := []string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}
	baudSelect := widget.NewSelect(bauds, func(v string) { s.baudRate = v })
	baudSelect.SetSelected(s.baudRate)
	if baudSelect.Selected == "" {
		baudSelect.SetSelected("9600")
		s.baudRate = "9600"
	}

	return container.NewVBox(title, desc, widget.NewSeparator(), widget.NewForm(
		widget.NewFormItem("Puerto COM", portSelect),
		widget.NewFormItem("Velocidad (baud)", baudSelect),
	))
}

func buildStep3(s *wizardState) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Puerto HTTP local", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	hint := widget.NewLabel(
		"El sistema web buscará el conector en localhost:[puerto].\n" +
			"Seleccioná el puerto asignado a esta báscula (recomendado: 8080).",
	)
	hint.Wrapping = fyne.TextWrapWord

	ports := []string{"8080", "8081", "8082", "8083", "8084", "8085"}
	portSelect := widget.NewSelect(ports, func(v string) { s.httpPort = v })
	if s.httpPort == "" {
		s.httpPort = "8080"
	}
	portSelect.SetSelected(s.httpPort)

	return container.NewVBox(title, hint, widget.NewSeparator(), widget.NewForm(
		widget.NewFormItem("Puerto HTTP", portSelect),
	))
}

func buildStep4(s *wizardState) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Confirmar configuración", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	summary := fmt.Sprintf(
		"Báscula:    %s\nCOM:        %s @ %s baud\nHTTP:       localhost:%s",
		s.scaleID, s.comPort, s.baudRate, s.httpPort,
	)
	label := widget.NewLabel(summary)
	label.TextStyle = fyne.TextStyle{Monospace: true}

	note := widget.NewLabel("Al confirmar se guardará la configuración y el servicio arrancará automáticamente.")
	note.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), label, widget.NewSeparator(), note)
}

// ── Validación ───────────────────────────────────────────────────────────────

func validateStep(step int, s *wizardState) error {
	switch step {
	case 0:
		if s.scaleID == "" {
			return fmt.Errorf("el ID de la báscula no puede estar vacío")
		}
		for _, c := range s.scaleID {
			if c == ' ' {
				return fmt.Errorf("el ID no puede contener espacios")
			}
		}
	case 2:
		p, err := strconv.Atoi(s.httpPort)
		if err != nil || p < 8080 || p > 8085 {
			return fmt.Errorf("el puerto HTTP debe ser uno de: 8080, 8081, 8082, 8083, 8084, 8085")
		}
	}
	return nil
}

// ── Guardar y arrancar ───────────────────────────────────────────────────────

func saveAndStart(connector *app.App, s *wizardState, w fyne.Window) error {
	baud, err := strconv.Atoi(s.baudRate)
	if err != nil {
		baud = 9600
	}

	values := map[string]string{
		"SCALE_ID":    s.scaleID,
		"SERIAL_PORT": s.comPort,
		"BAUD_RATE":   strconv.Itoa(baud),
		"SERVER_PORT": s.httpPort,
		"MOCK_SCALE":  "false",
	}

	envPath := config.EnvPath()
	if err := config.WriteEnv(envPath, values); err != nil {
		return fmt.Errorf("no se pudo guardar la configuración: %w", err)
	}

	// Recargar config y actualizar el conector.
	newCfg := config.Load()
	connector.Cfg = newCfg

	if err := connector.StartAll(); err != nil {
		return fmt.Errorf("error al iniciar el servicio: %w", err)
	}

	return nil
}
