package main

// Stubs temporales — reemplazados en Fases D y E.

import (
	"bascula-connector/internal/app"
	"bascula-connector/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func showWizard(fa fyne.App, connector *app.App, cfg config.Config) {
	w := fa.NewWindow("Configuración inicial")
	w.SetContent(container.NewVBox(
		widget.NewLabel("Wizard — próximamente"),
	))
	w.Resize(fyne.NewSize(480, 360))
	w.Show()
}

func showDashboard(fa fyne.App, connector *app.App) {
	w := fa.NewWindow("Báscula Connector")
	w.SetContent(container.NewVBox(
		widget.NewLabel("Dashboard — próximamente"),
	))
	w.Resize(fyne.NewSize(480, 400))
	w.ShowAndRun()
}
