package gui

import (
	_ "embed"

	"bascula-connector/internal/app"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

//go:embed assets/tray_green.png
var iconGreenBytes []byte

//go:embed assets/tray_yellow.png
var iconYellowBytes []byte

//go:embed assets/tray_red.png
var iconRedBytes []byte

var (
	iconGreen  = fyne.NewStaticResource("tray_green.png", iconGreenBytes)
	iconYellow = fyne.NewStaticResource("tray_yellow.png", iconYellowBytes)
	iconRed    = fyne.NewStaticResource("tray_red.png", iconRedBytes)
)

// SetupTray configura el ícono y menú del system tray.
// mainWindow es la ventana principal que se muestra/oculta desde el tray.
// No hace nada en sistemas que no soporten desktop.App (e.g. mobile).
func SetupTray(fa fyne.App, connector *app.App, mainWindow fyne.Window) {
	deskApp, ok := fa.(desktop.App)
	if !ok {
		return
	}

	deskApp.SetSystemTrayIcon(iconGreen)
	deskApp.SetSystemTrayMenu(fyne.NewMenu("Báscula Connector",
		fyne.NewMenuItem("Abrir panel", func() {
			mainWindow.Show()
			mainWindow.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Salir", func() {
			connector.StopAll()
			fa.Quit()
		}),
	))

	// Actualizar ícono del tray según estado del conector.
	go func() {
		ch := connector.Subscribe()
		defer connector.Unsubscribe(ch)
		for st := range ch {
			switch {
			case st.COMConnected && st.HTTPRunning:
				deskApp.SetSystemTrayIcon(iconGreen)
			case st.COMReconnecting:
				deskApp.SetSystemTrayIcon(iconYellow)
			default:
				deskApp.SetSystemTrayIcon(iconRed)
			}
		}
	}()
}
