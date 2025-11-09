// Package android provides Android bindings for the Pejelagarto translator.
// This package is designed to be used with gomobile bind to create an Android AAR/APK.
package android

import (
	"pejelagarto-translator/internal/translator"
)

// Translator provides translation functions for Android.
type Translator struct{}

// NewTranslator creates a new Translator instance.
func NewTranslator() *Translator {
	return &Translator{}
}

// TranslateToPejelagarto translates human text to Pejelagarto.
func (t *Translator) TranslateToPejelagarto(input string) string {
	return translator.TranslateToPejelagarto(input)
}

// TranslateFromPejelagarto translates Pejelagarto text to human.
func (t *Translator) TranslateFromPejelagarto(input string) string {
	return translator.TranslateFromPejelagarto(input)
}

// ToPejelagarto is a convenience function for direct translation to Pejelagarto.
func ToPejelagarto(input string) string {
	return translator.TranslateToPejelagarto(input)
}

// FromPejelagarto is a convenience function for direct translation from Pejelagarto.
func FromPejelagarto(input string) string {
	return translator.TranslateFromPejelagarto(input)
}
