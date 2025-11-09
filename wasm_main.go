//go:build frontend

package main

import (
	"syscall/js"
	
	"pejelagarto-translator/internal/translator"
)

// This file contains the main() function for WASM builds (when -tags frontend is used)
// The translation logic in main.go is still available since it has no build constraints

func main() {
	// Export translation functions to JavaScript
	js.Global().Set("GoTranslateToPejelagarto", js.FuncOf(goTranslateToPejalagartoJS))
	js.Global().Set("GoTranslateFromPejelagarto", js.FuncOf(goTranslateFromPejalagartoJS))
	
	// Signal that WASM is ready
	js.Global().Set("WASMReady", js.ValueOf(true))
	
	// Keep the program running
	select {}
}

// goTranslateToPejalagartoJS wraps TranslateToPejelagarto for JavaScript
func goTranslateToPejalagartoJS(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: expected 1 argument")
	}
	
	input := args[0].String()
	result := translator.TranslateToPejelagarto(input)
	return js.ValueOf(result)
}

// goTranslateFromPejalagartoJS wraps TranslateFromPejelagarto for JavaScript
func goTranslateFromPejalagartoJS(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: expected 1 argument")
	}
	
	input := args[0].String()
	result := translator.TranslateFromPejelagarto(input)
	return js.ValueOf(result)
}
