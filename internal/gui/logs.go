package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowLogs abre una ventana con el contenido del log de hoy.
// Devuelve la ventana para que el llamador pueda gestionar su ciclo de vida.
func ShowLogs(fa fyne.App, logDir string) fyne.Window {
	w := fa.NewWindow("Logs — Lector de básculas")
	w.Resize(fyne.NewSize(700, 400))
	w.CenterOnScreen()

	entry := widget.NewMultiLineEntry()
	entry.Disable()
	entry.TextStyle = fyne.TextStyle{Monospace: true}

	load := func() {
		logFile := filepath.Join(logDir, fmt.Sprintf("bascula-%s.log", time.Now().Format("2006-01-02")))
		data, err := os.ReadFile(logFile)
		if err != nil {
			entry.SetText(fmt.Sprintf("No se encontró el archivo de log:\n%s", logFile))
			return
		}
		lines := strings.Split(string(data), "\n")
		const maxLines = 200
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
		entry.SetText(strings.Join(lines, "\n"))
		// scroll al final
		entry.CursorRow = len(lines) - 1
	}

	load()

	refreshBtn := widget.NewButton("↻  Actualizar", func() { load() })
	refreshBtn.Importance = widget.LowImportance

	w.SetContent(container.NewBorder(
		nil,
		container.NewHBox(refreshBtn),
		nil, nil,
		container.NewScroll(entry),
	))
	w.Show()
	return w
}
