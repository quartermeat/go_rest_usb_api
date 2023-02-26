package main

import (
	_ "image/png"

	"github.com/faiface/pixel/pixelgl"
	"github.com/quartermeat/go_rest_usb_api/app"
)

func main() {
	pixelgl.Run(app.App)
	//scratch.Run()
}
