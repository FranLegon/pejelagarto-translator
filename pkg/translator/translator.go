// Package translator provides translation functions for Android via gomobile bind
package translator

import (
	internalTranslator "pejelagarto-translator/internal/translator"
)

// Translator is a simple wrapper for translation operations
type Translator struct{}

// New creates a new Translator instance
func New() *Translator {
	return &Translator{}
}

// TranslateToPejelagarto translates English text to Pejelagarto language
func (t *Translator) TranslateToPejelagarto(text string) string {
	return internalTranslator.TranslateToPejelagarto(text)
}

// TranslateFromPejelagarto translates Pejelagarto text to English
func (t *Translator) TranslateFromPejelagarto(text string) string {
	return internalTranslator.TranslateFromPejelagarto(text)
}

// Package-level functions for direct calls
func TranslateToPejelagarto(text string) string {
	return internalTranslator.TranslateToPejelagarto(text)
}

func TranslateFromPejelagarto(text string) string {
	return internalTranslator.TranslateFromPejelagarto(text)
}
