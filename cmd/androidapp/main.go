// Pejelagarto Translator Android App
// A simple Android app for translating between Human and Pejelagarto languages.
package main

import (
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"

	"pejelagarto-translator/internal/translator"
)

var (
	images *glutil.Images
	sz     size.Event
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{})
			}
		}
	})
}

func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx)
	
	// Test the translator
	testInput := "Hello World"
	testOutput := translator.TranslateToPejelagarto(testInput)
	log.Printf("Translation test: %s -> %s", testInput, testOutput)
}

func onStop(glctx gl.Context) {
	if images != nil {
		images.Release()
	}
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0.2, 0.3, 0.4, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
}
