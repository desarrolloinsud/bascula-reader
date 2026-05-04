package main

import (
	"bascula-connector/internal/app"
	"bascula-connector/internal/config"
	guipkg "bascula-connector/internal/gui"

	"fyne.io/fyne/v2"
)

func showWizard(fa fyne.App, connector *app.App, cfg config.Config) {
	guipkg.ShowWizard(fa, connector, func() {
		showDashboard(fa, connector)
	})
}

func showDashboard(fa fyne.App, connector *app.App) {
	w := guipkg.ShowDashboard(fa, connector)
	guipkg.SetupTray(fa, connector, w)
}
