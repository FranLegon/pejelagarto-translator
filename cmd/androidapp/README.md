# Pejelagarto Translator Android App

A standalone Android application for the Pejelagarto translator.

## Building

```bash
export PATH=$PATH:$(go env GOPATH)/bin
export ANDROID_NDK_HOME=/usr/local/lib/android/sdk/ndk/26.3.11579264
export ANDROID_HOME=/usr/local/lib/android/sdk
gomobile build -androidapi=21 -target=android -o bin/pejelagarto-translator.apk ./cmd/androidapp
```

## Installation

Install on an Android device:

```bash
adb install bin/pejelagarto-translator.apk
```

## Features

- Simple OpenGL-based interface
- Demonstrates translation from Human to Pejelagarto
- Supports all Android architectures (ARM, ARM64, x86, x86_64)
- Minimum Android version: 5.0 (API 21)

## Files

- `main.go` - Main application code with OpenGL rendering
- `AndroidManifest.xml` - Android manifest (package: com.pejelagarto.translator)
