package main

// showDashboard es stub temporal — reemplazado en Fase E.

import (
	"bascula-connector/internal/app"
	"bascula-connector/internal/config"
	guipkg "bascula-connector/internal/gui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func showWizard(fa fyne.App, connector *app.App, cfg config.Config) {
	guipkg.ShowWizard(fa, connector, func() {
		showDashboard(fa, connector)
	})
}

func showDashboard(fa fyne.App, connector *app.App) {
	w := fa.NewWindow("Báscula Connector")
	w.SetContent(container.NewVBox(
		widget.NewLabel("Dashboard — próximamente (Fase E)"),
	))
	w.Resize(fyne.NewSize(480, 400))
	w.SetMaster()
	w.Show()
}
