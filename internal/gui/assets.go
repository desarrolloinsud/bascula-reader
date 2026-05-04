package gui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/logo.png
var logoBytes []byte

//go:embed assets/patio.jpg
var patioBytes []byte

var (
	resLogo  = fyne.NewStaticResource("logo.png", logoBytes)
	resPatio = fyne.NewStaticResource("patio.jpg", patioBytes)
)
